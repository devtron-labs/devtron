package action

import (
	"context"
	"errors"
	client2 "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean4 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactApproval/read"
	util2 "github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
)

type ArtifactApprovalActionService interface {
	PerformDeploymentApprovalAction(userId int32, approvalActionRequest bean4.UserApprovalActionRequest) error
}

type ArtifactApprovalActionServiceImpl struct {
	logger                          *zap.SugaredLogger
	artifactApprovalDataReadService read.ArtifactApprovalDataReadService
	cdTriggerService                devtronApps.TriggerService
	eventClient                     client2.EventClient
	eventFactory                    client2.EventFactory
	appArtifactManager              pipeline.AppArtifactManager
	deploymentApprovalRepository    pipelineConfig.DeploymentApprovalRepository
	ciArtifactRepository            repository.CiArtifactRepository
	pipelineRepository              pipelineConfig.PipelineRepository
}

func NewArtifactApprovalActionServiceImpl(logger *zap.SugaredLogger,
	artifactApprovalDataReadService read.ArtifactApprovalDataReadService,
	cdTriggerService devtronApps.TriggerService, eventClient client2.EventClient,
	eventFactory client2.EventFactory, appArtifactManager pipeline.AppArtifactManager,
	deploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository,
	ciArtifactRepository repository.CiArtifactRepository,
	pipelineRepository pipelineConfig.PipelineRepository) *ArtifactApprovalActionServiceImpl {
	return &ArtifactApprovalActionServiceImpl{
		logger:                          logger,
		artifactApprovalDataReadService: artifactApprovalDataReadService,
		cdTriggerService:                cdTriggerService,
		eventClient:                     eventClient,
		eventFactory:                    eventFactory,
		appArtifactManager:              appArtifactManager,
		deploymentApprovalRepository:    deploymentApprovalRepository,
		ciArtifactRepository:            ciArtifactRepository,
		pipelineRepository:              pipelineRepository,
	}
}

func (impl *ArtifactApprovalActionServiceImpl) PerformDeploymentApprovalAction(userId int32, approvalActionRequest bean4.UserApprovalActionRequest) error {
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
			approvalDataForArtifacts, err := impl.artifactApprovalDataReadService.FetchApprovalDataForArtifacts([]int{artifactId}, pipelineId, approvalConfig.RequiredCount)
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

func (impl *ArtifactApprovalActionServiceImpl) performNotificationApprovalAction(approvalActionRequest bean4.UserApprovalActionRequest, userId int32) {
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
