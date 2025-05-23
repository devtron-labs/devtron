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

package devtronApps

import (
	"bufio"
	"context"
	"os"
	"time"

	"github.com/devtron-labs/common-lib/async"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	util5 "github.com/devtron-labs/common-lib/utils/k8s"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	client2 "github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/client/argocdServer"
	client "github.com/devtron-labs/devtron/client/events"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	repository4 "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	bean4 "github.com/devtron-labs/devtron/pkg/app/bean"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	userBean "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/build/git/gitMaterial/read"
	pipeline2 "github.com/devtron-labs/devtron/pkg/build/pipeline"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	repository5 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common"
	bean9 "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/publish"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest/service"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/out"
	"github.com/devtron-labs/devtron/pkg/executor"
	"github.com/devtron-labs/devtron/pkg/imageDigestPolicy"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/plugin"
	security2 "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning"
	read2 "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/read"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	globalUtil "github.com/devtron-labs/devtron/util"
	util2 "github.com/devtron-labs/devtron/util/event"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
)

/*
files in this package are -
HandlerService.go - containing If and impl with common used code
HandlerService_ent.go - containing ent If and impl with common used code
deployStageHandlerCode.go - code related to deploy stage trigger
deployStageHandlerCode_ent.go - ent code related to deploy stage trigger
preStageHandlerCode.go - code related to pre stage trigger
preStageHandlerCode_ent.go - ent code related to pre stage trigger
postStageHandlerCode.go - code related to post stage trigger
postStageHandlerCode_ent.go - ent code related to post stage trigger
prePostWfAndLogsCode.go - code containing pre/post wf handling(abort) and logs related code
*/

type HandlerService interface {
	TriggerPostStage(request bean.TriggerRequest) (*bean4.ManifestPushTemplate, error)
	TriggerPreStage(request bean.TriggerRequest) (*bean4.ManifestPushTemplate, error)

	TriggerAutoCDOnPreStageSuccess(triggerContext bean.TriggerContext, cdPipelineId, ciArtifactId, workflowId int) error

	TriggerStageForBulk(triggerRequest bean.TriggerRequest) error

	ManualCdTrigger(triggerContext bean.TriggerContext, overrideRequest *bean3.ValuesOverrideRequest, userMetadata *userBean.UserMetadata) (int, string, *bean4.ManifestPushTemplate, error)
	TriggerAutomaticDeployment(request bean.TriggerRequest) error

	TriggerRelease(ctx context.Context, overrideRequest *bean3.ValuesOverrideRequest, envDeploymentConfig *bean9.DeploymentConfig, triggeredAt time.Time, triggeredBy int32) (releaseNo int, manifestPushTemplate *bean4.ManifestPushTemplate, err error)

	CancelStage(workflowRunnerId int, forceAbort bool, userId int32) (int, error)
	DownloadCdWorkflowArtifacts(buildId int) (*os.File, error)
	GetRunningWorkflowLogs(environmentId int, pipelineId int, workflowId int) (*bufio.Reader, func() error, error)
}

