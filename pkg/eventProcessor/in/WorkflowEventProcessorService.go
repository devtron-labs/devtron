package in

import (
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	pubsub2 "github.com/devtron-labs/devtron/api/router/pubsub"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean4 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/deployedApp"
	bean6 "github.com/devtron-labs/devtron/pkg/deployment/deployedApp/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	bean5 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/adapter"
	bean3 "github.com/devtron-labs/devtron/pkg/workflow/cd/bean"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	util "github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"k8s.io/utils/pointer"
	"strconv"
	"time"
)

type WorkflowEventProcessorImpl struct {
	logger                  *zap.SugaredLogger
	pubSubClient            *pubsub.PubSubClientServiceImpl
	cdWorkflowService       cd.CdWorkflowService
	cdWorkflowRunnerService cd.CdWorkflowRunnerService
	workflowDagExecutor     pipeline.WorkflowDagExecutor
	argoUserService         argo.ArgoUserService
	ciHandler               pipeline.CiHandler
	cdHandler               pipeline.CdHandler
	eventFactory            client.EventFactory
	eventClient             client.EventClient
	cdTriggerService        devtronApps.TriggerService
	deployedAppService      deployedApp.DeployedAppService
	webhookService          pipeline.WebhookService
	validator               *validator.Validate
	ciEventConfig           *pubsub2.CiEventConfig
	//repositories import to be removed
	pipelineRepository   pipelineConfig.PipelineRepository
	ciArtifactRepository repository.CiArtifactRepository
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository
}

func NewWorkflowEventProcessorImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	cdWorkflowService cd.CdWorkflowService,
	cdWorkflowRunnerService cd.CdWorkflowRunnerService,
	workflowDagExecutor pipeline.WorkflowDagExecutor,
	argoUserService argo.ArgoUserService,
	ciHandler pipeline.CiHandler, cdHandler pipeline.CdHandler,
	eventFactory client.EventFactory, eventClient client.EventClient,
	cdTriggerService devtronApps.TriggerService,
	deployedAppService deployedApp.DeployedAppService,
	webhookService pipeline.WebhookService,
	validator *validator.Validate,
	ciEventConfig *pubsub2.CiEventConfig,
	pipelineRepository pipelineConfig.PipelineRepository,
	ciArtifactRepository repository.CiArtifactRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository) (*WorkflowEventProcessorImpl, error) {
	impl := &WorkflowEventProcessorImpl{
		logger:                  logger,
		pubSubClient:            pubSubClient,
		cdWorkflowService:       cdWorkflowService,
		cdWorkflowRunnerService: cdWorkflowRunnerService,
		argoUserService:         argoUserService,
		ciHandler:               ciHandler,
		cdHandler:               cdHandler,
		eventFactory:            eventFactory,
		eventClient:             eventClient,
		workflowDagExecutor:     workflowDagExecutor,
		cdTriggerService:        cdTriggerService,
		deployedAppService:      deployedAppService,
		webhookService:          webhookService,
		validator:               validator,
		ciEventConfig:           ciEventConfig,
		pipelineRepository:      pipelineRepository,
		ciArtifactRepository:    ciArtifactRepository,
		cdWorkflowRepository:    cdWorkflowRepository,
	}
	return impl, nil
}

