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
	"github.com/devtron-labs/devtron/pkg/deployment/common"
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
	error2 "github.com/devtron-labs/devtron/util/error"
	eventUtil "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
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
	pipelineRepository      pipelineConfig.PipelineRepository
	ciArtifactRepository    repository.CiArtifactRepository
	cdWorkflowRepository    pipelineConfig.CdWorkflowRepository
	deploymentConfigService common.DeploymentConfigService
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
	deploymentConfigService common.DeploymentConfigService) (*WorkflowEventProcessorImpl, error) {
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
		deploymentConfigService:         deploymentConfigService,
	}
	appServiceConfig, err := app.GetAppServiceConfig()
	if err != nil {
		return nil, err
	}
	impl.appServiceConfig = appServiceConfig
	// handle incomplete deployment requests after restart
	go impl.ProcessIncompleteDeploymentReq()
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

		if wfr.Status != string(v1alpha1.NodeSucceeded) {
			impl.logger.Debugw("event received from ci runner, updating workflow runner status as succeeded", "savedWorkflowRunnerId", wfr.Id, "oldStatus", wfr.Status, "podStatus", wfr.PodStatus)
			err = impl.cdWorkflowRunnerService.UpdateWfrStatus(wfr, string(v1alpha1.NodeSucceeded), 1)
			if err != nil {
				impl.logger.Errorw("update cd-wf-runner failed for id ", "cdWfrId", wfr.Id, "err", err)
				return
			}
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
		ctx, err := impl.argoUserService.GetACDContext(context.Background())
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
		ctx := context.Background()
		cdAsyncInstallReq, err := impl.extractAsyncCdDeployRequestFromEventMsg(ctx, msg)
		if err != nil {
			impl.logger.Errorw("err on extracting override request, SubscribeDevtronAsyncInstallRequest", "err", err)
			return
		}
		_ = impl.ProcessConcurrentAsyncDeploymentReq(ctx, cdAsyncInstallReq)
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
		cdAsyncInstallReq := &bean.UserDeploymentRequest{}
		err := json.Unmarshal([]byte(msg.Data), cdAsyncInstallReq)
		if err != nil {
			return fmt.Sprintf("error in unmarshalling CD Pipeline %s install request nats message", topicType), []interface{}{"err", err}
		}
		return fmt.Sprintf("got message for devtron chart %s install", topicType), []interface{}{"userDeploymentRequestId", cdAsyncInstallReq.Id}
	}
}