type HandlerServiceImpl struct {
	logger                              *zap.SugaredLogger
	cdWorkflowCommonService             cd.CdWorkflowCommonService
	gitOpsManifestPushService           publish.GitOpsPushService
	gitOpsConfigReadService             config.GitOpsConfigReadService
	argoK8sClient                       argocdServer.ArgoK8sClient
	ACDConfig                           *argocdServer.ACDConfig
	argoClientWrapperService            argocdServer.ArgoClientWrapperService
	pipelineStatusTimelineService       status.PipelineStatusTimelineService
	chartTemplateService                util.ChartTemplateService
	eventFactory                        client.EventFactory
	eventClient                         client.EventClient
	globalEnvVariables                  *globalUtil.GlobalEnvVariables
	workflowEventPublishService         out.WorkflowEventPublishService
	manifestCreationService             manifest.ManifestCreationService
	deployedConfigurationHistoryService history.DeployedConfigurationHistoryService
	pipelineStageService                pipeline.PipelineStageService
	globalPluginService                 plugin.GlobalPluginService
	customTagService                    pipeline.CustomTagService
	pluginInputVariableParser           pipeline.PluginInputVariableParser
	prePostCdScriptHistoryService       history.PrePostCdScriptHistoryService
	scopedVariableManager               variables.ScopedVariableCMCSManager
	imageDigestPolicyService            imageDigestPolicy.ImageDigestPolicyService
	userService                         user.UserService
	config                              *types.CdConfig
	helmAppService                      client2.HelmAppService
	imageScanService                    security2.ImageScanService
	enforcerUtil                        rbac.EnforcerUtil
	userDeploymentRequestService        service.UserDeploymentRequestService
	helmAppClient                       gRPC.HelmAppClient //TODO refactoring: use helm app service instead
	appRepository                       appRepository.AppRepository
	ciPipelineMaterialRepository        pipelineConfig.CiPipelineMaterialRepository
	imageScanHistoryReadService         read2.ImageScanHistoryReadService
	imageScanDeployInfoService          security2.ImageScanDeployInfoService
	imageScanDeployInfoReadService      read2.ImageScanDeployInfoReadService
	pipelineRepository                  pipelineConfig.PipelineRepository
	pipelineOverrideRepository          chartConfig.PipelineOverrideRepository
	manifestPushConfigRepository        repository.ManifestPushConfigRepository
	chartRepository                     chartRepoRepository.ChartRepository
	envRepository                       repository2.EnvironmentRepository
	cdWorkflowRepository                pipelineConfig.CdWorkflowRepository
	ciWorkflowRepository                pipelineConfig.CiWorkflowRepository
	ciArtifactRepository                repository3.CiArtifactRepository
	ciTemplateService                   pipeline2.CiTemplateReadService
	gitMaterialReadService              read.GitMaterialReadService
	appLabelRepository                  pipelineConfig.AppLabelRepository
	ciPipelineRepository                pipelineConfig.CiPipelineRepository
	appWorkflowRepository               appWorkflow.AppWorkflowRepository
	dockerArtifactStoreRepository       repository4.DockerArtifactStoreRepository
	K8sUtil                             *util5.K8sServiceImpl
	transactionUtilImpl                 *sql.TransactionUtilImpl
	deploymentConfigService             common.DeploymentConfigService
	ciCdPipelineOrchestrator            pipeline.CiCdPipelineOrchestrator
	gitOperationService                 git.GitOperationService
	attributeService                    attributes.AttributesService
	clusterRepository                   repository5.ClusterRepository
	cdWorkflowRunnerService             cd.CdWorkflowRunnerService
	clusterService                      cluster.ClusterService
	ciLogService                        pipeline.CiLogService
	workflowService                     executor.WorkflowService
	blobConfigStorageService            pipeline.BlobStorageConfigService
	asyncRunnable                       *async.Runnable
}

