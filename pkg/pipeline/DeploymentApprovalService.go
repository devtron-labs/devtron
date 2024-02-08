package pipeline

import (
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	bean4 "github.com/devtron-labs/devtron/pkg/bean"
	"go.uber.org/zap"
	"time"
)

type DeploymentApprovalService interface {
	FetchApprovalPendingArtifacts(pipelineId, limit, offset, requiredApprovals int, searchString string) ([]bean4.CiArtifactBean, int, error)
	FetchApprovalDataForArtifacts(artifactIds []int, pipelineId int, requiredApprovals int) (map[int]*pipelineConfig.UserApprovalMetadata, error)
}

type DeploymentApprovalServiceImpl struct {
	logger                       *zap.SugaredLogger
	userService                  user.UserService
	deploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository
	ciArtifactRepository         repository.CiArtifactRepository
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	appWorkflowRepository        appWorkflow.AppWorkflowRepository
}

func NewDeploymentApprovalServiceImpl(logger *zap.SugaredLogger,
	userService user.UserService,
	deploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository,
	ciArtifactRepository repository.CiArtifactRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	appWorkflowRepository appWorkflow.AppWorkflowRepository) *DeploymentApprovalServiceImpl {
	return &DeploymentApprovalServiceImpl{
		logger:                       logger,
		userService:                  userService,
		deploymentApprovalRepository: deploymentApprovalRepository,
		ciArtifactRepository:         ciArtifactRepository,
		ciPipelineRepository:         ciPipelineRepository,
		appWorkflowRepository:        appWorkflowRepository,
	}
}

func (impl *DeploymentApprovalServiceImpl) FetchApprovalPendingArtifacts(pipelineId, limit, offset, requiredApprovals int, searchString string) ([]bean4.CiArtifactBean, int, error) {

	var ciArtifacts []bean4.CiArtifactBean
	deploymentApprovalRequests, totalCount, err := impl.deploymentApprovalRepository.FetchApprovalPendingArtifacts(pipelineId, limit, offset, requiredApprovals, searchString)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching approval request data", "pipelineId", pipelineId, "err", err)
		return ciArtifacts, 0, err
	}

	var artifactIds []int
	for _, request := range deploymentApprovalRequests {
		artifactIds = append(artifactIds, request.ArtifactId)
	}

	if len(artifactIds) > 0 {
		deploymentApprovalRequests, err = impl.getLatestDeploymentByArtifactIds(pipelineId, deploymentApprovalRequests, artifactIds)
		if err != nil {
			impl.logger.Errorw("error occurred while fetching FetchLatestDeploymentByArtifactIds", "pipelineId", pipelineId, "artifactIds", artifactIds, "err", err)
			return nil, 0, err
		}
	}

	for _, request := range deploymentApprovalRequests {

		mInfo, err := parseMaterialInfo([]byte(request.CiArtifact.MaterialInfo), request.CiArtifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("Error in parsing artifact material info", "err", err)
		}

		var artifact bean4.CiArtifactBean
		ciArtifact := request.CiArtifact
		artifact.Id = ciArtifact.Id
		artifact.Image = ciArtifact.Image
		artifact.ImageDigest = ciArtifact.ImageDigest
		artifact.MaterialInfo = mInfo
		artifact.DataSource = ciArtifact.DataSource
		artifact.Deployed = ciArtifact.Deployed
		artifact.Scanned = ciArtifact.Scanned
		artifact.ScanEnabled = ciArtifact.ScanEnabled
		artifact.CiPipelineId = ciArtifact.PipelineId
		artifact.DeployedTime = formatDate(ciArtifact.DeployedTime, bean4.LayoutRFC3339)
		if ciArtifact.WorkflowId != nil {
			artifact.WfrId = *ciArtifact.WorkflowId
		}
		artifact.CiPipelineId = ciArtifact.PipelineId
		ciArtifacts = append(ciArtifacts, artifact)
	}

	return ciArtifacts, totalCount, err
}

func (impl *DeploymentApprovalServiceImpl) FetchApprovalDataForArtifacts(artifactIds []int, pipelineId int, requiredApprovals int) (map[int]*pipelineConfig.UserApprovalMetadata, error) {
	artifactIdVsApprovalMetadata := make(map[int]*pipelineConfig.UserApprovalMetadata)
	deploymentApprovalRequests, err := impl.deploymentApprovalRepository.FetchApprovalDataForArtifacts(artifactIds, pipelineId)
	if err != nil {
		return artifactIdVsApprovalMetadata, err
	}

	var requestedUserIds []int32
	for _, approvalRequest := range deploymentApprovalRequests {
		requestedUserIds = append(requestedUserIds, approvalRequest.CreatedBy)
	}

	userInfos, err := impl.userService.GetByIds(requestedUserIds)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching users", "requestedUserIds", requestedUserIds, "err", err)
		return artifactIdVsApprovalMetadata, err
	}
	userInfoMap := make(map[int32]bean3.UserInfo)
	for _, userInfo := range userInfos {
		userId := userInfo.Id
		userInfoMap[userId] = userInfo
	}

	for _, approvalRequest := range deploymentApprovalRequests {
		artifactId := approvalRequest.ArtifactId
		requestedUserId := approvalRequest.CreatedBy
		if userInfo, ok := userInfoMap[requestedUserId]; ok {
			approvalRequest.UserEmail = userInfo.EmailId
		}
		approvalMetadata := approvalRequest.ConvertToApprovalMetadata()
		if approvalRequest.GetApprovedCount() >= requiredApprovals {
			approvalMetadata.ApprovalRuntimeState = pipelineConfig.ApprovedApprovalState
		} else {
			approvalMetadata.ApprovalRuntimeState = pipelineConfig.RequestedApprovalState
		}
		artifactIdVsApprovalMetadata[artifactId] = approvalMetadata
	}
	return artifactIdVsApprovalMetadata, nil

}

func (impl *DeploymentApprovalServiceImpl) getLatestDeploymentByArtifactIds(pipelineId int, deploymentApprovalRequests []*pipelineConfig.DeploymentApprovalRequest, artifactIds []int) ([]*pipelineConfig.DeploymentApprovalRequest, error) {
	var latestDeployedArtifacts []*pipelineConfig.DeploymentApprovalRequest
	var err error
	if len(artifactIds) > 0 {
		latestDeployedArtifacts, err = impl.deploymentApprovalRepository.FetchLatestDeploymentByArtifactIds(pipelineId, artifactIds)
		if err != nil {
			impl.logger.Errorw("error occurred while fetching FetchLatestDeploymentByArtifactIds", "pipelineId", pipelineId, "artifactIds", artifactIds, "err", err)
			return nil, err
		}
	}
	latestDeployedArtifactsMap := make(map[int]time.Time, 0)
	for _, artifact := range latestDeployedArtifacts {
		latestDeployedArtifactsMap[artifact.ArtifactId] = artifact.AuditLog.CreatedOn
	}

	for _, request := range deploymentApprovalRequests {
		if deployedTime, ok := latestDeployedArtifactsMap[request.ArtifactId]; ok {
			request.CiArtifact.Deployed = true
			request.CiArtifact.DeployedTime = deployedTime
		}
	}

	return deploymentApprovalRequests, nil
}
