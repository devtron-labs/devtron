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
	"github.com/devtron-labs/devtron/pkg/deployment/manifest"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	bean7 "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/out"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	"strconv"
	"strings"
	"sync"
	"time"

	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/common-lib/utils/k8s/health"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	gitSensorClient "github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/imageDigestPolicy"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/plugin"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	"github.com/devtron-labs/devtron/pkg/variables"
	repository5 "github.com/devtron-labs/devtron/pkg/variables/repository"
	util4 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"k8s.io/utils/strings/slices"

	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	history2 "github.com/devtron-labs/devtron/pkg/pipeline/history"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	util3 "github.com/devtron-labs/devtron/pkg/util"

	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	util2 "github.com/devtron-labs/devtron/util/event"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type WorkflowDagExecutor interface {
	HandleCiSuccessEvent(triggerContext TriggerContext, artifact *repository.CiArtifact, async bool, triggeredBy int32) error
	HandleWebhookExternalCiEvent(artifact *repository.CiArtifact, triggeredBy int32, externalCiId int, auth func(token string, projectObject string, envObject string) bool, token string) (bool, error)
	HandlePreStageSuccessEvent(triggerContext TriggerContext, cdStageCompleteEvent bean7.CdStageCompleteEvent) error
	HandleDeploymentSuccessEvent(triggerContext TriggerContext, pipelineOverride *chartConfig.PipelineOverride) error
	HandlePostStageSuccessEvent(triggerContext TriggerContext, cdWorkflowId int, cdPipelineId int, triggeredBy int32, pluginRegistryImageDetails map[string][]string) error

	TriggerPostStage(request TriggerRequest) error
	TriggerPreStage(request TriggerRequest) error
	TriggerDeployment(request TriggerRequest) error
	ManualCdTrigger(triggerContext TriggerContext, overrideRequest *bean.ValuesOverrideRequest) (int, error)
	TriggerStageForBulk(triggerRequest TriggerRequest) error

	TriggerBulkHibernateAsync(request StopDeploymentGroupRequest, ctx context.Context) (interface{}, error)
	StopStartApp(triggerContext TriggerContext, stopRequest *StopAppRequest) (int, error)

	MarkCurrentDeploymentFailed(runner *pipelineConfig.CdWorkflowRunner, releaseErr error, triggeredBy int32) error
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
	argoUserService               argo.ArgoUserService
	cdPipelineStatusTimelineRepo  pipelineConfig.PipelineStatusTimelineRepository
	pipelineStatusTimelineService status.PipelineStatusTimelineService
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
	workflowEventPublishService         out.WorkflowEventPublishService
	manifestCreationService             manifest.ManifestCreationService
	deploymentTemplateService           deploymentTemplate.DeploymentTemplateService
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

type TriggerRequest struct {
	CdWf                  *pipelineConfig.CdWorkflow
	Pipeline              *pipelineConfig.Pipeline
	Artifact              *repository.CiArtifact
	ApplyAuth             bool
	TriggeredBy           int32
	RefCdWorkflowRunnerId int
	TriggerContext
}

type TriggerContext struct {
	// Context is a context object to be passed to the pipeline trigger
	// +optional
	Context context.Context
	// ReferenceId is a unique identifier for the workflow runner
	// refer pipelineConfig.CdWorkflowRunner
	ReferenceId *string
}

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
	argoUserService argo.ArgoUserService,
	cdPipelineStatusTimelineRepo pipelineConfig.PipelineStatusTimelineRepository,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
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
	workflowEventPublishService out.WorkflowEventPublishService,
	manifestCreationService manifest.ManifestCreationService,
	deploymentTemplateService deploymentTemplate.DeploymentTemplateService,
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
		argoUserService:               argoUserService,
		cdPipelineStatusTimelineRepo:  cdPipelineStatusTimelineRepo,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
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
		workflowEventPublishService:         workflowEventPublishService,
		manifestCreationService:             manifestCreationService,
		deploymentTemplateService:           deploymentTemplateService,
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
			if err := impl.MarkCurrentDeploymentFailed(cdWfr, fmt.Errorf("Deployment timeout: release %s took more than %d mins", appIdentifier.ReleaseName, impl.appServiceConfig.DevtronChartInstallRequestTimeout), overrideRequest.UserId); err != nil {
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
		if err := impl.MarkCurrentDeploymentFailed(cdWfr, errors.New(pipelineConfig.NEW_DEPLOYMENT_INITIATED), overrideRequest.UserId); err != nil {
			impl.logger.Errorw("error while updating current runner status to failed, handleAsyncTriggerReleaseError", "cdWfr", cdWfr.Id, "err", err)
		}
		return
	case "":
		return
	default:
		if err := impl.MarkCurrentDeploymentFailed(cdWfr, releaseErr, overrideRequest.UserId); err != nil {
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
		if err = impl.MarkCurrentDeploymentFailed(cdWfr, errors.New("CD pipeline has been deleted"), triggeredBy); err != nil {
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
		err := impl.MarkCurrentDeploymentFailed(cdWfr, errors.New(pipelineConfig.NEW_DEPLOYMENT_INITIATED), overrideRequest.UserId)
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
		err1 := impl.updatePreviousDeploymentStatus(cdWfr, overrideRequest.PipelineId, CDAsyncInstallNatsMessage.TriggeredAt, overrideRequest.UserId)
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

func (impl *WorkflowDagExecutorImpl) HandleCiSuccessEvent(triggerContext TriggerContext, artifact *repository.CiArtifact, async bool, triggeredBy int32) error {
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
		triggerRequest := TriggerRequest{
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
		triggerRequest := TriggerRequest{
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

func (impl *WorkflowDagExecutorImpl) triggerIfAutoStageCdPipeline(request TriggerRequest) error {

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
			err = impl.TriggerPreStage(request) // TODO handle error here
			return err
		}
	} else if request.Pipeline.TriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC {
		// trigger deployment
		impl.logger.Debugw("trigger cd for pipeline", "artifactId", request.Artifact.Id, "pipelineId", request.Pipeline.Id)
		err = impl.TriggerDeployment(request)
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

func (impl *WorkflowDagExecutorImpl) TriggerStageForBulk(triggerRequest TriggerRequest) error {

	preStage, err := impl.getPipelineStage(triggerRequest.Pipeline.Id, repository4.PIPELINE_STAGE_TYPE_PRE_CD)
	if err != nil {
		return err
	}

	//handle corrupt data (https://github.com/devtron-labs/devtron/issues/3826)
	err, deleted := impl.deleteCorruptedPipelineStage(preStage, triggerRequest.TriggeredBy)
	if err != nil {
		impl.logger.Errorw("error in deleteCorruptedPipelineStage ", "cdPipelineId", triggerRequest.Pipeline.Id, "err", err, "preStage", preStage, "triggeredBy", triggerRequest.TriggeredBy)
		return err
	}

	triggerRequest.TriggerContext.Context = context.Background()
	if len(triggerRequest.Pipeline.PreStageConfig) > 0 || (preStage != nil && !deleted) {
		//pre stage exists
		impl.logger.Debugw("trigger pre stage for pipeline", "artifactId", triggerRequest.Artifact.Id, "pipelineId", triggerRequest.Pipeline.Id)
		triggerRequest.RefCdWorkflowRunnerId = 0
		err = impl.TriggerPreStage(triggerRequest) // TODO handle error here
		return err
	} else {
		// trigger deployment
		impl.logger.Debugw("trigger cd for pipeline", "artifactId", triggerRequest.Artifact.Id, "pipelineId", triggerRequest.Pipeline.Id)
		err = impl.TriggerDeployment(triggerRequest)
		return err
	}
}

func (impl *WorkflowDagExecutorImpl) HandlePreStageSuccessEvent(triggerContext TriggerContext, cdStageCompleteEvent bean7.CdStageCompleteEvent) error {
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
			triggerRequest := TriggerRequest{
				CdWf:           cdWorkflow,
				Pipeline:       pipeline,
				Artifact:       ciArtifact,
				ApplyAuth:      applyAuth,
				TriggeredBy:    cdStageCompleteEvent.TriggeredBy,
				TriggerContext: triggerContext,
			}
			triggerRequest.TriggerContext.Context = context.Background()
			err = impl.TriggerDeployment(triggerRequest)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) TriggerPreStage(request TriggerRequest) error {
	//setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()
	triggeredBy := request.TriggeredBy
	artifact := request.Artifact
	pipeline := request.Pipeline
	ctx := request.TriggerContext.Context
	//in case of pre stage manual trigger auth is already applied and for auto triggers there is no need for auth check here
	cdWf := request.CdWf
	var err error
	if cdWf == nil {
		cdWf = &pipelineConfig.CdWorkflow{
			CiArtifactId: artifact.Id,
			PipelineId:   pipeline.Id,
			AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: 1, UpdatedOn: triggeredAt, UpdatedBy: 1},
		}
		err = impl.cdWorkflowRepository.SaveWorkFlow(ctx, cdWf)
		if err != nil {
			return err
		}
	}
	cdWorkflowExecutorType := impl.config.GetWorkflowExecutorType()
	runner := &pipelineConfig.CdWorkflowRunner{
		Name:                  pipeline.Name,
		WorkflowType:          bean.CD_WORKFLOW_TYPE_PRE,
		ExecutorType:          cdWorkflowExecutorType,
		Status:                pipelineConfig.WorkflowStarting, // starting PreStage
		TriggeredBy:           triggeredBy,
		StartedOn:             triggeredAt,
		Namespace:             impl.config.GetDefaultNamespace(),
		BlobStorageEnabled:    impl.config.BlobStorageEnabled,
		CdWorkflowId:          cdWf.Id,
		LogLocation:           fmt.Sprintf("%s/%s%s-%s/main.log", impl.config.GetDefaultBuildLogsKeyPrefix(), strconv.Itoa(cdWf.Id), string(bean.CD_WORKFLOW_TYPE_PRE), pipeline.Name),
		AuditLog:              sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: 1, UpdatedOn: triggeredAt, UpdatedBy: 1},
		RefCdWorkflowRunnerId: request.RefCdWorkflowRunnerId,
		ReferenceId:           request.TriggerContext.ReferenceId,
	}
	var env *repository2.Environment
	if pipeline.RunPreStageInEnv {
		_, span := otel.Tracer("orchestrator").Start(ctx, "envRepository.FindById")
		env, err = impl.envRepository.FindById(pipeline.EnvironmentId)
		span.End()
		if err != nil {
			impl.logger.Errorw(" unable to find env ", "err", err)
			return err
		}
		impl.logger.Debugw("env", "env", env)
		runner.Namespace = env.Namespace
	}
	_, span := otel.Tracer("orchestrator").Start(ctx, "cdWorkflowRepository.SaveWorkFlowRunner")
	_, err = impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
	span.End()
	if err != nil {
		return err
	}

	//checking vulnerability for the selected image
	isVulnerable, err := impl.GetArtifactVulnerabilityStatus(artifact, pipeline, ctx)
	if err != nil {
		impl.logger.Errorw("error in getting Artifact vulnerability status, TriggerPreStage", "err", err)
		return err
	}
	if isVulnerable {
		// if image vulnerable, update timeline status and return
		runner.Status = pipelineConfig.WorkflowFailed
		runner.Message = pipelineConfig.FOUND_VULNERABILITY
		runner.FinishedOn = time.Now()
		runner.UpdatedOn = time.Now()
		runner.UpdatedBy = triggeredBy
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
		if err != nil {
			impl.logger.Errorw("error in updating wfr status due to vulnerable image", "err", err)
			return err
		}
		return fmt.Errorf("found vulnerability for image digest %s", artifact.ImageDigest)
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "buildWFRequest")
	cdStageWorkflowRequest, err := impl.buildWFRequest(runner, cdWf, pipeline, triggeredBy)
	span.End()
	if err != nil {
		return err
	}
	cdStageWorkflowRequest.StageType = types.PRE
	// handling copyContainerImage plugin specific logic
	imagePathReservationIds, err := impl.SetCopyContainerImagePluginDataInWorkflowRequest(cdStageWorkflowRequest, pipeline.Id, types.PRE, artifact)
	if err != nil {
		runner.Status = pipelineConfig.WorkflowFailed
		runner.Message = err.Error()
		_ = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
		return err
	} else {
		runner.ImagePathReservationIds = imagePathReservationIds
		_ = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "cdWorkflowService.SubmitWorkflow")
	cdStageWorkflowRequest.Pipeline = pipeline
	cdStageWorkflowRequest.Env = env
	cdStageWorkflowRequest.Type = bean3.CD_WORKFLOW_PIPELINE_TYPE
	_, err = impl.cdWorkflowService.SubmitWorkflow(cdStageWorkflowRequest)
	span.End()
	err = impl.sendPreStageNotification(ctx, cdWf, pipeline)
	if err != nil {
		return err
	}
	//creating cd config history entry
	_, span = otel.Tracer("orchestrator").Start(ctx, "prePostCdScriptHistoryService.CreatePrePostCdScriptHistory")
	err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, nil, repository3.PRE_CD_TYPE, true, triggeredBy, triggeredAt)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in creating pre cd script entry", "err", err, "pipeline", pipeline)
		return err
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) SetCopyContainerImagePluginDataInWorkflowRequest(cdStageWorkflowRequest *types.WorkflowRequest, pipelineId int, pipelineStage string, artifact *repository.CiArtifact) ([]int, error) {
	copyContainerImagePluginId, err := impl.globalPluginService.GetRefPluginIdByRefPluginName(COPY_CONTAINER_IMAGE)
	var imagePathReservationIds []int
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting copyContainerImage plugin id", "err", err)
		return imagePathReservationIds, err
	}
	for _, step := range cdStageWorkflowRequest.PrePostDeploySteps {
		if copyContainerImagePluginId != 0 && step.RefPluginId == copyContainerImagePluginId {
			var pipelineStageEntityType int
			if pipelineStage == types.PRE {
				pipelineStageEntityType = bean3.EntityTypePreCD
			} else {
				pipelineStageEntityType = bean3.EntityTypePostCD
			}
			customTagId := -1
			var DockerImageTag string

			customTag, err := impl.customTagService.GetActiveCustomTagByEntityKeyAndValue(pipelineStageEntityType, strconv.Itoa(pipelineId))
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error in fetching custom tag data", "err", err)
				return imagePathReservationIds, err
			}

			if !customTag.Enabled {
				// case when custom tag is not configured - source image tag will be taken as docker image tag
				pluginTriggerImageSplit := strings.Split(artifact.Image, ":")
				DockerImageTag = pluginTriggerImageSplit[len(pluginTriggerImageSplit)-1]
			} else {
				// for copyContainerImage plugin parse destination images and save its data in image path reservation table
				customTagDbObject, customDockerImageTag, err := impl.customTagService.GetCustomTag(pipelineStageEntityType, strconv.Itoa(pipelineId))
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in fetching custom tag by entity key and value for CD", "err", err)
					return imagePathReservationIds, err
				}
				if customTagDbObject != nil && customTagDbObject.Id > 0 {
					customTagId = customTagDbObject.Id
				}
				DockerImageTag = customDockerImageTag
			}

			var sourceDockerRegistryId string
			if artifact.DataSource == repository.PRE_CD || artifact.DataSource == repository.POST_CD || artifact.DataSource == repository.POST_CI {
				if artifact.CredentialsSourceType == repository.GLOBAL_CONTAINER_REGISTRY {
					sourceDockerRegistryId = artifact.CredentialSourceValue
				}
			} else {
				sourceDockerRegistryId = cdStageWorkflowRequest.DockerRegistryId
			}
			registryDestinationImageMap, registryCredentialMap, err := impl.pluginInputVariableParser.HandleCopyContainerImagePluginInputVariables(step.InputVars, DockerImageTag, cdStageWorkflowRequest.CiArtifactDTO.Image, sourceDockerRegistryId)
			if err != nil {
				impl.logger.Errorw("error in parsing copyContainerImage input variable", "err", err)
				return imagePathReservationIds, err
			}
			var destinationImages []string
			for _, images := range registryDestinationImageMap {
				for _, image := range images {
					destinationImages = append(destinationImages, image)
				}
			}
			// fetch already saved artifacts to check if they are already present
			savedCIArtifacts, err := impl.ciArtifactRepository.FindCiArtifactByImagePaths(destinationImages)
			if err != nil {
				impl.logger.Errorw("error in fetching artifacts by image path", "err", err)
				return imagePathReservationIds, err
			}
			if len(savedCIArtifacts) > 0 {
				// if already present in ci artifact, return "image path already in use error"
				return imagePathReservationIds, bean3.ErrImagePathInUse
			}
			imagePathReservationIds, err = impl.ReserveImagesGeneratedAtPlugin(customTagId, registryDestinationImageMap)
			if err != nil {
				impl.logger.Errorw("error in reserving image", "err", err)
				return imagePathReservationIds, err
			}
			cdStageWorkflowRequest.RegistryDestinationImageMap = registryDestinationImageMap
			cdStageWorkflowRequest.RegistryCredentialMap = registryCredentialMap
			var pluginArtifactStage string
			if pipelineStage == types.PRE {
				pluginArtifactStage = repository.PRE_CD
			} else {
				pluginArtifactStage = repository.POST_CD
			}
			cdStageWorkflowRequest.PluginArtifactStage = pluginArtifactStage
		}
	}
	return imagePathReservationIds, nil
}

func (impl *WorkflowDagExecutorImpl) sendPreStageNotification(ctx context.Context, cdWf *pipelineConfig.CdWorkflow, pipeline *pipelineConfig.Pipeline) error {
	wfr, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(ctx, cdWf.Id, bean.CD_WORKFLOW_TYPE_PRE)
	if err != nil {
		return err
	}

	event := impl.eventFactory.Build(util2.Trigger, &pipeline.Id, pipeline.AppId, &pipeline.EnvironmentId, util2.CD)
	impl.logger.Debugw("event PreStageTrigger", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, &wfr, 0, bean.CD_WORKFLOW_TYPE_PRE)
	_, span := otel.Tracer("orchestrator").Start(ctx, "eventClient.WriteNotificationEvent")
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	span.End()
	if evtErr != nil {
		impl.logger.Errorw("CD trigger event not sent", "error", evtErr)
	}
	return nil
}

func convert(ts string) (*time.Time, error) {
	//layout := "2006-01-02T15:04:05Z"
	t, err := time.Parse(bean2.LayoutRFC3339, ts)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (impl *WorkflowDagExecutorImpl) TriggerPostStage(request TriggerRequest) error {
	//setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()
	triggeredBy := request.TriggeredBy
	pipeline := request.Pipeline
	cdWf := request.CdWf

	runner := &pipelineConfig.CdWorkflowRunner{
		Name:                  pipeline.Name,
		WorkflowType:          bean.CD_WORKFLOW_TYPE_POST,
		ExecutorType:          impl.config.GetWorkflowExecutorType(),
		Status:                pipelineConfig.WorkflowStarting, // starting PostStage
		TriggeredBy:           triggeredBy,
		StartedOn:             triggeredAt,
		Namespace:             impl.config.GetDefaultNamespace(),
		BlobStorageEnabled:    impl.config.BlobStorageEnabled,
		CdWorkflowId:          cdWf.Id,
		LogLocation:           fmt.Sprintf("%s/%s%s-%s/main.log", impl.config.GetDefaultBuildLogsKeyPrefix(), strconv.Itoa(cdWf.Id), string(bean.CD_WORKFLOW_TYPE_POST), pipeline.Name),
		AuditLog:              sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: triggeredBy, UpdatedOn: triggeredAt, UpdatedBy: triggeredBy},
		RefCdWorkflowRunnerId: request.RefCdWorkflowRunnerId,
		ReferenceId:           request.TriggerContext.ReferenceId,
	}
	var env *repository2.Environment
	var err error
	if pipeline.RunPostStageInEnv {
		env, err = impl.envRepository.FindById(pipeline.EnvironmentId)
		if err != nil {
			impl.logger.Errorw(" unable to find env ", "err", err)
			return err
		}
		runner.Namespace = env.Namespace
	}

	_, err = impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
	if err != nil {
		return err
	}

	if cdWf.CiArtifact == nil || cdWf.CiArtifact.Id == 0 {
		cdWf.CiArtifact, err = impl.ciArtifactRepository.Get(cdWf.CiArtifactId)
		if err != nil {
			impl.logger.Errorw("error fetching artifact data", "err", err)
			return err
		}
	}
	// Migration of deprecated DataSource Type
	if cdWf.CiArtifact.IsMigrationRequired() {
		migrationErr := impl.ciArtifactRepository.MigrateToWebHookDataSourceType(cdWf.CiArtifact.Id)
		if migrationErr != nil {
			impl.logger.Warnw("unable to migrate deprecated DataSource", "artifactId", cdWf.CiArtifact.Id)
		}
	}
	//checking vulnerability for the selected image
	isVulnerable, err := impl.GetArtifactVulnerabilityStatus(cdWf.CiArtifact, pipeline, context.Background())
	if err != nil {
		impl.logger.Errorw("error in getting Artifact vulnerability status, TriggerPostStage", "err", err)
		return err
	}
	if isVulnerable {
		// if image vulnerable, update timeline status and return
		runner.Status = pipelineConfig.WorkflowFailed
		runner.Message = pipelineConfig.FOUND_VULNERABILITY
		runner.FinishedOn = time.Now()
		runner.UpdatedOn = time.Now()
		runner.UpdatedBy = triggeredBy
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
		if err != nil {
			impl.logger.Errorw("error in updating wfr status due to vulnerable image", "err", err)
			return err
		}
		return fmt.Errorf("found vulnerability for image digest %s", cdWf.CiArtifact.ImageDigest)
	}

	cdStageWorkflowRequest, err := impl.buildWFRequest(runner, cdWf, pipeline, triggeredBy)
	if err != nil {
		impl.logger.Errorw("error in building wfRequest", "err", err, "runner", runner, "cdWf", cdWf, "pipeline", pipeline)
		return err
	}
	cdStageWorkflowRequest.StageType = types.POST
	cdStageWorkflowRequest.Pipeline = pipeline
	cdStageWorkflowRequest.Env = env
	cdStageWorkflowRequest.Type = bean3.CD_WORKFLOW_PIPELINE_TYPE
	// handling plugin specific logic

	pluginImagePathReservationIds, err := impl.SetCopyContainerImagePluginDataInWorkflowRequest(cdStageWorkflowRequest, pipeline.Id, types.POST, cdWf.CiArtifact)
	if err != nil {
		runner.Status = pipelineConfig.WorkflowFailed
		runner.Message = err.Error()
		_ = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
		return err
	}

	_, err = impl.cdWorkflowService.SubmitWorkflow(cdStageWorkflowRequest)
	if err != nil {
		impl.logger.Errorw("error in submitting workflow", "err", err, "cdStageWorkflowRequest", cdStageWorkflowRequest, "pipeline", pipeline, "env", env)
		return err
	}

	wfr, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(context.Background(), cdWf.Id, bean.CD_WORKFLOW_TYPE_POST)
	if err != nil {
		impl.logger.Errorw("error in getting wfr by workflowId and runnerType", "err", err, "wfId", cdWf.Id)
		return err
	}
	wfr.ImagePathReservationIds = pluginImagePathReservationIds
	err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(&wfr)
	if err != nil {
		impl.logger.Error("error in updating image path reservation ids in cd workflow runner", "err", "err")
	}

	event := impl.eventFactory.Build(util2.Trigger, &pipeline.Id, pipeline.AppId, &pipeline.EnvironmentId, util2.CD)
	impl.logger.Debugw("event Cd Post Trigger", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, &wfr, 0, bean.CD_WORKFLOW_TYPE_POST)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("CD trigger event not sent", "error", evtErr)
	}
	//creating cd config history entry
	err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, nil, repository3.POST_CD_TYPE, true, triggeredBy, triggeredAt)
	if err != nil {
		impl.logger.Errorw("error in creating post cd script entry", "err", err, "pipeline", pipeline)
		return err
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) ReserveImagesGeneratedAtPlugin(customTagId int, registryImageMap map[string][]string) ([]int, error) {
	var imagePathReservationIds []int
	for _, images := range registryImageMap {
		for _, image := range images {
			imagePathReservationData, err := impl.customTagService.ReserveImagePath(image, customTagId)
			if err != nil {
				impl.logger.Errorw("Error in marking custom tag reserved", "err", err)
				return imagePathReservationIds, err
			}
			if imagePathReservationData != nil {
				imagePathReservationIds = append(imagePathReservationIds, imagePathReservationData.Id)
			}
		}
	}
	return imagePathReservationIds, nil
}

