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

package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	bean6 "github.com/devtron-labs/devtron/api/helm-app/bean"
	client2 "github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/pkg/build/artifacts"
	bean4 "github.com/devtron-labs/devtron/pkg/deployment/deployedApp/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	bean5 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	bean7 "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	"sync"
	"time"

	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	gitSensorClient "github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/imageDigestPolicy"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/plugin"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	"github.com/devtron-labs/devtron/pkg/variables"
	util4 "github.com/devtron-labs/devtron/util"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"k8s.io/utils/strings/slices"

	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	history2 "github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type WorkflowDagExecutor interface {
	HandleCiSuccessEvent(triggerContext bean5.TriggerContext, artifact *repository.CiArtifact, async bool, triggeredBy int32) error
	HandleWebhookExternalCiEvent(artifact *repository.CiArtifact, triggeredBy int32, externalCiId int, auth func(token string, projectObject string, envObject string) bool, token string) (bool, error)
	HandlePreStageSuccessEvent(triggerContext bean5.TriggerContext, cdStageCompleteEvent bean7.CdStageCompleteEvent) error
	HandleDeploymentSuccessEvent(triggerContext bean5.TriggerContext, pipelineOverride *chartConfig.PipelineOverride) error
	HandlePostStageSuccessEvent(triggerContext bean5.TriggerContext, cdWorkflowId int, cdPipelineId int, triggeredBy int32, pluginRegistryImageDetails map[string][]string) error

	TriggerBulkHibernateAsync(request StopDeploymentGroupRequest, ctx context.Context) (interface{}, error)

	UpdateWorkflowRunnerStatusForDeployment(appIdentifier *client2.AppIdentifier, wfr *pipelineConfig.CdWorkflowRunner, skipReleaseNotFound bool) bool
	OnDeleteCdPipelineEvent(pipelineId int, triggeredBy int32)
	GetTriggerValidateFuncs() []pubsub.ValidateMsg
}

type WorkflowDagExecutorImpl struct {
	logger                        *zap.SugaredLogger
	pipelineRepository            pipelineConfig.PipelineRepository
	cdWorkflowRepository          pipelineConfig.CdWorkflowRepository
	pubsubClient                  *pubsub.PubSubClientServiceImpl
	cdWorkflowService             WorkflowService
	ciPipelineRepository          pipelineConfig.CiPipelineRepository
	materialRepository            pipelineConfig.MaterialRepository
	pipelineOverrideRepository    chartConfig.PipelineOverrideRepository
	ciArtifactRepository          repository.CiArtifactRepository
	user                          user.UserService
	enforcerUtil                  rbac.EnforcerUtil
	groupRepository               repository.DeploymentGroupRepository
	envRepository                 repository2.EnvironmentRepository
	eventFactory                  client.EventFactory
	eventClient                   client.EventClient
	cvePolicyRepository           security.CvePolicyRepository
	scanResultRepository          security.ImageScanResultRepository
	appWorkflowRepository         appWorkflow.AppWorkflowRepository
	prePostCdScriptHistoryService history2.PrePostCdScriptHistoryService
	CiTemplateRepository          pipelineConfig.CiTemplateRepository
	ciWorkflowRepository          pipelineConfig.CiWorkflowRepository
	appLabelRepository            pipelineConfig.AppLabelRepository
	gitSensorGrpcClient           gitSensorClient.Client
	pipelineStageService          PipelineStageService
	config                        *types.CdConfig
	appServiceConfig              *app.AppServiceConfig
	globalPluginService           plugin.GlobalPluginService

	scopedVariableManager     variables.ScopedVariableCMCSManager
	pluginInputVariableParser PluginInputVariableParser

	devtronAsyncHelmInstallRequestMap  map[int]bool
	devtronAsyncHelmInstallRequestLock *sync.Mutex
	devtronAppReleaseContextMap        map[int]DevtronAppReleaseContextType
	devtronAppReleaseContextMapLock    *sync.Mutex

	helmAppService           client2.HelmAppService
	customTagService         CustomTagService
	imageDigestPolicyService imageDigestPolicy.ImageDigestPolicyService

	cdWorkflowCommonService cd.CdWorkflowCommonService
	cdTriggerService        devtronApps.TriggerService

	deployedConfigurationHistoryService history2.DeployedConfigurationHistoryService
	manifestCreationService             manifest.ManifestCreationService
	commonArtifactService               artifacts.CommonArtifactService
}