func (impl *WorkflowEventProcessorImpl) extractAsyncCdDeployRequestFromEventMsg(ctx context.Context, msg *model.PubSubMsg) (*bean.UserDeploymentRequest, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestServiceImpl.SaveNewDeployment")
	defer span.End()
	cdAsyncInstallReq := &bean.UserDeploymentRequest{}
	err := json.Unmarshal([]byte(msg.Data), cdAsyncInstallReq)
	if err != nil {
		impl.logger.Errorw("error in unmarshalling CD async install request nats message", "err", err)
		return nil, err
	}
	if cdAsyncInstallReq.Id == 0 && cdAsyncInstallReq.ValuesOverrideRequest == nil {
		impl.logger.Errorw("invalid async cd pipeline deployment request", "msg", msg.Data)
		return nil, fmt.Errorf("invalid async cd pipeline deployment request")
	}
	if cdAsyncInstallReq.Id != 0 {
		// getting the latest UserDeploymentRequest for the pipeline
		latestCdAsyncInstallReq, err := impl.userDeploymentRequestService.GetLatestAsyncCdDeployRequestForPipeline(newCtx, cdAsyncInstallReq.Id)
		if err != nil {
			impl.logger.Errorw("error in fetching userDeploymentRequest by id", "userDeploymentRequestId", cdAsyncInstallReq.Id, "err", err)
			return nil, err
		}
		// will process the latest UserDeploymentRequest irrespective of the received UserDeploymentRequest.Id
		// overriding cdAsyncInstallReq with the latest UserDeploymentRequest for the pipeline
		cdAsyncInstallReq = latestCdAsyncInstallReq
	}
	// handling cdAsyncInstallReq.ValuesOverrideRequest for backward compatibility
	err = impl.setAdditionalDataInAsyncInstallReq(newCtx, cdAsyncInstallReq)
	if err != nil {
		impl.logger.Errorw("error in setting additional data to UserDeploymentRequest", "err", err)
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
		envDeploymentConfig, err := impl.deploymentConfigService.GetConfigEvenIfInactive(pipeline.AppId, pipeline.EnvironmentId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in fetching environment deployment config by appId and envId", "appId", pipeline.AppId, "envId", pipeline.EnvironmentId, "err", err)
			return
		}
		var deploymentAppType string
		if err == pg.ErrNoRows {
			deploymentAppType = pipeline.DeploymentAppType
		} else {
			deploymentAppType = envDeploymentConfig.DeploymentAppType
		}
		if util3.IsHelmApp(deploymentAppType) || util3.IsAcdApp(deploymentAppType) {
			impl.RemoveReleaseContextForPipeline(cdPipelineDeleteEvent.PipelineId, cdPipelineDeleteEvent.TriggeredBy)
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
		cdWfr, dbErr := impl.cdWorkflowRepository.FindLatestByPipelineIdAndRunnerType(pipelineId, apiBean.CD_WORKFLOW_TYPE_DEPLOY)
		if dbErr != nil {
			impl.logger.Errorw("err on fetching cd workflow runner", "pipelineId", pipelineId, "err", dbErr)
		} else if dbErr = impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(&cdWfr, errors.New("CD pipeline has been deleted"), 1); dbErr != nil {
			impl.logger.Errorw("error while updating current runner status to failed", "cdWfr", cdWfr.Id, "err", dbErr)
		}
		return nil, err
	}
	return pipelineModel, nil
}

func (impl *WorkflowEventProcessorImpl) ProcessIncompleteDeploymentReq() {
	ctx := context.Background()
	cdAsyncInstallRequests, err := impl.userDeploymentRequestService.GetAllInCompleteRequests(ctx)
	if err != nil {
		impl.logger.Errorw("error in fetching all in complete userDeploymentRequests", "err", err)
		return
	}
	count := len(cdAsyncInstallRequests)
	if count > 0 {
		impl.logger.Infow("found incomplete deployment requests", "count", count)
	} else {
		impl.logger.Infow("no incomplete deployment requests to be processed, skipping")
	}
	for index, cdAsyncInstallReq := range cdAsyncInstallRequests {
		impl.logger.Infow("processing incomplete deployment request", "cdAsyncInstallReq", cdAsyncInstallReq, "request sequence", index+1)
		err = impl.setAdditionalDataInAsyncInstallReq(ctx, cdAsyncInstallReq)
		if err != nil {
			impl.logger.Errorw("error in setting additional data to UserDeploymentRequest, skipping", "err", err)
			continue
		}
		err = impl.ProcessConcurrentAsyncDeploymentReq(ctx, cdAsyncInstallReq)
		if err != nil {
			impl.logger.Errorw("error in processing incomplete deployment request", "cdAsyncInstallReq", cdAsyncInstallReq, "err", err)
		} else {
			impl.logger.Infow("successfully processed deployment request", "cdAsyncInstallReq", cdAsyncInstallReq)
		}
	}
	if count > 0 {
		impl.logger.Infow("successfully processed all incomplete deployment requests")
	}
	return
}

func (impl *WorkflowEventProcessorImpl) getDevtronAppReleaseContextWithLock(ctx context.Context,
	cdAsyncInstallReq *bean.UserDeploymentRequest, cdWfr *pipelineConfig.CdWorkflowRunner) (context.Context, bool, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "WorkflowEventProcessorImpl.getDevtronAppReleaseContextWithLock")
	defer span.End()
	cdWfrId := cdAsyncInstallReq.ValuesOverrideRequest.WfrId
	pipelineId := cdAsyncInstallReq.ValuesOverrideRequest.PipelineId
	userId := cdAsyncInstallReq.ValuesOverrideRequest.UserId
	impl.devtronAppReleaseContextMapLock.Lock()
	defer impl.devtronAppReleaseContextMapLock.Unlock()
	isValidRequest, err := impl.validateConcurrentOrInvalidRequest(ctx, cdWfr, cdAsyncInstallReq.Id, pipelineId, userId)
	if err != nil {
		impl.logger.Errorw("error, validateConcurrentOrInvalidRequest", "err", err, "cdWfrId", cdWfrId, "cdWfrStatus", cdWfr.Status, "pipelineId", pipelineId)
		return nil, false, err
	}
	if !isValidRequest {
		impl.logger.Debugw("skipping devtron async install request", "req", cdAsyncInstallReq.ValuesOverrideRequest)
		return nil, true, nil
	}
	ctxWithTimeOut, cancelParentCtx := context.WithTimeout(ctx, impl.getTimeOutByDeploymentType(cdAsyncInstallReq.ValuesOverrideRequest.DeploymentAppType))
	releaseContext, cancelWithCause := context.WithCancelCause(ctxWithTimeOut)
	impl.UpdateReleaseContextForPipeline(releaseContext, pipelineId, cdWfrId, cancelWithCause, cancelParentCtx)
	return releaseContext, false, err
}