func (impl *WorkflowDagExecutorImpl) buildArtifactLocationForS3(cdWorkflowConfig *pipelineConfig.CdWorkflowConfig, cdWf *pipelineConfig.CdWorkflow, runner *pipelineConfig.CdWorkflowRunner) (string, string, string) {
	cdArtifactLocationFormat := cdWorkflowConfig.CdArtifactLocationFormat
	if cdArtifactLocationFormat == "" {
		cdArtifactLocationFormat = impl.config.GetArtifactLocationFormat()
	}
	if cdWorkflowConfig.LogsBucket == "" {
		cdWorkflowConfig.LogsBucket = impl.config.GetDefaultBuildLogsBucket()
	}
	ArtifactLocation := fmt.Sprintf("s3://%s/"+impl.config.GetDefaultArtifactKeyPrefix()+"/"+cdArtifactLocationFormat, cdWorkflowConfig.LogsBucket, cdWf.Id, runner.Id)
	artifactFileName := fmt.Sprintf(impl.config.GetDefaultArtifactKeyPrefix()+"/"+cdArtifactLocationFormat, cdWf.Id, runner.Id)
	return ArtifactLocation, cdWorkflowConfig.LogsBucket, artifactFileName
}

func (impl *WorkflowDagExecutorImpl) getDeployStageDetails(pipelineId int) (pipelineConfig.CdWorkflowRunner, string, int, error) {
	deployStageWfr := pipelineConfig.CdWorkflowRunner{}
	//getting deployment pipeline latest wfr by pipelineId
	deployStageWfr, err := impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipelineId, bean.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.logger.Errorw("error in getting latest status of deploy type wfr by pipelineId", "err", err, "pipelineId", pipelineId)
		return deployStageWfr, "", 0, err
	}
	deployStageTriggeredByUserEmail, err := impl.user.GetEmailById(deployStageWfr.TriggeredBy)
	if err != nil {
		impl.logger.Errorw("error in getting user email by id", "err", err, "userId", deployStageWfr.TriggeredBy)
		return deployStageWfr, "", 0, err
	}
	pipelineReleaseCounter, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(pipelineId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching latest release counter for pipeline", "pipelineId", pipelineId, "err", err)
		return deployStageWfr, "", 0, err
	}
	return deployStageWfr, deployStageTriggeredByUserEmail, pipelineReleaseCounter, nil
}

