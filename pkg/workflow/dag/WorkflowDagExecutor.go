/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package dag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	bean6 "github.com/devtron-labs/devtron/api/helm-app/bean"
	client2 "github.com/devtron-labs/devtron/api/helm-app/service"
	client "github.com/devtron-labs/devtron/client/events"
	bean4 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/build/artifacts"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	bean5 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	bean7 "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	bean2 "github.com/devtron-labs/devtron/pkg/workflow/dag/bean"
	util2 "github.com/devtron-labs/devtron/util/event"
	"strings"
	"sync"
	"time"

	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	util4 "github.com/devtron-labs/devtron/util"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"k8s.io/utils/strings/slices"

	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type WorkflowDagExecutor interface {
	HandleCiSuccessEvent(triggerContext bean5.TriggerContext, ciPipelineId int, request *bean2.CiArtifactWebhookRequest, imagePushedAt *time.Time) (id int, err error)
	HandlePreStageSuccessEvent(triggerContext bean5.TriggerContext, cdStageCompleteEvent bean7.CdStageCompleteEvent) error
	HandleDeploymentSuccessEvent(triggerContext bean5.TriggerContext, pipelineOverride *chartConfig.PipelineOverride) error
	HandlePostStageSuccessEvent(triggerContext bean5.TriggerContext, cdWorkflowId int, cdPipelineId int, triggeredBy int32, pluginRegistryImageDetails map[string][]string) error
	HandleCdStageReTrigger(runner *pipelineConfig.CdWorkflowRunner) error
	HandleCiStepFailedEvent(ciPipelineId int, request *bean2.CiArtifactWebhookRequest) (err error)
	HandleExternalCiWebhook(externalCiId int, request *bean2.CiArtifactWebhookRequest,
		auth func(token string, projectObject string, envObject string) bool, token string) (id int, err error)

	UpdateWorkflowRunnerStatusForDeployment(appIdentifier *client2.AppIdentifier, wfr *pipelineConfig.CdWorkflowRunner, skipReleaseNotFound bool) bool
	OnDeleteCdPipelineEvent(pipelineId int, triggeredBy int32)

	BuildCiArtifactRequestForWebhook(event pipeline.ExternalCiWebhookDto) (*bean2.CiArtifactWebhookRequest, error)
}

type WorkflowDagExecutorImpl struct {
	logger                       *zap.SugaredLogger
	pipelineRepository           pipelineConfig.PipelineRepository
	cdWorkflowRepository         pipelineConfig.CdWorkflowRepository
	pubsubClient                 *pubsub.PubSubClientServiceImpl
	ciArtifactRepository         repository.CiArtifactRepository
	enforcerUtil                 rbac.EnforcerUtil
	appWorkflowRepository        appWorkflow.AppWorkflowRepository
	pipelineStageService         pipeline.PipelineStageService
	ciWorkflowRepository         pipelineConfig.CiWorkflowRepository
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	pipelineStageRepository      repository4.PipelineStageRepository
	globalPluginRepository       repository2.GlobalPluginRepository
	deploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository
	config                       *types.CdConfig
	ciConfig                     *types.CiConfig
	appServiceConfig             *app.AppServiceConfig
	eventClient                  client.EventClient
	eventFactory                 client.EventFactory
	customTagService             pipeline.CustomTagService

	devtronAsyncHelmInstallRequestMap  map[int]bool
	devtronAsyncHelmInstallRequestLock *sync.Mutex
	devtronAppReleaseContextMap        map[int]DevtronAppReleaseContextType
	devtronAppReleaseContextMapLock    *sync.Mutex

	helmAppService client2.HelmAppService

	cdWorkflowCommonService cd.CdWorkflowCommonService
	cdTriggerService        devtronApps.TriggerService

	manifestCreationService manifest.ManifestCreationService
	commonArtifactService   artifacts.CommonArtifactService
}

type DevtronAppReleaseContextType struct {
	CancelContext context.CancelFunc
	RunnerId      int
}

func NewWorkflowDagExecutorImpl(Logger *zap.SugaredLogger, pipelineRepository pipelineConfig.PipelineRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pubsubClient *pubsub.PubSubClientServiceImpl,
	ciArtifactRepository repository.CiArtifactRepository,
	enforcerUtil rbac.EnforcerUtil,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	pipelineStageService pipeline.PipelineStageService,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	pipelineStageRepository repository4.PipelineStageRepository,
	globalPluginRepository repository2.GlobalPluginRepository,
	deploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository,
	eventClient client.EventClient,
	eventFactory client.EventFactory,
	helmAppService client2.HelmAppService,
	pipelineConfigListenerService pipeline.PipelineConfigListenerService,
	cdWorkflowCommonService cd.CdWorkflowCommonService,
	cdTriggerService devtronApps.TriggerService,
	manifestCreationService manifest.ManifestCreationService,
	commonArtifactService artifacts.CommonArtifactService) *WorkflowDagExecutorImpl {
	wde := &WorkflowDagExecutorImpl{logger: Logger,
		pipelineRepository:           pipelineRepository,
		cdWorkflowRepository:         cdWorkflowRepository,
		pubsubClient:                 pubsubClient,
		ciArtifactRepository:         ciArtifactRepository,
		enforcerUtil:                 enforcerUtil,
		appWorkflowRepository:        appWorkflowRepository,
		pipelineStageService:         pipelineStageService,
		ciWorkflowRepository:         ciWorkflowRepository,
		ciPipelineRepository:         ciPipelineRepository,
		pipelineStageRepository:      pipelineStageRepository,
		globalPluginRepository:       globalPluginRepository,
		deploymentApprovalRepository: deploymentApprovalRepository,
		eventClient:                  eventClient,
		eventFactory:                 eventFactory,

		devtronAsyncHelmInstallRequestMap:  make(map[int]bool),
		devtronAsyncHelmInstallRequestLock: &sync.Mutex{},
		devtronAppReleaseContextMap:        make(map[int]DevtronAppReleaseContextType),
		devtronAppReleaseContextMapLock:    &sync.Mutex{},
		helmAppService:                     helmAppService,
		cdWorkflowCommonService:            cdWorkflowCommonService,
		cdTriggerService:                   cdTriggerService,
		manifestCreationService:            manifestCreationService,
		commonArtifactService:              commonArtifactService,
	}
	config, err := types.GetCdConfig()
	if err != nil {
		return nil
	}
	wde.config = config
	ciConfig, err := types.GetCiConfig()
	if err != nil {
		return nil
	}
	wde.ciConfig = ciConfig
	appServiceConfig, err := app.GetAppServiceConfig()
	if err != nil {
		return nil
	}
	wde.appServiceConfig = appServiceConfig
	err = wde.SubscribeDevtronAsyncHelmInstallRequest()
	if err != nil {
		return nil
	}
	pipelineConfigListenerService.RegisterPipelineDeleteListener(wde)
	return wde
}

