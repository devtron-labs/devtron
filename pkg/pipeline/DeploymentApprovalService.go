package pipeline

import (
	"context"
	"errors"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	client2 "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	bean4 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	util2 "github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
	"time"
)

type DeploymentApprovalService interface {
	PerformDeploymentApprovalAction(userId int32, approvalActionRequest bean4.UserApprovalActionRequest) error
	FetchApprovalPendingArtifacts(pipelineId, limit, offset, requiredApprovals int, searchString string) ([]bean4.CiArtifactBean, int, error)
	FetchApprovalDataForArtifacts(artifactIds []int, pipelineId int, requiredApprovals int) (map[int]*pipelineConfig.UserApprovalMetadata, error)
}

type DeploymentApprovalServiceImpl struct {
	logger                       *zap.SugaredLogger
	userService                  user.UserService
	cdTriggerService             devtronApps.TriggerService
	eventClient                  client2.EventClient
	eventFactory                 client2.EventFactory
	appArtifactManager           AppArtifactManager
	deploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository
	ciArtifactRepository         repository.CiArtifactRepository
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	appWorkflowRepository        appWorkflow.AppWorkflowRepository
	pipelineRepository           pipelineConfig.PipelineRepository
}

func NewDeploymentApprovalServiceImpl(logger *zap.SugaredLogger,
	userService user.UserService,
	cdTriggerService devtronApps.TriggerService,
	eventClient client2.EventClient,
	eventFactory client2.EventFactory,
	appArtifactManager AppArtifactManager,
	deploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository,
	ciArtifactRepository repository.CiArtifactRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	pipelineRepository pipelineConfig.PipelineRepository) *DeploymentApprovalServiceImpl {
	return &DeploymentApprovalServiceImpl{
		logger:                       logger,
		userService:                  userService,
		cdTriggerService:             cdTriggerService,
		eventClient:                  eventClient,
		eventFactory:                 eventFactory,
		appArtifactManager:           appArtifactManager,
		deploymentApprovalRepository: deploymentApprovalRepository,
		ciArtifactRepository:         ciArtifactRepository,
		ciPipelineRepository:         ciPipelineRepository,
		appWorkflowRepository:        appWorkflowRepository,
		pipelineRepository:           pipelineRepository,
	}
}