func (impl *WorkflowEventProcessorImpl) ProcessConcurrentAsyncDeploymentReq(ctx context.Context, cdAsyncInstallReq *bean.UserDeploymentRequest) error {
	cdWfrId := cdAsyncInstallReq.ValuesOverrideRequest.WfrId
	pipelineId := cdAsyncInstallReq.ValuesOverrideRequest.PipelineId
	cdWfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(cdWfrId)
	if err != nil {
		impl.logger.Errorw("err on fetching cd workflow runner by id", "err", err, "cdWfrId", cdWfrId)
		return err
	}
	acdCtx, err := impl.argoUserService.GetACDContext(ctx)
	if err != nil {
		impl.logger.Errorw("error in creating ArgoCd context", "err", err)
		return err
	}
	releaseContext, skipRequest, err := impl.getDevtronAppReleaseContextWithLock(acdCtx, cdAsyncInstallReq, cdWfr)
	if err != nil {
		impl.logger.Errorw("error, getDevtronAppReleaseContextWithLock", "err", err, "cdWfrId", cdWfrId, "cdWfrStatus", cdWfr.Status, "pipelineId", pipelineId)
		return err
	}
	if skipRequest {
		impl.logger.Debugw("skipping async deployment request", "req", cdAsyncInstallReq.ValuesOverrideRequest)
		return nil
	}
	defer impl.cleanUpDevtronAppReleaseContextMap(pipelineId, cdWfrId)
	err = impl.workflowDagExecutor.ProcessDevtronAsyncInstallRequest(cdAsyncInstallReq, releaseContext)
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

func (impl *WorkflowEventProcessorImpl) validateConcurrentOrInvalidRequest(ctx context.Context, cdWfr *pipelineConfig.CdWorkflowRunner, userDeploymentRequestId, pipelineId int, userId int32) (isValidRequest bool, err error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "WorkflowEventProcessorImpl.validateConcurrentOrInvalidRequest")
	defer span.End()
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
				impl.logger.Warnw("skipped deployment as the workflow runner status is already in terminal state, validateConcurrentOrInvalidRequest", "cdWfrId", cdWfr.Id, "status", cdWfr.Status)
				return isValidRequest, nil
			}
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
			if !isLatestRequest {
				impl.logger.Warnw("skipped deployment as the workflow runner is not the latest one", "cdWfrId", cdWfr.Id)
				err := impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(cdWfr, pipelineConfig.ErrorDeploymentSuperseded, userId)
				if err != nil {
					impl.logger.Errorw("error while updating current runner status to failed, validateConcurrentOrInvalidRequest", "cdWfr", cdWfr.Id, "err", err)
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

func (impl *WorkflowEventProcessorImpl) UpdateReleaseContextForPipeline(ctx context.Context, pipelineId, cdWfrId int, cancelWithCause context.CancelCauseFunc, cancelParentCtx context.CancelFunc) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "WorkflowEventProcessorImpl.UpdateReleaseContextForPipeline")
	defer span.End()
	if releaseContext, ok := impl.devtronAppReleaseContextMap[pipelineId]; ok {
		impl.logger.Infow("new deployment has been triggered with a running deployment in progress!", "aborting deployment for pipelineId", pipelineId)
		// abort previous running release
		releaseContext.CancelContext(pipelineConfig.ErrorDeploymentSuperseded)
		// cancelling parent context
		releaseContext.CancelParentContext()
	}
	impl.devtronAppReleaseContextMap[pipelineId] = bean.DevtronAppReleaseContextType{
		CancelParentContext: cancelParentCtx,
		CancelContext:       cancelWithCause,
		RunnerId:            cdWfrId,
	}
}

