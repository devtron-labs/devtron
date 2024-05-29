/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package in

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	apiBean "github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util3 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	userBean "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/deployedApp"
	deploymentBean "github.com/devtron-labs/devtron/pkg/deployment/deployedApp/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	triggerAdapter "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/adapter"
	triggerBean "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest/service"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	eventProcessorBean "github.com/devtron-labs/devtron/pkg/eventProcessor/out/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/adapter"
	cdWorkflowBean "github.com/devtron-labs/devtron/pkg/workflow/cd/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/dag"
	wrokflowDagBean "github.com/devtron-labs/devtron/pkg/workflow/dag/bean"
	globalUtil "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	cronUtil "github.com/devtron-labs/devtron/util/cron"
	eventUtil "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"gopkg.in/go-playground/validator.v9"
	"k8s.io/utils/pointer"
	"strconv"
	"sync"
	"time"
)

type WorkflowEventProcessorImpl struct {
	logger                       *zap.SugaredLogger
	pubSubClient                 *pubsub.PubSubClientServiceImpl
	cdWorkflowService            cd.CdWorkflowService
	cdWorkflowRunnerService      cd.CdWorkflowRunnerService
	workflowDagExecutor          dag.WorkflowDagExecutor
	argoUserService              argo.ArgoUserService
	ciHandler                    pipeline.CiHandler
	cdHandler                    pipeline.CdHandler
	eventFactory                 client.EventFactory
	eventClient                  client.EventClient
	cdTriggerService             devtronApps.TriggerService
	deployedAppService           deployedApp.DeployedAppService
	webhookService               pipeline.WebhookService
	validator                    *validator.Validate
	globalEnvVariables           *globalUtil.GlobalEnvVariables
	cdWorkflowCommonService      cd.CdWorkflowCommonService
	cdPipelineConfigService      pipeline.CdPipelineConfigService
	userDeploymentRequestService service.UserDeploymentRequestService

	devtronAppReleaseContextMap     map[int]bean.DevtronAppReleaseContextType
	devtronAppReleaseContextMapLock *sync.Mutex
	appServiceConfig                *app.AppServiceConfig

	// repositories import to be removed
	pipelineRepository   pipelineConfig.PipelineRepository
	ciArtifactRepository repository.CiArtifactRepository
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository
}

func NewWorkflowEventProcessorImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	cdWorkflowService cd.CdWorkflowService,
	cdWorkflowRunnerService cd.CdWorkflowRunnerService,
	workflowDagExecutor dag.WorkflowDagExecutor,
	argoUserService argo.ArgoUserService,
	ciHandler pipeline.CiHandler, cdHandler pipeline.CdHandler,
	eventFactory client.EventFactory, eventClient client.EventClient,
	cdTriggerService devtronApps.TriggerService,
	deployedAppService deployedApp.DeployedAppService,
	webhookService pipeline.WebhookService,
	validator *validator.Validate,
	envVariables *globalUtil.EnvironmentVariables,
	cdWorkflowCommonService cd.CdWorkflowCommonService,
	cdPipelineConfigService pipeline.CdPipelineConfigService,
	userDeploymentRequestService service.UserDeploymentRequestService,
	pipelineRepository pipelineConfig.PipelineRepository,
	ciArtifactRepository repository.CiArtifactRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	cronLogger *cronUtil.CronLoggerImpl) (*WorkflowEventProcessorImpl, error) {
	impl := &WorkflowEventProcessorImpl{
		logger:                          logger,
		pubSubClient:                    pubSubClient,
		cdWorkflowService:               cdWorkflowService,
		cdWorkflowRunnerService:         cdWorkflowRunnerService,
		argoUserService:                 argoUserService,
		ciHandler:                       ciHandler,
		cdHandler:                       cdHandler,
		eventFactory:                    eventFactory,
		eventClient:                     eventClient,
		workflowDagExecutor:             workflowDagExecutor,
		cdTriggerService:                cdTriggerService,
		deployedAppService:              deployedAppService,
		webhookService:                  webhookService,
		validator:                       validator,
		globalEnvVariables:              envVariables.GlobalEnvVariables,
		cdWorkflowCommonService:         cdWorkflowCommonService,
		cdPipelineConfigService:         cdPipelineConfigService,
		userDeploymentRequestService:    userDeploymentRequestService,
		devtronAppReleaseContextMap:     make(map[int]bean.DevtronAppReleaseContextType),
		devtronAppReleaseContextMapLock: &sync.Mutex{},
		pipelineRepository:              pipelineRepository,
		ciArtifactRepository:            ciArtifactRepository,
		cdWorkflowRepository:            cdWorkflowRepository,
	}
	appServiceConfig, err := app.GetAppServiceConfig()
	if err != nil {
		return nil, err
	}
	impl.appServiceConfig = appServiceConfig
	newCron := cron.New(
		cron.WithChain(cron.Recover(cronLogger)))
	newCron.Start()
	_, err = newCron.AddFunc(fmt.Sprintf("@every %dm", appServiceConfig.DevtronChartArgoCdInstallRequestTimeout), impl.ProcessIncompleteDeploymentReq)
	if err != nil {
		logger.Errorw("error while configure cron job for ci workflow status update", "err", err)
		return impl, err
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
		triggerContext := triggerBean.TriggerContext{
			ReferenceId: pointer.String(msg.MsgId),
		}
		if wfr.WorkflowType == apiBean.CD_WORKFLOW_TYPE_PRE {
			impl.logger.Debugw("received pre stage success event for workflow runner ", "wfId", strconv.Itoa(wfr.Id))
			err = impl.workflowDagExecutor.HandlePreStageSuccessEvent(triggerContext, cdStageCompleteEvent)
			if err != nil {
				impl.logger.Errorw("deployment success event error", "err", err)
				return
			}
		} else if wfr.WorkflowType == apiBean.CD_WORKFLOW_TYPE_POST {
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

	validations := impl.cdWorkflowCommonService.GetTriggerValidateFuncs()

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
		wf := &cdWorkflowBean.CdWorkflowDto{
			Id:           cdWorkflow.Id,
			CiArtifactId: cdWorkflow.CiArtifactId,
			PipelineId:   cdWorkflow.PipelineId,
			UserId:       userBean.SYSTEM_USER_ID,
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
		triggerContext := triggerBean.TriggerContext{
			ReferenceId: pointer.String(msg.MsgId),
		}

		triggerRequest := triggerBean.TriggerRequest{
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

	validations := impl.cdWorkflowCommonService.GetTriggerValidateFuncs()
	return impl.pubSubClient.Subscribe(pubsub.BULK_DEPLOY_TOPIC, callback, loggerFunc, validations...)
}

func (impl *WorkflowEventProcessorImpl) SubscribeHibernateBulkAction() error {
	callback := func(msg *model.PubSubMsg) {
		deploymentGroupAppWithEnv := new(eventProcessorBean.DeploymentGroupAppWithEnv)
		err := json.Unmarshal([]byte(msg.Data), deploymentGroupAppWithEnv)
		if err != nil {
			impl.logger.Error("Error while unmarshalling deploymentGroupAppWithEnv json object", err)
			return
		}

		stopAppRequest := &deploymentBean.StopAppRequest{
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
		deploymentGroupAppWithEnv := new(eventProcessorBean.DeploymentGroupAppWithEnv)
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
			eventType := eventUtil.EventType(0)
			if wfrStatus == string(v1alpha1.NodeSucceeded) {
				eventType = eventUtil.Success
			} else if wfrStatus == string(v1alpha1.NodeFailed) || wfrStatus == string(v1alpha1.NodeError) {
				eventType = eventUtil.Fail
			}

			if wfr != nil && executors.CheckIfReTriggerRequired(wfrStatus, wfStatus.Message, wfr.Status) {
				err = impl.workflowDagExecutor.HandleCdStageReTrigger(wfr)
				if err != nil {
					//check if this log required or not
					impl.logger.Errorw("error in HandleCdStageReTrigger", "error", err)
				}
			}

			if wfr.WorkflowType == apiBean.CD_WORKFLOW_TYPE_PRE || wfr.WorkflowType == apiBean.CD_WORKFLOW_TYPE_POST {
				event := impl.eventFactory.Build(eventType, &wfr.CdWorkflow.PipelineId, wfr.CdWorkflow.Pipeline.AppId, &wfr.CdWorkflow.Pipeline.EnvironmentId, eventUtil.CD)
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
		err := json.Unmarshal([]byte(msg.Data), &ciCompleteEvent)
		if err != nil {
			impl.logger.Error("error while unmarshalling json data", "error", err)
			return
		}
		impl.logger.Debugw("ci complete event for ci", "ciPipelineId", ciCompleteEvent.PipelineId)
		req, err := impl.BuildCiArtifactRequest(ciCompleteEvent)
		if err != nil {
			return
		}

		triggerContext := triggerBean.TriggerContext{
			Context:     context.Background(),
			ReferenceId: pointer.String(msg.MsgId),
		}

		if ciCompleteEvent.FailureReason != "" {
			req.FailureReason = ciCompleteEvent.FailureReason
			err := impl.workflowDagExecutor.HandleCiStepFailedEvent(ciCompleteEvent.PipelineId, req)
			if err != nil {
				impl.logger.Error("Error while sending event for CI failure for pipelineID: ",
					ciCompleteEvent.PipelineId, "request: ", req, "error: ", err)
				return
			}
		} else if ciCompleteEvent.ImageDetailsFromCR != nil {
			if len(ciCompleteEvent.ImageDetailsFromCR.ImageDetails) > 0 {
				imageDetails := globalUtil.GetReverseSortedImageDetails(ciCompleteEvent.ImageDetailsFromCR.ImageDetails)
				digestWorkflowMap, err := impl.webhookService.HandleMultipleImagesFromEvent(imageDetails, *ciCompleteEvent.WorkflowId)
				if err != nil {
					impl.logger.Errorw("error in getting digest workflow map", "err", err, "workflowId", ciCompleteEvent.WorkflowId)
					return
				}
				for _, detail := range imageDetails {
					if detail.ImageTags == nil {
						continue
					}
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
			globalUtil.TriggerCIMetrics(ciCompleteEvent.Metrics, impl.globalEnvVariables.ExposeCiMetrics, ciCompleteEvent.PipelineName, ciCompleteEvent.AppName)
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

func (impl *WorkflowEventProcessorImpl) ValidateAndHandleCiSuccessEvent(triggerContext triggerBean.TriggerContext, ciPipelineId int, request *wrokflowDagBean.CiArtifactWebhookRequest, imagePushedAt *time.Time) (int, error) {
	validationErr := impl.validator.Struct(request)
	if validationErr != nil {
		impl.logger.Errorw("validation err, HandleCiSuccessEvent", "err", validationErr, "payload", request)
		return 0, validationErr
	}
	buildArtifactId, err := impl.workflowDagExecutor.HandleCiSuccessEvent(triggerContext, ciPipelineId, request, imagePushedAt)
	if err != nil {
		impl.logger.Error("Error while sending event for CI success for pipelineID",
			ciPipelineId, "request", request, "error", err)
		return 0, err
	}
	return buildArtifactId, nil
}

func (impl *WorkflowEventProcessorImpl) BuildCiArtifactRequest(event bean.CiCompleteEvent) (*wrokflowDagBean.CiArtifactWebhookRequest, error) {
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

	request := &wrokflowDagBean.CiArtifactWebhookRequest{
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

func (impl *WorkflowEventProcessorImpl) BuildCIArtifactRequestForImageFromCR(imageDetails types.ImageDetail, region string, event bean.CiCompleteEvent, workflowId int) (*wrokflowDagBean.CiArtifactWebhookRequest, error) {
	if event.TriggeredBy == 0 {
		event.TriggeredBy = 1 // system triggered event
	}
	request := &wrokflowDagBean.CiArtifactWebhookRequest{
		Image:              globalUtil.ExtractEcrImage(*imageDetails.RegistryId, region, *imageDetails.RepositoryName, imageDetails.ImageTags[0]),
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

func (impl *WorkflowEventProcessorImpl) SubscribeDevtronAsyncInstallRequest() error {
	callback := func(msg *model.PubSubMsg) {
		cdAsyncInstallReq, err := impl.extractAsyncCdDeployRequestFromEventMsg(msg)
		if err != nil {
			impl.logger.Errorw("err on extracting override request, SubscribeDevtronAsyncInstallRequest", "err", err)
			return
		}
		_ = impl.ProcessConcurrentAsyncDeploymentReq(cdAsyncInstallReq)
		return
	}

	err := impl.pubSubClient.Subscribe(pubsub.DEVTRON_CHART_INSTALL_TOPIC, callback, getAsyncDeploymentLoggerFunc(fmt.Sprintf("async Helm")))
	if err != nil {
		impl.logger.Error(err)
		return err
	}

	err = impl.pubSubClient.Subscribe(pubsub.DEVTRON_CHART_PRIORITY_INSTALL_TOPIC, callback, getAsyncDeploymentLoggerFunc(fmt.Sprintf("priority async Helm")))
	if err != nil {
		impl.logger.Error(err)
		return err
	}

	err = impl.pubSubClient.Subscribe(pubsub.DEVTRON_CHART_GITOPS_INSTALL_TOPIC, callback, getAsyncDeploymentLoggerFunc(fmt.Sprintf("async ArgoCd")))
	if err != nil {
		impl.logger.Error(err)
		return err
	}

	err = impl.pubSubClient.Subscribe(pubsub.DEVTRON_CHART_GITOPS_PRIORITY_INSTALL_TOPIC, callback, getAsyncDeploymentLoggerFunc(fmt.Sprintf("priority async ArgoCd")))
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}

func getAsyncDeploymentLoggerFunc(topicType string) pubsub.LoggerFunc {
	return func(msg model.PubSubMsg) (string, []interface{}) {
		cdAsyncInstallReq := &bean.AsyncCdDeployRequest{}
		err := json.Unmarshal([]byte(msg.Data), cdAsyncInstallReq)
		if err != nil {
			return fmt.Sprintf("error in unmarshalling CD Pipeline %s install request nats message", topicType), []interface{}{"err", err}
		}
		return fmt.Sprintf("got message for devtron chart %s install", topicType), []interface{}{"appId", cdAsyncInstallReq.ValuesOverrideRequest.AppId, "pipelineId", cdAsyncInstallReq.ValuesOverrideRequest.PipelineId, "artifactId", cdAsyncInstallReq.ValuesOverrideRequest.CiArtifactId}
	}
}

func (impl *WorkflowEventProcessorImpl) extractAsyncCdDeployRequestFromEventMsg(msg *model.PubSubMsg) (*bean.AsyncCdDeployRequest, error) {
	cdAsyncInstallReq := &bean.AsyncCdDeployRequest{}
	err := json.Unmarshal([]byte(msg.Data), cdAsyncInstallReq)
	if err != nil {
		impl.logger.Errorw("error in unmarshalling CD async install request nats message", "err", err)
		return nil, err
	}
	if cdAsyncInstallReq.UserDeploymentRequestId != 0 && cdAsyncInstallReq.ValuesOverrideRequest == nil {
		impl.logger.Errorw("invalid async cd pipeline deployment request", "msg", msg.Data)
		return nil, fmt.Errorf("invalid async cd pipeline deployment request")
	}
	if cdAsyncInstallReq.UserDeploymentRequestId != 0 {
		cdAsyncInstallReq, err = impl.userDeploymentRequestService.GetAsyncCdDeployRequestById(cdAsyncInstallReq.UserDeploymentRequestId)
		if err != nil {
			impl.logger.Errorw("error in fetching userDeploymentRequest by id", "userDeploymentRequestId", cdAsyncInstallReq.UserDeploymentRequestId, "err", err)
			return nil, err
		}
	}
	err = impl.setAdditionalDataInAsyncInstallReq(cdAsyncInstallReq)
	if err != nil {
		impl.logger.Errorw("error in setting additional data to AsyncCdDeployRequest", "err", err)
		return nil, err
	}
	return cdAsyncInstallReq, nil
}

func (impl *WorkflowEventProcessorImpl) SubscribeCDPipelineDeleteEvent() error {
	callback := func(msg *model.PubSubMsg) {
		cdPipelineDeleteEvent := &eventProcessorBean.CdPipelineDeleteEvent{}
		err := json.Unmarshal([]byte(msg.Data), cdPipelineDeleteEvent)
		if err != nil {
			impl.logger.Errorw("error while unmarshalling cdPipelineDeleteEvent object", "err", err, "msg", msg.Data)
			return
		}
		pipeline, err := impl.pipelineRepository.FindByIdEvenIfInactive(cdPipelineDeleteEvent.PipelineId)
		if err != nil {
			impl.logger.Errorw("error in fetching pipeline by pipelineId", "err", err, "pipelineId", cdPipelineDeleteEvent.PipelineId)
			return
		}
		if util3.IsHelmApp(pipeline.DeploymentAppType) || util3.IsAcdApp(pipeline.DeploymentAppType) {
			impl.RemoveReleaseContextForPipeline(cdPipelineDeleteEvent.PipelineId, cdPipelineDeleteEvent.TriggeredBy)
			err = impl.userDeploymentRequestService.UpdateStatusOnPipelineDelete(cdPipelineDeleteEvent.PipelineId)
			if err != nil {
				impl.logger.Errorw("error while terminating userDeploymentRequest for deleted pipeline",
					"pipelineId", cdPipelineDeleteEvent.PipelineId, "err", err)
			}
			// there is a possibility that when the pipeline was deleted, async request nats message was not consumed completely and could have led to dangling deployment app
			// trying to delete deployment app once
			err = impl.cdPipelineConfigService.DeleteHelmTypePipelineDeploymentApp(context.Background(), true, pipeline)
			if err != nil {
				impl.logger.Errorw("error, DeleteHelmTypePipelineDeploymentApp", "pipelineId", pipeline.Id)
			}
		}
	}
	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		cdStageCompleteEvent := &bean.CdStageCompleteEvent{}
		err := json.Unmarshal([]byte(msg.Data), cdStageCompleteEvent)
		if err != nil {
			return "error while unmarshalling cdPipelineDeleteEvent object", []interface{}{"err", err, "msg", msg.Data}
		}
		return "got message for cd pipeline deletion", []interface{}{"request", cdStageCompleteEvent}
	}

	err := impl.pubSubClient.Subscribe(pubsub.CD_PIPELINE_DELETE_EVENT_TOPIC, callback, loggerFunc)
	if err != nil {
		impl.logger.Error("error", "err", err)
		return err
	}
	return nil
}

func (impl *WorkflowEventProcessorImpl) getPipelineModelById(pipelineId int) (*pipelineConfig.Pipeline, error) {
	pipelineModel, err := impl.pipelineRepository.FindById(pipelineId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching pipelineModel by pipelineId", "pipelineId", pipelineId, "err", err)
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) || pipelineModel == nil || pipelineModel.Id == 0 {
		impl.logger.Warnw("invalid request received pipeline not active, terminating all userDeploymentRequest", "pipelineId", pipelineId, "err", err)
		err1 := impl.userDeploymentRequestService.UpdateStatusOnPipelineDelete(pipelineId)
		if err1 != nil {
			impl.logger.Errorw("error while terminating userDeploymentRequest for deleted pipeline",
				"pipelineId", pipelineId, "err", err1)
		}
		return nil, err
	}
	return pipelineModel, nil
}

func (impl *WorkflowEventProcessorImpl) ProcessIncompleteDeploymentReq() {
	cdAsyncInstallRequests, err := impl.userDeploymentRequestService.GetAllInCompleteRequests()
	if err != nil {
		impl.logger.Errorw("error in fetching all in complete userDeploymentRequests", "err", err)
		return
	}
	for _, cdAsyncInstallReq := range cdAsyncInstallRequests {
		err = impl.setAdditionalDataInAsyncInstallReq(cdAsyncInstallReq)
		if err != nil {
			impl.logger.Errorw("error in setting additional data to AsyncCdDeployRequest, skipping", "err", err)
			continue
		}
		err = impl.ProcessConcurrentAsyncDeploymentReq(cdAsyncInstallReq)
		if err != nil {
			impl.logger.Errorw("error in processing incomplete deployment request", "cdAsyncInstallReq", cdAsyncInstallReq, "err", err)
		} else {
			impl.logger.Infow("successfully processed deployment request", "cdAsyncInstallReq", cdAsyncInstallReq)
		}
	}
	return
}

func (impl *WorkflowEventProcessorImpl) ProcessConcurrentAsyncDeploymentReq(cdAsyncInstallReq *bean.AsyncCdDeployRequest) error {
	cdWfrId := cdAsyncInstallReq.ValuesOverrideRequest.WfrId
	pipelineId := cdAsyncInstallReq.ValuesOverrideRequest.PipelineId
	userId := cdAsyncInstallReq.ValuesOverrideRequest.UserId
	cdWfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(cdWfrId)
	if err != nil {
		impl.logger.Errorw("err on fetching cd workflow runner by id", "err", err, "cdWfrId", cdWfrId)
		return err
	}
	isValidRequest, err := impl.handleConcurrentOrInvalidRequest(cdWfr, cdAsyncInstallReq.UserDeploymentRequestId, pipelineId, userId)
	if err != nil {
		impl.logger.Errorw("error, handleConcurrentOrInvalidRequest", "err", err, "cdWfrId", cdWfrId, "cdWfrStatus", cdWfr.Status, "pipelineId", pipelineId)
		return err
	}
	if !isValidRequest {
		impl.logger.Debugw("skipping async helm install request", "req", cdAsyncInstallReq.ValuesOverrideRequest)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), impl.getTimeOutByDeploymentType(cdAsyncInstallReq.ValuesOverrideRequest.DeploymentAppType))
	defer cancel()
	newCtx, cancelCause := context.WithCancelCause(ctx)
	// setting Cancel Cause -> nil for successfully processed request
	defer cancelCause(nil)
	impl.UpdateReleaseContextForPipeline(pipelineId, cdWfrId, cancelCause)
	defer impl.cleanUpDevtronAppReleaseContextMap(pipelineId, cdWfrId)
	err = impl.workflowDagExecutor.ProcessDevtronAsyncInstallRequest(cdAsyncInstallReq, newCtx)
	if err != nil {
		impl.logger.Errorw("error, ProcessDevtronAsyncInstallRequest", "err", err, "req", cdAsyncInstallReq)
		return err
	}
	return nil
}

func (impl *WorkflowEventProcessorImpl) getTimeOutByDeploymentType(deploymentType string) time.Duration {
	switch deploymentType {
	case triggerBean.Helm:
		return time.Duration(impl.appServiceConfig.DevtronChartHelmInstallRequestTimeout) * time.Minute
	case triggerBean.ArgoCd:
		return time.Duration(impl.appServiceConfig.DevtronChartArgoCdInstallRequestTimeout) * time.Minute
	}
	return time.Duration(0)
}

func (impl *WorkflowEventProcessorImpl) handleConcurrentOrInvalidRequest(cdWfr *pipelineConfig.CdWorkflowRunner, userDeploymentRequestId, pipelineId int, userId int32) (isValidRequest bool, err error) {
	var isLatestRequest bool
	if userDeploymentRequestId != 0 {
		isLatestRequest, err = impl.userDeploymentRequestService.IsLatestForPipelineId(userDeploymentRequestId, pipelineId)
		if err != nil {
			impl.logger.Errorw("error, CheckIfWfrLatest", "err", err, "cdWfrId", cdWfr.Id)
			return isValidRequest, err
		}
	} else {
		isLatestRequest, err = impl.cdWorkflowRunnerService.CheckIfWfrLatest(cdWfr.Id, pipelineId)
		if err != nil {
			impl.logger.Errorw("error, CheckIfWfrLatest", "err", err, "cdWfrId", cdWfr.Id)
			return isValidRequest, err
		}
	}
	impl.devtronAppReleaseContextMapLock.Lock()
	defer impl.devtronAppReleaseContextMapLock.Unlock()
	if releaseContext, ok := impl.devtronAppReleaseContextMap[pipelineId]; ok {
		if releaseContext.RunnerId == cdWfr.Id {
			// request in process for same wfrId, skipping and doing nothing
			// earlier we used to check if wfrStatus is in starting then only skip, removed that
			return isValidRequest, nil
		} else {
			// request in process but for other wfrId
			// skip if the cdWfr.Status is already in a terminal state
			skipCDWfrStatusList := append(pipelineConfig.WfrTerminalStatusList, pipelineConfig.WorkflowInProgress)
			if slices.Contains(skipCDWfrStatusList, cdWfr.Status) {
				impl.logger.Warnw("skipped deployment as the workflow runner status is already in terminal state, handleConcurrentOrInvalidRequest", "cdWfrId", cdWfr.Id, "status", cdWfr.Status)
				return isValidRequest, nil
			}
			if !isLatestRequest {
				impl.logger.Warnw("skipped deployment as the workflow runner is not the latest one", "cdWfrId", cdWfr.Id)
				err := impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(cdWfr, errors.New(pipelineConfig.NEW_DEPLOYMENT_INITIATED), userId)
				if err != nil {
					impl.logger.Errorw("error while updating current runner status to failed, handleConcurrentOrInvalidRequest", "cdWfr", cdWfr.Id, "err", err)
					return isValidRequest, err
				}
				return isValidRequest, nil
			}
		}
	} else {
		// no request in process for pipeline, continue
	}

	return true, nil
}

func (impl *WorkflowEventProcessorImpl) UpdateReleaseContextForPipeline(pipelineId, cdWfrId int, cancel context.CancelCauseFunc) {
	impl.devtronAppReleaseContextMapLock.Lock()
	defer impl.devtronAppReleaseContextMapLock.Unlock()
	if releaseContext, ok := impl.devtronAppReleaseContextMap[pipelineId]; ok {
		// Abort previous running release
		impl.logger.Infow("new deployment has been triggered with a running deployment in progress!", "aborting deployment for pipelineId", pipelineId)
		releaseContext.CancelContext(errors.New(pipelineConfig.NEW_DEPLOYMENT_INITIATED))
	}
	impl.devtronAppReleaseContextMap[pipelineId] = bean.DevtronAppReleaseContextType{
		CancelContext: cancel,
		RunnerId:      cdWfrId,
	}
}

func (impl *WorkflowEventProcessorImpl) cleanUpDevtronAppReleaseContextMap(pipelineId, wfrId int) {
	if impl.isReleaseContextExistsForPipeline(pipelineId, wfrId) {
		impl.devtronAppReleaseContextMapLock.Lock()
		defer impl.devtronAppReleaseContextMapLock.Unlock()
		if _, ok := impl.devtronAppReleaseContextMap[pipelineId]; ok {
			delete(impl.devtronAppReleaseContextMap, pipelineId)
		}
	}
}

func (impl *WorkflowEventProcessorImpl) isReleaseContextExistsForPipeline(pipelineId, cdWfrId int) bool {
	impl.devtronAppReleaseContextMapLock.Lock()
	defer impl.devtronAppReleaseContextMapLock.Unlock()
	if releaseContext, ok := impl.devtronAppReleaseContextMap[pipelineId]; ok {
		return releaseContext.RunnerId == cdWfrId
	}
	return false
}

func (impl *WorkflowEventProcessorImpl) RemoveReleaseContextForPipeline(pipelineId int, triggeredBy int32) {
	impl.devtronAppReleaseContextMapLock.Lock()
	defer impl.devtronAppReleaseContextMapLock.Unlock()
	if releaseContext, ok := impl.devtronAppReleaseContextMap[pipelineId]; ok {
		//Abort previous running release
		impl.logger.Infow("CD pipeline has been deleted with a running deployment in progress!", "aborting deployment for pipelineId", pipelineId)
		cdWfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(releaseContext.RunnerId)
		if err != nil {
			impl.logger.Errorw("err on fetching cd workflow runner, RemoveReleaseContextForPipeline", "err", err)
		}
		if err = impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(cdWfr, errors.New("CD pipeline has been deleted"), triggeredBy); err != nil {
			impl.logger.Errorw("error while updating current runner status to failed, RemoveReleaseContextForPipeline", "cdWfr", cdWfr.Id, "err", err)
		}
		releaseContext.CancelContext(errors.New(pipelineConfig.PIPELINE_DELETED))
		delete(impl.devtronAppReleaseContextMap, pipelineId)
	}
	return
}

func (impl *WorkflowEventProcessorImpl) setAdditionalDataInAsyncInstallReq(cdAsyncInstallReq *bean.AsyncCdDeployRequest) error {
	pipelineModel, err := impl.getPipelineModelById(cdAsyncInstallReq.ValuesOverrideRequest.PipelineId)
	if err != nil {
		return err
	}
	triggerAdapter.SetPipelineFieldsInOverrideRequest(cdAsyncInstallReq.ValuesOverrideRequest, pipelineModel)
	if cdAsyncInstallReq.ValuesOverrideRequest.DeploymentType == models.DEPLOYMENTTYPE_UNKNOWN {
		cdAsyncInstallReq.ValuesOverrideRequest.DeploymentType = models.DEPLOYMENTTYPE_DEPLOY
	}
	return nil
}