func (impl *DeploymentApprovalServiceImpl) PerformDeploymentApprovalAction(userId int32, approvalActionRequest bean4.UserApprovalActionRequest) error {
	approvalActionType := approvalActionRequest.ActionType
	artifactId := approvalActionRequest.ArtifactId
	approvalRequestId := approvalActionRequest.ApprovalRequestId
	if approvalActionType == bean4.APPROVAL_APPROVE_ACTION {
		// fetch approval request data, same user should not be Approval requester
		approvalRequest, err := impl.deploymentApprovalRepository.FetchWithPipelineAndArtifactDetails(approvalRequestId)
		if err != nil {
			return &bean4.DeploymentApprovalValidationError{
				Err:           errors.New("failed to fetch approval request data"),
				ApprovalState: bean4.RequestCancelled,
			}

		}
		if approvalRequest.ArtifactDeploymentTriggered == true {
			return &bean4.DeploymentApprovalValidationError{
				Err:           errors.New("deployment has already been triggered for this request"),
				ApprovalState: bean4.AlreadyApproved,
			}
		}
		if approvalRequest.CreatedBy == userId {
			return errors.New("requester cannot be an approver")
		}

		// fetch artifact metadata, who triggered this build
		ciArtifact, err := impl.ciArtifactRepository.Get(artifactId)
		if err != nil {
			impl.logger.Errorw("error occurred while fetching workflow data for artifact", "artifactId", artifactId, "userId", userId, "err", err)
			return errors.New("failed to fetch workflow for artifact data")
		}
		if ciArtifact.CreatedBy == userId {
			return errors.New("user who triggered the build cannot be an approver")
		}
		deploymentApprovalData := &pipelineConfig.DeploymentApprovalUserData{
			ApprovalRequestId: approvalRequestId,
			UserId:            userId,
			UserResponse:      pipelineConfig.APPROVED,
		}
		deploymentApprovalData.CreatedBy = userId
		deploymentApprovalData.UpdatedBy = userId
		err = impl.deploymentApprovalRepository.SaveDeploymentUserData(deploymentApprovalData)
		if err != nil {
			impl.logger.Errorw("error occurred while saving user approval data", "approvalRequestId", approvalRequestId, "err", err)
			return &bean4.DeploymentApprovalValidationError{
				Err:           err,
				ApprovalState: bean4.AlreadyApproved,
			}
		}
		// trigger deployment if approved and pipeline type is automatic
		pipeline := approvalRequest.Pipeline
		if pipeline.TriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC {
			pipelineId := approvalRequest.PipelineId
			approvalConfig, err := pipeline.GetApprovalConfig()
			if err != nil {
				impl.logger.Errorw("error occurred while fetching approval config", "pipelineId", pipelineId, "config", pipeline.UserApprovalConfig, "err", err)
				return nil
			}
			approvalDataForArtifacts, err := impl.FetchApprovalDataForArtifacts([]int{artifactId}, pipelineId, approvalConfig.RequiredCount)
			if err != nil {
				impl.logger.Errorw("error occurred while fetching approval data for artifacts", "artifactId", artifactId, "pipelineId", pipelineId, "config", pipeline.UserApprovalConfig, "err", err)
				return nil
			}
			if approvedData, ok := approvalDataForArtifacts[artifactId]; ok && approvedData.ApprovalRuntimeState == pipelineConfig.ApprovedApprovalState {
				// trigger deployment
				triggerRequest := bean.TriggerRequest{
					CdWf:        nil,
					Pipeline:    pipeline,
					Artifact:    approvalRequest.CiArtifact,
					TriggeredBy: 1,
					TriggerContext: bean.TriggerContext{
						Context: context.Background(),
					},
				}
				err = impl.cdTriggerService.TriggerAutomaticDeployment(triggerRequest)
				if err != nil {
					impl.logger.Errorw("error occurred while triggering deployment", "pipelineId", pipelineId, "artifactId", artifactId, "err", err)
					return errors.New("auto deployment failed, please try manually")
				}
			}
		}

	} else if approvalActionType == bean4.APPROVAL_REQUEST_ACTION {
		pipelineId := approvalActionRequest.PipelineId
		deploymentApprovalRequest := &pipelineConfig.DeploymentApprovalRequest{
			PipelineId: pipelineId,
			ArtifactId: artifactId,
			Active:     true,
		}
		deploymentApprovalRequest.CreatedBy = userId
		deploymentApprovalRequest.UpdatedBy = userId
		err := impl.deploymentApprovalRepository.Save(deploymentApprovalRequest)
		if err != nil {
			impl.logger.Errorw("error occurred while submitting approval request", "pipelineId", pipelineId, "artifactId", artifactId, "err", err)
			return err
		}
		approvalActionRequest.ApprovalRequestId = deploymentApprovalRequest.Id
		go impl.performNotificationApprovalAction(approvalActionRequest, userId)

	} else {
		// fetch if cd wf runner is present then user cannot cancel the request, as deployment has been triggered already
		approvalRequest, err := impl.deploymentApprovalRepository.FetchById(approvalRequestId)
		if err != nil {
			return errors.New("failed to fetch approval request data")
		}
		if approvalRequest.CreatedBy != userId {
			return errors.New("request cannot be cancelled as not initiated by the same")
		}
		if approvalRequest.ArtifactDeploymentTriggered {
			return errors.New("request cannot be cancelled as deployment is already been made for this request")
		}
		approvalRequest.Active = false
		err = impl.deploymentApprovalRepository.Update(approvalRequest)
		if err != nil {
			impl.logger.Errorw("error occurred while updating approval request", "pipelineId", approvalRequest.PipelineId, "artifactId", artifactId, "err", err)
			return err
		}
	}
	return nil
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

func (impl *DeploymentApprovalServiceImpl) performNotificationApprovalAction(approvalActionRequest bean4.UserApprovalActionRequest, userId int32) {
	if len(approvalActionRequest.ApprovalNotificationConfig.EmailIds) == 0 {
		return
	}
	eventType := util2.Approval
	var events []client2.Event
	pipeline, err := impl.pipelineRepository.FindById(approvalActionRequest.PipelineId)
	if err != nil {
		impl.logger.Errorw("error occurred while updating approval request", "pipelineId", pipeline, "pipeline", pipeline, "err", err)
	}
	event := impl.eventFactory.Build(eventType, &approvalActionRequest.PipelineId, approvalActionRequest.AppId, &pipeline.EnvironmentId, "")
	imageComment, imageTagNames, err := impl.appArtifactManager.GetImageTagsAndComment(approvalActionRequest.ArtifactId)
	if err != nil {
		impl.logger.Errorw("error in fetching tags and comment", "artifactId", approvalActionRequest.ArtifactId)
	}
	events = impl.eventFactory.BuildExtraApprovalData(event, approvalActionRequest, pipeline, userId, imageTagNames, imageComment.Comment)
	for _, evnt := range events {
		_, evtErr := impl.eventClient.WriteNotificationEvent(evnt)
		if evtErr != nil {
			impl.logger.Errorw("unable to send approval event", "error", evtErr)
		}
	}

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