const (
	GIT_COMMIT_HASH_PREFIX       = "GIT_COMMIT_HASH"
	GIT_SOURCE_TYPE_PREFIX       = "GIT_SOURCE_TYPE"
	GIT_SOURCE_VALUE_PREFIX      = "GIT_SOURCE_VALUE"
	GIT_SOURCE_COUNT             = "GIT_SOURCE_COUNT"
	APP_LABEL_KEY_PREFIX         = "APP_LABEL_KEY"
	APP_LABEL_VALUE_PREFIX       = "APP_LABEL_VALUE"
	APP_LABEL_COUNT              = "APP_LABEL_COUNT"
	CHILD_CD_ENV_NAME_PREFIX     = "CHILD_CD_ENV_NAME"
	CHILD_CD_CLUSTER_NAME_PREFIX = "CHILD_CD_CLUSTER_NAME"
	CHILD_CD_COUNT               = "CHILD_CD_COUNT"
	DEVTRON_SYSTEM_USER_ID       = 1
	ARGOCD_REFRESH_ERROR         = "Error in refreshing argocd app"
)

type DevtronAppReleaseContextType struct {
	CancelContext context.CancelFunc
	RunnerId      int
}

//
//type bean5.TriggerRequest struct {
//	CdWf                  *pipelineConfig.CdWorkflow
//	Pipeline              *pipelineConfig.Pipeline
//	Artifact              *repository.CiArtifact
//	ApplyAuth             bool
//	TriggeredBy           int32
//	RefCdWorkflowRunnerId int
//	TriggerContext
//}
//
//type bean5.TriggerContext struct {
//	// Context is a context object to be passed to the pipeline trigger
//	// +optional
//	Context context.Context
//	// ReferenceId is a unique identifier for the workflow runner
//	// refer pipelineConfig.CdWorkflowRunner
//	ReferenceId *string
//}