func NewHandlerServiceImpl(logger *zap.SugaredLogger,
	cdWorkflowCommonService cd.CdWorkflowCommonService,
	gitOpsManifestPushService publish.GitOpsPushService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	argoK8sClient argocdServer.ArgoK8sClient,
	ACDConfig *argocdServer.ACDConfig,
	argoClientWrapperService argocdServer.ArgoClientWrapperService,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	chartTemplateService util.ChartTemplateService,
	workflowEventPublishService out.WorkflowEventPublishService,
	manifestCreationService manifest.ManifestCreationService,
	deployedConfigurationHistoryService history.DeployedConfigurationHistoryService,
	pipelineStageService pipeline.PipelineStageService,
	globalPluginService plugin.GlobalPluginService,
	customTagService pipeline.CustomTagService,
	pluginInputVariableParser pipeline.PluginInputVariableParser,
	prePostCdScriptHistoryService history.PrePostCdScriptHistoryService,
	scopedVariableManager variables.ScopedVariableCMCSManager,
	imageDigestPolicyService imageDigestPolicy.ImageDigestPolicyService,
	userService user.UserService,
	helmAppService client2.HelmAppService,
	enforcerUtil rbac.EnforcerUtil,
	userDeploymentRequestService service.UserDeploymentRequestService,
	helmAppClient gRPC.HelmAppClient,
	eventFactory client.EventFactory,
	eventClient client.EventClient,
	envVariables *globalUtil.EnvironmentVariables,
	appRepository appRepository.AppRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	imageScanHistoryReadService read2.ImageScanHistoryReadService,
	imageScanDeployInfoReadService read2.ImageScanDeployInfoReadService,
	imageScanDeployInfoService security2.ImageScanDeployInfoService,
	pipelineRepository pipelineConfig.PipelineRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	manifestPushConfigRepository repository.ManifestPushConfigRepository,
	chartRepository chartRepoRepository.ChartRepository,
	envRepository repository2.EnvironmentRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	ciArtifactRepository repository3.CiArtifactRepository,
	ciTemplateService pipeline2.CiTemplateReadService,
	gitMaterialReadService read.GitMaterialReadService,
	appLabelRepository pipelineConfig.AppLabelRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	dockerArtifactStoreRepository repository4.DockerArtifactStoreRepository,
	imageScanService security2.ImageScanService,
	K8sUtil *util5.K8sServiceImpl,
	transactionUtilImpl *sql.TransactionUtilImpl,
	deploymentConfigService common.DeploymentConfigService,
	ciCdPipelineOrchestrator pipeline.CiCdPipelineOrchestrator,
	gitOperationService git.GitOperationService,
	attributeService attributes.AttributesService,
	clusterRepository repository5.ClusterRepository,
	cdWorkflowRunnerService cd.CdWorkflowRunnerService,
	clusterService cluster.ClusterService,
	ciLogService pipeline.CiLogService,
	workflowService executor.WorkflowService,
	blobConfigStorageService pipeline.BlobStorageConfigService,
	asyncRunnable *async.Runnable) (*HandlerServiceImpl, error) {
	impl := &HandlerServiceImpl{
		logger:                              logger,
		cdWorkflowCommonService:             cdWorkflowCommonService,
		gitOpsManifestPushService:           gitOpsManifestPushService,
		gitOpsConfigReadService:             gitOpsConfigReadService,
		argoK8sClient:                       argoK8sClient,
		ACDConfig:                           ACDConfig,
		argoClientWrapperService:            argoClientWrapperService,
		pipelineStatusTimelineService:       pipelineStatusTimelineService,
		chartTemplateService:                chartTemplateService,
		workflowEventPublishService:         workflowEventPublishService,
		manifestCreationService:             manifestCreationService,
		deployedConfigurationHistoryService: deployedConfigurationHistoryService,
		pipelineStageService:                pipelineStageService,
		globalPluginService:                 globalPluginService,
		customTagService:                    customTagService,
		pluginInputVariableParser:           pluginInputVariableParser,
		prePostCdScriptHistoryService:       prePostCdScriptHistoryService,
		scopedVariableManager:               scopedVariableManager,
		imageDigestPolicyService:            imageDigestPolicyService,
		userService:                         userService,
		helmAppService:                      helmAppService,
		enforcerUtil:                        enforcerUtil,
		eventFactory:                        eventFactory,
		eventClient:                         eventClient,

		globalEnvVariables:             envVariables.GlobalEnvVariables,
		userDeploymentRequestService:   userDeploymentRequestService,
		helmAppClient:                  helmAppClient,
		appRepository:                  appRepository,
		ciPipelineMaterialRepository:   ciPipelineMaterialRepository,
		imageScanHistoryReadService:    imageScanHistoryReadService,
		imageScanDeployInfoReadService: imageScanDeployInfoReadService,
		imageScanDeployInfoService:     imageScanDeployInfoService,
		pipelineRepository:             pipelineRepository,
		pipelineOverrideRepository:     pipelineOverrideRepository,
		manifestPushConfigRepository:   manifestPushConfigRepository,
		chartRepository:                chartRepository,
		envRepository:                  envRepository,
		cdWorkflowRepository:           cdWorkflowRepository,
		ciWorkflowRepository:           ciWorkflowRepository,
		ciArtifactRepository:           ciArtifactRepository,
		ciTemplateService:              ciTemplateService,
		gitMaterialReadService:         gitMaterialReadService,
		appLabelRepository:             appLabelRepository,
		ciPipelineRepository:           ciPipelineRepository,
		appWorkflowRepository:          appWorkflowRepository,
		dockerArtifactStoreRepository:  dockerArtifactStoreRepository,

		imageScanService: imageScanService,
		K8sUtil:          K8sUtil,

		transactionUtilImpl: transactionUtilImpl,

		deploymentConfigService:  deploymentConfigService,
		ciCdPipelineOrchestrator: ciCdPipelineOrchestrator,
		gitOperationService:      gitOperationService,
		attributeService:         attributeService,
		cdWorkflowRunnerService:  cdWorkflowRunnerService,

		clusterRepository:        clusterRepository,
		clusterService:           clusterService,
		ciLogService:             ciLogService,
		workflowService:          workflowService,
		blobConfigStorageService: blobConfigStorageService,
		asyncRunnable:            asyncRunnable,
	}
	config, err := types.GetCdConfig()
	if err != nil {
		return nil, err
	}
	impl.config = config
	return impl, nil
}

