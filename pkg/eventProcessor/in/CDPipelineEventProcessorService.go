package in

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	"github.com/devtron-labs/devtron/pkg/workflow/status"
	"github.com/devtron-labs/devtron/util/argo"
	"go.uber.org/zap"
	"k8s.io/utils/pointer"
)

type CDPipelineEventProcessorImpl struct {
	logger                  *zap.SugaredLogger
	pubSubClient            *pubsub.PubSubClientServiceImpl
	cdWorkflowCommonService cd.CdWorkflowCommonService
	workflowStatusService   status.WorkflowStatusService
	cdTriggerService        devtronApps.TriggerService
	argoUserService         argo.ArgoUserService

	pipelineRepository     pipelineConfig.PipelineRepository
	installedAppRepository repository2.InstalledAppRepository
}

func NewCDPipelineEventProcessorImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	cdWorkflowCommonService cd.CdWorkflowCommonService,
	workflowStatusService status.WorkflowStatusService,
	cdTriggerService devtronApps.TriggerService,
	argoUserService argo.ArgoUserService,
	pipelineRepository pipelineConfig.PipelineRepository,
	installedAppRepository repository2.InstalledAppRepository) *CDPipelineEventProcessorImpl {
	cdPipelineEventProcessorImpl := &CDPipelineEventProcessorImpl{
		logger:                  logger,
		pubSubClient:            pubSubClient,
		cdWorkflowCommonService: cdWorkflowCommonService,
		workflowStatusService:   workflowStatusService,
		cdTriggerService:        cdTriggerService,
		argoUserService:         argoUserService,
		pipelineRepository:      pipelineRepository,
		installedAppRepository:  installedAppRepository,
	}
	return cdPipelineEventProcessorImpl
}

func (impl *CDPipelineEventProcessorImpl) SubscribeCDBulkTriggerTopic() error {

	callback := func(msg *model.PubSubMsg) {
		event := &bean.BulkCDDeployEvent{}
		err := json.Unmarshal([]byte(msg.Data), event)
		if err != nil {
			impl.logger.Errorw("Error unmarshalling received event", "topic", pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC, "msg", msg.Data, "err", err)
			return
		}
		event.ValuesOverrideRequest.UserId = event.UserId
		// trigger
		ctx, err := impl.argoUserService.BuildACDContext()
		if err != nil {
			impl.logger.Errorw("error in creating acd context", "err", err)
			return
		}
		triggerContext := bean2.TriggerContext{
			ReferenceId: pointer.String(msg.MsgId),
			Context:     ctx,
		}
		_, err = impl.cdTriggerService.ManualCdTrigger(triggerContext, event.ValuesOverrideRequest)
		if err != nil {
			impl.logger.Errorw("Error triggering CD", "topic", pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC, "msg", msg.Data, "err", err)
		}
	}
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		event := &bean.BulkCDDeployEvent{}
		err := json.Unmarshal([]byte(msg.Data), event)
		if err != nil {
			return "error unmarshalling received event", []interface{}{"msg", msg.Data, "err", err}
		}
		return "got message for trigger cd in bulk", []interface{}{"pipelineId", event.ValuesOverrideRequest.PipelineId, "appId", event.ValuesOverrideRequest.AppId, "cdWorkflowType", event.ValuesOverrideRequest.CdWorkflowType, "ciArtifactId", event.ValuesOverrideRequest.CiArtifactId}
	}
	validations := impl.cdWorkflowCommonService.GetTriggerValidateFuncs()
	err := impl.pubSubClient.Subscribe(pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC, callback, loggerFunc, validations...)
	if err != nil {
		impl.logger.Error("failed to subscribe to NATS topic", "topic", pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC, "err", err)
		return err
	}
	return nil
}

func (impl *CDPipelineEventProcessorImpl) SubscribeArgoTypePipelineSyncEvent() error {
	callback := func(msg *model.PubSubMsg) {
		statusUpdateEvent := bean.ArgoPipelineStatusSyncEvent{}
		var err error
		var cdPipeline *pipelineConfig.Pipeline
		var installedApp repository2.InstalledApps

		err = json.Unmarshal([]byte(msg.Data), &statusUpdateEvent)
		if err != nil {
			impl.logger.Errorw("unmarshal error on argo pipeline status update event", "err", err)
			return
		}

		if statusUpdateEvent.IsAppStoreApplication {
			installedApp, err = impl.installedAppRepository.GetInstalledAppByInstalledAppVersionId(statusUpdateEvent.InstalledAppVersionId)
			if err != nil {
				impl.logger.Errorw("error in getting installedAppVersion by id", "err", err, "id", statusUpdateEvent.PipelineId)
				return
			}
		} else {
			cdPipeline, err = impl.pipelineRepository.FindById(statusUpdateEvent.PipelineId)
			if err != nil {
				impl.logger.Errorw("error in getting cdPipeline by id", "err", err, "id", statusUpdateEvent.PipelineId)
				return
			}
		}

		triggerContext := bean2.TriggerContext{
			ReferenceId: pointer.String(msg.MsgId),
		}

		err, _ = impl.workflowStatusService.UpdatePipelineTimelineAndStatusByLiveApplicationFetch(triggerContext, cdPipeline, installedApp, statusUpdateEvent.UserId)
		if err != nil {
			impl.logger.Errorw("error on argo pipeline status update", "err", err, "msg", msg.Data)
			return
		}
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		statusUpdateEvent := bean.ArgoPipelineStatusSyncEvent{}
		err := json.Unmarshal([]byte(msg.Data), &statusUpdateEvent)
		if err != nil {
			return "unmarshal error on argo pipeline status update event", []interface{}{"err", err}
		}
		return "got message for argo pipeline status update", []interface{}{"pipelineId", statusUpdateEvent.PipelineId, "installedAppVersionId", statusUpdateEvent.InstalledAppVersionId, "isAppStoreApplication", statusUpdateEvent.IsAppStoreApplication}
	}

	validations := impl.cdWorkflowCommonService.GetTriggerValidateFuncs()
	err := impl.pubSubClient.Subscribe(pubsub.ARGO_PIPELINE_STATUS_UPDATE_TOPIC, callback, loggerFunc, validations...)
	if err != nil {
		impl.logger.Errorw("error in subscribing to argo application status update topic", "err", err)
		return err
	}
	return nil
}