func (impl *WorkflowEventProcessorImpl) cleanUpDevtronAppReleaseContextMap(pipelineId, wfrId int) {
	if impl.isReleaseContextExistsForPipeline(pipelineId, wfrId) {
		impl.devtronAppReleaseContextMapLock.Lock()
		defer impl.devtronAppReleaseContextMapLock.Unlock()
		if releaseContext, ok := impl.devtronAppReleaseContextMap[pipelineId]; ok {
			// cancelling child context. setting cancel cause -> nil for successfully processed request
			releaseContext.CancelContext(nil)
			// cancelling parent context
			releaseContext.CancelParentContext()
			delete(impl.devtronAppReleaseContextMap, pipelineId)
		}
	}
}

func (impl *WorkflowEventProcessorImpl) ShutDownDevtronAppReleaseContext() {
	impl.devtronAppReleaseContextMapLock.Lock()
	defer impl.devtronAppReleaseContextMapLock.Unlock()
	impl.logger.Infow("cancelling devtron app deployment context", "count", len(impl.devtronAppReleaseContextMap))
	for _, devtronAppReleaseContext := range impl.devtronAppReleaseContextMap {
		// cancelling child context. setting cancel cause -> error.ServerShutDown
		devtronAppReleaseContext.CancelContext(error2.ServerShutDown)
		// cancelling parent context
		devtronAppReleaseContext.CancelParentContext()
	}
	impl.devtronAppReleaseContextMap = make(map[int]bean.DevtronAppReleaseContextType)
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
		// cancelling child context. setting cancel cause -> pipeline deleted
		releaseContext.CancelContext(errors.New(pipelineConfig.PIPELINE_DELETED))
		// cancelling parent context
		releaseContext.CancelParentContext()
		delete(impl.devtronAppReleaseContextMap, pipelineId)
	}
	return
}

func (impl *WorkflowEventProcessorImpl) setAdditionalDataInAsyncInstallReq(ctx context.Context, cdAsyncInstallReq *bean.UserDeploymentRequest) error {
	_, span := otel.Tracer("orchestrator").Start(ctx, "WorkflowEventProcessorImpl.setAdditionalDataInAsyncInstallReq")
	defer span.End()
	pipelineModel, err := impl.getPipelineModelById(cdAsyncInstallReq.ValuesOverrideRequest.PipelineId)
	if err != nil {
		return err
	}
	envDeploymentConfig, err := impl.deploymentConfigService.GetConfigForDevtronApps(pipelineModel.AppId, pipelineModel.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in fetching environment deployment config by appId and envId", "appId", pipelineModel.AppId, "envId", pipelineModel.EnvironmentId, "err", err)
		return err
	}
	triggerAdapter.SetPipelineFieldsInOverrideRequest(cdAsyncInstallReq.ValuesOverrideRequest, pipelineModel, envDeploymentConfig)
	if cdAsyncInstallReq.ValuesOverrideRequest.DeploymentType == models.DEPLOYMENTTYPE_UNKNOWN {
		cdAsyncInstallReq.ValuesOverrideRequest.DeploymentType = models.DEPLOYMENTTYPE_DEPLOY
	}
	cdAsyncInstallReq.ValuesOverrideRequest.UserId = cdAsyncInstallReq.TriggeredBy
	return nil
}