func (impl *WorkflowDagExecutorImpl) HandleCdStageReTrigger(runner *pipelineConfig.CdWorkflowRunner) error {
	// do not re-trigger if retries = 0
	if !impl.config.WorkflowRetriesEnabled() {
		impl.logger.Debugw("cd stage workflow re-triggering is not enabled")
		return nil
	}

	impl.logger.Infow("re triggering cd stage ", "runnerId", runner.Id)
	var err error
	// add comment for this logic
	if runner.RefCdWorkflowRunnerId != 0 {
		runner, err = impl.cdWorkflowRepository.FindWorkflowRunnerById(runner.RefCdWorkflowRunnerId)
		if err != nil {
			impl.logger.Errorw("error in FindWorkflowRunnerById by id ", "err", err, "wfrId", runner.RefCdWorkflowRunnerId)
			return err
		}
	}
	retryCnt, err := impl.cdWorkflowRepository.FindRetriedWorkflowCountByReferenceId(runner.Id)
	if err != nil {
		impl.logger.Errorw("error in FindRetriedWorkflowCountByReferenceId ", "err", err, "cdWorkflowRunnerId", runner.Id)
		return err
	}

	if retryCnt >= impl.config.MaxCdWorkflowRunnerRetries {
		impl.logger.Infow("maximum retries for this workflow are exhausted, not re-triggering again", "retries", retryCnt, "wfrId", runner.Id)
		return nil
	}

	triggerRequest := bean5.TriggerRequest{
		CdWf:                  runner.CdWorkflow,
		Pipeline:              runner.CdWorkflow.Pipeline,
		Artifact:              runner.CdWorkflow.CiArtifact,
		TriggeredBy:           1,
		ApplyAuth:             false,
		RefCdWorkflowRunnerId: runner.Id,
		TriggerContext: bean5.TriggerContext{
			Context: context.Background(),
		},
	}

	if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		err = impl.cdTriggerService.TriggerPreStage(triggerRequest)
		if err != nil {
			impl.logger.Errorw("error in TriggerPreStage ", "err", err, "cdWorkflowRunnerId", runner.Id)
			return err
		}
	} else if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
		err = impl.cdTriggerService.TriggerPostStage(triggerRequest)
		if err != nil {
			impl.logger.Errorw("error in TriggerPostStage ", "err", err, "cdWorkflowRunnerId", runner.Id)
			return err
		}
	}

	impl.logger.Infow("cd stage re triggered for runner", "runnerId", runner.Id)
	return nil
}

func (impl *WorkflowDagExecutorImpl) extractOverrideRequestFromCDAsyncInstallEvent(msg *model.PubSubMsg) (*bean7.AsyncCdDeployEvent, *client2.AppIdentifier, error) {
	CDAsyncInstallNatsMessage := &bean7.AsyncCdDeployEvent{}
	err := json.Unmarshal([]byte(msg.Data), CDAsyncInstallNatsMessage)
	if err != nil {
		impl.logger.Errorw("error in unmarshalling CD async install request nats message", "err", err)
		return nil, nil, err
	}
	pipeline, err := impl.pipelineRepository.FindById(CDAsyncInstallNatsMessage.ValuesOverrideRequest.PipelineId)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline by pipelineId", "err", err)
		return nil, nil, err
	}
	devtronApps.SetPipelineFieldsInOverrideRequest(CDAsyncInstallNatsMessage.ValuesOverrideRequest, pipeline)
	if CDAsyncInstallNatsMessage.ValuesOverrideRequest.DeploymentType == models.DEPLOYMENTTYPE_UNKNOWN {
		CDAsyncInstallNatsMessage.ValuesOverrideRequest.DeploymentType = models.DEPLOYMENTTYPE_DEPLOY
	}
	appIdentifier := &client2.AppIdentifier{
		ClusterId:   pipeline.Environment.ClusterId,
		Namespace:   pipeline.Environment.Namespace,
		ReleaseName: pipeline.DeploymentAppName,
	}
	return CDAsyncInstallNatsMessage, appIdentifier, nil
}

// UpdateWorkflowRunnerStatusForDeployment will update CD workflow runner based on release status and app status
func (impl *WorkflowDagExecutorImpl) UpdateWorkflowRunnerStatusForDeployment(appIdentifier *client2.AppIdentifier, wfr *pipelineConfig.CdWorkflowRunner, skipReleaseNotFound bool) bool {
	helmInstalledDevtronApp, err := impl.helmAppService.GetApplicationAndReleaseStatus(context.Background(), appIdentifier)
	if err != nil {
		impl.logger.Errorw("error in getting helm app release status", "appIdentifier", appIdentifier, "err", err)
		// Handle release not found errors
		if skipReleaseNotFound && util.GetGRPCErrorDetailedMessage(err) != bean6.ErrReleaseNotFound {
			// skip this error and continue for next workflow status
			impl.logger.Warnw("found error, skipping helm apps status update for this trigger", "appIdentifier", appIdentifier, "err", err)
			return false
		}
		// If release not found, mark the deployment as failure
		wfr.Status = pipelineConfig.WorkflowFailed
		wfr.Message = util.GetGRPCErrorDetailedMessage(err)
		wfr.FinishedOn = time.Now()
		return true
	}

	switch helmInstalledDevtronApp.GetReleaseStatus() {
	case serverBean.HelmReleaseStatusSuperseded:
		// If release status is superseded, mark the deployment as failure
		wfr.Status = pipelineConfig.WorkflowFailed
		wfr.Message = pipelineConfig.NEW_DEPLOYMENT_INITIATED
		wfr.FinishedOn = time.Now()
		return true
	case serverBean.HelmReleaseStatusFailed:
		// If release status is failed, mark the deployment as failure
		wfr.Status = pipelineConfig.WorkflowFailed
		wfr.Message = helmInstalledDevtronApp.GetDescription()
		wfr.FinishedOn = time.Now()
		return true
	case serverBean.HelmReleaseStatusDeployed:
		//skip if there is no deployment after wfr.StartedOn and continue for next workflow status
		if helmInstalledDevtronApp.GetLastDeployed().AsTime().Before(wfr.StartedOn) {
			impl.logger.Warnw("release mismatched, skipping helm apps status update for this trigger", "appIdentifier", appIdentifier, "err", err)
			return false
		}

		if helmInstalledDevtronApp.GetApplicationStatus() == application.Healthy {
			// mark the deployment as succeed
			wfr.Status = pipelineConfig.WorkflowSucceeded
			wfr.FinishedOn = time.Now()
			return true
		}
	}
	if wfr.Status == pipelineConfig.WorkflowInProgress {
		return false
	}
	wfr.Status = pipelineConfig.WorkflowInProgress
	return true
}

func (impl *WorkflowDagExecutorImpl) handleAsyncTriggerReleaseError(releaseErr error, cdWfr *pipelineConfig.CdWorkflowRunner, overrideRequest *bean.ValuesOverrideRequest, appIdentifier *client2.AppIdentifier) {
	releaseErrString := util.GetGRPCErrorDetailedMessage(releaseErr)
	switch releaseErrString {
	case context.DeadlineExceeded.Error():
		// if context deadline is exceeded fetch release status and UpdateWorkflowRunnerStatusForDeployment
		if isWfrUpdated := impl.UpdateWorkflowRunnerStatusForDeployment(appIdentifier, cdWfr, false); !isWfrUpdated {
			// updating cdWfr to failed
			if err := impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(cdWfr, fmt.Errorf("Deployment timeout: release %s took more than %d mins", appIdentifier.ReleaseName, impl.appServiceConfig.DevtronChartInstallRequestTimeout), overrideRequest.UserId); err != nil {
				impl.logger.Errorw("error while updating current runner status to failed, handleAsyncTriggerReleaseError", "cdWfr", cdWfr.Id, "err", err)
			}
		}
		cdWfr.UpdatedBy = 1
		cdWfr.UpdatedOn = time.Now()
		err := impl.cdWorkflowRepository.UpdateWorkFlowRunner(cdWfr)
		if err != nil {
			impl.logger.Errorw("error on update cd workflow runner", "wfr", cdWfr, "err", err)
			return
		}
		cdMetrics := util4.CDMetrics{
			AppName:         cdWfr.CdWorkflow.Pipeline.DeploymentAppName,
			Status:          cdWfr.Status,
			DeploymentType:  cdWfr.CdWorkflow.Pipeline.DeploymentAppType,
			EnvironmentName: cdWfr.CdWorkflow.Pipeline.Environment.Name,
			Time:            time.Since(cdWfr.StartedOn).Seconds() - time.Since(cdWfr.FinishedOn).Seconds(),
		}
		util4.TriggerCDMetrics(cdMetrics, impl.config.ExposeCDMetrics)
		impl.logger.Infow("updated workflow runner status for helm app", "wfr", cdWfr)
		return
	case context.Canceled.Error():
		if err := impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(cdWfr, errors.New(pipelineConfig.NEW_DEPLOYMENT_INITIATED), overrideRequest.UserId); err != nil {
			impl.logger.Errorw("error while updating current runner status to failed, handleAsyncTriggerReleaseError", "cdWfr", cdWfr.Id, "err", err)
		}
		return
	case "":
		return
	default:
		if err := impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(cdWfr, releaseErr, overrideRequest.UserId); err != nil {
			impl.logger.Errorw("error while updating current runner status to failed, handleAsyncTriggerReleaseError", "cdWfr", cdWfr.Id, "err", err)
		}
		return
	}
}