func isExtraVariableDynamic(variableName string, webhookAndCiData *gitSensorClient.WebhookAndCiData) bool {
	if strings.Contains(variableName, GIT_COMMIT_HASH_PREFIX) || strings.Contains(variableName, GIT_SOURCE_TYPE_PREFIX) || strings.Contains(variableName, GIT_SOURCE_VALUE_PREFIX) ||
		strings.Contains(variableName, APP_LABEL_VALUE_PREFIX) || strings.Contains(variableName, APP_LABEL_KEY_PREFIX) ||
		strings.Contains(variableName, CHILD_CD_ENV_NAME_PREFIX) || strings.Contains(variableName, CHILD_CD_CLUSTER_NAME_PREFIX) ||
		strings.Contains(variableName, CHILD_CD_COUNT) || strings.Contains(variableName, APP_LABEL_COUNT) || strings.Contains(variableName, GIT_SOURCE_COUNT) ||
		webhookAndCiData != nil {

		return true
	}
	return false
}

func setExtraEnvVariableInDeployStep(deploySteps []*bean3.StepObject, extraEnvVariables map[string]string, webhookAndCiData *gitSensorClient.WebhookAndCiData) {
	for _, deployStep := range deploySteps {
		for variableKey, variableValue := range extraEnvVariables {
			if isExtraVariableDynamic(variableKey, webhookAndCiData) && deployStep.StepType == "INLINE" {
				extraInputVar := &bean3.VariableObject{
					Name:                  variableKey,
					Format:                "STRING",
					Value:                 variableValue,
					VariableType:          bean3.VARIABLE_TYPE_REF_GLOBAL,
					ReferenceVariableName: variableKey,
				}
				deployStep.InputVars = append(deployStep.InputVars, extraInputVar)
			}
		}
	}
}
func (impl *WorkflowDagExecutorImpl) buildWFRequest(runner *pipelineConfig.CdWorkflowRunner, cdWf *pipelineConfig.CdWorkflow, cdPipeline *pipelineConfig.Pipeline, triggeredBy int32) (*types.WorkflowRequest, error) {
	cdWorkflowConfig, err := impl.cdWorkflowRepository.FindConfigByPipelineId(cdPipeline.Id)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}

	workflowExecutor := runner.ExecutorType

	artifact, err := impl.ciArtifactRepository.Get(cdWf.CiArtifactId)
	if err != nil {
		return nil, err
	}
	// Migration of deprecated DataSource Type
	if artifact.IsMigrationRequired() {
		migrationErr := impl.ciArtifactRepository.MigrateToWebHookDataSourceType(artifact.Id)
		if migrationErr != nil {
			impl.logger.Warnw("unable to migrate deprecated DataSource", "artifactId", artifact.Id)
		}
	}
	ciMaterialInfo, err := repository.GetCiMaterialInfo(artifact.MaterialInfo, artifact.DataSource)
	if err != nil {
		impl.logger.Errorw("parsing error", "err", err)
		return nil, err
	}

	var ciProjectDetails []bean3.CiProjectDetails
	var ciPipeline *pipelineConfig.CiPipeline
	if cdPipeline.CiPipelineId > 0 {
		ciPipeline, err = impl.ciPipelineRepository.FindById(cdPipeline.CiPipelineId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("cannot find ciPipelineRequest", "err", err)
			return nil, err
		}

		for _, m := range ciPipeline.CiPipelineMaterials {
			// git material should be active in this case
			if m == nil || m.GitMaterial == nil || !m.GitMaterial.Active {
				continue
			}
			var ciMaterialCurrent repository.CiMaterialInfo
			for _, ciMaterial := range ciMaterialInfo {
				if ciMaterial.Material.GitConfiguration.URL == m.GitMaterial.Url {
					ciMaterialCurrent = ciMaterial
					break
				}
			}
			gitMaterial, err := impl.materialRepository.FindById(m.GitMaterialId)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("could not fetch git materials", "err", err)
				return nil, err
			}

			ciProjectDetail := bean3.CiProjectDetails{
				GitRepository:   ciMaterialCurrent.Material.GitConfiguration.URL,
				MaterialName:    gitMaterial.Name,
				CheckoutPath:    gitMaterial.CheckoutPath,
				FetchSubmodules: gitMaterial.FetchSubmodules,
				SourceType:      m.Type,
				SourceValue:     m.Value,
				Type:            string(m.Type),
				GitOptions: bean3.GitOptions{
					UserName:      gitMaterial.GitProvider.UserName,
					Password:      gitMaterial.GitProvider.Password,
					SshPrivateKey: gitMaterial.GitProvider.SshPrivateKey,
					AccessToken:   gitMaterial.GitProvider.AccessToken,
					AuthMode:      gitMaterial.GitProvider.AuthMode,
				},
			}

			if len(ciMaterialCurrent.Modifications) > 0 {
				ciProjectDetail.CommitHash = ciMaterialCurrent.Modifications[0].Revision
				ciProjectDetail.Author = ciMaterialCurrent.Modifications[0].Author
				ciProjectDetail.GitTag = ciMaterialCurrent.Modifications[0].Tag
				ciProjectDetail.Message = ciMaterialCurrent.Modifications[0].Message
				commitTime, err := convert(ciMaterialCurrent.Modifications[0].ModifiedTime)
				if err != nil {
					return nil, err
				}
				ciProjectDetail.CommitTime = commitTime.Format(bean2.LayoutRFC3339)
			} else if ciPipeline.PipelineType == bean3.CI_JOB {
				// This has been done to resolve unmarshalling issue in ci-runner, in case of no commit time(eg- polling container images)
				ciProjectDetail.CommitTime = time.Time{}.Format(bean2.LayoutRFC3339)
			} else {
				impl.logger.Debugw("devtronbug#1062", ciPipeline.Id, cdPipeline.Id)
				return nil, fmt.Errorf("modifications not found for %d", ciPipeline.Id)
			}

			// set webhook data
			if m.Type == pipelineConfig.SOURCE_TYPE_WEBHOOK && len(ciMaterialCurrent.Modifications) > 0 {
				webhookData := ciMaterialCurrent.Modifications[0].WebhookData
				ciProjectDetail.WebhookData = pipelineConfig.WebhookData{
					Id:              webhookData.Id,
					EventActionType: webhookData.EventActionType,
					Data:            webhookData.Data,
				}
			}

			ciProjectDetails = append(ciProjectDetails, ciProjectDetail)
		}
	}
	var stageYaml string
	var deployStageWfr pipelineConfig.CdWorkflowRunner
	var deployStageTriggeredByUserEmail string
	var pipelineReleaseCounter int
	var preDeploySteps []*bean3.StepObject
	var postDeploySteps []*bean3.StepObject
	var refPluginsData []*bean3.RefPluginObject
	//if pipeline_stage_steps present for pre-CD or post-CD then no need to add stageYaml to cdWorkflowRequest in that
	//case add PreDeploySteps and PostDeploySteps to cdWorkflowRequest, this is done for backward compatibility
	pipelineStage, err := impl.getPipelineStage(cdPipeline.Id, runner.WorkflowType.WorkflowTypeToStageType())
	if err != nil {
		return nil, err
	}
	env, err := impl.envRepository.FindById(cdPipeline.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in getting environment by id", "err", err)
		return nil, err
	}

	//Scope will pick the environment of CD pipeline irrespective of in-cluster mode,
	//since user sees the environment of the CD pipeline
	scope := resourceQualifiers.Scope{
		AppId:     cdPipeline.App.Id,
		EnvId:     env.Id,
		ClusterId: env.ClusterId,
		SystemMetadata: &resourceQualifiers.SystemMetadata{
			EnvironmentName: env.Name,
			ClusterName:     env.Cluster.ClusterName,
			Namespace:       env.Namespace,
			Image:           artifact.Image,
			ImageTag:        util3.GetImageTagFromImage(artifact.Image),
		},
	}
	if pipelineStage != nil {
		var variableSnapshot map[string]string
		if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
			//preDeploySteps, _, refPluginsData, err = impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(cdPipeline.Id, cdStage)
			prePostAndRefPluginResponse, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(cdPipeline.Id, preCdStage, scope)
			if err != nil {
				impl.logger.Errorw("error in getting pre, post & refPlugin steps data for wf request", "err", err, "cdPipelineId", cdPipeline.Id)
				return nil, err
			}
			preDeploySteps = prePostAndRefPluginResponse.PreStageSteps
			refPluginsData = prePostAndRefPluginResponse.RefPluginData
			variableSnapshot = prePostAndRefPluginResponse.VariableSnapshot
		} else if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
			//_, postDeploySteps, refPluginsData, err = impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(cdPipeline.Id, cdStage)
			prePostAndRefPluginResponse, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(cdPipeline.Id, postCdStage, scope)
			if err != nil {
				impl.logger.Errorw("error in getting pre, post & refPlugin steps data for wf request", "err", err, "cdPipelineId", cdPipeline.Id)
				return nil, err
			}
			postDeploySteps = prePostAndRefPluginResponse.PostStageSteps
			refPluginsData = prePostAndRefPluginResponse.RefPluginData
			variableSnapshot = prePostAndRefPluginResponse.VariableSnapshot
			deployStageWfr, deployStageTriggeredByUserEmail, pipelineReleaseCounter, err = impl.getDeployStageDetails(cdPipeline.Id)
			if err != nil {
				impl.logger.Errorw("error in getting deployStageWfr, deployStageTriggeredByUser and pipelineReleaseCounter wf request", "err", err, "cdPipelineId", cdPipeline.Id)
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("unsupported workflow triggerd")
		}

		//Save Scoped VariableSnapshot
		var variableSnapshotHistories = util4.GetBeansPtr(
			repository5.GetSnapshotBean(runner.Id, repository5.HistoryReferenceTypeCDWORKFLOWRUNNER, variableSnapshot))
		if len(variableSnapshotHistories) > 0 {
			err = impl.scopedVariableManager.SaveVariableHistoriesForTrigger(variableSnapshotHistories, runner.TriggeredBy)
			if err != nil {
				impl.logger.Errorf("Not able to save variable snapshot for CD trigger %s %d %s", err, runner.Id, variableSnapshot)
			}
		}
	} else {
		//in this case no plugin script is not present for this cdPipeline hence going with attaching preStage or postStage config
		if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
			stageYaml = cdPipeline.PreStageConfig
		} else if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
			stageYaml = cdPipeline.PostStageConfig
			deployStageWfr, deployStageTriggeredByUserEmail, pipelineReleaseCounter, err = impl.getDeployStageDetails(cdPipeline.Id)
			if err != nil {
				impl.logger.Errorw("error in getting deployStageWfr, deployStageTriggeredByUser and pipelineReleaseCounter wf request", "err", err, "cdPipelineId", cdPipeline.Id)
				return nil, err
			}

		} else {
			return nil, fmt.Errorf("unsupported workflow triggerd")
		}
	}

	digestConfigurationRequest := imageDigestPolicy.DigestPolicyConfigurationRequest{PipelineId: cdPipeline.Id}
	digestPolicyConfigurations, err := impl.imageDigestPolicyService.GetDigestPolicyConfigurations(digestConfigurationRequest)
	if err != nil {
		impl.logger.Errorw("error in checking if isImageDigestPolicyConfiguredForPipeline", "err", err, "pipelineId", cdPipeline.Id)
		return nil, err
	}
	image := artifact.Image
	if digestPolicyConfigurations.UseDigestForTrigger() {
		image = ReplaceImageTagWithDigest(image, artifact.ImageDigest)
	}
	cdStageWorkflowRequest := &types.WorkflowRequest{
		EnvironmentId:         cdPipeline.EnvironmentId,
		AppId:                 cdPipeline.AppId,
		WorkflowId:            cdWf.Id,
		WorkflowRunnerId:      runner.Id,
		WorkflowNamePrefix:    strconv.Itoa(runner.Id) + "-" + runner.Name,
		WorkflowPrefixForLog:  strconv.Itoa(cdWf.Id) + string(runner.WorkflowType) + "-" + runner.Name,
		CdImage:               impl.config.GetDefaultImage(),
		CdPipelineId:          cdWf.PipelineId,
		TriggeredBy:           triggeredBy,
		StageYaml:             stageYaml,
		CiProjectDetails:      ciProjectDetails,
		Namespace:             runner.Namespace,
		ActiveDeadlineSeconds: impl.config.GetDefaultTimeout(),
		CiArtifactDTO: types.CiArtifactDTO{
			Id:           artifact.Id,
			PipelineId:   artifact.PipelineId,
			Image:        artifact.Image,
			ImageDigest:  artifact.ImageDigest,
			MaterialInfo: artifact.MaterialInfo,
			DataSource:   artifact.DataSource,
			WorkflowId:   artifact.WorkflowId,
		},
		OrchestratorHost:  impl.config.OrchestratorHost,
		OrchestratorToken: impl.config.OrchestratorToken,
		CloudProvider:     impl.config.CloudProvider,
		WorkflowExecutor:  workflowExecutor,
		RefPlugins:        refPluginsData,
		Scope:             scope,
	}

	extraEnvVariables := make(map[string]string)
	if env != nil {
		extraEnvVariables[plugin.CD_PIPELINE_ENV_NAME_KEY] = env.Name
		if env.Cluster != nil {
			extraEnvVariables[plugin.CD_PIPELINE_CLUSTER_NAME_KEY] = env.Cluster.ClusterName
		}
	}
	ciWf, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflowByArtifactId(artifact.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting ciWf by artifactId", "err", err, "artifactId", artifact.Id)
		return nil, err
	}
	var webhookAndCiData *gitSensorClient.WebhookAndCiData
	if ciWf != nil && ciWf.GitTriggers != nil {
		i := 1
		var gitCommitEnvVariables []types.GitMetadata

		for ciPipelineMaterialId, gitTrigger := range ciWf.GitTriggers {
			extraEnvVariables[fmt.Sprintf("%s_%d", GIT_COMMIT_HASH_PREFIX, i)] = gitTrigger.Commit
			extraEnvVariables[fmt.Sprintf("%s_%d", GIT_SOURCE_TYPE_PREFIX, i)] = string(gitTrigger.CiConfigureSourceType)
			extraEnvVariables[fmt.Sprintf("%s_%d", GIT_SOURCE_VALUE_PREFIX, i)] = gitTrigger.CiConfigureSourceValue

			gitCommitEnvVariables = append(gitCommitEnvVariables, types.GitMetadata{
				GitCommitHash:  gitTrigger.Commit,
				GitSourceType:  string(gitTrigger.CiConfigureSourceType),
				GitSourceValue: gitTrigger.CiConfigureSourceValue,
			})

			// CODE-BLOCK starts - store extra environment variables if webhook
			if gitTrigger.CiConfigureSourceType == pipelineConfig.SOURCE_TYPE_WEBHOOK {
				webhookDataId := gitTrigger.WebhookData.Id
				if webhookDataId > 0 {
					webhookDataRequest := &gitSensorClient.WebhookDataRequest{
						Id:                   webhookDataId,
						CiPipelineMaterialId: ciPipelineMaterialId,
					}
					webhookAndCiData, err = impl.gitSensorGrpcClient.GetWebhookData(context.Background(), webhookDataRequest)
					if err != nil {
						impl.logger.Errorw("err while getting webhook data from git-sensor", "err", err, "webhookDataRequest", webhookDataRequest)
						return nil, err
					}
					if webhookAndCiData != nil {
						for extEnvVariableKey, extEnvVariableVal := range webhookAndCiData.ExtraEnvironmentVariables {
							extraEnvVariables[extEnvVariableKey] = extEnvVariableVal
						}
					}
				}
			}
			// CODE_BLOCK ends

			i++
		}
		gitMetadata, err := json.Marshal(&gitCommitEnvVariables)
		if err != nil {
			impl.logger.Errorw("err while marshaling git metdata", "err", err)
			return nil, err
		}
		extraEnvVariables[plugin.GIT_METADATA] = string(gitMetadata)

		extraEnvVariables[GIT_SOURCE_COUNT] = strconv.Itoa(len(ciWf.GitTriggers))
	}

	childCdIds, err := impl.appWorkflowRepository.FindChildCDIdsByParentCDPipelineId(cdPipeline.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting child cdPipelineIds by parent cdPipelineId", "err", err, "parent cdPipelineId", cdPipeline.Id)
		return nil, err
	}
	if len(childCdIds) > 0 {
		childPipelines, err := impl.pipelineRepository.FindByIdsIn(childCdIds)
		if err != nil {
			impl.logger.Errorw("error in getting pipelines by ids", "err", err, "ids", childCdIds)
			return nil, err
		}
		var childCdEnvVariables []types.ChildCdMetadata
		for i, childPipeline := range childPipelines {
			extraEnvVariables[fmt.Sprintf("%s_%d", CHILD_CD_ENV_NAME_PREFIX, i+1)] = childPipeline.Environment.Name
			extraEnvVariables[fmt.Sprintf("%s_%d", CHILD_CD_CLUSTER_NAME_PREFIX, i+1)] = childPipeline.Environment.Cluster.ClusterName

			childCdEnvVariables = append(childCdEnvVariables, types.ChildCdMetadata{
				ChildCdEnvName:     childPipeline.Environment.Name,
				ChildCdClusterName: childPipeline.Environment.Cluster.ClusterName,
			})
		}
		childCdEnvVariablesMetadata, err := json.Marshal(&childCdEnvVariables)
		if err != nil {
			impl.logger.Errorw("err while marshaling childCdEnvVariables", "err", err)
			return nil, err
		}
		extraEnvVariables[plugin.CHILD_CD_METADATA] = string(childCdEnvVariablesMetadata)

		extraEnvVariables[CHILD_CD_COUNT] = strconv.Itoa(len(childPipelines))
	}
	if ciPipeline != nil && ciPipeline.Id > 0 {
		extraEnvVariables["APP_NAME"] = ciPipeline.App.AppName
		cdStageWorkflowRequest.DockerUsername = ciPipeline.CiTemplate.DockerRegistry.Username
		cdStageWorkflowRequest.DockerPassword = ciPipeline.CiTemplate.DockerRegistry.Password
		cdStageWorkflowRequest.AwsRegion = ciPipeline.CiTemplate.DockerRegistry.AWSRegion
		cdStageWorkflowRequest.DockerConnection = ciPipeline.CiTemplate.DockerRegistry.Connection
		cdStageWorkflowRequest.DockerCert = ciPipeline.CiTemplate.DockerRegistry.Cert
		cdStageWorkflowRequest.AccessKey = ciPipeline.CiTemplate.DockerRegistry.AWSAccessKeyId
		cdStageWorkflowRequest.SecretKey = ciPipeline.CiTemplate.DockerRegistry.AWSSecretAccessKey
		cdStageWorkflowRequest.DockerRegistryType = string(ciPipeline.CiTemplate.DockerRegistry.RegistryType)
		cdStageWorkflowRequest.DockerRegistryURL = ciPipeline.CiTemplate.DockerRegistry.RegistryURL
		cdStageWorkflowRequest.DockerRegistryId = ciPipeline.CiTemplate.DockerRegistry.Id
		cdStageWorkflowRequest.CiPipelineType = ciPipeline.PipelineType
	} else if cdPipeline.AppId > 0 {
		ciTemplate, err := impl.CiTemplateRepository.FindByAppId(cdPipeline.AppId)
		if err != nil {
			return nil, err
		}
		extraEnvVariables["APP_NAME"] = ciTemplate.App.AppName
		cdStageWorkflowRequest.DockerUsername = ciTemplate.DockerRegistry.Username
		cdStageWorkflowRequest.DockerPassword = ciTemplate.DockerRegistry.Password
		cdStageWorkflowRequest.AwsRegion = ciTemplate.DockerRegistry.AWSRegion
		cdStageWorkflowRequest.DockerConnection = ciTemplate.DockerRegistry.Connection
		cdStageWorkflowRequest.DockerCert = ciTemplate.DockerRegistry.Cert
		cdStageWorkflowRequest.AccessKey = ciTemplate.DockerRegistry.AWSAccessKeyId
		cdStageWorkflowRequest.SecretKey = ciTemplate.DockerRegistry.AWSSecretAccessKey
		cdStageWorkflowRequest.DockerRegistryType = string(ciTemplate.DockerRegistry.RegistryType)
		cdStageWorkflowRequest.DockerRegistryURL = ciTemplate.DockerRegistry.RegistryURL
		appLabels, err := impl.appLabelRepository.FindAllByAppId(cdPipeline.AppId)
		cdStageWorkflowRequest.DockerRegistryId = ciTemplate.DockerRegistry.Id
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting labels by appId", "err", err, "appId", cdPipeline.AppId)
			return nil, err
		}
		var appLabelEnvVariables []types.AppLabelMetadata
		for i, appLabel := range appLabels {
			extraEnvVariables[fmt.Sprintf("%s_%d", APP_LABEL_KEY_PREFIX, i+1)] = appLabel.Key
			extraEnvVariables[fmt.Sprintf("%s_%d", APP_LABEL_VALUE_PREFIX, i+1)] = appLabel.Value
			appLabelEnvVariables = append(appLabelEnvVariables, types.AppLabelMetadata{
				AppLabelKey:   appLabel.Key,
				AppLabelValue: appLabel.Value,
			})
		}
		if len(appLabels) > 0 {
			extraEnvVariables[APP_LABEL_COUNT] = strconv.Itoa(len(appLabels))
			appLabelEnvVariablesMetadata, err := json.Marshal(&appLabelEnvVariables)
			if err != nil {
				impl.logger.Errorw("err while marshaling appLabelEnvVariables", "err", err)
				return nil, err
			}
			extraEnvVariables[plugin.APP_LABEL_METADATA] = string(appLabelEnvVariablesMetadata)

		}
	}
	cdStageWorkflowRequest.ExtraEnvironmentVariables = extraEnvVariables
	cdStageWorkflowRequest.DeploymentTriggerTime = deployStageWfr.StartedOn
	cdStageWorkflowRequest.DeploymentTriggeredBy = deployStageTriggeredByUserEmail

	if pipelineReleaseCounter > 0 {
		cdStageWorkflowRequest.DeploymentReleaseCounter = pipelineReleaseCounter
	}
	if cdWorkflowConfig.CdCacheRegion == "" {
		cdWorkflowConfig.CdCacheRegion = impl.config.GetDefaultCdLogsBucketRegion()
	}

	if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		//populate input variables of steps with extra env variables
		setExtraEnvVariableInDeployStep(preDeploySteps, extraEnvVariables, webhookAndCiData)
		cdStageWorkflowRequest.PrePostDeploySteps = preDeploySteps
	} else if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
		setExtraEnvVariableInDeployStep(postDeploySteps, extraEnvVariables, webhookAndCiData)
		cdStageWorkflowRequest.PrePostDeploySteps = postDeploySteps
	}
	cdStageWorkflowRequest.BlobStorageConfigured = runner.BlobStorageEnabled
	switch cdStageWorkflowRequest.CloudProvider {
	case types.BLOB_STORAGE_S3:
		//No AccessKey is used for uploading artifacts, instead IAM based auth is used
		cdStageWorkflowRequest.CdCacheRegion = cdWorkflowConfig.CdCacheRegion
		cdStageWorkflowRequest.CdCacheLocation = cdWorkflowConfig.CdCacheBucket
		cdStageWorkflowRequest.ArtifactLocation, cdStageWorkflowRequest.CiArtifactBucket, cdStageWorkflowRequest.CiArtifactFileName = impl.buildArtifactLocationForS3(cdWorkflowConfig, cdWf, runner)
		cdStageWorkflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			AccessKey:                  impl.config.BlobStorageS3AccessKey,
			Passkey:                    impl.config.BlobStorageS3SecretKey,
			EndpointUrl:                impl.config.BlobStorageS3Endpoint,
			IsInSecure:                 impl.config.BlobStorageS3EndpointInsecure,
			CiCacheBucketName:          cdWorkflowConfig.CdCacheBucket,
			CiCacheRegion:              cdWorkflowConfig.CdCacheRegion,
			CiCacheBucketVersioning:    impl.config.BlobStorageS3BucketVersioned,
			CiArtifactBucketName:       cdStageWorkflowRequest.CiArtifactBucket,
			CiArtifactRegion:           cdWorkflowConfig.CdCacheRegion,
			CiArtifactBucketVersioning: impl.config.BlobStorageS3BucketVersioned,
			CiLogBucketName:            impl.config.GetDefaultBuildLogsBucket(),
			CiLogRegion:                impl.config.GetDefaultCdLogsBucketRegion(),
			CiLogBucketVersioning:      impl.config.BlobStorageS3BucketVersioned,
		}
	case types.BLOB_STORAGE_GCP:
		cdStageWorkflowRequest.GcpBlobConfig = &blob_storage.GcpBlobConfig{
			CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
			ArtifactBucketName:     impl.config.GetDefaultBuildLogsBucket(),
			LogBucketName:          impl.config.GetDefaultBuildLogsBucket(),
		}
		cdStageWorkflowRequest.ArtifactLocation = impl.buildDefaultArtifactLocation(cdWorkflowConfig, cdWf, runner)
		cdStageWorkflowRequest.CiArtifactFileName = cdStageWorkflowRequest.ArtifactLocation
	case types.BLOB_STORAGE_AZURE:
		cdStageWorkflowRequest.AzureBlobConfig = &blob_storage.AzureBlobConfig{
			Enabled:               true,
			AccountName:           impl.config.AzureAccountName,
			BlobContainerCiCache:  impl.config.AzureBlobContainerCiCache,
			AccountKey:            impl.config.AzureAccountKey,
			BlobContainerCiLog:    impl.config.AzureBlobContainerCiLog,
			BlobContainerArtifact: impl.config.AzureBlobContainerCiLog,
		}
		cdStageWorkflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			EndpointUrl:     impl.config.AzureGatewayUrl,
			IsInSecure:      impl.config.AzureGatewayConnectionInsecure,
			CiLogBucketName: impl.config.AzureBlobContainerCiLog,
			CiLogRegion:     "",
			AccessKey:       impl.config.AzureAccountName,
		}
		cdStageWorkflowRequest.ArtifactLocation = impl.buildDefaultArtifactLocation(cdWorkflowConfig, cdWf, runner)
		cdStageWorkflowRequest.CiArtifactFileName = cdStageWorkflowRequest.ArtifactLocation
	default:
		if impl.config.BlobStorageEnabled {
			return nil, fmt.Errorf("blob storage %s not supported", cdStageWorkflowRequest.CloudProvider)
		}
	}
	cdStageWorkflowRequest.DefaultAddressPoolBaseCidr = impl.config.GetDefaultAddressPoolBaseCidr()
	cdStageWorkflowRequest.DefaultAddressPoolSize = impl.config.GetDefaultAddressPoolSize()
	return cdStageWorkflowRequest, nil
}