func (impl *WorkflowEventProcessorImpl) SubscribeCDStageCompleteEvent() error {
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
		triggerContext := bean5.TriggerContext{
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
		triggerContext := bean5.TriggerContext{
			ReferenceId: pointer.String(msg.MsgId),
		}

		triggerRequest := bean5.TriggerRequest{
			CdWf:           adapter.ConvertCdWorkflowDtoToDbObj(wf), //TODO: update object from db to dto
			Artifact:       artifact,
			Pipeline:       pipelineObj,
			TriggeredBy:    cdWorkflow.CreatedBy,
			ApplyAuth:      false,
			TriggerContext: triggerContext,
		}
		err = impl.cdTriggerService.TriggerStageForBulk(triggerRequest)
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

		stopAppRequest := &bean6.StopAppRequest{
			AppId:         deploymentGroupAppWithEnv.AppId,
			EnvironmentId: deploymentGroupAppWithEnv.EnvironmentId,
			UserId:        deploymentGroupAppWithEnv.UserId,
			RequestType:   deploymentGroupAppWithEnv.RequestType,
			ReferenceId:   pointer.String(msg.MsgId),
		}
		ctx, err := impl.argoUserService.BuildACDContext()
		if err != nil {
			impl.logger.Errorw("error in creating acd sync context", "err", err)
			return
		}
		_, err = impl.deployedAppService.StopStartApp(ctx, stopAppRequest)
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

func (impl *WorkflowEventProcessorImpl) SubscribeCIWorkflowStatusUpdate() error {
	callback := func(msg *model.PubSubMsg) {
		wfStatus := v1alpha1.WorkflowStatus{}
		err := json.Unmarshal([]byte(msg.Data), &wfStatus)
		if err != nil {
			impl.logger.Errorw("error while unmarshalling wf status update", "err", err, "msg", msg.Data)
			return
		}

		err = impl.ciHandler.CheckAndReTriggerCI(wfStatus)
		if err != nil {
			impl.logger.Errorw("error in checking and re triggering ci", "err", err)
			//don't return as we have to update the workflow status
		}

		_, err = impl.ciHandler.UpdateWorkflow(wfStatus)
		if err != nil {
			impl.logger.Errorw("error on update workflow status", "err", err, "msg", msg.Data)
			return
		}

	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		wfStatus := v1alpha1.WorkflowStatus{}
		err := json.Unmarshal([]byte(msg.Data), &wfStatus)
		if err != nil {
			return "error while unmarshalling wf status update", []interface{}{"err", err, "msg", msg.Data}
		}

		workflowName, status, _, message, _, _ := pipeline.ExtractWorkflowStatus(wfStatus)
		return "got message for ci workflow status update ", []interface{}{"workflowName", workflowName, "status", status, "message", message}
	}

	err := impl.pubSubClient.Subscribe(pubsub.WORKFLOW_STATUS_UPDATE_TOPIC, callback, loggerFunc)

	if err != nil {
		impl.logger.Error("err", err)
		return err
	}
	return nil
}

func (impl *WorkflowEventProcessorImpl) SubscribeCDWorkflowStatusUpdate() error {
	callback := func(msg *model.PubSubMsg) {
		wfStatus := v1alpha1.WorkflowStatus{}
		err := json.Unmarshal([]byte(msg.Data), &wfStatus)
		if err != nil {
			impl.logger.Error("Error while unmarshalling wfStatus json object", "error", err)
			return
		}

		wfrId, wfrStatus, err := impl.cdHandler.UpdateWorkflow(wfStatus)
		impl.logger.Debugw("UpdateWorkflow", "wfrId", wfrId, "wfrStatus", wfrStatus)
		if err != nil {
			impl.logger.Error("err", err)
			return
		}

		wfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(wfrId)
		if err != nil {
			impl.logger.Errorw("could not get wf runner", "err", err)
			return
		}
		if wfrStatus == string(v1alpha1.NodeFailed) || wfrStatus == string(v1alpha1.NodeError) {
			if len(wfr.ImagePathReservationIds) > 0 {
				err := impl.cdHandler.DeactivateImageReservationPathsOnFailure(wfr.ImagePathReservationIds)
				if err != nil {
					impl.logger.Errorw("error in removing image path reservation ")
				}
			}
		}
		if wfrStatus == string(v1alpha1.NodeSucceeded) || wfrStatus == string(v1alpha1.NodeFailed) || wfrStatus == string(v1alpha1.NodeError) {
			eventType := util.EventType(0)
			if wfrStatus == string(v1alpha1.NodeSucceeded) {
				eventType = util.Success
			} else if wfrStatus == string(v1alpha1.NodeFailed) || wfrStatus == string(v1alpha1.NodeError) {
				eventType = util.Fail
			}

			if wfr != nil && executors.CheckIfReTriggerRequired(wfrStatus, wfStatus.Message, wfr.Status) {
				err = impl.cdHandler.HandleCdStageReTrigger(wfr)
				if err != nil {
					//check if this log required or not
					impl.logger.Errorw("error in HandleCdStageReTrigger", "error", err)
				}
			}

			if wfr.WorkflowType == bean2.CD_WORKFLOW_TYPE_PRE || wfr.WorkflowType == bean2.CD_WORKFLOW_TYPE_POST {
				event := impl.eventFactory.Build(eventType, &wfr.CdWorkflow.PipelineId, wfr.CdWorkflow.Pipeline.AppId, &wfr.CdWorkflow.Pipeline.EnvironmentId, util.CD)
				impl.logger.Debugw("event pre stage", "event", event)
				event = impl.eventFactory.BuildExtraCDData(event, wfr, 0, wfr.WorkflowType)
				_, evtErr := impl.eventClient.WriteNotificationEvent(event)
				if evtErr != nil {
					impl.logger.Errorw("CD stage post fail or success event unable to sent", "error", evtErr)
				}
			}
		}
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		wfStatus := v1alpha1.WorkflowStatus{}
		err := json.Unmarshal([]byte(msg.Data), &wfStatus)
		if err != nil {
			return "error while unmarshalling wfStatus json object", []interface{}{"error", err}
		}
		workflowName, status, _, message, _, _ := pipeline.ExtractWorkflowStatus(wfStatus)
		return "got message for cd workflow status", []interface{}{"workflowName", workflowName, "status", status, "message", message}
	}

	err := impl.pubSubClient.Subscribe(pubsub.CD_WORKFLOW_STATUS_UPDATE, callback, loggerFunc)
	if err != nil {
		impl.logger.Error("err", err)
		return err
	}
	return nil
}

func (impl *WorkflowEventProcessorImpl) SubscribeCICompleteEvent() error {
	callback := func(msg *model.PubSubMsg) {
		ciCompleteEvent := bean.CiCompleteEvent{}
		err := json.Unmarshal([]byte(string(msg.Data)), &ciCompleteEvent)
		if err != nil {
			impl.logger.Error("error while unmarshalling json data", "error", err)
			return
		}
		impl.logger.Debugw("ci complete event for ci", "ciPipelineId", ciCompleteEvent.PipelineId)
		req, err := impl.BuildCiArtifactRequest(ciCompleteEvent)
		if err != nil {
			return
		}

		triggerContext := bean5.TriggerContext{
			ReferenceId: pointer.String(msg.MsgId),
		}

		if ciCompleteEvent.FailureReason != "" {
			req.FailureReason = ciCompleteEvent.FailureReason
			err := impl.webhookService.HandleCiStepFailedEvent(ciCompleteEvent.PipelineId, req)
			if err != nil {
				impl.logger.Error("Error while sending event for CI failure for pipelineID: ",
					ciCompleteEvent.PipelineId, "request: ", req, "error: ", err)
				return
			}
		} else if ciCompleteEvent.ImageDetailsFromCR != nil {
			if len(ciCompleteEvent.ImageDetailsFromCR.ImageDetails) > 0 {
				imageDetails := util2.GetReverseSortedImageDetails(ciCompleteEvent.ImageDetailsFromCR.ImageDetails)
				digestWorkflowMap, err := impl.webhookService.HandleMultipleImagesFromEvent(imageDetails, *ciCompleteEvent.WorkflowId)
				if err != nil {
					impl.logger.Errorw("error in getting digest workflow map", "err", err, "workflowId", ciCompleteEvent.WorkflowId)
					return
				}
				for _, detail := range imageDetails {
					request, err := impl.BuildCIArtifactRequestForImageFromCR(detail, ciCompleteEvent.ImageDetailsFromCR.Region, ciCompleteEvent, digestWorkflowMap[*detail.ImageDigest].Id)
					if err != nil {
						impl.logger.Error("Error while creating request for pipelineID", "pipelineId", ciCompleteEvent.PipelineId, "err", err)
						return
					}
					resp, err := impl.ValidateAndHandleCiSuccessEvent(triggerContext, ciCompleteEvent.PipelineId, request, detail.ImagePushedAt)
					if err != nil {
						return
					}
					impl.logger.Debug("response of handle ci success event for multiple images from plugin", "resp", resp)
				}
			}

		} else {
			util2.TriggerCIMetrics(ciCompleteEvent.Metrics, impl.ciEventConfig.ExposeCiMetrics, ciCompleteEvent.PipelineName, ciCompleteEvent.AppName)
			resp, err := impl.ValidateAndHandleCiSuccessEvent(triggerContext, ciCompleteEvent.PipelineId, req, &time.Time{})
			if err != nil {
				return
			}
			impl.logger.Debug(resp)
		}
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		ciCompleteEvent := bean.CiCompleteEvent{}
		err := json.Unmarshal([]byte(msg.Data), &ciCompleteEvent)
		if err != nil {
			return "error while unmarshalling json data", []interface{}{"error", err}
		}
		return "got message for ci-completion", []interface{}{"ciPipelineId", ciCompleteEvent.PipelineId, "workflowId", ciCompleteEvent.WorkflowId}
	}

	validations := impl.webhookService.GetTriggerValidateFuncs()
	err := impl.pubSubClient.Subscribe(pubsub.CI_COMPLETE_TOPIC, callback, loggerFunc, validations...)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}

func (impl *WorkflowEventProcessorImpl) ValidateAndHandleCiSuccessEvent(triggerContext bean5.TriggerContext, ciPipelineId int, request *pipeline.CiArtifactWebhookRequest, imagePushedAt *time.Time) (int, error) {
	validationErr := impl.validator.Struct(request)
	if validationErr != nil {
		impl.logger.Errorw("validation err, HandleCiSuccessEvent", "err", validationErr, "payload", request)
		return 0, validationErr
	}
	buildArtifactId, err := impl.webhookService.HandleCiSuccessEvent(triggerContext, ciPipelineId, request, imagePushedAt)
	if err != nil {
		impl.logger.Error("Error while sending event for CI success for pipelineID",
			ciPipelineId, "request", request, "error", err)
		return 0, err
	}
	return buildArtifactId, nil
}

func (impl *WorkflowEventProcessorImpl) BuildCiArtifactRequest(event bean.CiCompleteEvent) (*pipeline.CiArtifactWebhookRequest, error) {
	var ciMaterialInfos []repository.CiMaterialInfo
	for _, p := range event.CiProjectDetails {
		var modifications []repository.Modification

		var branch string
		var tag string
		var webhookData repository.WebhookData
		if p.SourceType == pipelineConfig.SOURCE_TYPE_BRANCH_FIXED {
			branch = p.SourceValue
		} else if p.SourceType == pipelineConfig.SOURCE_TYPE_WEBHOOK {
			webhookData = repository.WebhookData{
				Id:              p.WebhookData.Id,
				EventActionType: p.WebhookData.EventActionType,
				Data:            p.WebhookData.Data,
			}
		}

		modification := repository.Modification{
			Revision:     p.CommitHash,
			ModifiedTime: p.CommitTime,
			Author:       p.Author,
			Branch:       branch,
			Tag:          tag,
			WebhookData:  webhookData,
			Message:      p.Message,
		}

		modifications = append(modifications, modification)
		ciMaterialInfo := repository.CiMaterialInfo{
			Material: repository.Material{
				GitConfiguration: repository.GitConfiguration{
					URL: p.GitRepository,
				},
				Type: event.MaterialType,
			},
			Changed:       true,
			Modifications: modifications,
		}
		ciMaterialInfos = append(ciMaterialInfos, ciMaterialInfo)
	}

	materialBytes, err := json.Marshal(ciMaterialInfos)
	if err != nil {
		impl.logger.Errorw("cannot build ci artifact req", "err", err)
		return nil, err
	}
	rawMaterialInfo := json.RawMessage(materialBytes)
	fmt.Printf("Raw Message : %s\n", rawMaterialInfo)

	if event.TriggeredBy == 0 {
		event.TriggeredBy = 1 // system triggered event
	}

	request := &pipeline.CiArtifactWebhookRequest{
		Image:                         event.DockerImage,
		ImageDigest:                   event.Digest,
		DataSource:                    event.DataSource,
		PipelineName:                  event.PipelineName,
		MaterialInfo:                  rawMaterialInfo,
		UserId:                        event.TriggeredBy,
		WorkflowId:                    event.WorkflowId,
		IsArtifactUploaded:            event.IsArtifactUploaded,
		PluginRegistryArtifactDetails: event.PluginRegistryArtifactDetails,
		PluginArtifactStage:           event.PluginArtifactStage,
	}
	// if DataSource is empty, repository.WEBHOOK is considered as default
	if request.DataSource == "" {
		request.DataSource = repository.WEBHOOK
	}
	return request, nil
}

func (impl *WorkflowEventProcessorImpl) BuildCIArtifactRequestForImageFromCR(imageDetails types.ImageDetail, region string, event bean.CiCompleteEvent, workflowId int) (*pipeline.CiArtifactWebhookRequest, error) {
	if event.TriggeredBy == 0 {
		event.TriggeredBy = 1 // system triggered event
	}
	request := &pipeline.CiArtifactWebhookRequest{
		Image:              util2.ExtractEcrImage(*imageDetails.RegistryId, region, *imageDetails.RepositoryName, imageDetails.ImageTags[0]),
		ImageDigest:        *imageDetails.ImageDigest,
		DataSource:         event.DataSource,
		PipelineName:       event.PipelineName,
		UserId:             event.TriggeredBy,
		WorkflowId:         &workflowId,
		IsArtifactUploaded: event.IsArtifactUploaded,
	}
	if request.DataSource == "" {
		request.DataSource = repository.WEBHOOK
	}
	return request, nil
}