func (impl *WorkflowDagExecutorImpl) handleIfPreviousRunnerTriggerRequest(currentRunner *pipelineConfig.CdWorkflowRunner, userId int32) (bool, error) {
	exists, err := impl.cdWorkflowRepository.IsLatestCDWfr(currentRunner.Id, currentRunner.CdWorkflow.PipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err on fetching latest cd workflow runner, SubscribeDevtronAsyncHelmInstallRequest", "err", err)
		return false, err
	}
	return exists, nil
}

func (impl *WorkflowDagExecutorImpl) UpdateReleaseContextForPipeline(pipelineId, cdWfrId int, cancel context.CancelFunc) {
	impl.devtronAppReleaseContextMapLock.Lock()
	defer impl.devtronAppReleaseContextMapLock.Unlock()
	if releaseContext, ok := impl.devtronAppReleaseContextMap[pipelineId]; ok {
		//Abort previous running release
		impl.logger.Infow("new deployment has been triggered with a running deployment in progress!", "aborting deployment for pipelineId", pipelineId)
		releaseContext.CancelContext()
	}
	impl.devtronAppReleaseContextMap[pipelineId] = DevtronAppReleaseContextType{
		CancelContext: cancel,
		RunnerId:      cdWfrId,
	}
}

func (impl *WorkflowDagExecutorImpl) RemoveReleaseContextForPipeline(pipelineId int, triggeredBy int32) {
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
		releaseContext.CancelContext()
		delete(impl.devtronAppReleaseContextMap, pipelineId)
	}
	return
}

func (impl *WorkflowDagExecutorImpl) OnDeleteCdPipelineEvent(pipelineId int, triggeredBy int32) {
	impl.logger.Debugw("CD pipeline delete event received", "pipelineId", pipelineId, "deletedBy", triggeredBy)
	impl.RemoveReleaseContextForPipeline(pipelineId, triggeredBy)
	return
}

func (impl *WorkflowDagExecutorImpl) isReleaseContextExistsForPipeline(pipelineId, cdWfrId int) bool {
	impl.devtronAppReleaseContextMapLock.Lock()
	defer impl.devtronAppReleaseContextMapLock.Unlock()
	if releaseContext, ok := impl.devtronAppReleaseContextMap[pipelineId]; ok {
		return releaseContext.RunnerId == cdWfrId
	}
	return false
}

func (impl *WorkflowDagExecutorImpl) handleConcurrentRequest(wfrId int) bool {
	impl.devtronAsyncHelmInstallRequestLock.Lock()
	defer impl.devtronAsyncHelmInstallRequestLock.Unlock()
	if _, exists := impl.devtronAsyncHelmInstallRequestMap[wfrId]; exists {
		//request is in process already, Skip here
		return true
	}
	impl.devtronAsyncHelmInstallRequestMap[wfrId] = true
	return false
}

func (impl *WorkflowDagExecutorImpl) cleanUpDevtronAppReleaseContextMap(pipelineId, wfrId int) {
	if impl.isReleaseContextExistsForPipeline(pipelineId, wfrId) {
		impl.devtronAppReleaseContextMapLock.Lock()
		defer impl.devtronAppReleaseContextMapLock.Unlock()
		if _, ok := impl.devtronAppReleaseContextMap[pipelineId]; ok {
			delete(impl.devtronAppReleaseContextMap, pipelineId)
		}
	}
}

func (impl *WorkflowDagExecutorImpl) cleanUpDevtronAsyncHelmInstallRequest(pipelineId, wfrId int) {
	impl.devtronAsyncHelmInstallRequestLock.Lock()
	defer impl.devtronAsyncHelmInstallRequestLock.Unlock()
	if _, exists := impl.devtronAsyncHelmInstallRequestMap[wfrId]; exists {
		//request is in process already, Skip here
		delete(impl.devtronAsyncHelmInstallRequestMap, wfrId)
	}
	impl.cleanUpDevtronAppReleaseContextMap(pipelineId, wfrId)
}

func (impl *WorkflowDagExecutorImpl) processDevtronAsyncHelmInstallRequest(CDAsyncInstallNatsMessage *bean7.AsyncCdDeployEvent, appIdentifier *client2.AppIdentifier) {
	overrideRequest := CDAsyncInstallNatsMessage.ValuesOverrideRequest
	cdWfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(overrideRequest.WfrId)
	if err != nil {
		impl.logger.Errorw("err on fetching cd workflow runner, processDevtronAsyncHelmInstallRequest", "err", err)
		return
	}

	// skip if the cdWfr.Status is already in a terminal state
	skipCDWfrStatusList := pipelineConfig.WfrTerminalStatusList
	skipCDWfrStatusList = append(skipCDWfrStatusList, pipelineConfig.WorkflowInProgress)
	if slices.Contains(skipCDWfrStatusList, cdWfr.Status) {
		impl.logger.Warnw("skipped deployment as the workflow runner status is already in terminal state, processDevtronAsyncHelmInstallRequest", "cdWfrId", cdWfr.Id, "status", cdWfr.Status)
		return
	}

	//skip if the cdWfr is not the latest one
	exists, err := impl.handleIfPreviousRunnerTriggerRequest(cdWfr, overrideRequest.UserId)
	if err != nil {
		impl.logger.Errorw("err in validating latest cd workflow runner, processDevtronAsyncHelmInstallRequest", "err", err)
		return
	}
	if exists {
		impl.logger.Warnw("skipped deployment as the workflow runner is not the latest one", "cdWfrId", cdWfr.Id)
		err := impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(cdWfr, errors.New(pipelineConfig.NEW_DEPLOYMENT_INITIATED), overrideRequest.UserId)
		if err != nil {
			impl.logger.Errorw("error while updating current runner status to failed, processDevtronAsyncHelmInstallRequest", "cdWfr", cdWfr.Id, "err", err)
			return
		}
		return
	}

	if cdWfr.Status == pipelineConfig.WorkflowStarting && impl.isReleaseContextExistsForPipeline(overrideRequest.PipelineId, cdWfr.Id) {
		impl.logger.Warnw("event redelivered! deployment is currently in progress, processDevtronAsyncHelmInstallRequest", "cdWfrId", cdWfr.Id, "status", cdWfr.Status)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(impl.appServiceConfig.DevtronChartInstallRequestTimeout)*time.Minute)
	defer cancel()

	impl.UpdateReleaseContextForPipeline(overrideRequest.PipelineId, cdWfr.Id, cancel)
	//update workflow runner status, used in app workflow view
	err = impl.cdWorkflowCommonService.UpdateCDWorkflowRunnerStatus(ctx, overrideRequest, CDAsyncInstallNatsMessage.TriggeredAt, pipelineConfig.WorkflowStarting, "")
	if err != nil {
		impl.logger.Errorw("error in updating the workflow runner status, processDevtronAsyncHelmInstallRequest", "cdWfrId", cdWfr.Id, "err", err)
		return
	}
	// build merged values and save PCO history for the release
	valuesOverrideResponse, builtChartPath, err := impl.manifestCreationService.BuildManifestForTrigger(overrideRequest, CDAsyncInstallNatsMessage.TriggeredAt, ctx)
	if err != nil {
		return
	}

	_, span := otel.Tracer("orchestrator").Start(ctx, "WorkflowDagExecutorImpl.TriggerRelease")
	releaseId, _, releaseErr := impl.cdTriggerService.TriggerRelease(overrideRequest, valuesOverrideResponse, builtChartPath, ctx, CDAsyncInstallNatsMessage.TriggeredAt, CDAsyncInstallNatsMessage.TriggeredBy)
	span.End()
	if releaseErr != nil {
		impl.handleAsyncTriggerReleaseError(releaseErr, cdWfr, overrideRequest, appIdentifier)
	} else {
		impl.logger.Infow("pipeline triggered successfully !!", "cdPipelineId", overrideRequest.PipelineId, "artifactId", overrideRequest.CiArtifactId, "releaseId", releaseId)
		// Update previous deployment runner status (in transaction): Failed
		_, span = otel.Tracer("orchestrator").Start(ctx, "updatePreviousDeploymentStatus")
		err1 := impl.cdWorkflowCommonService.UpdatePreviousDeploymentStatus(cdWfr, overrideRequest.PipelineId, CDAsyncInstallNatsMessage.TriggeredAt, overrideRequest.UserId)
		span.End()
		if err1 != nil {
			impl.logger.Errorw("error while update previous cd workflow runners, processDevtronAsyncHelmInstallRequest", "err", err, "runner", cdWfr, "pipelineId", overrideRequest.PipelineId)
			return
		}
	}
}