func NewWorkflowDagExecutorImpl(Logger *zap.SugaredLogger, pipelineRepository pipelineConfig.PipelineRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pubsubClient *pubsub.PubSubClientServiceImpl,
	cdWorkflowService WorkflowService,
	ciArtifactRepository repository.CiArtifactRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	materialRepository pipelineConfig.MaterialRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	user user.UserService,
	groupRepository repository.DeploymentGroupRepository,
	envRepository repository2.EnvironmentRepository,
	enforcerUtil rbac.EnforcerUtil, eventFactory client.EventFactory,
	eventClient client.EventClient, cvePolicyRepository security.CvePolicyRepository,
	scanResultRepository security.ImageScanResultRepository,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	prePostCdScriptHistoryService history2.PrePostCdScriptHistoryService,
	CiTemplateRepository pipelineConfig.CiTemplateRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	appLabelRepository pipelineConfig.AppLabelRepository, gitSensorGrpcClient gitSensorClient.Client,
	pipelineStageService PipelineStageService,
	globalPluginService plugin.GlobalPluginService,
	pluginInputVariableParser PluginInputVariableParser,
	scopedVariableManager variables.ScopedVariableCMCSManager,
	helmAppService client2.HelmAppService,
	pipelineConfigListenerService PipelineConfigListenerService,
	customTagService CustomTagService,
	imageDigestPolicyService imageDigestPolicy.ImageDigestPolicyService,
	cdWorkflowCommonService cd.CdWorkflowCommonService,
	cdTriggerService devtronApps.TriggerService,
	deployedConfigurationHistoryService history2.DeployedConfigurationHistoryService,
	manifestCreationService manifest.ManifestCreationService,
	commonArtifactService artifacts.CommonArtifactService) *WorkflowDagExecutorImpl {
	wde := &WorkflowDagExecutorImpl{logger: Logger,
		pipelineRepository:            pipelineRepository,
		cdWorkflowRepository:          cdWorkflowRepository,
		pubsubClient:                  pubsubClient,
		cdWorkflowService:             cdWorkflowService,
		ciPipelineRepository:          ciPipelineRepository,
		ciArtifactRepository:          ciArtifactRepository,
		materialRepository:            materialRepository,
		pipelineOverrideRepository:    pipelineOverrideRepository,
		user:                          user,
		enforcerUtil:                  enforcerUtil,
		groupRepository:               groupRepository,
		envRepository:                 envRepository,
		eventFactory:                  eventFactory,
		eventClient:                   eventClient,
		cvePolicyRepository:           cvePolicyRepository,
		scanResultRepository:          scanResultRepository,
		appWorkflowRepository:         appWorkflowRepository,
		prePostCdScriptHistoryService: prePostCdScriptHistoryService,
		CiTemplateRepository:          CiTemplateRepository,
		ciWorkflowRepository:          ciWorkflowRepository,
		appLabelRepository:            appLabelRepository,
		gitSensorGrpcClient:           gitSensorGrpcClient,
		pipelineStageService:          pipelineStageService,
		scopedVariableManager:         scopedVariableManager,
		globalPluginService:           globalPluginService,
		pluginInputVariableParser:     pluginInputVariableParser,

		devtronAsyncHelmInstallRequestMap:   make(map[int]bool),
		devtronAsyncHelmInstallRequestLock:  &sync.Mutex{},
		devtronAppReleaseContextMap:         make(map[int]DevtronAppReleaseContextType),
		devtronAppReleaseContextMapLock:     &sync.Mutex{},
		helmAppService:                      helmAppService,
		customTagService:                    customTagService,
		imageDigestPolicyService:            imageDigestPolicyService,
		cdWorkflowCommonService:             cdWorkflowCommonService,
		cdTriggerService:                    cdTriggerService,
		deployedConfigurationHistoryService: deployedConfigurationHistoryService,
		manifestCreationService:             manifestCreationService,
		commonArtifactService:               commonArtifactService,
	}
	config, err := types.GetCdConfig()
	if err != nil {
		return nil
	}
	wde.config = config
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
	impl.SetPipelineFieldsInOverrideRequest(CDAsyncInstallNatsMessage.ValuesOverrideRequest, pipeline)
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

	_, span := otel.Tracer("orchestrator").Start(ctx, "appService.TriggerRelease")
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

func (impl *WorkflowDagExecutorImpl) HandleCiSuccessEvent(triggerContext bean5.TriggerContext, artifact *repository.CiArtifact, async bool, triggeredBy int32) error {
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

func (impl *WorkflowDagExecutorImpl) HandleWebhookExternalCiEvent(artifact *repository.CiArtifact, triggeredBy int32, externalCiId int, auth func(token string, projectObject string, envObject string) bool, token string) (bool, error) {
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
		if pipeline.TriggerType == pipelineConfig.TRIGGER_TYPE_MANUAL {
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
		impl.logger.Debugw("trigger cd for pipeline", "artifactId", request.Artifact.Id, "pipelineId", request.Pipeline.Id)
		err = impl.cdTriggerService.TriggerAutomaticDeployment(request)
		return err
	}
	return nil
}

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
		pipeline, err := impl.pipelineRepository.FindById(cdStageCompleteEvent.CdPipelineId)
		if err != nil {
			return err
		}
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
		PreCDArtifacts, err := impl.commonArtifactService.SavePluginArtifacts(ciArtifact, cdStageCompleteEvent.PluginRegistryArtifactDetails, pipeline.Id, repository.PRE_CD, cdStageCompleteEvent.TriggeredBy)
		if err != nil {
			impl.logger.Errorw("error in saving plugin artifacts", "err", err)
			return err
		}
		if pipeline.TriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC {
			if len(PreCDArtifacts) > 0 {
				ciArtifact = PreCDArtifacts[0] // deployment will be trigger with artifact copied by plugin
			}
			cdWorkflow, err := impl.cdWorkflowRepository.FindById(cdStageCompleteEvent.WorkflowId)
			if err != nil {
				return err
			}
			//passing applyAuth as false since this event is for auto trigger and user who already has access to this cd can trigger pre cd also
			applyAuth := false
			if cdStageCompleteEvent.TriggeredBy != 1 {
				applyAuth = true
			}
			triggerRequest := bean5.TriggerRequest{
				CdWf:           cdWorkflow,
				Pipeline:       pipeline,
				Artifact:       ciArtifact,
				ApplyAuth:      applyAuth,
				TriggeredBy:    cdStageCompleteEvent.TriggeredBy,
				TriggerContext: triggerContext,
			}
			triggerRequest.TriggerContext.Context = context.Background()
			err = impl.cdTriggerService.TriggerAutomaticDeployment(triggerRequest)
			if err != nil {
				return err
			}
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
		//find pipeline by cdPipeline ID
		pipeline, err := impl.pipelineRepository.FindById(cdPipelineMapping.ComponentId)
		if err != nil {
			impl.logger.Errorw("error in getting cd pipeline by id", "err", err, "pipelineId", cdPipelineMapping.ComponentId)
			return err
		}
		//finding ci artifact by ciPipelineID and pipelineId
		//TODO : confirm values for applyAuth, async & triggeredBy

		triggerRequest := bean5.TriggerRequest{
			CdWf:           nil,
			Pipeline:       pipeline,
			Artifact:       ciArtifact,
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

type StopDeploymentGroupRequest struct {
	DeploymentGroupId int               `json:"deploymentGroupId" validate:"required"`
	UserId            int32             `json:"userId"`
	RequestType       bean4.RequestType `json:"requestType" validate:"oneof=START STOP"`
}

type DeploymentGroupAppWithEnv struct {
	EnvironmentId     int               `json:"environmentId"`
	DeploymentGroupId int               `json:"deploymentGroupId"`
	AppId             int               `json:"appId"`
	Active            bool              `json:"active"`
	UserId            int32             `json:"userId"`
	RequestType       bean4.RequestType `json:"requestType" validate:"oneof=START STOP"`
}

func (impl *WorkflowDagExecutorImpl) TriggerBulkHibernateAsync(request StopDeploymentGroupRequest, ctx context.Context) (interface{}, error) {
	dg, err := impl.groupRepository.FindByIdWithApp(request.DeploymentGroupId)
	if err != nil {
		impl.logger.Errorw("error while fetching dg", "err", err)
		return nil, err
	}

	for _, app := range dg.DeploymentGroupApps {
		deploymentGroupAppWithEnv := &DeploymentGroupAppWithEnv{
			AppId:             app.AppId,
			EnvironmentId:     dg.EnvironmentId,
			DeploymentGroupId: dg.Id,
			Active:            dg.Active,
			UserId:            request.UserId,
			RequestType:       request.RequestType,
		}

		data, err := json.Marshal(deploymentGroupAppWithEnv)
		if err != nil {
			impl.logger.Errorw("error while writing app stop event to nats ", "app", app.AppId, "deploymentGroup", app.DeploymentGroupId, "err", err)
		} else {
			err = impl.pubsubClient.Publish(pubsub.BULK_HIBERNATE_TOPIC, string(data))
			if err != nil {
				impl.logger.Errorw("Error while publishing request", "topic", pubsub.BULK_HIBERNATE_TOPIC, "error", err)
			}
		}
	}
	return nil, nil
}

func (impl *WorkflowDagExecutorImpl) SetPipelineFieldsInOverrideRequest(overrideRequest *bean.ValuesOverrideRequest, pipeline *pipelineConfig.Pipeline) {
	overrideRequest.PipelineId = pipeline.Id
	overrideRequest.PipelineName = pipeline.Name
	overrideRequest.EnvId = pipeline.EnvironmentId
	overrideRequest.EnvName = pipeline.Environment.Name
	overrideRequest.ClusterId = pipeline.Environment.ClusterId
	overrideRequest.AppId = pipeline.AppId
	overrideRequest.AppName = pipeline.App.AppName
	overrideRequest.DeploymentAppType = pipeline.DeploymentAppType
}

// write integration/unit test for each function

// canInitiateTrigger checks if the current trigger request with natsMsgId haven't already initiated the trigger.
// throws error if the request is already processed.
func (impl *WorkflowDagExecutorImpl) canInitiateTrigger(natsMsgId string) (bool, error) {
	if natsMsgId == "" {
		return true, nil
	}
	exists, err := impl.cdWorkflowRepository.CheckWorkflowRunnerByReferenceId(natsMsgId)
	if err != nil {
		impl.logger.Errorw("error in fetching cd workflow runner using reference_id", "referenceId", natsMsgId, "err", err)
		return false, errors.New("error in fetching cd workflow runner")
	}

	if exists {
		impl.logger.Errorw("duplicate pre stage trigger request as there is already a workflow runner object created by this message")
		return false, errors.New("duplicate pre stage trigger request, this request was already processed")
	}

	return true, nil
}

// GetTriggerValidateFuncs gets all the required validation funcs
func (impl *WorkflowDagExecutorImpl) GetTriggerValidateFuncs() []pubsub.ValidateMsg {

	var duplicateTriggerValidateFunc pubsub.ValidateMsg = func(msg model.PubSubMsg) bool {
		if msg.MsgDeliverCount == 1 {
			// first time message got delivered, always validate this.
			return true
		}

		// message is redelivered, check if the message is already processed.
		if ok, err := impl.canInitiateTrigger(msg.MsgId); !ok || err != nil {
			impl.logger.Warnw("duplicate trigger condition, duplicate message", "msgId", msg.MsgId, "err", err)
			return false
		}
		return true
	}

	return []pubsub.ValidateMsg{duplicateTriggerValidateFunc}

}