func ReplaceImageTagWithDigest(image, digest string) string {
	imageWithoutTag := strings.Split(image, ":")[0]
	imageWithDigest := fmt.Sprintf("%s@%s", imageWithoutTag, digest)
	return imageWithDigest
}

func (impl *WorkflowDagExecutorImpl) buildDefaultArtifactLocation(cdWorkflowConfig *pipelineConfig.CdWorkflowConfig, savedWf *pipelineConfig.CdWorkflow, runner *pipelineConfig.CdWorkflowRunner) string {
	cdArtifactLocationFormat := cdWorkflowConfig.CdArtifactLocationFormat
	if cdArtifactLocationFormat == "" {
		cdArtifactLocationFormat = impl.config.GetArtifactLocationFormat()
	}
	ArtifactLocation := fmt.Sprintf("%s/"+cdArtifactLocationFormat, impl.config.GetDefaultArtifactKeyPrefix(), savedWf.Id, runner.Id)
	return ArtifactLocation
}

func (impl *WorkflowDagExecutorImpl) HandleDeploymentSuccessEvent(triggerContext TriggerContext, pipelineOverride *chartConfig.PipelineOverride) error {
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

			triggerRequest := TriggerRequest{
				CdWf:                  cdWorkflow,
				Pipeline:              pipelineOverride.Pipeline,
				TriggeredBy:           triggeredByUser,
				TriggerContext:        triggerContext,
				RefCdWorkflowRunnerId: 0,
			}
			triggerRequest.TriggerContext.Context = context.Background()
			err = impl.TriggerPostStage(triggerRequest)
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

func (impl *WorkflowDagExecutorImpl) HandlePostStageSuccessEvent(triggerContext TriggerContext, cdWorkflowId int, cdPipelineId int, triggeredBy int32, pluginRegistryImageDetails map[string][]string) error {
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

		triggerRequest := TriggerRequest{
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

// Only used for auto trigger
func (impl *WorkflowDagExecutorImpl) TriggerDeployment(request TriggerRequest) error {
	//in case of manual trigger auth is already applied and for auto triggers there is no need for auth check here
	triggeredBy := request.TriggeredBy
	pipeline := request.Pipeline
	artifact := request.Artifact

	//setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()
	cdWf := request.CdWf

	if cdWf == nil || (cdWf != nil && cdWf.CiArtifactId != artifact.Id) {
		// cdWf != nil && cdWf.CiArtifactId != artifact.Id for auto trigger case when deployment is triggered with image generated by plugin
		cdWf = &pipelineConfig.CdWorkflow{
			CiArtifactId: artifact.Id,
			PipelineId:   pipeline.Id,
			AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: 1, UpdatedOn: triggeredAt, UpdatedBy: 1},
		}
		err := impl.cdWorkflowRepository.SaveWorkFlow(context.Background(), cdWf)
		if err != nil {
			return err
		}
	}

	runner := &pipelineConfig.CdWorkflowRunner{
		Name:         pipeline.Name,
		WorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
		ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM,
		Status:       pipelineConfig.WorkflowInitiated, // deployment Initiated for auto trigger
		TriggeredBy:  1,
		StartedOn:    triggeredAt,
		Namespace:    impl.config.GetDefaultNamespace(),
		CdWorkflowId: cdWf.Id,
		AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: triggeredBy, UpdatedOn: triggeredAt, UpdatedBy: triggeredBy},
		ReferenceId:  request.TriggerContext.ReferenceId,
	}
	savedWfr, err := impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
	if err != nil {
		return err
	}
	runner.CdWorkflow = &pipelineConfig.CdWorkflow{
		Pipeline: pipeline,
	}
	// creating cd pipeline status timeline for deployment initialisation
	timeline := &pipelineConfig.PipelineStatusTimeline{
		CdWorkflowRunnerId: runner.Id,
		Status:             pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED,
		StatusDetail:       "Deployment initiated successfully.",
		StatusTime:         time.Now(),
		AuditLog: sql.AuditLog{
			CreatedBy: 1,
			CreatedOn: time.Now(),
			UpdatedBy: 1,
			UpdatedOn: time.Now(),
		},
	}
	isAppStore := false
	err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, isAppStore)
	if err != nil {
		impl.logger.Errorw("error in creating timeline status for deployment initiation", "err", err, "timeline", timeline)
	}
	//checking vulnerability for deploying image
	isVulnerable := false
	if len(artifact.ImageDigest) > 0 {
		var cveStores []*security.CveStore
		imageScanResult, err := impl.scanResultRepository.FindByImageDigest(artifact.ImageDigest)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error fetching image digest", "digest", artifact.ImageDigest, "err", err)
			return err
		}
		for _, item := range imageScanResult {
			cveStores = append(cveStores, &item.CveStore)
		}
		env, err := impl.envRepository.FindById(pipeline.EnvironmentId)
		if err != nil {
			impl.logger.Errorw("error while fetching env", "err", err)
			return err
		}
		blockCveList, err := impl.cvePolicyRepository.GetBlockedCVEList(cveStores, env.ClusterId, pipeline.EnvironmentId, pipeline.AppId, false)
		if err != nil {
			impl.logger.Errorw("error while fetching blocked cve list", "err", err)
			return err
		}
		if len(blockCveList) > 0 {
			isVulnerable = true
		}
	}
	if isVulnerable == true {
		if err = impl.MarkCurrentDeploymentFailed(runner, errors.New(pipelineConfig.FOUND_VULNERABILITY), triggeredBy); err != nil {
			impl.logger.Errorw("error while updating current runner status to failed, TriggerDeployment", "wfrId", runner.Id, "err", err)
		}
		return nil
	}

	releaseErr := impl.TriggerCD(artifact, cdWf.Id, savedWfr.Id, pipeline, triggeredAt)
	// if releaseErr found, then the mark current deployment Failed and return
	if releaseErr != nil {
		err := impl.MarkCurrentDeploymentFailed(runner, releaseErr, triggeredBy)
		if err != nil {
			impl.logger.Errorw("error while updating current runner status to failed, updatePreviousDeploymentStatus", "cdWfr", runner.Id, "err", err)
		}
		return releaseErr
	}
	//skip updatePreviousDeploymentStatus if Async Install is enabled; handled inside SubscribeDevtronAsyncHelmInstallRequest
	if !impl.cdTriggerService.IsDevtronAsyncInstallModeEnabled(pipeline.DeploymentAppType) {
		err1 := impl.updatePreviousDeploymentStatus(runner, pipeline.Id, triggeredAt, triggeredBy)
		if err1 != nil {
			impl.logger.Errorw("error while update previous cd workflow runners", "err", err, "runner", runner, "pipelineId", pipeline.Id)
			return err1
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) updatePreviousDeploymentStatus(currentRunner *pipelineConfig.CdWorkflowRunner, pipelineId int, triggeredAt time.Time, triggeredBy int32) error {
	// Initiating DB transaction
	dbConnection := impl.cdWorkflowRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error on update status, txn begin failed", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	//update [n,n-1] statuses as failed if not terminal
	terminalStatus := []string{string(health.HealthStatusHealthy), pipelineConfig.WorkflowAborted, pipelineConfig.WorkflowFailed, pipelineConfig.WorkflowSucceeded}
	previousNonTerminalRunners, err := impl.cdWorkflowRepository.FindPreviousCdWfRunnerByStatus(pipelineId, currentRunner.Id, terminalStatus)
	if err != nil {
		impl.logger.Errorw("error fetching previous wf runner, updating cd wf runner status,", "err", err, "currentRunner", currentRunner)
		return err
	} else if len(previousNonTerminalRunners) == 0 {
		impl.logger.Errorw("no previous runner found in updating cd wf runner status,", "err", err, "currentRunner", currentRunner)
		return nil
	}

	var timelines []*pipelineConfig.PipelineStatusTimeline
	for _, previousRunner := range previousNonTerminalRunners {
		if previousRunner.Status == string(health.HealthStatusHealthy) ||
			previousRunner.Status == pipelineConfig.WorkflowSucceeded ||
			previousRunner.Status == pipelineConfig.WorkflowAborted ||
			previousRunner.Status == pipelineConfig.WorkflowFailed {
			//terminal status return
			impl.logger.Infow("skip updating cd wf runner status as previous runner status is", "status", previousRunner.Status)
			continue
		}
		impl.logger.Infow("updating cd wf runner status as previous runner status is", "status", previousRunner.Status)
		previousRunner.FinishedOn = triggeredAt
		previousRunner.Message = pipelineConfig.NEW_DEPLOYMENT_INITIATED
		previousRunner.Status = pipelineConfig.WorkflowFailed
		previousRunner.UpdatedOn = time.Now()
		previousRunner.UpdatedBy = triggeredBy
		timeline := &pipelineConfig.PipelineStatusTimeline{
			CdWorkflowRunnerId: previousRunner.Id,
			Status:             pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_SUPERSEDED,
			StatusDetail:       "This deployment is superseded.",
			StatusTime:         time.Now(),
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: time.Now(),
				UpdatedBy: 1,
				UpdatedOn: time.Now(),
			},
		}
		timelines = append(timelines, timeline)
	}

	err = impl.cdWorkflowRepository.UpdateWorkFlowRunners(previousNonTerminalRunners)
	if err != nil {
		impl.logger.Errorw("error updating cd wf runner status", "err", err, "previousNonTerminalRunners", previousNonTerminalRunners)
		return err
	}
	err = impl.cdPipelineStatusTimelineRepo.SaveTimelinesWithTxn(timelines, tx)
	if err != nil {
		impl.logger.Errorw("error updating pipeline status timelines", "err", err, "timelines", timelines)
		return err
	}
	//commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in db transaction commit", "err", err)
		return err
	}
	return nil

}

