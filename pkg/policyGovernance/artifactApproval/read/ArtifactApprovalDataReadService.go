package read

import (
	"encoding/json"
	"fmt"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	bean4 "github.com/devtron-labs/devtron/pkg/bean"
	"go.uber.org/zap"
	"strings"
	"time"
)

type ArtifactApprovalDataReadService interface {
	FetchApprovalPendingArtifacts(pipelineId, limit, offset, requiredApprovals int, searchString string) ([]bean4.CiArtifactBean, int, error)
	FetchApprovalDataForArtifacts(artifactIds []int, pipelineId int, requiredApprovals int) (map[int]*pipelineConfig.UserApprovalMetadata, error)
}

type ArtifactApprovalDataReadServiceImpl struct {
	logger                       *zap.SugaredLogger
	userService                  user.UserService
	deploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository
}

func NewArtifactApprovalDataReadServiceImpl(logger *zap.SugaredLogger,
	userService user.UserService,
	deploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository) *ArtifactApprovalDataReadServiceImpl {
	return &ArtifactApprovalDataReadServiceImpl{
		logger:                       logger,
		userService:                  userService,
		deploymentApprovalRepository: deploymentApprovalRepository,
	}
}

func (impl *ArtifactApprovalDataReadServiceImpl) FetchApprovalPendingArtifacts(pipelineId, limit, offset, requiredApprovals int, searchString string) ([]bean4.CiArtifactBean, int, error) {

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

func (impl *ArtifactApprovalDataReadServiceImpl) FetchApprovalDataForArtifacts(artifactIds []int, pipelineId int, requiredApprovals int) (map[int]*pipelineConfig.UserApprovalMetadata, error) {
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

func (impl *ArtifactApprovalDataReadServiceImpl) getLatestDeploymentByArtifactIds(pipelineId int, deploymentApprovalRequests []*pipelineConfig.DeploymentApprovalRequest, artifactIds []int) ([]*pipelineConfig.DeploymentApprovalRequest, error) {
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

func parseMaterialInfo(materialInfo json.RawMessage, source string) (json.RawMessage, error) {
	if source != repository.GOCD && source != repository.CI_RUNNER && source != repository.WEBHOOK && source != repository.EXT && source != repository.PRE_CD && source != repository.POST_CD && source != repository.POST_CI {
		return nil, fmt.Errorf("datasource: %s not supported", source)
	}
	var ciMaterials []repository.CiMaterialInfo
	err := json.Unmarshal(materialInfo, &ciMaterials)
	if err != nil {
		println("material info", materialInfo)
		println("unmarshal error for material info", "err", err)
	}
	var scmMapList []map[string]string

	for _, material := range ciMaterials {
		scmMap := map[string]string{}
		var url string
		if material.Material.Type == "git" {
			url = material.Material.GitConfiguration.URL
		} else if material.Material.Type == "scm" {
			url = material.Material.ScmConfiguration.URL
		} else {
			return nil, fmt.Errorf("unknown material type:%s ", material.Material.Type)
		}
		if material.Modifications != nil && len(material.Modifications) > 0 {
			_modification := material.Modifications[0]

			revision := _modification.Revision
			url = strings.TrimSpace(url)

			_webhookDataStr := ""
			_webhookDataByteArr, err := json.Marshal(_modification.WebhookData)
			if err == nil {
				_webhookDataStr = string(_webhookDataByteArr)
			}

			scmMap["url"] = url
			scmMap["revision"] = revision
			scmMap["modifiedTime"] = _modification.ModifiedTime
			scmMap["author"] = _modification.Author
			scmMap["message"] = _modification.Message
			scmMap["tag"] = _modification.Tag
			scmMap["webhookData"] = _webhookDataStr
			scmMap["branch"] = _modification.Branch
		}
		scmMapList = append(scmMapList, scmMap)
	}
	mInfo, err := json.Marshal(scmMapList)
	return mInfo, err
}

func formatDate(t time.Time, layout string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(layout)
}