func (impl *HandlerServiceImpl) writeCDTriggerEvent(overrideRequest *bean3.ValuesOverrideRequest, artifact *repository3.CiArtifact, releaseId, pipelineOverrideId, wfrId int) {

	event, err := impl.eventFactory.Build(util2.Trigger, &overrideRequest.PipelineId, overrideRequest.AppId, &overrideRequest.EnvId, util2.CD)
	if err != nil {
		impl.logger.Errorw("error in building cd trigger event", "cdPipelineId", overrideRequest.PipelineId, "err", err)
	}
	impl.logger.Debugw("event WriteCDTriggerEvent", "event", event)
	wfr := impl.getEnrichedWorkflowRunner(overrideRequest, artifact, wfrId)
	event = impl.eventFactory.BuildExtraCDData(event, wfr, pipelineOverrideId, bean3.CD_WORKFLOW_TYPE_DEPLOY)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("CD trigger event not sent", "error", evtErr)
	}
	deploymentEvent := app.DeploymentEvent{
		ApplicationId:      overrideRequest.AppId,
		EnvironmentId:      overrideRequest.EnvId, // check for production Environment
		ReleaseId:          releaseId,
		PipelineOverrideId: pipelineOverrideId,
		TriggerTime:        time.Now(),
		CiArtifactId:       overrideRequest.CiArtifactId,
	}

	ciPipelineMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(artifact.PipelineId)
	if err != nil {
		impl.logger.Errorw("error in ")
	}
	materialInfoMap, mErr := artifact.ParseMaterialInfo()
	if mErr != nil {
		impl.logger.Errorw("material info map error", mErr)
		return
	}
	for _, ciPipelineMaterial := range ciPipelineMaterials {
		hash := materialInfoMap[ciPipelineMaterial.GitMaterial.Url]
		pipelineMaterialInfo := &app.PipelineMaterialInfo{PipelineMaterialId: ciPipelineMaterial.Id, CommitHash: hash}
		deploymentEvent.PipelineMaterials = append(deploymentEvent.PipelineMaterials, pipelineMaterialInfo)
	}
	impl.logger.Infow("triggering deployment event", "event", deploymentEvent)
	err = impl.eventClient.WriteNatsEvent(pubsub.CD_SUCCESS, deploymentEvent)
	if err != nil {
		impl.logger.Errorw("error in writing cd trigger event", "err", err)
	}
}
