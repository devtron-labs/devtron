package in

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean4 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/adapter"
	bean3 "github.com/devtron-labs/devtron/pkg/workflow/cd/bean"
	"github.com/devtron-labs/devtron/util/argo"
	"go.uber.org/zap"
	"k8s.io/utils/pointer"
	"strconv"
)

type WorkflowEventProcessorImpl struct {
	logger                  *zap.SugaredLogger
	pubSubClient            *pubsub.PubSubClientServiceImpl
	cdWorkflowService       cd.CdWorkflowService
	cdWorkflowRunnerService cd.CdWorkflowRunnerService
	workflowDagExecutor     pipeline.WorkflowDagExecutor
	argoUserService         argo.ArgoUserService
	//repositories import, to be removed
	pipelineRepository   pipelineConfig.PipelineRepository
	ciArtifactRepository repository.CiArtifactRepository
}

func NewWorkflowEventProcessorImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	cdWorkflowService cd.CdWorkflowService,
	cdWorkflowRunnerService cd.CdWorkflowRunnerService,
	workflowDagExecutor pipeline.WorkflowDagExecutor,
	argoUserService argo.ArgoUserService,
	pipelineRepository pipelineConfig.PipelineRepository,
	ciArtifactRepository repository.CiArtifactRepository) (*WorkflowEventProcessorImpl, error) {
	impl := &WorkflowEventProcessorImpl{
		logger:                  logger,
		pubSubClient:            pubSubClient,
		cdWorkflowService:       cdWorkflowService,
		cdWorkflowRunnerService: cdWorkflowRunnerService,
		argoUserService:         argoUserService,
		workflowDagExecutor:     workflowDagExecutor,
		pipelineRepository:      pipelineRepository,
		ciArtifactRepository:    ciArtifactRepository,
	}
	return impl, nil
}

func (impl *WorkflowEventProcessorImpl) SubscribeCdStageCompleteEvent() error {
	callback := func(msg *model.PubSubMsg) {
		cdStageCompleteEvent := bean.CdStageCompleteEvent{}
		err := json.Unmarshal([]byte(msg.Data), &cdStageCompleteEvent)
		if err != nil {
			impl.logger.Errorw("error while unmarshalling cdStageCompleteEvent object", "err", err, "msg", msg.Data)
			return
		}
		wfr, err := impl.cdWorkflowRunnerService.FindWorkflowRunnerById(cdStageCompleteEvent.WorkflowRunnerId)
		if err != nil {
			impl.logger.Errorw("could not get wf runner", "err", err)
			return
		}
		triggerContext := pipeline.TriggerContext{
			ReferenceId: pointer.String(msg.MsgId),
		}
		if wfr.WorkflowType == bean2.CD_WORKFLOW_TYPE_PRE {
			impl.logger.Debugw("received pre stage success event for workflow runner ", "wfId", strconv.Itoa(wfr.Id))
			err = impl.workflowDagExecutor.HandlePreStageSuccessEvent(triggerContext, cdStageCompleteEvent)
			if err != nil {
				impl.logger.Errorw("deployment success event error", "err", err)
				return
			}
		} else if wfr.WorkflowType == bean2.CD_WORKFLOW_TYPE_POST {
			impl.logger.Debugw("received post stage success event for workflow runner ", "wfId", strconv.Itoa(wfr.Id))
			err = impl.workflowDagExecutor.HandlePostStageSuccessEvent(triggerContext, wfr.CdWorkflowId, cdStageCompleteEvent.CdPipelineId, cdStageCompleteEvent.TriggeredBy, cdStageCompleteEvent.PluginRegistryArtifactDetails)
			if err != nil {
				impl.logger.Errorw("deployment success event error", "err", err)
				return
			}
		}
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		cdStageCompleteEvent := bean.CdStageCompleteEvent{}
		err := json.Unmarshal([]byte(msg.Data), &cdStageCompleteEvent)
		if err != nil {
			return "error while unmarshalling cdStageCompleteEvent object", []interface{}{"err", err, "msg", msg.Data}
		}
		return "got message for cd stage completion", []interface{}{"workflowRunnerId", cdStageCompleteEvent.WorkflowRunnerId, "workflowId", cdStageCompleteEvent.WorkflowId, "cdPipelineId", cdStageCompleteEvent.CdPipelineId}
	}

	validations := impl.workflowDagExecutor.GetTriggerValidateFuncs()

	err := impl.pubSubClient.Subscribe(pubsub.CD_STAGE_COMPLETE_TOPIC, callback, loggerFunc, validations...)
	if err != nil {
		impl.logger.Error("error", "err", err)
		return err
	}
	return nil
}