func (impl *WorkflowDagExecutorImpl) SubscribeDevtronAsyncHelmInstallRequest() error {
	callback := func(msg *model.PubSubMsg) {
		CDAsyncInstallNatsMessage, appIdentifier, err := impl.extractOverrideRequestFromCDAsyncInstallEvent(msg)
		if err != nil {
			impl.logger.Errorw("err on extracting override request, SubscribeDevtronAsyncHelmInstallRequest", "err", err)
			return
		}
		if skip := impl.handleConcurrentRequest(CDAsyncInstallNatsMessage.ValuesOverrideRequest.WfrId); skip {
			impl.logger.Warnw("concurrent request received, SubscribeDevtronAsyncHelmInstallRequest", "WfrId", CDAsyncInstallNatsMessage.ValuesOverrideRequest.WfrId)
			return
		}
		defer impl.cleanUpDevtronAsyncHelmInstallRequest(CDAsyncInstallNatsMessage.ValuesOverrideRequest.PipelineId, CDAsyncInstallNatsMessage.ValuesOverrideRequest.WfrId)
		impl.processDevtronAsyncHelmInstallRequest(CDAsyncInstallNatsMessage, appIdentifier)
		return
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		CDAsyncInstallNatsMessage := &bean7.AsyncCdDeployEvent{}
		err := json.Unmarshal([]byte(msg.Data), CDAsyncInstallNatsMessage)
		if err != nil {
			return "error in unmarshalling CD async install request nats message", []interface{}{"err", err}
		}
		return "got message for devtron chart install", []interface{}{"appId", CDAsyncInstallNatsMessage.ValuesOverrideRequest.AppId, "pipelineId", CDAsyncInstallNatsMessage.ValuesOverrideRequest.PipelineId, "artifactId", CDAsyncInstallNatsMessage.ValuesOverrideRequest.CiArtifactId}
	}

	err := impl.pubsubClient.Subscribe(pubsub.DEVTRON_CHART_INSTALL_TOPIC, callback, loggerFunc)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) handleCiSuccessEvent(triggerContext bean5.TriggerContext, artifact *repository.CiArtifact, async bool, triggeredBy int32) error {
	//1. get cd pipelines
	//2. get config
	//3. trigger wf/ deployment
	var pipelineID int
	if artifact.DataSource == repository.POST_CI {
		pipelineID = artifact.ComponentId
	} else {
		// TODO: need to migrate artifact.PipelineId for dataSource="CI_RUNNER" also to component_id
		pipelineID = artifact.PipelineId
	}
	pipelines, err := impl.pipelineRepository.FindByParentCiPipelineId(pipelineID)
	if err != nil {
		impl.logger.Errorw("error in fetching cd pipeline", "pipelineId", artifact.PipelineId, "err", err)
		return err
	}
	for _, pipeline := range pipelines {
		triggerRequest := bean5.TriggerRequest{
			CdWf:           nil,
			Pipeline:       pipeline,
			Artifact:       artifact,
			TriggeredBy:    triggeredBy,
			TriggerContext: triggerContext,
		}
		err = impl.triggerIfAutoStageCdPipeline(triggerRequest)
		if err != nil {
			impl.logger.Debugw("error on trigger cd pipeline", "err", err)
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) handleWebhookExternalCiEvent(artifact *repository.CiArtifact, triggeredBy int32, externalCiId int, auth func(token string, projectObject string, envObject string) bool, token string) (bool, error) {
	hasAnyTriggered := false
	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFCDMappingByExternalCiId(externalCiId)
	if err != nil {
		impl.logger.Errorw("error in fetching cd pipeline", "pipelineId", artifact.PipelineId, "err", err)
		return hasAnyTriggered, err
	}

	var pipelines []*pipelineConfig.Pipeline
	for _, appWorkflowMapping := range appWorkflowMappings {
		pipeline, err := impl.pipelineRepository.FindById(appWorkflowMapping.ComponentId)
		if err != nil {
			impl.logger.Errorw("error in fetching cd pipeline", "pipelineId", artifact.PipelineId, "err", err)
			return hasAnyTriggered, err
		}
		projectObject := impl.enforcerUtil.GetAppRBACNameByAppId(pipeline.AppId)
		envObject := impl.enforcerUtil.GetAppRBACByAppIdAndPipelineId(pipeline.AppId, pipeline.Id)
		if !auth(token, projectObject, envObject) {
			err = &util.ApiError{Code: "401", HttpStatusCode: 401, UserMessage: "Unauthorized"}
			return hasAnyTriggered, err
		}
		if pipeline.ApprovalNodeConfigured() {
			impl.logger.Warnw("approval node configured, so skipping pipeline for approval", "pipeline", pipeline)
			continue
		}
		if pipeline.IsManualTrigger() {
			impl.logger.Warnw("skipping deployment for manual trigger for webhook", "pipeline", pipeline)
			continue
		}
		pipelines = append(pipelines, pipeline)
	}

	for _, pipeline := range pipelines {
		//applyAuth=false, already auth applied for this flow
		triggerRequest := bean5.TriggerRequest{
			CdWf:        nil,
			Pipeline:    pipeline,
			Artifact:    artifact,
			ApplyAuth:   false,
			TriggeredBy: triggeredBy,
		}
		err = impl.triggerIfAutoStageCdPipeline(triggerRequest)
		if err != nil {
			impl.logger.Debugw("error on trigger cd pipeline", "err", err)
			return hasAnyTriggered, err
		}
		hasAnyTriggered = true
	}

	return hasAnyTriggered, err
}

// if stage is present with 0 stage steps, delete the stage
// handle corrupt data (https://github.com/devtron-labs/devtron/issues/3826)
func (impl *WorkflowDagExecutorImpl) deleteCorruptedPipelineStage(pipelineStage *repository4.PipelineStage, triggeredBy int32) (error, bool) {
	if pipelineStage != nil {
		stageReq := &bean3.PipelineStageDto{
			Id:   pipelineStage.Id,
			Type: pipelineStage.Type,
		}
		err, deleted := impl.pipelineStageService.DeletePipelineStageIfReq(stageReq, triggeredBy)
		if err != nil {
			impl.logger.Errorw("error in deleting the corrupted pipeline stage", "err", err, "pipelineStageReq", stageReq)
			return err, false
		}
		return nil, deleted
	}
	return nil, false
}

func (impl *WorkflowDagExecutorImpl) triggerIfAutoStageCdPipeline(request bean5.TriggerRequest) error {

	preStage, err := impl.getPipelineStage(request.Pipeline.Id, repository4.PIPELINE_STAGE_TYPE_PRE_CD)
	if err != nil {
		return err
	}

	//handle corrupt data (https://github.com/devtron-labs/devtron/issues/3826)
	err, deleted := impl.deleteCorruptedPipelineStage(preStage, request.TriggeredBy)
	if err != nil {
		impl.logger.Errorw("error in deleteCorruptedPipelineStage ", "cdPipelineId", request.Pipeline.Id, "err", err, "preStage", preStage, "triggeredBy", request.TriggeredBy)
		return err
	}

	request.TriggerContext.Context = context.Background()
	if len(request.Pipeline.PreStageConfig) > 0 || (preStage != nil && !deleted) {
		// pre stage exists
		if request.Pipeline.PreTriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC {
			impl.logger.Debugw("trigger pre stage for pipeline", "artifactId", request.Artifact.Id, "pipelineId", request.Pipeline.Id)
			err = impl.cdTriggerService.TriggerPreStage(request) // TODO handle error here
			return err
		}
	} else if request.Pipeline.TriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC {
		// trigger deployment
		if request.Pipeline.ApprovalNodeConfigured() {
			impl.logger.Warnw("approval node configured, so skipping pipeline for approval", "pipeline", request.Pipeline)
			return nil
		}
		impl.logger.Debugw("trigger cd for pipeline", "artifactId", request.Artifact.Id, "pipelineId", request.Pipeline.Id)
		err = impl.cdTriggerService.TriggerAutomaticDeployment(request)
		return err
	}
	return nil
}

// this function is for internal use only, this doesn't always guarantee pipeline stage even if pre/post-cd stage is configured
// because for old pipelines their pre-cd and post-cd data is stored in pipeline table in yaml format.
// TODO: remove
func (impl *WorkflowDagExecutorImpl) getPipelineStage(pipelineId int, stageType repository4.PipelineStageType) (*repository4.PipelineStage, error) {
	stage, err := impl.pipelineStageService.GetCdStageByCdPipelineIdAndStageType(pipelineId, stageType)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching CD pipeline stage", "cdPipelineId", pipelineId, "stage ", stage, "err", err)
		return nil, err
	}
	return stage, nil
}

func (impl *WorkflowDagExecutorImpl) HandlePreStageSuccessEvent(triggerContext bean5.TriggerContext, cdStageCompleteEvent bean7.CdStageCompleteEvent) error {
	wfRunner, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(cdStageCompleteEvent.WorkflowRunnerId)
	if err != nil {
		return err
	}
	if wfRunner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		ciArtifact, err := impl.ciArtifactRepository.Get(cdStageCompleteEvent.CiArtifactDTO.Id)
		if err != nil {
			return err
		}
		// Migration of deprecated DataSource Type
		if ciArtifact.IsMigrationRequired() {
			migrationErr := impl.ciArtifactRepository.MigrateToWebHookDataSourceType(ciArtifact.Id)
			if migrationErr != nil {
				impl.logger.Warnw("unable to migrate deprecated DataSource", "artifactId", ciArtifact.Id)
			}
		}
		PreCDArtifacts, err := impl.commonArtifactService.SavePluginArtifacts(ciArtifact, cdStageCompleteEvent.PluginRegistryArtifactDetails, cdStageCompleteEvent.CdPipelineId, repository.PRE_CD, cdStageCompleteEvent.TriggeredBy)
		if err != nil {
			impl.logger.Errorw("error in saving plugin artifacts", "err", err)
			return err
		}
		ciArtifactId := 0
		if len(PreCDArtifacts) > 0 {
			ciArtifactId = PreCDArtifacts[0].Id // deployment will be trigger with artifact copied by plugin
		} else {
			ciArtifactId = cdStageCompleteEvent.CiArtifactDTO.Id
		}
		err = impl.cdTriggerService.TriggerAutoCDOnPreStageSuccess(triggerContext, cdStageCompleteEvent.CdPipelineId, ciArtifactId, cdStageCompleteEvent.WorkflowId, cdStageCompleteEvent.TriggeredBy)
		if err != nil {
			impl.logger.Errorw("error in triggering cd on pre cd succcess", "err", err)
			return err
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) HandleDeploymentSuccessEvent(triggerContext bean5.TriggerContext, pipelineOverride *chartConfig.PipelineOverride) error {
	if pipelineOverride == nil {
		return fmt.Errorf("invalid request, pipeline override not found")
	}
	cdWorkflow, err := impl.cdWorkflowRepository.FindById(pipelineOverride.CdWorkflowId)
	if err != nil {
		impl.logger.Errorw("error in fetching cd workflow by id", "pipelineOverride", pipelineOverride)
		return err
	}

	postStage, err := impl.getPipelineStage(pipelineOverride.PipelineId, repository4.PIPELINE_STAGE_TYPE_POST_CD)
	if err != nil {
		return err
	}

	var triggeredByUser int32 = 1
	//handle corrupt data (https://github.com/devtron-labs/devtron/issues/3826)
	err, deleted := impl.deleteCorruptedPipelineStage(postStage, triggeredByUser)
	if err != nil {
		impl.logger.Errorw("error in deleteCorruptedPipelineStage ", "err", err, "preStage", postStage, "triggeredBy", triggeredByUser)
		return err
	}

	if len(pipelineOverride.Pipeline.PostStageConfig) > 0 || (postStage != nil && !deleted) {
		if pipelineOverride.Pipeline.PostTriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC &&
			pipelineOverride.DeploymentType != models.DEPLOYMENTTYPE_STOP &&
			pipelineOverride.DeploymentType != models.DEPLOYMENTTYPE_START {

			triggerRequest := bean5.TriggerRequest{
				CdWf:                  cdWorkflow,
				Pipeline:              pipelineOverride.Pipeline,
				TriggeredBy:           triggeredByUser,
				TriggerContext:        triggerContext,
				RefCdWorkflowRunnerId: 0,
			}
			triggerRequest.TriggerContext.Context = context.Background()
			err = impl.cdTriggerService.TriggerPostStage(triggerRequest)
			if err != nil {
				impl.logger.Errorw("error in triggering post stage after successful deployment event", "err", err, "cdWorkflow", cdWorkflow)
				return err
			}
		}
	} else {
		// to trigger next pre/cd, if any
		// finding children cd by pipeline id
		err = impl.HandlePostStageSuccessEvent(triggerContext, cdWorkflow.Id, pipelineOverride.PipelineId, 1, nil)
		if err != nil {
			impl.logger.Errorw("error in triggering children cd after successful deployment event", "parentCdPipelineId", pipelineOverride.PipelineId)
			return err
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) HandlePostStageSuccessEvent(triggerContext bean5.TriggerContext, cdWorkflowId int, cdPipelineId int, triggeredBy int32, pluginRegistryImageDetails map[string][]string) error {
	// finding children cd by pipeline id
	cdPipelinesMapping, err := impl.appWorkflowRepository.FindWFCDMappingByParentCDPipelineId(cdPipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting mapping of cd pipelines by parent cd pipeline id", "err", err, "parentCdPipelineId", cdPipelineId)
		return err
	}
	ciArtifact, err := impl.ciArtifactRepository.GetArtifactByCdWorkflowId(cdWorkflowId)
	if err != nil {
		impl.logger.Errorw("error in finding artifact by cd workflow id", "err", err, "cdWorkflowId", cdWorkflowId)
		return err
	}

	linkedMappings, linkedArtifactsMap, err := impl.processLinkedCDPipelines(cdPipelineId, ciArtifact, triggeredBy)
	if err != nil {
		impl.logger.Errorw("error in processing linked cd pipelines", "err", err, "cdPipelineId", cdPipelineId)
		return err
	}
	cdPipelinesMapping = append(cdPipelinesMapping, linkedMappings...)

	if len(pluginRegistryImageDetails) > 0 {
		PostCDArtifacts, err := impl.commonArtifactService.SavePluginArtifacts(ciArtifact, pluginRegistryImageDetails, cdPipelineId, repository.POST_CD, triggeredBy)
		if err != nil {
			impl.logger.Errorw("error in saving plugin artifacts", "err", err)
			return err
		}
		if len(PostCDArtifacts) > 0 {
			ciArtifact = PostCDArtifacts[0]
		}
	}
	for _, cdPipelineMapping := range cdPipelinesMapping {
		// find pipeline by cdPipeline ID
		pipeline, err := impl.pipelineRepository.FindById(cdPipelineMapping.ComponentId)
		if err != nil {
			impl.logger.Errorw("error in getting cd pipeline by id", "err", err, "pipelineId", cdPipelineMapping.ComponentId)
			return err
		}
		// finding ci artifact by ciPipelineID and pipelineId
		// TODO : confirm values for applyAuth, async & triggeredBy

		triggerArtifact := ciArtifact
		if artifact, ok := linkedArtifactsMap[cdPipelineMapping.ParentId]; ok {
			triggerArtifact = artifact
		}

		triggerRequest := bean5.TriggerRequest{
			CdWf:           nil,
			Pipeline:       pipeline,
			Artifact:       triggerArtifact,
			TriggeredBy:    triggeredBy,
			TriggerContext: triggerContext,
		}

		err = impl.triggerIfAutoStageCdPipeline(triggerRequest)
		if err != nil {
			impl.logger.Errorw("error in triggering cd pipeline after successful post stage", "err", err, "pipelineId", pipeline.Id)
			return err
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) HandleCiSuccessEvent(triggerContext bean5.TriggerContext, ciPipelineId int, request *bean2.CiArtifactWebhookRequest, imagePushedAt *time.Time) (id int, err error) {
	impl.logger.Infow("webhook for artifact save", "req", request)
	if request.WorkflowId != nil {
		savedWorkflow, err := impl.ciWorkflowRepository.FindById(*request.WorkflowId)
		if err != nil {
			impl.logger.Errorw("cannot get saved wf", "err", err)
			return 0, err
		}
		// if workflow already cancelled then return, this state arises when user force aborts a ci
		if savedWorkflow.Status == executors.WorkflowCancel {
			return 0, err
		}
		savedWorkflow.Status = string(v1alpha1.NodeSucceeded)
		impl.logger.Debugw("updating workflow ", "savedWorkflow", savedWorkflow)
		err = impl.ciWorkflowRepository.UpdateWorkFlow(savedWorkflow)
		if err != nil {
			impl.logger.Errorw("update wf failed for id ", "err", err)
			return 0, err
		}
	}

	pipeline, err := impl.ciPipelineRepository.FindByCiAndAppDetailsById(ciPipelineId)
	if request.PipelineName == "" {
		request.PipelineName = pipeline.Name
	}
	if err != nil {
		impl.logger.Errorw("unable to find pipeline", "name", request.PipelineName, "err", err)
		return 0, err
	}
	materialJson, err := request.MaterialInfo.MarshalJSON()
	if err != nil {
		impl.logger.Errorw("unable to marshal material metadata", "err", err)
		return 0, err
	}
	dst := new(bytes.Buffer)
	err = json.Compact(dst, materialJson)
	if err != nil {
		return 0, err
	}
	materialJson = dst.Bytes()
	createdOn := time.Now()
	updatedOn := time.Now()
	if !imagePushedAt.IsZero() {
		createdOn = *imagePushedAt
	}
	buildArtifact := &repository.CiArtifact{
		Image:              request.Image,
		ImageDigest:        request.ImageDigest,
		MaterialInfo:       string(materialJson),
		DataSource:         request.DataSource,
		PipelineId:         pipeline.Id,
		WorkflowId:         request.WorkflowId,
		ScanEnabled:        pipeline.ScanEnabled,
		Scanned:            false,
		IsArtifactUploaded: request.IsArtifactUploaded,
		AuditLog:           sql.AuditLog{CreatedBy: request.UserId, UpdatedBy: request.UserId, CreatedOn: createdOn, UpdatedOn: updatedOn},
	}
	plugin, err := impl.globalPluginRepository.GetPluginByName(bean3.VULNERABILITY_SCANNING_PLUGIN)
	if err != nil || len(plugin) == 0 {
		impl.logger.Errorw("error in getting image scanning plugin", "err", err)
		return 0, err
	}
	isScanPluginConfigured, err := impl.pipelineStageRepository.CheckPluginExistsInCiPipeline(pipeline.Id, string(repository4.PIPELINE_STAGE_TYPE_POST_CI), plugin[0].Id)
	if err != nil {
		impl.logger.Errorw("error in getting ci pipeline plugin", "err", err, "pipelineId", pipeline.Id, "pluginId", plugin[0].Id)
		return 0, err
	}
	if pipeline.ScanEnabled || isScanPluginConfigured {
		buildArtifact.Scanned = true
		buildArtifact.ScanEnabled = true
	}
	if err = impl.ciArtifactRepository.Save(buildArtifact); err != nil {
		impl.logger.Errorw("error in saving material", "err", err)
		return 0, err
	}

	var pluginArtifacts []*repository.CiArtifact
	for registry, artifacts := range request.PluginRegistryArtifactDetails {
		for _, image := range artifacts {
			if pipeline.PipelineType == bean3.CI_JOB && image == "" {
				continue
			}
			pluginArtifact := &repository.CiArtifact{
				Image:                 image,
				ImageDigest:           request.ImageDigest,
				MaterialInfo:          string(materialJson),
				DataSource:            request.PluginArtifactStage,
				ComponentId:           pipeline.Id,
				PipelineId:            pipeline.Id,
				AuditLog:              sql.AuditLog{CreatedBy: request.UserId, UpdatedBy: request.UserId, CreatedOn: createdOn, UpdatedOn: updatedOn},
				CredentialsSourceType: repository.GLOBAL_CONTAINER_REGISTRY,
				CredentialSourceValue: registry,
				ParentCiArtifact:      buildArtifact.Id,
				Scanned:               buildArtifact.Scanned,
				ScanEnabled:           buildArtifact.ScanEnabled,
			}
			pluginArtifacts = append(pluginArtifacts, pluginArtifact)
		}
	}
	if len(pluginArtifacts) > 0 {
		_, err = impl.ciArtifactRepository.SaveAll(pluginArtifacts)
		if err != nil {
			impl.logger.Errorw("error while saving ci artifacts", "err", err)
			return 0, err
		}
	}

	childrenCi, err := impl.ciPipelineRepository.FindByParentCiPipelineId(ciPipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetching childern ci ", "err", err)
		return 0, err
	}

	var ciArtifactArr []*repository.CiArtifact
	for _, ci := range childrenCi {
		ciArtifact := &repository.CiArtifact{
			Image:              request.Image,
			ImageDigest:        request.ImageDigest,
			MaterialInfo:       string(materialJson),
			DataSource:         request.DataSource,
			PipelineId:         ci.Id,
			ParentCiArtifact:   buildArtifact.Id,
			ScanEnabled:        ci.ScanEnabled,
			Scanned:            false,
			IsArtifactUploaded: request.IsArtifactUploaded,
			AuditLog:           sql.AuditLog{CreatedBy: request.UserId, UpdatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now()},
		}
		if ci.ScanEnabled {
			ciArtifact.Scanned = true
		}
		ciArtifactArr = append(ciArtifactArr, ciArtifact)
	}

	impl.logger.Debugw("saving ci artifacts", "art", ciArtifactArr)
	if len(ciArtifactArr) > 0 {
		_, err = impl.ciArtifactRepository.SaveAll(ciArtifactArr)
		if err != nil {
			impl.logger.Errorw("error while saving ci artifacts", "err", err)
			return 0, err
		}
	}
	if len(pluginArtifacts) == 0 {
		ciArtifactArr = append(ciArtifactArr, buildArtifact)
	} else {
		ciArtifactArr = append(ciArtifactArr, pluginArtifacts[0])
	}
	go impl.WriteCISuccessEvent(request, pipeline, buildArtifact)
	async := false

	// execute auto trigger in batch on CI success event
	totalCIArtifactCount := len(ciArtifactArr)
	batchSize := impl.ciConfig.CIAutoTriggerBatchSize
	// handling to avoid infinite loop
	if batchSize <= 0 {
		batchSize = 1
	}
	start := time.Now()
	impl.logger.Infow("Started: auto trigger for children Stage/CD pipelines", "Artifact count", totalCIArtifactCount)
	for i := 0; i < totalCIArtifactCount; {
		// requests left to process
		remainingBatch := totalCIArtifactCount - i
		if remainingBatch < batchSize {
			batchSize = remainingBatch
		}
		var wg sync.WaitGroup
		for j := 0; j < batchSize; j++ {
			wg.Add(1)
			index := i + j
			go func(index int) {
				defer wg.Done()
				ciArtifact := ciArtifactArr[index]
				// handle individual CiArtifact success event
				err = impl.handleCiSuccessEvent(triggerContext, ciArtifact, async, request.UserId)
				if err != nil {
					impl.logger.Errorw("error on handle  ci success event", "ciArtifactId", ciArtifact.Id, "err", err)
				}
			}(index)
		}
		wg.Wait()
		i += batchSize
	}
	impl.logger.Debugw("Completed: auto trigger for children Stage/CD pipelines", "Time taken", time.Since(start).Seconds())
	return buildArtifact.Id, err
}

func (impl *WorkflowDagExecutorImpl) WriteCISuccessEvent(request *bean2.CiArtifactWebhookRequest, pipeline *pipelineConfig.CiPipeline, artifact *repository.CiArtifact) {
	event := impl.eventFactory.Build(util2.Success, &pipeline.Id, pipeline.AppId, nil, util2.CI)
	event.CiArtifactId = artifact.Id
	if artifact.WorkflowId != nil {
		event.CiWorkflowRunnerId = *artifact.WorkflowId
	}
	event.UserId = int(request.UserId)
	event = impl.eventFactory.BuildExtraCIData(event, nil, artifact.Image)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("error in writing event", "err", evtErr)
	}
}

func (impl *WorkflowDagExecutorImpl) HandleCiStepFailedEvent(ciPipelineId int, request *bean2.CiArtifactWebhookRequest) (err error) {

	savedWorkflow, err := impl.ciWorkflowRepository.FindById(*request.WorkflowId)
	if err != nil {
		impl.logger.Errorw("cannot get saved wf", "wf ID: ", *request.WorkflowId, "err", err)
		return err
	}

	pipeline, err := impl.ciPipelineRepository.FindByCiAndAppDetailsById(ciPipelineId)
	if err != nil {
		impl.logger.Errorw("unable to find pipeline", "ID", ciPipelineId, "err", err)
		return err
	}

	go func() {
		if len(savedWorkflow.ImagePathReservationIds) > 0 {
			err = impl.customTagService.DeactivateImagePathReservationByImageIds(savedWorkflow.ImagePathReservationIds)
			if err != nil {
				impl.logger.Errorw("unable to deactivate impage_path_reservation ", err)
			}
		}
	}()

	go impl.WriteCIStepFailedEvent(pipeline, request, savedWorkflow)
	return nil
}

func (impl *WorkflowDagExecutorImpl) WriteCIStepFailedEvent(pipeline *pipelineConfig.CiPipeline, request *bean2.CiArtifactWebhookRequest, ciWorkflow *pipelineConfig.CiWorkflow) {
	event := impl.eventFactory.Build(util2.Fail, &pipeline.Id, pipeline.AppId, nil, util2.CI)
	material := &client.MaterialTriggerInfo{}
	material.GitTriggers = ciWorkflow.GitTriggers
	event.CiWorkflowRunnerId = ciWorkflow.Id
	event.UserId = int(ciWorkflow.TriggeredBy)
	event = impl.eventFactory.BuildExtraCIData(event, material, request.Image)
	event.CiArtifactId = 0
	event.Payload.FailureReason = request.FailureReason
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("error in writing event: ", event, "error: ", evtErr)
	}
}

func (impl *WorkflowDagExecutorImpl) HandleExternalCiWebhook(externalCiId int, request *bean2.CiArtifactWebhookRequest,
	auth func(token string, projectObject string, envObject string) bool, token string) (id int, err error) {
	externalCiPipeline, err := impl.ciPipelineRepository.FindExternalCiById(externalCiId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching external ci", "err", err)
		return 0, err
	}
	if externalCiPipeline.Id == 0 {
		impl.logger.Errorw("invalid external ci id", "externalCiId", externalCiId, "err", err)
		return 0, &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "invalid external ci id"}
	}

	impl.logger.Infow("request of webhook external ci", "req", request)
	materialJson, err := request.MaterialInfo.MarshalJSON()
	if err != nil {
		impl.logger.Errorw("unable to marshal material metadata", "err", err)
		return 0, err
	}
	dst := new(bytes.Buffer)
	err = json.Compact(dst, materialJson)
	if err != nil {
		impl.logger.Errorw("parsing error", "err", err)
		return 0, err
	}
	materialJson = dst.Bytes()
	artifact := &repository.CiArtifact{
		Image:                request.Image,
		ImageDigest:          request.ImageDigest,
		MaterialInfo:         string(materialJson),
		DataSource:           request.DataSource,
		WorkflowId:           request.WorkflowId,
		ExternalCiPipelineId: externalCiId,
		ScanEnabled:          false,
		Scanned:              false,
		IsArtifactUploaded:   request.IsArtifactUploaded,
		AuditLog:             sql.AuditLog{CreatedBy: request.UserId, UpdatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now()},
	}
	if err = impl.ciArtifactRepository.Save(artifact); err != nil {
		impl.logger.Errorw("error in saving material", "err", err)
		return 0, err
	}

	hasAnyTriggered, err := impl.handleWebhookExternalCiEvent(artifact, request.UserId, externalCiId, auth, token)
	if err != nil {
		impl.logger.Errorw("error on handle ext ci webhook", "err", err)
		// if none of the child node has been triggered
		if !hasAnyTriggered {
			if err1 := impl.ciArtifactRepository.Delete(artifact); err1 != nil {
				impl.logger.Errorw("error in rollback artifact", "err", err1)
				return 0, err1
			}
		}
	}
	return artifact.Id, err
}

// TODO: move in adapter
func (impl *WorkflowDagExecutorImpl) BuildCiArtifactRequestForWebhook(event pipeline.ExternalCiWebhookDto) (*bean2.CiArtifactWebhookRequest, error) {
	ciMaterialInfos := make([]repository.CiMaterialInfo, 0)
	if event.MaterialType == "" {
		event.MaterialType = "git"
	}
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

	request := &bean2.CiArtifactWebhookRequest{
		Image:              event.DockerImage,
		ImageDigest:        event.Digest,
		DataSource:         event.DataSource,
		PipelineName:       event.PipelineName,
		MaterialInfo:       rawMaterialInfo,
		UserId:             event.TriggeredBy,
		WorkflowId:         event.WorkflowId,
		IsArtifactUploaded: event.IsArtifactUploaded,
	}
	// if DataSource is empty, repository.WEBHOOK is considered as default
	if request.DataSource == "" {
		request.DataSource = repository.WEBHOOK
	}
	return request, nil
}

//TO check where to put, got from oss enterprise diff

func getCopiedArtifact(ciArtifact *repository.CiArtifact, pipelineId int, userId int32) *repository.CiArtifact {
	artifact := &repository.CiArtifact{
		Image:              ciArtifact.Image,
		ImageDigest:        ciArtifact.ImageDigest,
		MaterialInfo:       ciArtifact.MaterialInfo,
		DataSource:         ciArtifact.DataSource,
		ScanEnabled:        ciArtifact.ScanEnabled,
		Scanned:            ciArtifact.Scanned,
		IsArtifactUploaded: ciArtifact.IsArtifactUploaded,
		ParentCiArtifact:   ciArtifact.Id,
		PipelineId:         pipelineId,
		AuditLog:           sql.AuditLog{CreatedBy: userId, UpdatedBy: userId, CreatedOn: time.Now(), UpdatedOn: time.Now()},
	}
	if ciArtifact.ParentCiArtifact > 0 {
		artifact.ParentCiArtifact = ciArtifact.ParentCiArtifact
	}
	if ciArtifact.ExternalCiPipelineId > 0 {
		artifact.ExternalCiPipelineId = ciArtifact.ExternalCiPipelineId
	}
	return artifact
}

func (impl *WorkflowDagExecutorImpl) saveArtifactsForLinkedCDPipelines(linkedCiPipelineIds []int, ciArtifact *repository.CiArtifact, triggeredBy int32) (map[int]*repository.CiArtifact, error) {

	ciPipelineIdToArtifacts := make(map[int]*repository.CiArtifact)
	// existingArtifacts are for rollback/redeployment cases, where artifacts already exits
	existingArtifacts, err := impl.ciArtifactRepository.GetArtifactsByCiPipelineIds(linkedCiPipelineIds)
	if err != nil {
		impl.logger.Errorw("error while fetching ci artifacts for linked CD pipelines", "err", err, "ciPipelineIds", linkedCiPipelineIds)
		return ciPipelineIdToArtifacts, err
	}

	ciIdToExistingArtifact := make(map[int]repository.CiArtifact)
	for _, artifact := range existingArtifacts {
		// need to compare image for idempotency
		// Skopeo images will have same digest but different images
		if ciArtifact.Image == artifact.Image {
			ciIdToExistingArtifact[artifact.PipelineId] = artifact
		}
	}

	var newCiArtifactArr []*repository.CiArtifact
	var existingCiArtifactArr []*repository.CiArtifact
	var existingArtifactsIds []int
	for _, pipelineId := range linkedCiPipelineIds {

		if existingArtifact, ok := ciIdToExistingArtifact[pipelineId]; !ok {
			artifact := getCopiedArtifact(ciArtifact, pipelineId, triggeredBy)
			newCiArtifactArr = append(newCiArtifactArr, artifact)
		} else {
			existingCiArtifactArr = append(existingCiArtifactArr, &existingArtifact)
			existingArtifactsIds = append(existingArtifactsIds, existingArtifact.Id)
		}
	}

	savedCIArtifacts, err := impl.ciArtifactRepository.SaveAll(newCiArtifactArr)
	if err != nil {
		impl.logger.Errorw("error while saving ci artifacts for linked CD pipelines", "err", err, "linkedCiPipelineIds", linkedCiPipelineIds)
		return ciPipelineIdToArtifacts, err
	}

	// not needed for now, need to uncomment in order to show tag for image running on parent
	// err = impl.ciArtifactRepository.UpdateLatestTimestamp(existingArtifactsIds)
	// if err != nil {
	//	impl.logger.Errorw("error while updating ci artifacts for linked CD pipelines", "err", err, "cdPipelineId", cdPipelineId)
	//	return nil, nil, err
	// }

	allArtifacts := append(savedCIArtifacts, existingCiArtifactArr...)
	for _, artifact := range allArtifacts {
		ciPipelineIdToArtifacts[artifact.PipelineId] = artifact
	}
	return ciPipelineIdToArtifacts, nil
}

func (impl *WorkflowDagExecutorImpl) getLinkedCDPipelines(cdPipelineId int) ([]*appWorkflow.AppWorkflowMapping, []int, error) {
	linkedCiPipelineIds := make([]int, 0)
	linkedPipelines, err := impl.ciPipelineRepository.FindByParentIdAndType(cdPipelineId, string(bean4.LINKED_CD))
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in finding linked CD pipelines", "err", err, "cdPipelineId", cdPipelineId)
		return nil, linkedCiPipelineIds, err
	}
	linkedCDMappings := make([]*appWorkflow.AppWorkflowMapping, 0)
	if len(linkedPipelines) == 0 {
		return linkedCDMappings, linkedCiPipelineIds, nil
	}

	for _, pipeline := range linkedPipelines {
		linkedCiPipelineIds = append(linkedCiPipelineIds, pipeline.Id)
	}

	mappings, err := impl.appWorkflowRepository.FindWFCDMappingByCIPipelineIds(linkedCiPipelineIds)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching linked CD pipelines for parent CI", "err", err, "ciPipelineIds", linkedCiPipelineIds)
		return nil, linkedCiPipelineIds, err
	}

	// will return empty if mappings is nil
	linkedCDMappings = append(linkedCDMappings, mappings...)
	return linkedCDMappings, linkedCiPipelineIds, nil
}

func (impl *WorkflowDagExecutorImpl) processLinkedCDPipelines(cdPipelineId int, ciArtifact *repository.CiArtifact, triggeredBy int32) ([]*appWorkflow.AppWorkflowMapping, map[int]*repository.CiArtifact, error) {
	linkedArtifactsMap := make(map[int]*repository.CiArtifact)
	linkedMappings, linkedCIPipelineIds, err := impl.getLinkedCDPipelines(cdPipelineId)
	if err != nil || len(linkedMappings) == 0 {
		return linkedMappings, linkedArtifactsMap, err
	}
	linkedArtifactsMap, err = impl.saveArtifactsForLinkedCDPipelines(linkedCIPipelineIds, ciArtifact, triggeredBy)
	return linkedMappings, linkedArtifactsMap, err
}

func (impl *WorkflowDagExecutorImpl) FetchApprovalPendingArtifacts(pipelineId, limit, offset, requiredApprovals int, searchString string) ([]bean4.CiArtifactBean, int, error) {

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

func (impl *WorkflowDagExecutorImpl) getLatestDeploymentByArtifactIds(pipelineId int, deploymentApprovalRequests []*pipelineConfig.DeploymentApprovalRequest, artifactIds []int) ([]*pipelineConfig.DeploymentApprovalRequest, error) {
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