type RequestType string

const START RequestType = "START"
const STOP RequestType = "STOP"

type StopAppRequest struct {
	AppId         int         `json:"appId" validate:"required"`
	EnvironmentId int         `json:"environmentId" validate:"required"`
	UserId        int32       `json:"userId"`
	RequestType   RequestType `json:"requestType" validate:"oneof=START STOP"`
}

type StopDeploymentGroupRequest struct {
	DeploymentGroupId int         `json:"deploymentGroupId" validate:"required"`
	UserId            int32       `json:"userId"`
	RequestType       RequestType `json:"requestType" validate:"oneof=START STOP"`
}

func (impl *WorkflowDagExecutorImpl) StopStartApp(triggerContext TriggerContext, stopRequest *StopAppRequest) (int, error) {
	pipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(stopRequest.AppId, stopRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline", "app", stopRequest.AppId, "env", stopRequest.EnvironmentId, "err", err)
		return 0, err
	}
	if len(pipelines) == 0 {
		return 0, fmt.Errorf("no pipeline found")
	}
	pipeline := pipelines[0]

	//find pipeline with default
	var pipelineIds []int
	for _, p := range pipelines {
		impl.logger.Debugw("adding pipelineId", "pipelineId", p.Id)
		pipelineIds = append(pipelineIds, p.Id)
		//FIXME
	}
	wf, err := impl.cdWorkflowRepository.FindLatestCdWorkflowByPipelineId(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error in fetching latest release", "err", err)
		return 0, err
	}
	stopTemplate := `{"replicaCount":0,"autoscaling":{"MinReplicas":0,"MaxReplicas":0 ,"enabled": false} }`
	overrideRequest := &bean.ValuesOverrideRequest{
		PipelineId:     pipeline.Id,
		AppId:          stopRequest.AppId,
		CiArtifactId:   wf.CiArtifactId,
		UserId:         stopRequest.UserId,
		CdWorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
	}
	if stopRequest.RequestType == STOP {
		overrideRequest.AdditionalOverride = json.RawMessage([]byte(stopTemplate))
		overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_STOP
	} else if stopRequest.RequestType == START {
		overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_START
	} else {
		return 0, fmt.Errorf("unsupported operation %s", stopRequest.RequestType)
	}
	id, err := impl.ManualCdTrigger(triggerContext, overrideRequest)
	if err != nil {
		impl.logger.Errorw("error in stopping app", "err", err, "appId", stopRequest.AppId, "envId", stopRequest.EnvironmentId)
		return 0, err
	}
	return id, err
}