func (impl *WorkflowEventProcessorImpl) SubscribeTriggerBulkAction() error {
	callback := func(msg *model.PubSubMsg) {
		cdWorkflow := new(pipelineConfig.CdWorkflow)
		err := json.Unmarshal([]byte(msg.Data), cdWorkflow)
		if err != nil {
			impl.logger.Error("Error while unmarshalling cdWorkflow json object", "error", err)
			return
		}
		wf := &bean3.CdWorkflowDto{
			Id:           cdWorkflow.Id,
			CiArtifactId: cdWorkflow.CiArtifactId,
			PipelineId:   cdWorkflow.PipelineId,
			UserId:       bean4.SYSTEM_USER_ID,
		}
		latest, err := impl.cdWorkflowService.CheckIfLatestWf(cdWorkflow.PipelineId, cdWorkflow.Id)
		if err != nil {
			impl.logger.Errorw("error in determining latest", "wf", cdWorkflow, "err", err)
			wf.WorkflowStatus = pipelineConfig.DEQUE_ERROR
			err = impl.cdWorkflowService.UpdateWorkFlow(wf)
			if err != nil {
				impl.logger.Errorw("error in updating wf", "err", err, "req", wf)
			}
			return
		}
		if !latest {
			wf.WorkflowStatus = pipelineConfig.DROPPED_STALE
			err = impl.cdWorkflowService.UpdateWorkFlow(wf)
			if err != nil {
				impl.logger.Errorw("error in updating wf", "err", err, "req", wf)
			}
			return
		}
		pipelineObj, err := impl.pipelineRepository.FindById(cdWorkflow.PipelineId)
		if err != nil {
			impl.logger.Errorw("error in fetching pipeline", "err", err)
			wf.WorkflowStatus = pipelineConfig.TRIGGER_ERROR
			err = impl.cdWorkflowService.UpdateWorkFlow(wf)
			if err != nil {
				impl.logger.Errorw("error in updating wf", "err", err, "req", wf)
			}
			return
		}
		artifact, err := impl.ciArtifactRepository.Get(cdWorkflow.CiArtifactId)
		if err != nil {
			impl.logger.Errorw("error in fetching artefact", "err", err)
			wf.WorkflowStatus = pipelineConfig.TRIGGER_ERROR
			err = impl.cdWorkflowService.UpdateWorkFlow(wf)
			if err != nil {
				impl.logger.Errorw("error in updating wf", "err", err, "req", wf)
			}
			return
		}
		// Migration of deprecated DataSource Type
		if artifact.IsMigrationRequired() {
			migrationErr := impl.ciArtifactRepository.MigrateToWebHookDataSourceType(artifact.Id)
			if migrationErr != nil {
				impl.logger.Warnw("unable to migrate deprecated DataSource", "artifactId", artifact.Id)
			}
		}
		triggerContext := pipeline.TriggerContext{
			ReferenceId: pointer.String(msg.MsgId),
		}

		triggerRequest := pipeline.TriggerRequest{
			CdWf:           adapter.ConvertCdWorkflowDtoToDbObj(wf), //TODO: update object from db to dto
			Artifact:       artifact,
			Pipeline:       pipelineObj,
			TriggeredBy:    cdWorkflow.CreatedBy,
			ApplyAuth:      false,
			TriggerContext: triggerContext,
		}
		err = impl.workflowDagExecutor.TriggerStageForBulk(triggerRequest)
		if err != nil {
			impl.logger.Errorw("error in cd trigger ", "err", err)
			wf.WorkflowStatus = pipelineConfig.TRIGGER_ERROR
		} else {
			wf.WorkflowStatus = pipelineConfig.WF_STARTED
		}
		err = impl.cdWorkflowService.UpdateWorkFlow(wf)
		if err != nil {
			impl.logger.Errorw("error in updating wf", "err", err, "req", wf)
		}
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		cdWorkflow := new(pipelineConfig.CdWorkflow)
		err := json.Unmarshal([]byte(msg.Data), cdWorkflow)
		if err != nil {
			return "error while unmarshalling cdWorkflow json object", []interface{}{"error", err}
		}
		return "got message for bulk deploy", []interface{}{"cdWorkflowId", cdWorkflow.Id}
	}

	validations := impl.workflowDagExecutor.GetTriggerValidateFuncs()
	return impl.pubSubClient.Subscribe(pubsub.BULK_DEPLOY_TOPIC, callback, loggerFunc, validations...)
}

func (impl *WorkflowEventProcessorImpl) SubscribeHibernateBulkAction() error {
	callback := func(msg *model.PubSubMsg) {
		deploymentGroupAppWithEnv := new(pipeline.DeploymentGroupAppWithEnv)
		err := json.Unmarshal([]byte(msg.Data), deploymentGroupAppWithEnv)
		if err != nil {
			impl.logger.Error("Error while unmarshalling deploymentGroupAppWithEnv json object", err)
			return
		}

		stopAppRequest := &pipeline.StopAppRequest{
			AppId:         deploymentGroupAppWithEnv.AppId,
			EnvironmentId: deploymentGroupAppWithEnv.EnvironmentId,
			UserId:        deploymentGroupAppWithEnv.UserId,
			RequestType:   deploymentGroupAppWithEnv.RequestType,
		}
		ctx, err := impl.argoUserService.BuildACDContext()
		if err != nil {
			impl.logger.Errorw("error in creating acd sync context", "err", err)
			return
		}
		triggerContext := pipeline.TriggerContext{
			ReferenceId: pointer.String(msg.MsgId),
			Context:     ctx,
		}
		_, err = impl.workflowDagExecutor.StopStartApp(triggerContext, stopAppRequest)
		if err != nil {
			impl.logger.Errorw("error in stop app request", "err", err)
			return
		}
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		deploymentGroupAppWithEnv := new(pipeline.DeploymentGroupAppWithEnv)
		err := json.Unmarshal([]byte(msg.Data), deploymentGroupAppWithEnv)
		if err != nil {
			return "error while unmarshalling deploymentGroupAppWithEnv json object", []interface{}{"err", err}
		}
		return "got message for bulk hibernate", []interface{}{"deploymentGroupId", deploymentGroupAppWithEnv.DeploymentGroupId, "appId", deploymentGroupAppWithEnv.AppId, "environmentId", deploymentGroupAppWithEnv.EnvironmentId}
	}

	err := impl.pubSubClient.Subscribe(pubsub.BULK_HIBERNATE_TOPIC, callback, loggerFunc)
	return err
}