func (impl *WorkflowDagExecutorImpl) GetArtifactVulnerabilityStatus(artifact *repository.CiArtifact, cdPipeline *pipelineConfig.Pipeline, ctx context.Context) (bool, error) {
	isVulnerable := false
	if len(artifact.ImageDigest) > 0 {
		var cveStores []*security.CveStore
		_, span := otel.Tracer("orchestrator").Start(ctx, "scanResultRepository.FindByImageDigest")
		imageScanResult, err := impl.scanResultRepository.FindByImageDigest(artifact.ImageDigest)
		span.End()
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error fetching image digest", "digest", artifact.ImageDigest, "err", err)
			return false, err
		}
		for _, item := range imageScanResult {
			cveStores = append(cveStores, &item.CveStore)
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "cvePolicyRepository.GetBlockedCVEList")
		if cdPipeline.Environment.ClusterId == 0 {
			envDetails, err := impl.envRepository.FindById(cdPipeline.EnvironmentId)
			if err != nil {
				impl.logger.Errorw("error fetching cluster details by env, GetArtifactVulnerabilityStatus", "envId", cdPipeline.EnvironmentId, "err", err)
				return false, err
			}
			cdPipeline.Environment = *envDetails
		}
		blockCveList, err := impl.cvePolicyRepository.GetBlockedCVEList(cveStores, cdPipeline.Environment.ClusterId, cdPipeline.EnvironmentId, cdPipeline.AppId, false)
		span.End()
		if err != nil {
			impl.logger.Errorw("error while fetching env", "err", err)
			return false, err
		}
		if len(blockCveList) > 0 {
			isVulnerable = true
		}
	}
	return isVulnerable, nil
}

func (impl *WorkflowDagExecutorImpl) ManualCdTrigger(triggerContext TriggerContext, overrideRequest *bean.ValuesOverrideRequest) (int, error) {
	//setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()
	releaseId := 0
	ctx := triggerContext.Context
	var err error
	_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineRepository.FindById")
	cdPipeline, err := impl.pipelineRepository.FindById(overrideRequest.PipelineId)
	span.End()
	if err != nil {
		impl.logger.Errorw("manual trigger request with invalid pipelineId, ManualCdTrigger", "pipelineId", overrideRequest.PipelineId, "err", err)
		return 0, err
	}
	impl.SetPipelineFieldsInOverrideRequest(overrideRequest, cdPipeline)

	switch overrideRequest.CdWorkflowType {
	case bean.CD_WORKFLOW_TYPE_PRE:
		_, span = otel.Tracer("orchestrator").Start(ctx, "ciArtifactRepository.Get")
		artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting CiArtifact", "CiArtifactId", overrideRequest.CiArtifactId, "err", err)
			return 0, err
		}
		// Migration of deprecated DataSource Type
		if artifact.IsMigrationRequired() {
			migrationErr := impl.ciArtifactRepository.MigrateToWebHookDataSourceType(artifact.Id)
			if migrationErr != nil {
				impl.logger.Warnw("unable to migrate deprecated DataSource", "artifactId", artifact.Id)
			}
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "TriggerPreStage")
		triggerRequest := TriggerRequest{
			CdWf:                  nil,
			Artifact:              artifact,
			Pipeline:              cdPipeline,
			TriggeredBy:           overrideRequest.UserId,
			ApplyAuth:             false,
			TriggerContext:        triggerContext,
			RefCdWorkflowRunnerId: 0,
		}
		err = impl.TriggerPreStage(triggerRequest)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in TriggerPreStage, ManualCdTrigger", "err", err)
			return 0, err
		}
	case bean.CD_WORKFLOW_TYPE_DEPLOY:
		if overrideRequest.DeploymentType == models.DEPLOYMENTTYPE_UNKNOWN {
			overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_DEPLOY
		}

		cdWf, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(ctx, overrideRequest.CdWorkflowId, bean.CD_WORKFLOW_TYPE_PRE)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in getting cdWorkflow, ManualCdTrigger", "CdWorkflowId", overrideRequest.CdWorkflowId, "err", err)
			return 0, err
		}

		cdWorkflowId := cdWf.CdWorkflowId
		if cdWf.CdWorkflowId == 0 {
			cdWf := &pipelineConfig.CdWorkflow{
				CiArtifactId: overrideRequest.CiArtifactId,
				PipelineId:   overrideRequest.PipelineId,
				AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
			}
			err := impl.cdWorkflowRepository.SaveWorkFlow(ctx, cdWf)
			if err != nil {
				impl.logger.Errorw("error in creating cdWorkflow, ManualCdTrigger", "PipelineId", overrideRequest.PipelineId, "err", err)
				return 0, err
			}
			cdWorkflowId = cdWf.Id
		}

		runner := &pipelineConfig.CdWorkflowRunner{
			Name:         cdPipeline.Name,
			WorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
			ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
			Status:       pipelineConfig.WorkflowInitiated, //deployment Initiated for manual trigger
			TriggeredBy:  overrideRequest.UserId,
			StartedOn:    triggeredAt,
			Namespace:    impl.config.GetDefaultNamespace(),
			CdWorkflowId: cdWorkflowId,
			AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
			ReferenceId:  triggerContext.ReferenceId,
		}
		savedWfr, err := impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
		overrideRequest.WfrId = savedWfr.Id
		if err != nil {
			impl.logger.Errorw("err in creating cdWorkflowRunner, ManualCdTrigger", "cdWorkflowId", cdWorkflowId, "err", err)
			return 0, err
		}
		runner.CdWorkflow = &pipelineConfig.CdWorkflow{
			Pipeline: cdPipeline,
		}
		overrideRequest.CdWorkflowId = cdWorkflowId
		// creating cd pipeline status timeline for deployment initialisation
		timeline := impl.pipelineStatusTimelineService.GetTimelineDbObjectByTimelineStatusAndTimelineDescription(savedWfr.Id, 0, pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED, pipelineConfig.TIMELINE_DESCRIPTION_DEPLOYMENT_INITIATED, overrideRequest.UserId, time.Now())
		_, span = otel.Tracer("orchestrator").Start(ctx, "cdPipelineStatusTimelineRepo.SaveTimelineForACDHelmApps")
		err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)

		span.End()
		if err != nil {
			impl.logger.Errorw("error in creating timeline status for deployment initiation, ManualCdTrigger", "err", err, "timeline", timeline)
		}

		//checking vulnerability for deploying image
		_, span = otel.Tracer("orchestrator").Start(ctx, "ciArtifactRepository.Get")
		artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting ciArtifact, ManualCdTrigger", "CiArtifactId", overrideRequest.CiArtifactId, "err", err)
			return 0, err
		}
		// Migration of deprecated DataSource Type
		if artifact.IsMigrationRequired() {
			migrationErr := impl.ciArtifactRepository.MigrateToWebHookDataSourceType(artifact.Id)
			if migrationErr != nil {
				impl.logger.Warnw("unable to migrate deprecated DataSource", "artifactId", artifact.Id)
			}
		}
		isVulnerable, err := impl.GetArtifactVulnerabilityStatus(artifact, cdPipeline, ctx)
		if err != nil {
			impl.logger.Errorw("error in getting Artifact vulnerability status, ManualCdTrigger", "err", err)
			return 0, err
		}

		if isVulnerable == true {
			// if image vulnerable, update timeline status and return
			if err = impl.MarkCurrentDeploymentFailed(runner, errors.New(pipelineConfig.FOUND_VULNERABILITY), overrideRequest.UserId); err != nil {
				impl.logger.Errorw("error while updating current runner status to failed, TriggerDeployment", "wfrId", runner.Id, "err", err)
			}
			return 0, fmt.Errorf("found vulnerability for image digest %s", artifact.ImageDigest)
		}

		// Deploy the release
		_, span = otel.Tracer("orchestrator").Start(ctx, "appService.TriggerRelease")
		var releaseErr error
		releaseId, _, releaseErr = impl.HandleCDTriggerRelease(overrideRequest, ctx, triggeredAt, overrideRequest.UserId)
		span.End()
		// if releaseErr found, then the mark current deployment Failed and return
		if releaseErr != nil {
			err := impl.MarkCurrentDeploymentFailed(runner, releaseErr, overrideRequest.UserId)
			if err != nil {
				impl.logger.Errorw("error while updating current runner status to failed, updatePreviousDeploymentStatus", "cdWfr", runner.Id, "err", err)
			}
			return 0, releaseErr
		}

		// skip updatePreviousDeploymentStatus if Async Install is enabled; handled inside SubscribeDevtronAsyncHelmInstallRequest
		if !impl.cdTriggerService.IsDevtronAsyncInstallModeEnabled(cdPipeline.DeploymentAppType) {
			// Update previous deployment runner status (in transaction): Failed
			_, span = otel.Tracer("orchestrator").Start(ctx, "updatePreviousDeploymentStatus")
			err1 := impl.updatePreviousDeploymentStatus(runner, cdPipeline.Id, triggeredAt, overrideRequest.UserId)
			span.End()
			if err1 != nil {
				impl.logger.Errorw("error while update previous cd workflow runners, ManualCdTrigger", "err", err, "runner", runner, "pipelineId", cdPipeline.Id)
				return 0, err1
			}
		}

		if overrideRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_MANIFEST_DOWNLOAD {
			runner := &pipelineConfig.CdWorkflowRunner{
				Id:           runner.Id,
				Name:         cdPipeline.Name,
				WorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
				ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
				TriggeredBy:  overrideRequest.UserId,
				StartedOn:    triggeredAt,
				Status:       pipelineConfig.WorkflowSucceeded,
				Namespace:    impl.config.GetDefaultNamespace(),
				CdWorkflowId: overrideRequest.CdWorkflowId,
				AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
			}
			updateErr := impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
			if updateErr != nil {
				impl.logger.Errorw("error in updating runner for manifest_download type, ManualCdTrigger", "CdWorkflowId", overrideRequest.CdWorkflowId, "err", err)
				return 0, updateErr
			}
		}

	case bean.CD_WORKFLOW_TYPE_POST:
		cdWfRunner, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(ctx, overrideRequest.CdWorkflowId, bean.CD_WORKFLOW_TYPE_DEPLOY)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("err in getting cdWorkflowRunner, ManualCdTrigger", "cdWorkflowId", overrideRequest.CdWorkflowId, "err", err)
			return 0, err
		}

		var cdWf *pipelineConfig.CdWorkflow
		if cdWfRunner.CdWorkflowId == 0 {
			cdWf = &pipelineConfig.CdWorkflow{
				CiArtifactId: overrideRequest.CiArtifactId,
				PipelineId:   overrideRequest.PipelineId,
				AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
			}
			err := impl.cdWorkflowRepository.SaveWorkFlow(ctx, cdWf)
			if err != nil {
				impl.logger.Errorw("error in creating cdWorkflow, ManualCdTrigger", "CdWorkflowId", overrideRequest.CdWorkflowId, "err", err)
				return 0, err
			}
		} else {
			_, span = otel.Tracer("orchestrator").Start(ctx, "cdWorkflowRepository.FindById")
			cdWf, err = impl.cdWorkflowRepository.FindById(overrideRequest.CdWorkflowId)
			span.End()
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("error in getting cdWorkflow, ManualCdTrigger", "CdWorkflowId", overrideRequest.CdWorkflowId, "err", err)
				return 0, err
			}
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "TriggerPostStage")
		triggerRequest := TriggerRequest{
			CdWf:                  cdWf,
			Pipeline:              cdPipeline,
			TriggeredBy:           overrideRequest.UserId,
			RefCdWorkflowRunnerId: 0,
			TriggerContext:        triggerContext,
		}
		err = impl.TriggerPostStage(triggerRequest)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in TriggerPostStage, ManualCdTrigger", "CdWorkflowId", cdWf.Id, "err", err)
			return 0, err
		}
	default:
		impl.logger.Errorw("invalid CdWorkflowType, ManualCdTrigger", "CdWorkflowType", overrideRequest.CdWorkflowType, "err", err)
		return 0, fmt.Errorf("invalid CdWorkflowType %s for the trigger request", string(overrideRequest.CdWorkflowType))
	}

	return releaseId, err
}

type DeploymentGroupAppWithEnv struct {
	EnvironmentId     int         `json:"environmentId"`
	DeploymentGroupId int         `json:"deploymentGroupId"`
	AppId             int         `json:"appId"`
	Active            bool        `json:"active"`
	UserId            int32       `json:"userId"`
	RequestType       RequestType `json:"requestType" validate:"oneof=START STOP"`
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

func extractTimelineFailedStatusDetails(err error) string {
	errorString := util.GetGRPCErrorDetailedMessage(err)
	switch errorString {
	case pipelineConfig.FOUND_VULNERABILITY:
		return pipelineConfig.TIMELINE_DESCRIPTION_VULNERABLE_IMAGE
	default:
		return util.GetTruncatedMessage(fmt.Sprintf("Deployment failed: %s", errorString), 255)
	}
}

func (impl *WorkflowDagExecutorImpl) MarkCurrentDeploymentFailed(runner *pipelineConfig.CdWorkflowRunner, releaseErr error, triggeredBy int32) error {
	err := impl.pipelineStatusTimelineService.MarkPipelineStatusTimelineFailed(runner.Id, extractTimelineFailedStatusDetails(releaseErr))
	if err != nil {
		impl.logger.Errorw("error updating CdPipelineStatusTimeline", "err", err, "releaseErr", releaseErr)
		return err
	}
	//update current WF with error status
	impl.logger.Errorw("error in triggering cd WF, setting wf status as fail ", "wfId", runner.Id, "err", releaseErr)
	runner.Status = pipelineConfig.WorkflowFailed
	runner.Message = util.GetGRPCErrorDetailedMessage(releaseErr)
	runner.FinishedOn = time.Now()
	runner.UpdatedOn = time.Now()
	runner.UpdatedBy = triggeredBy
	err1 := impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
	if err1 != nil {
		impl.logger.Errorw("error updating cd wf runner status", "err", releaseErr, "currentRunner", runner)
		return err1
	}
	util4.TriggerCDMetrics(pipelineConfig.GetTriggerMetricsFromRunnerObj(runner), impl.config.ExposeCDMetrics)
	return nil
}

func (impl *WorkflowDagExecutorImpl) HandleCDTriggerRelease(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context, triggeredAt time.Time, deployedBy int32) (releaseNo int, manifest []byte, err error) {
	if impl.cdTriggerService.IsDevtronAsyncInstallModeEnabled(overrideRequest.DeploymentAppType) {
		// asynchronous mode of installation starts
		return impl.workflowEventPublishService.TriggerHelmAsyncRelease(overrideRequest, ctx, triggeredAt, deployedBy)
	}
	// synchronous mode of installation starts

	valuesOverrideResponse, builtChartPath, err := impl.manifestCreationService.BuildManifestForTrigger(overrideRequest, triggeredAt, ctx)
	_, span := otel.Tracer("orchestrator").Start(ctx, "CreateHistoriesForDeploymentTrigger")
	err1 := impl.deployedConfigurationHistoryService.CreateHistoriesForDeploymentTrigger(valuesOverrideResponse.Pipeline, valuesOverrideResponse.PipelineStrategy, valuesOverrideResponse.EnvOverride, triggeredAt, deployedBy)
	if err1 != nil {
		impl.logger.Errorw("error in saving histories for trigger", "err", err1, "pipelineId", valuesOverrideResponse.Pipeline.Id, "wfrId", overrideRequest.WfrId)
	}
	span.End()
	if err != nil {
		impl.logger.Errorw("error in building merged manifest for trigger", "err", err)
		return releaseNo, manifest, err
	}
	return impl.cdTriggerService.TriggerRelease(overrideRequest, valuesOverrideResponse, builtChartPath, ctx, triggeredAt, deployedBy)
}

func (impl *WorkflowDagExecutorImpl) TriggerCD(artifact *repository.CiArtifact, cdWorkflowId, wfrId int, pipeline *pipelineConfig.Pipeline, triggeredAt time.Time) error {
	impl.logger.Debugw("automatic pipeline trigger attempt async", "artifactId", artifact.Id)

	return impl.triggerReleaseAsync(artifact, cdWorkflowId, wfrId, pipeline, triggeredAt)
}

func (impl *WorkflowDagExecutorImpl) triggerReleaseAsync(artifact *repository.CiArtifact, cdWorkflowId, wfrId int, pipeline *pipelineConfig.Pipeline, triggeredAt time.Time) error {
	err := impl.validateAndTrigger(pipeline, artifact, cdWorkflowId, wfrId, triggeredAt)
	if err != nil {
		impl.logger.Errorw("error in trigger for pipeline", "pipelineId", strconv.Itoa(pipeline.Id))
	}
	impl.logger.Debugw("trigger attempted for all pipeline ", "artifactId", artifact.Id)
	return err
}

func (impl *WorkflowDagExecutorImpl) validateAndTrigger(p *pipelineConfig.Pipeline, artifact *repository.CiArtifact, cdWorkflowId, wfrId int, triggeredAt time.Time) error {
	object := impl.enforcerUtil.GetAppRBACNameByAppId(p.AppId)
	envApp := strings.Split(object, "/")
	if len(envApp) != 2 {
		impl.logger.Error("invalid req, app and env not found from rbac")
		return errors.New("invalid req, app and env not found from rbac")
	}
	err := impl.releasePipeline(p, artifact, cdWorkflowId, wfrId, triggeredAt)
	return err
}

func (impl *WorkflowDagExecutorImpl) releasePipeline(pipeline *pipelineConfig.Pipeline, artifact *repository.CiArtifact, cdWorkflowId, wfrId int, triggeredAt time.Time) error {
	impl.logger.Debugw("triggering release for ", "cdPipelineId", pipeline.Id, "artifactId", artifact.Id)

	pipeline, err := impl.pipelineRepository.FindById(pipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline by pipelineId", "err", err)
		return err
	}

	request := &bean.ValuesOverrideRequest{
		PipelineId:           pipeline.Id,
		UserId:               artifact.CreatedBy,
		CiArtifactId:         artifact.Id,
		AppId:                pipeline.AppId,
		CdWorkflowId:         cdWorkflowId,
		ForceTrigger:         true,
		DeploymentWithConfig: bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED,
		WfrId:                wfrId,
	}
	impl.SetPipelineFieldsInOverrideRequest(request, pipeline)

	ctx, err := impl.argoUserService.BuildACDContext()
	if err != nil {
		impl.logger.Errorw("error in creating acd sync context", "pipelineId", pipeline.Id, "artifactId", artifact.Id, "err", err)
		return err
	}
	//setting deployedBy as 1(system user) since case of auto trigger
	id, _, err := impl.HandleCDTriggerRelease(request, ctx, triggeredAt, 1)
	if err != nil {
		impl.logger.Errorw("error in auto  cd pipeline trigger", "pipelineId", pipeline.Id, "artifactId", artifact.Id, "err", err)
	} else {
		impl.logger.Infow("pipeline successfully triggered ", "cdPipelineId", pipeline.Id, "artifactId", artifact.Id, "releaseId", id)
	}
	return err

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

func (impl *WorkflowDagExecutorImpl) updatePipeline(pipeline *pipelineConfig.Pipeline, userId int32) (bool, error) {
	err := impl.pipelineRepository.SetDeploymentAppCreatedInPipeline(true, pipeline.Id, userId)
	if err != nil {
		impl.logger.Errorw("error on updating cd pipeline for setting deployment app created", "err", err)
		return false, err
	}
	return true, nil
}

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
