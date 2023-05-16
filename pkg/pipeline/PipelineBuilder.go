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
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	client "github.com/devtron-labs/devtron/api/helm-app"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	appGroup2 "github.com/devtron-labs/devtron/pkg/appGroup"
	"github.com/devtron-labs/devtron/pkg/chart"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.opentelemetry.io/otel"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/bean"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"go.uber.org/zap"
)

var DefaultPipelineValue = []byte(`{"ConfigMaps":{"enabled":false},"ConfigSecrets":{"enabled":false},"ContainerPort":[],"EnvVariables":[],"GracePeriod":30,"LivenessProbe":{},"MaxSurge":1,"MaxUnavailable":0,"MinReadySeconds":60,"ReadinessProbe":{},"Spec":{"Affinity":{"Values":"nodes","key":""}},"app":"13","appMetrics":false,"args":{},"autoscaling":{},"command":{"enabled":false,"value":[]},"containers":[],"dbMigrationConfig":{"enabled":false},"deployment":{"strategy":{"rolling":{"maxSurge":"25%","maxUnavailable":1}}},"deploymentType":"ROLLING","env":"1","envoyproxy":{"configMapName":"","image":"","resources":{"limits":{"cpu":"50m","memory":"50Mi"},"requests":{"cpu":"50m","memory":"50Mi"}}},"image":{"pullPolicy":"IfNotPresent"},"ingress":{},"ingressInternal":{"annotations":{},"enabled":false,"host":"","path":"","tls":[]},"initContainers":[],"pauseForSecondsBeforeSwitchActive":30,"pipelineName":"","prometheus":{"release":"monitoring"},"rawYaml":[],"releaseVersion":"1","replicaCount":1,"resources":{"limits":{"cpu":"0.05","memory":"50Mi"},"requests":{"cpu":"0.01","memory":"10Mi"}},"secret":{"data":{},"enabled":false},"server":{"deployment":{"image":"","image_tag":""}},"service":{"annotations":{},"type":"ClusterIP"},"servicemonitor":{"additionalLabels":{}},"tolerations":[],"volumeMounts":[],"volumes":[],"waitForSecondsBeforeScalingDown":30}`)

type EcrConfig struct {
	EcrPrefix string `env:"ECR_REPO_NAME_PREFIX" envDefault:"test/"`
}

func GetEcrConfig() (*EcrConfig, error) {
	cfg := &EcrConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

type DeploymentServiceTypeConfig struct {
	IsInternalUse bool `env:"IS_INTERNAL_USE" envDefault:"false"`
}

func GetDeploymentServiceTypeConfig() (*DeploymentServiceTypeConfig, error) {
	cfg := &DeploymentServiceTypeConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

type PipelineBuilder interface {
	CreateCiPipeline(createRequest *bean.CiConfigRequest) (*bean.PipelineCreateResponse, error)
	CreateApp(request *bean.CreateAppDTO) (*bean.CreateAppDTO, error)
	CreateMaterialsForApp(request *bean.CreateMaterialDTO) (*bean.CreateMaterialDTO, error)
	UpdateMaterialsForApp(request *bean.UpdateMaterialDTO) (*bean.UpdateMaterialDTO, error)
	DeleteMaterial(request *bean.UpdateMaterialDTO) error
	DeleteApp(appId int, userId int32) error
	GetCiPipeline(appId int) (ciConfig *bean.CiConfigRequest, err error)
	GetTriggerViewCiPipeline(appId int) (*bean.TriggerViewCiConfig, error)
	GetExternalCi(appId int) (ciConfig []*bean.ExternalCiConfig, err error)
	GetExternalCiById(appId int, externalCiId int) (ciConfig *bean.ExternalCiConfig, err error)
	UpdateCiTemplate(updateRequest *bean.CiConfigRequest) (*bean.CiConfigRequest, error)
	PatchCiPipeline(request *bean.CiPatchRequest) (ciConfig *bean.CiConfigRequest, err error)
	CreateCdPipelines(cdPipelines *bean.CdPipelines, ctx context.Context) (*bean.CdPipelines, error)
	GetApp(appId int) (application *bean.CreateAppDTO, err error)
	PatchCdPipelines(cdPipelines *bean.CDPatchRequest, ctx context.Context) (*bean.CdPipelines, error)
	DeleteCdPipeline(pipeline *pipelineConfig.Pipeline, ctx context.Context, forceDelete bool, acdDelete bool, userId int32) (err error)
	ChangeDeploymentType(ctx context.Context, request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error)
	DeleteDeploymentAppsForEnvironment(ctx context.Context, environmentId int, currentDeploymentAppType bean.DeploymentType, exclusionList []int, includeApps []int, userId int32) (*bean.DeploymentAppTypeChangeResponse, error)
	DeleteDeploymentApps(ctx context.Context, pipelines []*pipelineConfig.Pipeline, userId int32) *bean.DeploymentAppTypeChangeResponse
	GetTriggerViewCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error)
	GetCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error)
	GetCdPipelinesForAppAndEnv(appId int, envId int) (cdPipelines *bean.CdPipelines, err error)
	/*	CreateCdPipelines(cdPipelines bean.CdPipelines) (*bean.CdPipelines, error)*/
	RetrieveArtifactsByCDPipeline(pipeline *pipelineConfig.Pipeline, stage bean2.WorkflowType) (*bean.CiArtifactResponse, error)
	RetrieveParentDetails(pipelineId int) (parentId int, parentType bean2.WorkflowType, err error)
	FetchArtifactForRollback(cdPipelineId, offset, limit int) (bean.CiArtifactResponse, error)
	FindAppsByTeamId(teamId int) ([]*AppBean, error)
	FindAppsByTeamName(teamName string) ([]AppBean, error)
	FindPipelineById(cdPipelineId int) (*pipelineConfig.Pipeline, error)
	FindAppAndEnvDetailsByPipelineId(cdPipelineId int) (*pipelineConfig.Pipeline, error)
	GetAppList() ([]AppBean, error)
	GetCiPipelineMin(appId int) ([]*bean.CiPipelineMin, error)

	FetchCDPipelineStrategy(appId int) (PipelineStrategiesResponse, error)
	GetCdPipelineById(pipelineId int) (cdPipeline *bean.CDPipelineConfigObject, err error)

	FetchConfigmapSecretsForCdStages(appId, envId, cdPipelineId int) (ConfigMapSecretsResponse, error)
	FindByIds(ids []*int) ([]*AppBean, error)
	GetCiPipelineById(pipelineId int) (ciPipeline *bean.CiPipeline, err error)

	GetMaterialsForAppId(appId int) []*bean.GitMaterial
	FindAllMatchesByAppName(appName string, appType helper.AppType) ([]*AppBean, error)
	GetEnvironmentByCdPipelineId(pipelineId int) (int, error)
	PatchRegexCiPipeline(request *bean.CiRegexPatchRequest) (err error)

	GetBulkActionImpactedPipelines(dto *bean.CdBulkActionRequestDto) ([]*pipelineConfig.Pipeline, error)
	PerformBulkActionOnCdPipelines(dto *bean.CdBulkActionRequestDto, impactedPipelines []*pipelineConfig.Pipeline, ctx context.Context, dryRun bool, userId int32) ([]*bean.CdBulkActionResponseDto, error)
	DeleteCiPipeline(request *bean.CiPatchRequest) (*bean.CiPipeline, error)
	IsGitOpsRequiredForCD(pipelineCreateRequest *bean.CdPipelines) bool
	SetPipelineDeploymentAppType(pipelineCreateRequest *bean.CdPipelines, isGitOpsConfigured bool)
	MarkGitOpsDevtronAppsDeletedWhereArgoAppIsDeleted(appId int, envId int, acdToken string, pipeline *pipelineConfig.Pipeline) (bool, error)
	GetCiPipelineByEnvironment(request appGroup2.AppGroupingRequest) ([]*bean.CiConfigRequest, error)
	GetCiPipelineByEnvironmentMin(request appGroup2.AppGroupingRequest) ([]*bean.CiPipelineMinResponse, error)
	GetCdPipelinesByEnvironment(request appGroup2.AppGroupingRequest) (cdPipelines *bean.CdPipelines, err error)
	GetCdPipelinesByEnvironmentMin(request appGroup2.AppGroupingRequest) (cdPipelines []*bean.CDPipelineConfigObject, err error)
	GetExternalCiByEnvironment(request appGroup2.AppGroupingRequest) (ciConfig []*bean.ExternalCiConfig, err error)
	GetEnvironmentListForAutocompleteFilter(envName string, clusterIds []int, offset int, size int, emailId string, checkAuthBatch func(emailId string, appObject []string, envObject []string) (map[string]bool, map[string]bool), ctx context.Context) (*cluster.AppGroupingResponse, error)
	GetAppListForEnvironment(request appGroup2.AppGroupingRequest) ([]*AppBean, error)
}
type PipelineBuilderImpl struct {
	logger                        *zap.SugaredLogger
	ciCdPipelineOrchestrator      CiCdPipelineOrchestrator
	dockerArtifactStoreRepository dockerRegistryRepository.DockerArtifactStoreRepository
	materialRepo                  pipelineConfig.MaterialRepository
	appRepo                       app2.AppRepository
	pipelineRepository            pipelineConfig.PipelineRepository
	propertiesConfigService       PropertiesConfigService
	//	ciTemplateRepository             pipelineConfig.CiTemplateRepository
	ciPipelineRepository             pipelineConfig.CiPipelineRepository
	application                      application.ServiceClient
	chartRepository                  chartRepoRepository.ChartRepository
	ciArtifactRepository             repository.CiArtifactRepository
	ecrConfig                        *EcrConfig
	envConfigOverrideRepository      chartConfig.EnvConfigOverrideRepository
	environmentRepository            repository2.EnvironmentRepository
	pipelineConfigRepository         chartConfig.PipelineConfigRepository
	mergeUtil                        util.MergeUtil
	appWorkflowRepository            appWorkflow.AppWorkflowRepository
	ciConfig                         *CiConfig
	cdWorkflowRepository             pipelineConfig.CdWorkflowRepository
	appService                       app.AppService
	imageScanResultRepository        security.ImageScanResultRepository
	GitFactory                       *util.GitFactory
	ArgoK8sClient                    argocdServer.ArgoK8sClient
	attributesService                attributes.AttributesService
	aCDAuthConfig                    *util3.ACDAuthConfig
	gitOpsRepository                 repository.GitOpsConfigRepository
	pipelineStrategyHistoryService   history.PipelineStrategyHistoryService
	prePostCiScriptHistoryService    history.PrePostCiScriptHistoryService
	prePostCdScriptHistoryService    history.PrePostCdScriptHistoryService
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService
	appLevelMetricsRepository        repository.AppLevelMetricsRepository
	pipelineStageService             PipelineStageService
	chartTemplateService             util.ChartTemplateService
	chartRefRepository               chartRepoRepository.ChartRefRepository
	chartService                     chart.ChartService
	helmAppService                   client.HelmAppService
	deploymentGroupRepository        repository.DeploymentGroupRepository
	ciPipelineMaterialRepository     pipelineConfig.CiPipelineMaterialRepository
	ciWorkflowRepository             pipelineConfig.CiWorkflowRepository
	//ciTemplateOverrideRepository     pipelineConfig.CiTemplateOverrideRepository
	//ciBuildConfigService CiBuildConfigService
	ciTemplateService                               CiTemplateService
	userService                                     user.UserService
	ciTemplateOverrideRepository                    pipelineConfig.CiTemplateOverrideRepository
	gitMaterialHistoryService                       history.GitMaterialHistoryService
	CiTemplateHistoryService                        history.CiTemplateHistoryService
	CiPipelineHistoryService                        history.CiPipelineHistoryService
	globalStrategyMetadataRepository                chartRepoRepository.GlobalStrategyMetadataRepository
	globalStrategyMetadataChartRefMappingRepository chartRepoRepository.GlobalStrategyMetadataChartRefMappingRepository
	deploymentConfig                                *DeploymentServiceTypeConfig
	appStatusRepository                             appStatus.AppStatusRepository
	ArgoUserService                                 argo.ArgoUserService
	workflowDagExecutor                             WorkflowDagExecutor
	enforcerUtil                                    rbac.EnforcerUtil
	appGroupService                                 appGroup2.AppGroupService
}

func NewPipelineBuilderImpl(logger *zap.SugaredLogger,
	ciCdPipelineOrchestrator CiCdPipelineOrchestrator,
	dockerArtifactStoreRepository dockerRegistryRepository.DockerArtifactStoreRepository,
	materialRepo pipelineConfig.MaterialRepository,
	pipelineGroupRepo app2.AppRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	propertiesConfigService PropertiesConfigService,
	ciTemplateRepository pipelineConfig.CiTemplateRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	application application.ServiceClient,
	chartRepository chartRepoRepository.ChartRepository,
	ciArtifactRepository repository.CiArtifactRepository,
	ecrConfig *EcrConfig,
	envConfigOverrideRepository chartConfig.EnvConfigOverrideRepository,
	environmentRepository repository2.EnvironmentRepository,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	mergeUtil util.MergeUtil,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	ciConfig *CiConfig,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	appService app.AppService,
	imageScanResultRepository security.ImageScanResultRepository,
	ArgoK8sClient argocdServer.ArgoK8sClient,
	GitFactory *util.GitFactory, attributesService attributes.AttributesService,
	aCDAuthConfig *util3.ACDAuthConfig, gitOpsRepository repository.GitOpsConfigRepository,
	pipelineStrategyHistoryService history.PipelineStrategyHistoryService,
	prePostCiScriptHistoryService history.PrePostCiScriptHistoryService,
	prePostCdScriptHistoryService history.PrePostCdScriptHistoryService,
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService,
	appLevelMetricsRepository repository.AppLevelMetricsRepository,
	pipelineStageService PipelineStageService, chartRefRepository chartRepoRepository.ChartRefRepository,
	chartTemplateService util.ChartTemplateService, chartService chart.ChartService,
	helmAppService client.HelmAppService,
	deploymentGroupRepository repository.DeploymentGroupRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	userService user.UserService,
	ciTemplateService CiTemplateService,
	ciTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository,
	gitMaterialHistoryService history.GitMaterialHistoryService,
	CiTemplateHistoryService history.CiTemplateHistoryService,
	CiPipelineHistoryService history.CiPipelineHistoryService,
	globalStrategyMetadataRepository chartRepoRepository.GlobalStrategyMetadataRepository,
	globalStrategyMetadataChartRefMappingRepository chartRepoRepository.GlobalStrategyMetadataChartRefMappingRepository,
	deploymentConfig *DeploymentServiceTypeConfig, appStatusRepository appStatus.AppStatusRepository,
	workflowDagExecutor WorkflowDagExecutor,
	enforcerUtil rbac.EnforcerUtil, ArgoUserService argo.ArgoUserService,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	appGroupService appGroup2.AppGroupService) *PipelineBuilderImpl {
	return &PipelineBuilderImpl{
		logger:                        logger,
		ciCdPipelineOrchestrator:      ciCdPipelineOrchestrator,
		dockerArtifactStoreRepository: dockerArtifactStoreRepository,
		materialRepo:                  materialRepo,
		appService:                    appService,
		appRepo:                       pipelineGroupRepo,
		pipelineRepository:            pipelineRepository,
		propertiesConfigService:       propertiesConfigService,
		//ciTemplateRepository:             ciTemplateRepository,
		ciPipelineRepository:             ciPipelineRepository,
		application:                      application,
		chartRepository:                  chartRepository,
		ciArtifactRepository:             ciArtifactRepository,
		ecrConfig:                        ecrConfig,
		envConfigOverrideRepository:      envConfigOverrideRepository,
		environmentRepository:            environmentRepository,
		pipelineConfigRepository:         pipelineConfigRepository,
		mergeUtil:                        mergeUtil,
		appWorkflowRepository:            appWorkflowRepository,
		ciConfig:                         ciConfig,
		cdWorkflowRepository:             cdWorkflowRepository,
		imageScanResultRepository:        imageScanResultRepository,
		ArgoK8sClient:                    ArgoK8sClient,
		GitFactory:                       GitFactory,
		attributesService:                attributesService,
		aCDAuthConfig:                    aCDAuthConfig,
		gitOpsRepository:                 gitOpsRepository,
		pipelineStrategyHistoryService:   pipelineStrategyHistoryService,
		prePostCiScriptHistoryService:    prePostCiScriptHistoryService,
		prePostCdScriptHistoryService:    prePostCdScriptHistoryService,
		deploymentTemplateHistoryService: deploymentTemplateHistoryService,
		appLevelMetricsRepository:        appLevelMetricsRepository,
		pipelineStageService:             pipelineStageService,
		chartTemplateService:             chartTemplateService,
		chartRefRepository:               chartRefRepository,
		chartService:                     chartService,
		helmAppService:                   helmAppService,
		deploymentGroupRepository:        deploymentGroupRepository,
		ciPipelineMaterialRepository:     ciPipelineMaterialRepository,
		ciTemplateService:                ciTemplateService,
		//ciTemplateOverrideRepository:     ciTemplateOverrideRepository,
		//ciBuildConfigService: ciBuildConfigService,
		userService:                                     userService,
		ciTemplateOverrideRepository:                    ciTemplateOverrideRepository,
		gitMaterialHistoryService:                       gitMaterialHistoryService,
		CiTemplateHistoryService:                        CiTemplateHistoryService,
		CiPipelineHistoryService:                        CiPipelineHistoryService,
		globalStrategyMetadataRepository:                globalStrategyMetadataRepository,
		globalStrategyMetadataChartRefMappingRepository: globalStrategyMetadataChartRefMappingRepository,
		deploymentConfig:                                deploymentConfig,
		appStatusRepository:                             appStatusRepository,
		ArgoUserService:                                 ArgoUserService,
		workflowDagExecutor:                             workflowDagExecutor,
		enforcerUtil:                                    enforcerUtil,
		ciWorkflowRepository:                            ciWorkflowRepository,
		appGroupService:                                 appGroupService,
	}
}

// internal use only
const (
	teamIdKey                string = "teamId"
	teamNameKey              string = "teamName"
	appIdKey                 string = "appId"
	appNameKey               string = "appName"
	environmentIdKey         string = "environmentId"
	environmentNameKey       string = "environmentName"
	environmentIdentifierKey string = "environmentIdentifier"
)

func formatDate(t time.Time, layout string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(layout)
}

func (impl PipelineBuilderImpl) CreateApp(request *bean.CreateAppDTO) (*bean.CreateAppDTO, error) {
	impl.logger.Debugw("app create request received", "req", request)

	res, err := impl.ciCdPipelineOrchestrator.CreateApp(request)
	if err != nil {
		impl.logger.Errorw("error in saving create app req", "req", request, "err", err)
	}
	return res, err
}

func (impl PipelineBuilderImpl) DeleteApp(appId int, userId int32) error {
	impl.logger.Debugw("app delete request received", "app", appId)
	err := impl.ciCdPipelineOrchestrator.DeleteApp(appId, userId)
	return err
}

func (impl PipelineBuilderImpl) CreateMaterialsForApp(request *bean.CreateMaterialDTO) (*bean.CreateMaterialDTO, error) {
	res, err := impl.ciCdPipelineOrchestrator.CreateMaterials(request)
	if err != nil {
		impl.logger.Errorw("error in saving create materials req", "req", request, "err", err)
	}
	return res, err
}

func (impl PipelineBuilderImpl) UpdateMaterialsForApp(request *bean.UpdateMaterialDTO) (*bean.UpdateMaterialDTO, error) {
	res, err := impl.ciCdPipelineOrchestrator.UpdateMaterial(request)
	if err != nil {
		impl.logger.Errorw("error in updating materials req", "req", request, "err", err)
	}
	return res, err
}

func (impl PipelineBuilderImpl) DeleteMaterial(request *bean.UpdateMaterialDTO) error {
	//finding ci pipelines for this app; if found any, will not delete git material
	pipelines, err := impl.ciPipelineRepository.FindByAppId(request.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in deleting git material", "gitMaterial", request.Material, "err", err)
		return err
	}
	if len(pipelines) > 0 {
		//pipelines are present, in this case we will check if this material is used in docker config
		//if it is used, then we won't delete
		ciTemplateBean, err := impl.ciTemplateService.FindByAppId(request.AppId)
		if err != nil && err == errors.NotFoundf(err.Error()) {
			impl.logger.Errorw("err in getting docker registry", "appId", request.AppId, "err", err)
			return err
		}
		if ciTemplateBean != nil {
			ciTemplate := ciTemplateBean.CiTemplate
			if ciTemplate != nil && ciTemplate.GitMaterialId == request.Material.Id {
				return fmt.Errorf("cannot delete git material, is being used in docker config")
			}
		}
	}
	existingMaterial, err := impl.materialRepo.FindById(request.Material.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete", "gitMaterial", request.Material)
		return err
	}
	existingMaterial.UpdatedOn = time.Now()
	existingMaterial.UpdatedBy = request.UserId

	err = impl.materialRepo.MarkMaterialDeleted(existingMaterial)

	if err != nil {
		impl.logger.Errorw("error in deleting git material", "gitMaterial", existingMaterial)
		return err
	}
	err = impl.gitMaterialHistoryService.MarkMaterialDeletedAndCreateHistory(existingMaterial)

	return nil
}

func (impl PipelineBuilderImpl) GetApp(appId int) (application *bean.CreateAppDTO, err error) {
	app, err := impl.appRepo.FindById(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching app", "id", appId, "err", err)
		return nil, err
	}
	application = &bean.CreateAppDTO{
		Id:      app.Id,
		AppName: app.AppName,
		TeamId:  app.TeamId,
		AppType: app.AppType,
	}
	if app.AppType == helper.ChartStoreApp {
		return application, nil
	}
	gitMaterials := impl.GetMaterialsForAppId(appId)
	application.Material = gitMaterials
	if app.AppType == helper.Job {
		app.AppName = app.DisplayName
	}
	application.AppType = app.AppType
	return application, nil
}

func (impl PipelineBuilderImpl) GetMaterialsForAppId(appId int) []*bean.GitMaterial {
	materials, err := impl.materialRepo.FindByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching materials", "appId", appId, "err", err)
	}

	ciTemplateBean, err := impl.ciTemplateService.FindByAppId(appId)
	if err != nil && err != errors.NotFoundf(err.Error()) {
		impl.logger.Errorw("err in getting ci-template", "appId", appId, "err", err)
	}

	var gitMaterials []*bean.GitMaterial
	for _, material := range materials {
		gitMaterial := &bean.GitMaterial{
			Url:             material.Url,
			Name:            material.Name[strings.Index(material.Name, "-")+1:],
			Id:              material.Id,
			GitProviderId:   material.GitProviderId,
			CheckoutPath:    material.CheckoutPath,
			FetchSubmodules: material.FetchSubmodules,
			FilterPattern:   material.FilterPattern,
		}
		//check if git material is deletable or not
		if ciTemplateBean != nil {
			ciTemplate := ciTemplateBean.CiTemplate
			if ciTemplate != nil && ciTemplate.GitMaterialId == material.Id {
				gitMaterial.IsUsedInCiConfig = true
			}
		}
		gitMaterials = append(gitMaterials, gitMaterial)
	}
	return gitMaterials
}

/*
   1. create pipelineGroup
   2. save material (add credential provider support)

*/

func (impl PipelineBuilderImpl) getDefaultArtifactStore(id string) (store *dockerRegistryRepository.DockerArtifactStore, err error) {
	if id == "" {
		impl.logger.Debugw("docker repo is empty adding default repo")
		store, err = impl.dockerArtifactStoreRepository.FindActiveDefaultStore()

	} else {
		store, err = impl.dockerArtifactStoreRepository.FindOne(id)
	}
	return
}

func (impl PipelineBuilderImpl) getCiTemplateVariables(appId int) (ciConfig *bean.CiConfigRequest, err error) {
	ciTemplateBean, err := impl.ciTemplateService.FindByAppId(appId)
	if err != nil && !errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}
	if errors.IsNotFound(err) {
		impl.logger.Debugw("no ci pipeline exists", "appId", appId, "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no ci pipeline exists"}
		return nil, err
	}
	template := ciTemplateBean.CiTemplate

	gitMaterials, err := impl.materialRepo.FindByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching git materials", "appId", appId, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		impl.logger.Debugw(" no git materials exists", "appId", appId, "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no git materials exists"}
		return nil, err
	}

	var materials []bean.Material
	for _, g := range gitMaterials {
		m := bean.Material{
			GitMaterialId: g.Id,
			MaterialName:  g.Name[strings.Index(g.Name, "-")+1:],
		}
		materials = append(materials, m)
	}

	var regHost string
	dockerRegistry := template.DockerRegistry
	if dockerRegistry != nil {
		regHost, err = dockerRegistry.GetRegistryLocation()
		if err != nil {
			impl.logger.Errorw("invalid reg url", "err", err)
			return nil, err
		}
	}
	ciConfig = &bean.CiConfigRequest{
		Id:                template.Id,
		AppId:             template.AppId,
		AppName:           template.App.AppName,
		DockerRepository:  template.DockerRepository,
		DockerRegistryUrl: regHost,
		CiBuildConfig:     ciTemplateBean.CiBuildConfig,
		Version:           template.Version,
		CiTemplateName:    template.TemplateName,
		Materials:         materials,
		UpdatedOn:         template.UpdatedOn,
		UpdatedBy:         template.UpdatedBy,
		CreatedBy:         template.CreatedBy,
		CreatedOn:         template.CreatedOn,
		CiGitMaterialId:   template.GitMaterialId,
	}
	if dockerRegistry != nil {
		ciConfig.DockerRegistry = dockerRegistry.Id
	}
	return ciConfig, err
}

func (impl PipelineBuilderImpl) getCiTemplateVariablesByAppIds(appIds []int) (map[int]*bean.CiConfigRequest, error) {
	ciConfigMap := make(map[int]*bean.CiConfigRequest)
	ciTemplateMap, err := impl.ciTemplateService.FindByAppIds(appIds)
	if err != nil && !errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "appIds", appIds, "err", err)
		return nil, err
	}
	if errors.IsNotFound(err) {
		impl.logger.Debugw("no ci pipeline exists", "appIds", appIds, "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no ci pipeline exists"}
		return nil, err
	}
	gitMaterialsMap := make(map[int][]*pipelineConfig.GitMaterial)
	allGitMaterials, err := impl.materialRepo.FindByAppIds(appIds)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching git materials", "appIds", appIds, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		impl.logger.Debugw(" no git materials exists", "appIds", appIds, "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no git materials exists"}
		return nil, err
	}
	for _, gitMaterial := range allGitMaterials {
		gitMaterialsMap[gitMaterial.AppId] = append(gitMaterialsMap[gitMaterial.AppId], gitMaterial)
	}
	for _, ciTemplate := range ciTemplateMap {
		template := ciTemplate.CiTemplate
		var materials []bean.Material
		gitMaterials := gitMaterialsMap[ciTemplate.CiTemplate.AppId]
		for _, g := range gitMaterials {
			m := bean.Material{
				GitMaterialId: g.Id,
				MaterialName:  g.Name[strings.Index(g.Name, "-")+1:],
			}
			materials = append(materials, m)
		}

		var regHost string
		dockerRegistry := template.DockerRegistry
		if dockerRegistry != nil {
			regHost, err = dockerRegistry.GetRegistryLocation()
			if err != nil {
				impl.logger.Errorw("invalid reg url", "err", err)
				return nil, err
			}
		}
		ciConfig := &bean.CiConfigRequest{
			Id:                template.Id,
			AppId:             template.AppId,
			AppName:           template.App.AppName,
			DockerRepository:  template.DockerRepository,
			DockerRegistryUrl: regHost,
			CiBuildConfig:     ciTemplate.CiBuildConfig,
			Version:           template.Version,
			CiTemplateName:    template.TemplateName,
			Materials:         materials,
			//UpdatedOn:         template.UpdatedOn,
			//UpdatedBy:         template.UpdatedBy,
			//CreatedBy:         template.CreatedBy,
			//CreatedOn:         template.CreatedOn,
		}
		if dockerRegistry != nil {
			ciConfig.DockerRegistry = dockerRegistry.Id
		}
		ciConfigMap[template.AppId] = ciConfig
	}
	return ciConfigMap, err
}

func (impl PipelineBuilderImpl) GetTriggerViewCiPipeline(appId int) (*bean.TriggerViewCiConfig, error) {

	triggerViewCiConfig := &bean.TriggerViewCiConfig{}

	ciConfig, err := impl.getCiTemplateVariables(appId)
	if err != nil {
		impl.logger.Debugw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}

	triggerViewCiConfig.CiGitMaterialId = ciConfig.CiBuildConfig.GitMaterialId

	// fetch pipelines
	pipelines, err := impl.ciPipelineRepository.FindByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}

	ciOverrideTemplateMap := make(map[int]*bean3.CiTemplateBean)
	ciTemplateBeanOverrides, err := impl.ciTemplateService.FindTemplateOverrideByAppId(appId)
	if err != nil {
		return nil, err
	}

	for _, templateBeanOverride := range ciTemplateBeanOverrides {
		ciTemplateOverride := templateBeanOverride.CiTemplateOverride
		ciOverrideTemplateMap[ciTemplateOverride.CiPipelineId] = templateBeanOverride
	}

	var ciPipelineResp []*bean.CiPipeline
	for _, pipeline := range pipelines {
		isLinkedCiPipeline := pipeline.IsExternal
		ciPipeline := &bean.CiPipeline{
			Id:                       pipeline.Id,
			Version:                  pipeline.Version,
			Name:                     pipeline.Name,
			Active:                   pipeline.Active,
			Deleted:                  pipeline.Deleted,
			IsManual:                 pipeline.IsManual,
			IsExternal:               isLinkedCiPipeline,
			ParentCiPipeline:         pipeline.ParentCiPipeline,
			ScanEnabled:              pipeline.ScanEnabled,
			IsDockerConfigOverridden: pipeline.IsDockerConfigOverridden,
		}
		if ciTemplateBean, ok := ciOverrideTemplateMap[pipeline.Id]; ok {
			templateOverride := ciTemplateBean.CiTemplateOverride
			ciPipeline.DockerConfigOverride = bean.DockerConfigOverride{
				DockerRegistry:   templateOverride.DockerRegistryId,
				DockerRepository: templateOverride.DockerRepository,
				CiBuildConfig:    ciTemplateBean.CiBuildConfig,
			}
		}
		for _, material := range pipeline.CiPipelineMaterials {
			// ignore those materials which have inactive git material
			if material == nil || material.GitMaterial == nil || !material.GitMaterial.Active {
				continue
			}
			ciMaterial := &bean.CiMaterial{
				Id:              material.Id,
				CheckoutPath:    material.CheckoutPath,
				Path:            material.Path,
				ScmId:           material.ScmId,
				GitMaterialId:   material.GitMaterialId,
				GitMaterialName: material.GitMaterial.Name[strings.Index(material.GitMaterial.Name, "-")+1:],
				ScmName:         material.ScmName,
				ScmVersion:      material.ScmVersion,
				IsRegex:         material.Regex != "",
				Source:          &bean.SourceTypeConfig{Type: material.Type, Value: material.Value, Regex: material.Regex},
			}
			ciPipeline.CiMaterial = append(ciPipeline.CiMaterial, ciMaterial)
		}
		linkedCis, err := impl.ciPipelineRepository.FindByParentCiPipelineId(ciPipeline.Id)
		if err != nil && !util.IsErrNoRows(err) {
			return nil, err
		}
		ciPipeline.LinkedCount = len(linkedCis)
		ciPipelineResp = append(ciPipelineResp, ciPipeline)
	}
	triggerViewCiConfig.CiPipelines = ciPipelineResp
	triggerViewCiConfig.Materials = ciConfig.Materials

	return triggerViewCiConfig, nil
}

func (impl PipelineBuilderImpl) GetCiPipeline(appId int) (ciConfig *bean.CiConfigRequest, err error) {
	ciConfig, err = impl.getCiTemplateVariables(appId)
	if err != nil {
		impl.logger.Debugw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}
	//TODO fill these variables
	//ciConfig.CiPipeline=
	//--------pipeline population start
	pipelines, err := impl.ciPipelineRepository.FindByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}

	if impl.ciConfig.ExternalCiWebhookUrl == "" {
		hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
		if err != nil {
			return nil, err
		}
		if hostUrl != nil {
			impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, ExternalCiWebhookPath)
		}
	}
	//map of ciPipelineId and their templateOverrideConfig
	ciOverrideTemplateMap := make(map[int]*bean3.CiTemplateBean)
	ciTemplateBeanOverrides, err := impl.ciTemplateService.FindTemplateOverrideByAppId(appId)
	if err != nil {
		return nil, err
	}

	for _, templateBeanOverride := range ciTemplateBeanOverrides {
		ciTemplateOverride := templateBeanOverride.CiTemplateOverride
		ciOverrideTemplateMap[ciTemplateOverride.CiPipelineId] = templateBeanOverride
	}
	var ciPipelineResp []*bean.CiPipeline
	for _, pipeline := range pipelines {

		dockerArgs := make(map[string]string)
		if len(pipeline.DockerArgs) > 0 {
			err := json.Unmarshal([]byte(pipeline.DockerArgs), &dockerArgs)
			if err != nil {
				impl.logger.Warnw("error in unmarshal", "err", err)
			}
		}

		var externalCiConfig bean.ExternalCiConfig

		ciPipelineScripts, err := impl.ciPipelineRepository.FindCiScriptsByCiPipelineId(pipeline.Id)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in fetching ci scripts")
			return nil, err
		}

		var beforeDockerBuildScripts []*bean.CiScript
		var afterDockerBuildScripts []*bean.CiScript
		for _, ciScript := range ciPipelineScripts {
			ciScriptResp := &bean.CiScript{
				Id:             ciScript.Id,
				Index:          ciScript.Index,
				Name:           ciScript.Name,
				Script:         ciScript.Script,
				OutputLocation: ciScript.OutputLocation,
			}
			if ciScript.Stage == BEFORE_DOCKER_BUILD {
				beforeDockerBuildScripts = append(beforeDockerBuildScripts, ciScriptResp)
			} else if ciScript.Stage == AFTER_DOCKER_BUILD {
				afterDockerBuildScripts = append(afterDockerBuildScripts, ciScriptResp)
			}
		}
		parentCiPipeline, err := impl.ciPipelineRepository.FindById(pipeline.ParentCiPipeline)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("err", err)
			return nil, err
		}
		ciPipeline := &bean.CiPipeline{
			Id:                       pipeline.Id,
			Version:                  pipeline.Version,
			Name:                     pipeline.Name,
			Active:                   pipeline.Active,
			Deleted:                  pipeline.Deleted,
			DockerArgs:               dockerArgs,
			IsManual:                 pipeline.IsManual,
			IsExternal:               pipeline.IsExternal,
			ParentCiPipeline:         pipeline.ParentCiPipeline,
			ParentAppId:              parentCiPipeline.AppId,
			ExternalCiConfig:         externalCiConfig,
			BeforeDockerBuildScripts: beforeDockerBuildScripts,
			AfterDockerBuildScripts:  afterDockerBuildScripts,
			ScanEnabled:              pipeline.ScanEnabled,
			IsDockerConfigOverridden: pipeline.IsDockerConfigOverridden,
		}
		if ciTemplateBean, ok := ciOverrideTemplateMap[pipeline.Id]; ok {
			templateOverride := ciTemplateBean.CiTemplateOverride
			ciPipeline.DockerConfigOverride = bean.DockerConfigOverride{
				DockerRegistry:   templateOverride.DockerRegistryId,
				DockerRepository: templateOverride.DockerRepository,
				CiBuildConfig:    ciTemplateBean.CiBuildConfig,
			}
		}
		for _, material := range pipeline.CiPipelineMaterials {
			// ignore those materials which have inactive git material
			if material == nil || material.GitMaterial == nil || !material.GitMaterial.Active {
				continue
			}
			ciMaterial := &bean.CiMaterial{
				Id:              material.Id,
				CheckoutPath:    material.CheckoutPath,
				Path:            material.Path,
				ScmId:           material.ScmId,
				GitMaterialId:   material.GitMaterialId,
				GitMaterialName: material.GitMaterial.Name[strings.Index(material.GitMaterial.Name, "-")+1:],
				ScmName:         material.ScmName,
				ScmVersion:      material.ScmVersion,
				IsRegex:         material.Regex != "",
				Source:          &bean.SourceTypeConfig{Type: material.Type, Value: material.Value, Regex: material.Regex},
			}
			ciPipeline.CiMaterial = append(ciPipeline.CiMaterial, ciMaterial)
		}
		linkedCis, err := impl.ciPipelineRepository.FindByParentCiPipelineId(ciPipeline.Id)
		if err != nil && !util.IsErrNoRows(err) {
			return nil, err
		}
		ciPipeline.LinkedCount = len(linkedCis)
		ciPipelineResp = append(ciPipelineResp, ciPipeline)
	}
	ciConfig.CiPipelines = ciPipelineResp
	//--------pipeline population end
	return ciConfig, err
}

func (impl PipelineBuilderImpl) GetExternalCi(appId int) (ciConfig []*bean.ExternalCiConfig, err error) {
	externalCiPipelines, err := impl.ciPipelineRepository.FindExternalCiByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
		return nil, err
	}

	hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
	if err != nil {
		impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
		return nil, err
	}
	if hostUrl != nil {
		impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, ExternalCiWebhookPath)
	}

	externalCiConfigs := make([]*bean.ExternalCiConfig, 0)

	var externalCiPipelineIds []int
	appWorkflowMappingsMap := make(map[int][]*appWorkflow.AppWorkflowMapping)

	for _, externalCiPipeline := range externalCiPipelines {
		externalCiPipelineIds = append(externalCiPipelineIds, externalCiPipeline.Id)
	}
	if len(externalCiPipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no external ci pipeline found"}
		return externalCiConfigs, err
	}
	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFCDMappingByExternalCiIdByIdsIn(externalCiPipelineIds)
	if err != nil {
		impl.logger.Errorw("Error in fetching app workflow mapping for CD pipeline by external CI ID", "err", err)
		return nil, err
	}

	for _, appWorkflowMapping := range appWorkflowMappings {
		appWorkflowMappingsMap[appWorkflowMapping.ParentId] = append(appWorkflowMappingsMap[appWorkflowMapping.ParentId], appWorkflowMapping)
	}

	for _, externalCiPipeline := range externalCiPipelines {
		externalCiConfig := &bean.ExternalCiConfig{
			Id:         externalCiPipeline.Id,
			WebhookUrl: fmt.Sprintf("%s/%d", impl.ciConfig.ExternalCiWebhookUrl, externalCiPipeline.Id),
			Payload:    impl.ciConfig.ExternalCiPayload,
			AccessKey:  "",
		}

		if _, ok := appWorkflowMappingsMap[externalCiPipeline.Id]; !ok {
			impl.logger.Errorw("unable to find app workflow cd mapping corresponding to external ci pipeline id")
			return nil, errors.New("unable to find app workflow cd mapping corresponding to external ci pipeline id")
		}

		var appWorkflowComponentIds []int
		var appIds []int

		CDPipelineMap := make(map[int]*pipelineConfig.Pipeline)
		appIdMap := make(map[int]*app2.App)

		for _, appWorkflowMappings := range appWorkflowMappings {
			appWorkflowComponentIds = append(appWorkflowComponentIds, appWorkflowMappings.ComponentId)
		}
		if len(appWorkflowComponentIds) == 0 {
			continue
		}
		cdPipelines, err := impl.pipelineRepository.FindAppAndEnvironmentAndProjectByPipelineIds(appWorkflowComponentIds)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
			return nil, err
		}
		for _, pipeline := range cdPipelines {
			CDPipelineMap[pipeline.Id] = pipeline
			appIds = append(appIds, pipeline.AppId)
		}
		if len(appIds) == 0 {
			continue
		}
		apps, err := impl.appRepo.FindAppAndProjectByIdsIn(appIds)
		for _, app := range apps {
			appIdMap[app.Id] = app
		}

		roleData := make(map[string]interface{})
		for _, appWorkflowMapping := range appWorkflowMappings {
			if _, ok := CDPipelineMap[appWorkflowMapping.ComponentId]; !ok {
				impl.logger.Errorw("error in getting cd pipeline data for workflow", "app workflow id", appWorkflowMapping.ComponentId, "err", err)
				return nil, errors.New("error in getting cd pipeline data for workflow")
			}
			cdPipeline := CDPipelineMap[appWorkflowMapping.ComponentId]

			if _, ok := roleData[teamIdKey]; !ok {
				if _, ok := appIdMap[cdPipeline.AppId]; !ok {
					impl.logger.Errorw("error in getting app data for pipeline", "app id", cdPipeline.AppId)
					return nil, errors.New("error in getting app data for pipeline")
				}
				app := appIdMap[cdPipeline.AppId]
				roleData[teamIdKey] = app.TeamId
				roleData[teamNameKey] = app.Team.Name
				roleData[appIdKey] = cdPipeline.AppId
				roleData[appNameKey] = cdPipeline.App.AppName
			}
			if _, ok := roleData[environmentNameKey]; !ok {
				roleData[environmentNameKey] = cdPipeline.Environment.Name
			} else {
				roleData[environmentNameKey] = fmt.Sprintf("%s,%s", roleData[environmentNameKey], cdPipeline.Environment.Name)
			}
			if _, ok := roleData[environmentIdentifierKey]; !ok {
				roleData[environmentIdentifierKey] = cdPipeline.Environment.EnvironmentIdentifier
			} else {
				roleData[environmentIdentifierKey] = fmt.Sprintf("%s,%s", roleData[environmentIdentifierKey], cdPipeline.Environment.EnvironmentIdentifier)
			}
		}

		externalCiConfig.ExternalCiConfigRole = bean.ExternalCiConfigRole{
			ProjectId:             roleData[teamIdKey].(int),
			ProjectName:           roleData[teamNameKey].(string),
			AppId:                 roleData[appIdKey].(int),
			AppName:               roleData[appNameKey].(string),
			EnvironmentName:       roleData[environmentNameKey].(string),
			EnvironmentIdentifier: roleData[environmentIdentifierKey].(string),
			Role:                  "Build and deploy",
		}
		externalCiConfigs = append(externalCiConfigs, externalCiConfig)
	}
	//--------pipeline population end
	return externalCiConfigs, err
}

func (impl PipelineBuilderImpl) GetExternalCiById(appId int, externalCiId int) (ciConfig *bean.ExternalCiConfig, err error) {

	externalCiPipeline, err := impl.ciPipelineRepository.FindExternalCiById(externalCiId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
		return nil, err
	}

	if externalCiPipeline.Id == 0 {
		impl.logger.Errorw("invalid external ci id", "externalCiId", externalCiId, "err", err)
		return nil, &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "invalid external ci id"}
	}

	hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
	if err != nil {
		impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
		return nil, err
	}
	if hostUrl != nil {
		impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, ExternalCiWebhookPath)
	}

	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFCDMappingByExternalCiId(externalCiPipeline.Id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
		return nil, err
	}

	roleData := make(map[string]interface{})
	for _, appWorkflowMapping := range appWorkflowMappings {
		cdPipeline, err := impl.pipelineRepository.FindById(appWorkflowMapping.ComponentId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
			return nil, err
		}
		if _, ok := roleData[teamIdKey]; !ok {
			app, err := impl.appRepo.FindAppAndProjectByAppId(cdPipeline.AppId)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
				return nil, err
			}
			roleData[teamIdKey] = app.TeamId
			roleData[teamNameKey] = app.Team.Name
			roleData[appIdKey] = cdPipeline.AppId
			roleData[appNameKey] = cdPipeline.App.AppName
		}
		if _, ok := roleData[environmentNameKey]; !ok {
			roleData[environmentNameKey] = cdPipeline.Environment.Name
		} else {
			roleData[environmentNameKey] = fmt.Sprintf("%s,%s", roleData[environmentNameKey], cdPipeline.Environment.Name)
		}
		if _, ok := roleData[environmentIdentifierKey]; !ok {
			roleData[environmentIdentifierKey] = cdPipeline.Environment.EnvironmentIdentifier
		} else {
			roleData[environmentIdentifierKey] = fmt.Sprintf("%s,%s", roleData[environmentIdentifierKey], cdPipeline.Environment.EnvironmentIdentifier)
		}
	}

	externalCiConfig := &bean.ExternalCiConfig{
		Id:         externalCiPipeline.Id,
		WebhookUrl: fmt.Sprintf("%s/%d", impl.ciConfig.ExternalCiWebhookUrl, externalCiId),
		Payload:    impl.ciConfig.ExternalCiPayload,
		AccessKey:  "",
	}
	externalCiConfig.ExternalCiConfigRole = bean.ExternalCiConfigRole{
		ProjectId:             roleData[teamIdKey].(int),
		ProjectName:           roleData[teamNameKey].(string),
		AppId:                 roleData[appIdKey].(int),
		AppName:               roleData[appNameKey].(string),
		EnvironmentName:       roleData[environmentNameKey].(string),
		EnvironmentIdentifier: roleData[environmentIdentifierKey].(string),
		Role:                  "Build and deploy",
	}
	externalCiConfig.Schema = impl.buildExternalCiWebhookSchema()
	externalCiConfig.PayloadOption = impl.buildPayloadOption()
	externalCiConfig.Responses = impl.buildResponses()
	//--------pipeline population end
	return externalCiConfig, err
}

func (impl PipelineBuilderImpl) GetCiPipelineMin(appId int) ([]*bean.CiPipelineMin, error) {
	pipelines, err := impl.ciPipelineRepository.FindByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows || len(pipelines) == 0 {
		impl.logger.Errorw("no ci pipeline found", "appId", appId, "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no ci pipeline found"}
		return nil, err
	}
	parentCiPipelines, linkedCiPipelineIds, err := impl.ciPipelineRepository.FindParentCiPipelineMapByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("err", err)
		return nil, err
	}
	pipelineParentCiMap := make(map[int]*pipelineConfig.CiPipeline)
	for index, item := range parentCiPipelines {
		pipelineParentCiMap[linkedCiPipelineIds[index]] = item
	}

	var ciPipelineResp []*bean.CiPipelineMin
	for _, pipeline := range pipelines {
		parentCiPipeline := pipelineConfig.CiPipeline{}
		pipelineType := bean.PipelineType(bean.NORMAL)

		if pipelineParentCiMap[pipeline.Id] != nil {
			parentCiPipeline = *pipelineParentCiMap[pipeline.Id]
			pipelineType = bean.PipelineType(bean.LINKED)
		} else if pipeline.IsExternal == true {
			pipelineType = bean.PipelineType(bean.EXTERNAL)
		}

		ciPipeline := &bean.CiPipelineMin{
			Id:               pipeline.Id,
			Name:             pipeline.Name,
			ParentCiPipeline: pipeline.ParentCiPipeline,
			ParentAppId:      parentCiPipeline.AppId,
			PipelineType:     pipelineType,
			ScanEnabled:      pipeline.ScanEnabled,
		}
		ciPipelineResp = append(ciPipelineResp, ciPipeline)
	}
	return ciPipelineResp, err
}

func (impl PipelineBuilderImpl) UpdateCiTemplate(updateRequest *bean.CiConfigRequest) (*bean.CiConfigRequest, error) {
	originalCiConf, err := impl.getCiTemplateVariables(updateRequest.AppId)
	if err != nil {
		impl.logger.Errorw("error in fetching original ciConfig for update", "appId", updateRequest.Id, "err", err)
		return nil, err
	}
	if originalCiConf.Version != updateRequest.Version {
		impl.logger.Errorw("stale version requested", "appId", updateRequest.Id, "old", originalCiConf.Version, "new", updateRequest.Version)
		return nil, fmt.Errorf("stale version of resource requested kindly refresh. requested: %s, found %s", updateRequest.Version, originalCiConf.Version)
	}
	dockerArtifaceStore, err := impl.dockerArtifactStoreRepository.FindOne(updateRequest.DockerRegistry)
	if err != nil {
		impl.logger.Errorw("error in fetching DockerRegistry  for update", "appId", updateRequest.Id, "err", err, "registry", updateRequest.DockerRegistry)
		return nil, err
	}
	regHost, err := dockerArtifaceStore.GetRegistryLocation()
	if err != nil {
		impl.logger.Errorw("invalid reg url", "err", err)
		return nil, err
	}

	var repo string
	if updateRequest.DockerRepository != "" {
		repo = updateRequest.DockerRepository
	} else {
		repo = originalCiConf.DockerRepository
	}

	if dockerArtifaceStore.RegistryType == dockerRegistryRepository.REGISTRYTYPE_ECR {
		err := impl.ciCdPipelineOrchestrator.CreateEcrRepo(repo, dockerArtifaceStore.AWSRegion, dockerArtifaceStore.AWSAccessKeyId, dockerArtifaceStore.AWSSecretAccessKey)
		if err != nil {
			impl.logger.Errorw("ecr repo creation failed while updating ci template", "repo", repo, "err", err)
			return nil, err
		}
	}

	originalCiConf.AfterDockerBuild = updateRequest.AfterDockerBuild
	originalCiConf.BeforeDockerBuild = updateRequest.BeforeDockerBuild
	//originalCiConf.CiBuildConfigBean = updateRequest.CiBuildConfigBean
	originalCiConf.DockerRegistry = updateRequest.DockerRegistry
	originalCiConf.DockerRepository = updateRequest.DockerRepository
	originalCiConf.DockerRegistryUrl = regHost

	//argByte, err := json.Marshal(originalCiConf.DockerBuildConfig.Args)
	//if err != nil {
	//	return nil, err
	//}
	afterByte, err := json.Marshal(originalCiConf.AfterDockerBuild)
	if err != nil {
		return nil, err
	}
	beforeByte, err := json.Marshal(originalCiConf.BeforeDockerBuild)
	if err != nil {
		return nil, err
	}
	//buildOptionsByte, err := json.Marshal(originalCiConf.DockerBuildConfig.DockerBuildOptions)
	//if err != nil {
	//	impl.logger.Errorw("error in marshaling dockerBuildOptions", "err", err)
	//	return nil, err
	//}
	ciBuildConfig := updateRequest.CiBuildConfig
	originalCiBuildConfig := originalCiConf.CiBuildConfig
	ciTemplate := &pipelineConfig.CiTemplate{
		//DockerfilePath:    originalCiConf.DockerBuildConfig.DockerfilePath,
		GitMaterialId:             ciBuildConfig.GitMaterialId,
		BuildContextGitMaterialId: ciBuildConfig.BuildContextGitMaterialId,
		//Args:              string(argByte),
		//TargetPlatform:    originalCiConf.DockerBuildConfig.TargetPlatform,
		AppId:             originalCiConf.AppId,
		BeforeDockerBuild: string(beforeByte),
		AfterDockerBuild:  string(afterByte),
		Version:           originalCiConf.Version,
		Id:                originalCiConf.Id,
		DockerRepository:  originalCiConf.DockerRepository,
		DockerRegistryId:  &originalCiConf.DockerRegistry,
		Active:            true,
		AuditLog: sql.AuditLog{
			CreatedOn: originalCiConf.CreatedOn,
			CreatedBy: originalCiConf.CreatedBy,
			UpdatedOn: time.Now(),
			UpdatedBy: updateRequest.UserId,
		},
	}

	ciBuildConfig.Id = originalCiBuildConfig.Id
	ciTemplateBean := &bean3.CiTemplateBean{
		CiTemplate:    ciTemplate,
		CiBuildConfig: ciBuildConfig,
		UserId:        updateRequest.UserId,
	}
	err = impl.ciTemplateService.Update(ciTemplateBean)
	if err != nil {
		return nil, err
	}

	originalCiConf.CiBuildConfig = ciBuildConfig

	err = impl.CiTemplateHistoryService.SaveHistory(ciTemplateBean, "update")

	if err != nil {
		impl.logger.Errorw("error in saving update history for ci template", "error", err)
	}

	return originalCiConf, nil
}

func (impl PipelineBuilderImpl) CreateCiPipeline(createRequest *bean.CiConfigRequest) (*bean.PipelineCreateResponse, error) {
	impl.logger.Debugw("pipeline create request received", "req", createRequest)

	//-----------fetch data
	app, err := impl.appRepo.FindById(createRequest.AppId)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline group", "groupId", createRequest.AppId, "err", err)
		return nil, err
	}
	//--ecr config
	createRequest.AppName = app.AppName
	if !createRequest.IsJob {
		store, err := impl.getDefaultArtifactStore(createRequest.DockerRegistry)
		if err != nil {
			impl.logger.Errorw("error in fetching docker store ", "id", createRequest.DockerRepository, "err", err)
			return nil, err
		}

		regHost, err := store.GetRegistryLocation()
		if err != nil {
			impl.logger.Errorw("invalid reg url", "err", err)
			return nil, err
		}
		createRequest.DockerRegistryUrl = regHost
		createRequest.DockerRegistry = store.Id

		var repo string
		if createRequest.DockerRepository != "" {
			repo = createRequest.DockerRepository
		} else {
			repo = impl.ecrConfig.EcrPrefix + app.AppName
		}

		if store.RegistryType == dockerRegistryRepository.REGISTRYTYPE_ECR {
			err := impl.ciCdPipelineOrchestrator.CreateEcrRepo(repo, store.AWSRegion, store.AWSAccessKeyId, store.AWSSecretAccessKey)
			if err != nil {
				impl.logger.Errorw("ecr repo creation failed while creating ci pipeline", "repo", repo, "err", err)
				return nil, err
			}
		}
		createRequest.DockerRepository = repo
	}
	//--ecr config	end
	//-- template config start

	//argByte, err := json.Marshal(createRequest.DockerBuildConfig.Args)
	//if err != nil {
	//	return nil, err
	//}
	afterByte, err := json.Marshal(createRequest.AfterDockerBuild)
	if err != nil {
		return nil, err
	}
	beforeByte, err := json.Marshal(createRequest.BeforeDockerBuild)
	if err != nil {
		return nil, err
	}
	buildConfig := createRequest.CiBuildConfig
	ciTemplate := &pipelineConfig.CiTemplate{
		//DockerRegistryId: createRequest.DockerRegistry,
		//DockerRepository: createRequest.DockerRepository,
		GitMaterialId:             buildConfig.GitMaterialId,
		BuildContextGitMaterialId: buildConfig.BuildContextGitMaterialId,
		//DockerfilePath:    createRequest.DockerBuildConfig.DockerfilePath,
		//Args:              string(argByte),
		//TargetPlatform:    createRequest.DockerBuildConfig.TargetPlatform,
		Active:            true,
		TemplateName:      createRequest.CiTemplateName,
		Version:           createRequest.Version,
		AppId:             createRequest.AppId,
		AfterDockerBuild:  string(afterByte),
		BeforeDockerBuild: string(beforeByte),
		AuditLog:          sql.AuditLog{CreatedOn: time.Now(), UpdatedOn: time.Now(), CreatedBy: createRequest.UserId, UpdatedBy: createRequest.UserId},
	}
	if !createRequest.IsJob {
		ciTemplate.DockerRegistryId = &createRequest.DockerRegistry
		ciTemplate.DockerRepository = createRequest.DockerRepository
	}

	ciTemplateBean := &bean3.CiTemplateBean{
		CiTemplate:    ciTemplate,
		CiBuildConfig: createRequest.CiBuildConfig,
	}
	err = impl.ciTemplateService.Save(ciTemplateBean)
	if err != nil {
		return nil, err
	}

	//-- template config end

	err = impl.CiTemplateHistoryService.SaveHistory(ciTemplateBean, "add")

	if err != nil {
		impl.logger.Errorw("error in saving audit logs of ci Template", "error", err)
	}

	createRequest.Id = ciTemplate.Id
	createRequest.CiTemplateName = ciTemplate.TemplateName
	if len(createRequest.CiPipelines) > 0 {
		conf, err := impl.addpipelineToTemplate(createRequest)
		if err != nil {
			impl.logger.Errorw("error in pipeline creation ", "err", err)
			return nil, err
		}
		impl.logger.Debugw("pipeline created ", "detail", conf)
	}
	createRes := &bean.PipelineCreateResponse{AppName: app.AppName, AppId: createRequest.AppId} //FIXME
	return createRes, nil
}

func (impl PipelineBuilderImpl) getGitMaterialsForApp(appId int) ([]*bean.GitMaterial, error) {
	materials, err := impl.materialRepo.FindByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching materials for app", "appId", appId, "err", err)
		return nil, err
	}
	var gitMaterials []*bean.GitMaterial

	for _, material := range materials {
		gitUrl := material.Url
		if material.GitProvider.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD ||
			material.GitProvider.AuthMode == repository.AUTH_MODE_ACCESS_TOKEN {
			u, err := url.Parse(gitUrl)
			if err != nil {
				return nil, err
			}
			var password string
			userName := material.GitProvider.UserName
			if material.GitProvider.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD {
				password = material.GitProvider.Password

			} else if material.GitProvider.AuthMode == repository.AUTH_MODE_ACCESS_TOKEN {
				password = material.GitProvider.AccessToken
				if userName == "" {
					userName = "devtron-boat"
				}
			}
			if userName == "" || password == "" {
				return nil, util.ApiError{}.ErrorfUser("invalid git credentials config")
			}
			u.User = url.UserPassword(userName, password)
			gitUrl = u.String()
		}
		gitMaterial := &bean.GitMaterial{
			Id:            material.Id,
			Url:           gitUrl,
			GitProviderId: material.GitProviderId,
			Name:          material.Name[strings.Index(material.Name, "-")+1:],
			CheckoutPath:  material.CheckoutPath,
		}
		gitMaterials = append(gitMaterials, gitMaterial)
	}
	return gitMaterials, nil
}

func (impl PipelineBuilderImpl) addpipelineToTemplate(createRequest *bean.CiConfigRequest) (resp *bean.CiConfigRequest, err error) {

	if createRequest.AppWorkflowId == 0 {
		// create workflow
		wf := &appWorkflow.AppWorkflow{
			Name:   fmt.Sprintf("wf-%d-%s", createRequest.AppId, util2.Generate(4)),
			AppId:  createRequest.AppId,
			Active: true,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				CreatedBy: createRequest.UserId,
				UpdatedBy: createRequest.UserId,
			},
		}
		savedAppWf, err := impl.appWorkflowRepository.SaveAppWorkflow(wf)
		if err != nil {
			impl.logger.Errorw("err", err)
			return nil, err
		}
		// workflow creation ends
		createRequest.AppWorkflowId = savedAppWf.Id
	}
	//single ci in same wf validation
	workflowMapping, err := impl.appWorkflowRepository.FindWFCIMappingByWorkflowId(createRequest.AppWorkflowId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching workflow mapping for ci validation", "err", err)
		return nil, err
	}
	if len(workflowMapping) > 0 {
		return nil, &util.ApiError{
			InternalMessage:   "pipeline already exists",
			UserDetailMessage: fmt.Sprintf("pipeline already exists in workflow"),
			UserMessage:       fmt.Sprintf("pipeline already exists in workflow")}
	}

	//pipeline name validation
	var pipelineNames []string
	for _, pipeline := range createRequest.CiPipelines {
		pipelineNames = append(pipelineNames, pipeline.Name)
	}
	if err != nil {
		impl.logger.Errorw("error in creating pipeline group", "err", err)
		return nil, err
	}
	createRequest, err = impl.ciCdPipelineOrchestrator.CreateCiConf(createRequest, createRequest.Id)
	if err != nil {
		return nil, err
	}
	return createRequest, err
}

func (impl PipelineBuilderImpl) PatchCiPipeline(request *bean.CiPatchRequest) (ciConfig *bean.CiConfigRequest, err error) {
	ciConfig, err = impl.getCiTemplateVariables(request.AppId)
	if err != nil {
		impl.logger.Errorw("err in fetching template for pipeline patch, ", "err", err, "appId", request.AppId)
		return nil, err
	}
	ciConfig.AppWorkflowId = request.AppWorkflowId
	ciConfig.UserId = request.UserId
	if request.CiPipeline != nil {
		ciConfig.ScanEnabled = request.CiPipeline.ScanEnabled
	}
	switch request.Action {
	case bean.CREATE:
		impl.logger.Debugw("create patch request")
		ciConfig.CiPipelines = []*bean.CiPipeline{request.CiPipeline} //request.CiPipeline
		res, err := impl.addpipelineToTemplate(ciConfig)
		if err != nil {
			impl.logger.Errorw("error in adding pipeline to template", "ciConf", ciConfig, "err", err)
			return nil, err
		}
		return res, nil
	case bean.UPDATE_SOURCE:
		return impl.patchCiPipelineUpdateSource(ciConfig, request.CiPipeline)
	case bean.DELETE:
		pipeline, err := impl.DeleteCiPipeline(request)
		if err != nil {
			return nil, err
		}
		ciConfig.CiPipelines = []*bean.CiPipeline{pipeline}
		return ciConfig, nil
	default:
		impl.logger.Errorw("unsupported operation ", "op", request.Action)
		return nil, fmt.Errorf("unsupported operation %s", request.Action)
	}

}

func (impl PipelineBuilderImpl) PatchRegexCiPipeline(request *bean.CiRegexPatchRequest) (err error) {
	var materials []*pipelineConfig.CiPipelineMaterial
	for _, material := range request.CiPipelineMaterial {
		materialDbObject, err := impl.ciPipelineMaterialRepository.GetById(material.Id)
		if err != nil {
			impl.logger.Errorw("err in fetching material, ", "err", err)
			return err
		}
		if materialDbObject.Regex != "" {
			if !impl.ciCdPipelineOrchestrator.CheckStringMatchRegex(materialDbObject.Regex, material.Value) {
				impl.logger.Errorw("not matching given regex, ", "err", err)
				return errors.New("not matching given regex")
			}
		}
		pipelineMaterial := &pipelineConfig.CiPipelineMaterial{
			Id:            material.Id,
			Value:         material.Value,
			CiPipelineId:  materialDbObject.CiPipelineId,
			Type:          pipelineConfig.SourceType(material.Type),
			Active:        true,
			GitMaterialId: materialDbObject.GitMaterialId,
			Regex:         materialDbObject.Regex,
			AuditLog:      sql.AuditLog{UpdatedBy: request.UserId, UpdatedOn: time.Now(), CreatedOn: time.Now(), CreatedBy: request.UserId},
		}
		materials = append(materials, pipelineMaterial)
	}

	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	err = impl.ciPipelineMaterialRepository.Update(tx, materials...)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	err = impl.ciCdPipelineOrchestrator.AddPipelineMaterialInGitSensor(materials)
	if err != nil {
		impl.logger.Errorf("error in saving pipelineMaterials in git sensor", "materials", materials, "err", err)
		return err
	}
	return nil
}
func (impl PipelineBuilderImpl) DeleteCiPipeline(request *bean.CiPatchRequest) (*bean.CiPipeline, error) {
	ciPipelineId := request.CiPipeline.Id
	//wf validation
	workflowMapping, err := impl.appWorkflowRepository.FindWFCDMappingByCIPipelineId(ciPipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching workflow mapping for ci validation", "err", err)
		return nil, err
	}
	if len(workflowMapping) > 0 {
		return nil, &util.ApiError{
			InternalMessage:   "cd pipeline exists for this CI",
			UserDetailMessage: fmt.Sprintf("cd pipeline exists for this CI"),
			UserMessage:       fmt.Sprintf("cd pipeline exists for this CI")}
	}

	pipeline, err := impl.ciPipelineRepository.FindById(ciPipelineId)
	if err != nil {
		impl.logger.Errorw("pipeline fetch err", "id", ciPipelineId, "err", err)
		return nil, err
	}
	appId := request.AppId
	if pipeline.AppId != appId {
		return nil, fmt.Errorf("invalid appid: %d pipelineId: %d mapping", appId, ciPipelineId)
	}

	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	err = impl.ciCdPipelineOrchestrator.DeleteCiPipeline(pipeline, request, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting pipeline db")
		return nil, err
	}

	//delete app workflow mapping
	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFCIMappingByCIPipelineId(pipeline.Id)
	for _, mapping := range appWorkflowMappings {
		err := impl.appWorkflowRepository.DeleteAppWorkflowMapping(mapping, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting workflow mapping", "err", err)
			return nil, err
		}
	}
	if request.CiPipeline.PreBuildStage != nil && request.CiPipeline.PreBuildStage.Id > 0 {
		//deleting pre stage
		err = impl.pipelineStageService.DeleteCiStage(request.CiPipeline.PreBuildStage, request.UserId, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting pre stage", "err", err, "preBuildStage", request.CiPipeline.PreBuildStage)
			return nil, err
		}
	}
	if request.CiPipeline.PostBuildStage != nil && request.CiPipeline.PostBuildStage.Id > 0 {
		//deleting post stage
		err = impl.pipelineStageService.DeleteCiStage(request.CiPipeline.PostBuildStage, request.UserId, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting post stage", "err", err, "postBuildStage", request.CiPipeline.PostBuildStage)
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	request.CiPipeline.Deleted = true
	request.CiPipeline.Name = pipeline.Name
	return request.CiPipeline, nil
	//delete pipeline
	//delete scm

}

func (impl PipelineBuilderImpl) patchCiPipelineUpdateSource(baseCiConfig *bean.CiConfigRequest, modifiedCiPipeline *bean.CiPipeline) (ciConfig *bean.CiConfigRequest, err error) {

	pipeline, err := impl.ciPipelineRepository.FindById(modifiedCiPipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline", "id", modifiedCiPipeline.Id, "err", err)
		return nil, err
	}

	cannotUpdate := false
	for _, material := range pipeline.CiPipelineMaterials {
		if material.ScmId != "" {
			cannotUpdate = true
		}
	}

	if cannotUpdate {
		//scm plugin material change scm object
		//material.ScmName
		return nil, fmt.Errorf("update of plugin scm material not supported")
	} else {
		modifiedCiPipeline.ScanEnabled = baseCiConfig.ScanEnabled
		modifiedCiPipeline, err = impl.ciCdPipelineOrchestrator.PatchMaterialValue(modifiedCiPipeline, baseCiConfig.UserId, pipeline)
		if err != nil {
			return nil, err
		}
		baseCiConfig.CiPipelines = append(baseCiConfig.CiPipelines, modifiedCiPipeline)
		return baseCiConfig, err
	}

}

func (impl PipelineBuilderImpl) IsGitopsConfigured() (bool, error) {

	isGitOpsConfigured := false
	gitOpsConfig, err := impl.gitOpsRepository.GetGitOpsConfigActive()

	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GetGitOpsConfigActive, error while getting", "err", err)
		return false, err
	}
	if gitOpsConfig != nil && gitOpsConfig.Id > 0 {
		isGitOpsConfigured = true
	}

	return isGitOpsConfigured, nil

}

func (impl PipelineBuilderImpl) ValidateCDPipelineRequest(pipelineCreateRequest *bean.CdPipelines, isGitOpsConfigured, haveAtleastOneGitOps bool) (bool, error) {

	if isGitOpsConfigured == false && haveAtleastOneGitOps {
		impl.logger.Errorw("Gitops not configured but selected in creating cd pipeline")
		err := &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: "Gitops integration is not installed/configured. Please install/configure gitops or use helm option.",
			UserMessage:     "Gitops integration is not installed/configured. Please install/configure gitops or use helm option.",
		}
		return false, err
	}

	envPipelineMap := make(map[int]string)
	for _, pipeline := range pipelineCreateRequest.Pipelines {
		if envPipelineMap[pipeline.EnvironmentId] != "" {
			err := &util.ApiError{
				HttpStatusCode:  http.StatusBadRequest,
				InternalMessage: "cd-pipelines already exist for this app and env, cannot create multiple cd-pipelines",
				UserMessage:     "cd-pipelines already exist for this app and env, cannot create multiple cd-pipelines",
			}
			return false, err
		}
		envPipelineMap[pipeline.EnvironmentId] = pipeline.Name

		existingCdPipelinesForEnv, pErr := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(pipelineCreateRequest.AppId, pipeline.EnvironmentId)
		if pErr != nil && !util.IsErrNoRows(pErr) {
			impl.logger.Errorw("error in fetching cd pipelines ", "err", pErr, "appId", pipelineCreateRequest.AppId)
			return false, pErr
		}
		if len(existingCdPipelinesForEnv) > 0 {
			err := &util.ApiError{
				HttpStatusCode:  http.StatusBadRequest,
				InternalMessage: "cd-pipelines already exist for this app and env, cannot create multiple cd-pipelines",
				UserMessage:     "cd-pipelines already exist for this app and env, cannot create multiple cd-pipelines",
			}
			return false, err
		}

		if len(pipeline.PreStage.Config) > 0 && !strings.Contains(pipeline.PreStage.Config, "beforeStages") {
			err := &util.ApiError{
				HttpStatusCode:  http.StatusBadRequest,
				InternalMessage: "invalid yaml config, must include - beforeStages",
				UserMessage:     "invalid yaml config, must include - beforeStages",
			}
			return false, err
		}
		if len(pipeline.PostStage.Config) > 0 && !strings.Contains(pipeline.PostStage.Config, "afterStages") {
			err := &util.ApiError{
				HttpStatusCode:  http.StatusBadRequest,
				InternalMessage: "invalid yaml config, must include - afterStages",
				UserMessage:     "invalid yaml config, must include - afterStages",
			}
			return false, err
		}
	}

	return true, nil

}

func (impl PipelineBuilderImpl) RegisterInACD(app *app2.App, pipelineCreateRequest *bean.CdPipelines, ctx context.Context) error {

	//if gitops configured create GIT repository and register into ACD
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(app.Id)
	if err != nil && pg.ErrNoRows != err {
		return err
	}
	gitOpsRepoName := impl.chartTemplateService.GetGitOpsRepoName(app.AppName)
	chartGitAttr, err := impl.chartTemplateService.CreateGitRepositoryForApp(gitOpsRepoName, chart.ReferenceTemplate, chart.ChartVersion, pipelineCreateRequest.UserId)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git ", "path", chartGitAttr.ChartLocation, "err", err)
		return err
	}
	err = impl.chartTemplateService.RegisterInArgo(chartGitAttr, ctx)
	if err != nil {
		impl.logger.Errorw("error while register git repo in argo", "err", err)
		emptyRepoErrorMessage := []string{"failed to get index: 404 Not Found", "remote repository is empty"}
		if strings.Contains(err.Error(), emptyRepoErrorMessage[0]) || strings.Contains(err.Error(), emptyRepoErrorMessage[1]) {
			// - found empty repository, create some file in repository
			err := impl.chartTemplateService.CreateReadmeInGitRepo(gitOpsRepoName, pipelineCreateRequest.UserId)
			if err != nil {
				impl.logger.Errorw("error in creating file in git repo", "err", err)
				return err
			}
			// - retry register in argo
			err = impl.chartTemplateService.RegisterInArgo(chartGitAttr, ctx)
			if err != nil {
				impl.logger.Errorw("error in re-try register in argo", "err", err)
				return err
			}
		} else {
			return err
		}
	}

	// here updating all the chart version git repo url, as per current implementation all are same git repo url but we have to update each row
	err = impl.updateGitRepoUrlInCharts(app.Id, chartGitAttr, pipelineCreateRequest.UserId)
	if err != nil {
		impl.logger.Errorw("error in updating git repo urls in charts", "appId", app.Id, "chartGitAttr", chartGitAttr, "err", err)
		return err

	}
	return nil
}

func (impl PipelineBuilderImpl) IsGitOpsRequiredForCD(pipelineCreateRequest *bean.CdPipelines) bool {

	// if deploymentAppType is not coming in request than hasAtLeastOneGitOps will be false

	haveAtLeastOneGitOps := false
	for _, pipeline := range pipelineCreateRequest.Pipelines {
		if pipeline.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD {
			haveAtLeastOneGitOps = true
		}
	}
	return haveAtLeastOneGitOps
}

func (impl PipelineBuilderImpl) SetPipelineDeploymentAppType(pipelineCreateRequest *bean.CdPipelines, isGitOpsConfigured bool) {
	isInternalUse := impl.deploymentConfig.IsInternalUse
	var globalDeploymentAppType string
	if !isInternalUse {
		if isGitOpsConfigured {
			globalDeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
		} else {
			globalDeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
		}
	} else {
		// if gitops or helm is option available, and deployment app type is not present in pipeline request/
		for _, pipeline := range pipelineCreateRequest.Pipelines {
			if pipeline.DeploymentAppType == "" {
				if isGitOpsConfigured {
					pipeline.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
				} else {
					pipeline.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
				}
			}
		}
	}
	for _, pipeline := range pipelineCreateRequest.Pipelines {
		if !isInternalUse {
			pipeline.DeploymentAppType = globalDeploymentAppType
		}
	}
}

func (impl PipelineBuilderImpl) CreateCdPipelines(pipelineCreateRequest *bean.CdPipelines, ctx context.Context) (*bean.CdPipelines, error) {

	isGitOpsConfigured, err := impl.IsGitopsConfigured()
	impl.SetPipelineDeploymentAppType(pipelineCreateRequest, isGitOpsConfigured)
	isGitOpsRequiredForCD := impl.IsGitOpsRequiredForCD(pipelineCreateRequest)
	app, err := impl.appRepo.FindById(pipelineCreateRequest.AppId)
	if err != nil {
		impl.logger.Errorw("app not found", "err", err, "appId", pipelineCreateRequest.AppId)
		return nil, err
	}
	_, err = impl.ValidateCDPipelineRequest(pipelineCreateRequest, isGitOpsConfigured, isGitOpsRequiredForCD)
	if err != nil {
		return nil, err
	}
	if isGitOpsConfigured && isGitOpsRequiredForCD {
		err = impl.RegisterInACD(app, pipelineCreateRequest, ctx)
		if err != nil {
			return nil, err
		}
	}

	for _, pipeline := range pipelineCreateRequest.Pipelines {

		id, err := impl.createCdPipeline(ctx, app, pipeline, pipelineCreateRequest.UserId)
		if err != nil {
			impl.logger.Errorw("error in creating pipeline", "name", pipeline.Name, "err", err)
			return nil, err
		}
		pipeline.Id = id
	}

	return pipelineCreateRequest, nil
}

func (impl PipelineBuilderImpl) PatchCdPipelines(cdPipelines *bean.CDPatchRequest, ctx context.Context) (*bean.CdPipelines, error) {
	pipelineRequest := &bean.CdPipelines{
		UserId:    cdPipelines.UserId,
		AppId:     cdPipelines.AppId,
		Pipelines: []*bean.CDPipelineConfigObject{cdPipelines.Pipeline},
	}
	switch cdPipelines.Action {
	case bean.CD_CREATE:
		return impl.CreateCdPipelines(pipelineRequest, ctx)
	case bean.CD_UPDATE:
		err := impl.updateCdPipeline(ctx, cdPipelines.Pipeline, cdPipelines.UserId)
		return pipelineRequest, err
	case bean.CD_DELETE:
		pipeline, err := impl.pipelineRepository.FindById(cdPipelines.Pipeline.Id)
		if err != nil {
			impl.logger.Errorw("error in getting cd pipeline by id", "err", err, "id", cdPipelines.Pipeline.Id)
			return pipelineRequest, err
		}
		err = impl.DeleteCdPipeline(pipeline, ctx, cdPipelines.ForceDelete, false, cdPipelines.UserId)
		return pipelineRequest, err
	case bean.CD_DELETE_PARTIAL:
		pipeline, err := impl.pipelineRepository.FindById(cdPipelines.Pipeline.Id)
		if err != nil {
			impl.logger.Errorw("error in getting cd pipeline by id", "err", err, "id", cdPipelines.Pipeline.Id)
			return pipelineRequest, err
		}
		err = impl.DeleteCdPipelinePartial(pipeline, ctx, cdPipelines.ForceDelete, cdPipelines.UserId)
		return pipelineRequest, err
	default:
		return nil, &util.ApiError{Code: "404", HttpStatusCode: 404, UserMessage: "operation not supported"}
	}
}

func (impl PipelineBuilderImpl) DeleteCdPipeline(pipeline *pipelineConfig.Pipeline, ctx context.Context, forceDelete, deleteFromAcd bool, userId int32) (err error) {
	//getting children CD pipeline details
	childNodes, err := impl.appWorkflowRepository.FindWFCDMappingByParentCDPipelineId(pipeline.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting children cd details", "err", err)
		return err
	} else if len(childNodes) > 0 {
		impl.logger.Debugw("cannot delete cd pipeline, contains children cd")
		return fmt.Errorf("Please delete children CD pipelines before deleting this pipeline.")
	}
	//getting deployment group for this pipeline
	deploymentGroupNames, err := impl.deploymentGroupRepository.GetNamesByAppIdAndEnvId(pipeline.EnvironmentId, pipeline.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting deployment group names by appId and envId", "err", err)
		return err
	} else if len(deploymentGroupNames) > 0 {
		groupNamesByte, err := json.Marshal(deploymentGroupNames)
		if err != nil {
			impl.logger.Errorw("error in marshaling deployment group names", "err", err, "deploymentGroupNames", deploymentGroupNames)
		}
		impl.logger.Debugw("cannot delete cd pipeline, is being used in deployment group")
		return fmt.Errorf("Please remove this CD pipeline from deployment groups : %s", string(groupNamesByte))
	}
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	if err = impl.ciCdPipelineOrchestrator.DeleteCdPipeline(pipeline.Id, tx); err != nil {
		impl.logger.Errorw("err in deleting pipeline from db", "id", pipeline, "err", err)
		return err
	}
	// delete entry in app_status table
	err = impl.appStatusRepository.Delete(tx, pipeline.AppId, pipeline.EnvironmentId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in deleting app_status from db", "appId", pipeline.AppId, "envId", pipeline.EnvironmentId, "err", err)
		return err
	}
	//delete app workflow mapping
	appWorkflowMapping, err := impl.appWorkflowRepository.FindWFCDMappingByCDPipelineId(pipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in deleting workflow mapping", "err", err)
		return err
	}
	if appWorkflowMapping.ParentType == appWorkflow.WEBHOOK {
		childNodes, err := impl.appWorkflowRepository.FindWFCDMappingByExternalCiId(appWorkflowMapping.ParentId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in fetching external ci", "err", err)
			return err
		}
		noOtherChildNodes := true
		for _, childNode := range childNodes {
			if appWorkflowMapping.Id != childNode.Id {
				noOtherChildNodes = false
			}
		}
		if noOtherChildNodes {
			externalCiPipeline, err := impl.ciPipelineRepository.FindExternalCiById(appWorkflowMapping.ParentId)
			if err != nil {
				impl.logger.Errorw("error in deleting workflow mapping", "err", err)
				return err
			}
			externalCiPipeline.Active = false
			externalCiPipeline.UpdatedOn = time.Now()
			externalCiPipeline.UpdatedBy = userId
			_, err = impl.ciPipelineRepository.UpdateExternalCi(externalCiPipeline, tx)
			if err != nil {
				impl.logger.Errorw("error in deleting workflow mapping", "err", err)
				return err
			}

			appWorkflow, err := impl.appWorkflowRepository.FindById(appWorkflowMapping.AppWorkflowId)
			if err != nil {
				impl.logger.Errorw("error in deleting workflow mapping", "err", err)
				return err
			}
			err = impl.appWorkflowRepository.DeleteAppWorkflow(appWorkflow, tx)
			if err != nil {
				impl.logger.Errorw("error in deleting workflow mapping", "err", err)
				return err
			}
		}
	}
	err = impl.appWorkflowRepository.DeleteAppWorkflowMapping(appWorkflowMapping, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting workflow mapping", "err", err)
		return err
	}

	if pipeline.PreStageConfig != "" {
		err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, tx, repository4.PRE_CD_TYPE, false, 0, time.Time{})
		if err != nil {
			impl.logger.Errorw("error in creating pre cd script entry", "err", err, "pipeline", pipeline)
			return err
		}
	}
	if pipeline.PostStageConfig != "" {
		err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, tx, repository4.POST_CD_TYPE, false, 0, time.Time{})
		if err != nil {
			impl.logger.Errorw("error in creating post cd script entry", "err", err, "pipeline", pipeline)
			return err
		}
	}
	//delete app from argo cd, if created
	if pipeline.DeploymentAppCreated == true {
		deploymentAppName := fmt.Sprintf("%s-%s", pipeline.App.AppName, pipeline.Environment.Name)
		if util.IsAcdApp(pipeline.DeploymentAppType) {
			//todo: provide option for cascading to user
			impl.logger.Debugw("acd app is already deleted for this pipeline", "pipeline", pipeline)
			if deleteFromAcd {
				cascadeDelete := true
				req := &application2.ApplicationDeleteRequest{
					Name:    &deploymentAppName,
					Cascade: &cascadeDelete,
				}
				if _, err := impl.application.Delete(ctx, req); err != nil {
					impl.logger.Errorw("err in deleting pipeline on argocd", "id", pipeline, "err", err)

					if forceDelete {
						impl.logger.Warnw("error while deletion of app in acd, continue to delete in db as this operation is force delete", "error", err)
					} else {
						//statusError, _ := err.(*errors2.StatusError)
						if strings.Contains(err.Error(), "code = NotFound") {
							err = &util.ApiError{
								UserMessage:     "Could not delete as application not found in argocd",
								InternalMessage: err.Error(),
							}
						} else {
							err = &util.ApiError{
								UserMessage:     "Could not delete application",
								InternalMessage: err.Error(),
							}
						}
						return err
					}
				}

				impl.logger.Infow("app deleted from argocd", "id", pipeline.Id, "pipelineName", pipeline.Name, "app", deploymentAppName)
			}
		} else if util.IsHelmApp(pipeline.DeploymentAppType) {
			appIdentifier := &client.AppIdentifier{
				ClusterId:   pipeline.Environment.ClusterId,
				ReleaseName: deploymentAppName,
				Namespace:   pipeline.Environment.Namespace,
			}
			deleteResponse, err := impl.helmAppService.DeleteApplication(ctx, appIdentifier)
			if forceDelete {
				impl.logger.Warnw("error while deletion of helm application, ignore error and delete from db since force delete req", "error", err, "pipelineId", pipeline.Id)
			} else {
				if err != nil {
					impl.logger.Errorw("error in deleting helm application", "error", err, "appIdentifier", appIdentifier)
					return err
				}
				if deleteResponse == nil || !deleteResponse.GetSuccess() {
					return errors.New("delete application response unsuccessful")
				}
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing db transaction", "err", err)
		return err
	}
	return nil
}

func (impl PipelineBuilderImpl) DeleteCdPipelinePartial(pipeline *pipelineConfig.Pipeline, ctx context.Context, forceDelete bool, userId int32) (err error) {
	//getting children CD pipeline details
	childNodes, err := impl.appWorkflowRepository.FindWFCDMappingByParentCDPipelineId(pipeline.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting children cd details", "err", err)
		return err
	} else if len(childNodes) > 0 {
		impl.logger.Debugw("cannot delete cd pipeline, contains children cd")
		return fmt.Errorf("Please delete children CD pipelines before deleting this pipeline.")
	}
	//getting deployment group for this pipeline
	deploymentGroupNames, err := impl.deploymentGroupRepository.GetNamesByAppIdAndEnvId(pipeline.EnvironmentId, pipeline.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting deployment group names by appId and envId", "err", err)
		return err
	} else if len(deploymentGroupNames) > 0 {
		groupNamesByte, err := json.Marshal(deploymentGroupNames)
		if err != nil {
			impl.logger.Errorw("error in marshaling deployment group names", "err", err, "deploymentGroupNames", deploymentGroupNames)
		}
		impl.logger.Debugw("cannot delete cd pipeline, is being used in deployment group")
		return fmt.Errorf("Please remove this CD pipeline from deployment groups : %s", string(groupNamesByte))
	}
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	//delete app from argo cd, if created
	if pipeline.DeploymentAppCreated == true && pipeline.DeploymentAppDeleteRequest == false {
		deploymentAppName := fmt.Sprintf("%s-%s", pipeline.App.AppName, pipeline.Environment.Name)
		if util.IsAcdApp(pipeline.DeploymentAppType) {
			//todo: provide option for cascading to user
			impl.logger.Debugw("acd app is already deleted for this pipeline", "pipeline", pipeline)
			cascadeDelete := true
			req := &application2.ApplicationDeleteRequest{
				Name:    &deploymentAppName,
				Cascade: &cascadeDelete,
			}
			if _, err := impl.application.Delete(ctx, req); err != nil {
				impl.logger.Errorw("err in deleting pipeline on argocd", "id", pipeline, "err", err)

				if forceDelete {
					impl.logger.Warnw("error while deletion of app in acd, continue to delete in db as this operation is force delete", "error", err)
				} else {
					//statusError, _ := err.(*errors2.StatusError)
					if strings.Contains(err.Error(), "code = NotFound") {
						err = &util.ApiError{
							UserMessage:     "Could not delete as application not found in argocd",
							InternalMessage: err.Error(),
						}
					} else {
						err = &util.ApiError{
							UserMessage:     "Could not delete application",
							InternalMessage: err.Error(),
						}
					}
					return err
				}
			}
			impl.logger.Infow("app deleted from argocd", "id", pipeline.Id, "pipelineName", pipeline.Name, "app", deploymentAppName)
			pipeline.DeploymentAppDeleteRequest = true
			err = impl.pipelineRepository.Update(pipeline, tx)
			if err != nil {
				impl.logger.Errorw("error in partially delete cd pipeline", "err", err)
				return err
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing db transaction", "err", err)
		return err
	}
	return nil
}

// ChangeDeploymentType takes in DeploymentAppTypeChangeRequest struct and
// deletes all the cd pipelines for that deployment type in all apps that belongs to
// that environment and updates the db with desired deployment app type
func (impl PipelineBuilderImpl) ChangeDeploymentType(ctx context.Context,
	request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error) {

	var response *bean.DeploymentAppTypeChangeResponse
	var deleteDeploymentType bean.DeploymentType
	var err error

	if request.DesiredDeploymentType == bean.ArgoCd {
		deleteDeploymentType = bean.Helm
	} else {
		deleteDeploymentType = bean.ArgoCd
	}

	// Force delete apps
	response, err = impl.DeleteDeploymentAppsForEnvironment(ctx,
		request.EnvId, deleteDeploymentType, request.ExcludeApps, request.IncludeApps, request.UserId)

	if err != nil {
		return nil, err
	}

	// Updating the env id and desired deployment app type received from request in the response
	response.EnvId = request.EnvId
	response.DesiredDeploymentType = request.DesiredDeploymentType
	response.TriggeredPipelines = make([]*bean.CdPipelineTrigger, 0)

	// Update the deployment app type to Helm and toggle deployment_app_created to false in db
	var cdPipelineIds []int
	for _, item := range response.SuccessfulPipelines {
		cdPipelineIds = append(cdPipelineIds, item.Id)
	}

	// If nothing to update in db
	if len(cdPipelineIds) == 0 {
		return response, nil
	}

	// Update in db
	err = impl.pipelineRepository.UpdateCdPipelineDeploymentAppInFilter(string(request.DesiredDeploymentType),
		cdPipelineIds, request.UserId)

	if err != nil {
		impl.logger.Errorw("failed to update deployment app type in db",
			"pipeline ids", cdPipelineIds,
			"desired deployment type", request.DesiredDeploymentType,
			"err", err)

		return response, nil
	}

	if !request.AutoTriggerDeployment {
		return response, nil
	}

	// Bulk trigger all the successfully changed pipelines (async)
	bulkTriggerRequest := make([]*BulkTriggerRequest, 0)

	pipelineIds := make([]int, 0, len(response.SuccessfulPipelines))
	for _, item := range response.SuccessfulPipelines {
		pipelineIds = append(pipelineIds, item.Id)
	}

	// Get all pipelines
	pipelines, err := impl.pipelineRepository.FindByIdsIn(pipelineIds)
	if err != nil {
		impl.logger.Errorw("failed to fetch pipeline details",
			"ids", pipelineIds,
			"err", err)

		return response, nil
	}

	for _, pipeline := range pipelines {

		artifactDetails, err := impl.RetrieveArtifactsByCDPipeline(pipeline, "DEPLOY")

		if err != nil {
			impl.logger.Errorw("failed to fetch artifact details for cd pipeline",
				"pipelineId", pipeline.Id,
				"appId", pipeline.AppId,
				"envId", pipeline.EnvironmentId,
				"err", err)

			return response, nil
		}

		if artifactDetails.LatestWfArtifactId == 0 || artifactDetails.LatestWfArtifactStatus == "" {
			continue
		}

		bulkTriggerRequest = append(bulkTriggerRequest, &BulkTriggerRequest{
			CiArtifactId: artifactDetails.LatestWfArtifactId,
			PipelineId:   pipeline.Id,
		})
		response.TriggeredPipelines = append(response.TriggeredPipelines, &bean.CdPipelineTrigger{
			CiArtifactId: artifactDetails.LatestWfArtifactId,
			PipelineId:   pipeline.Id,
		})
	}

	// pg panics if empty slice is passed as an argument
	if len(bulkTriggerRequest) == 0 {
		return response, nil
	}

	// Trigger
	_, err = impl.workflowDagExecutor.TriggerBulkDeploymentAsync(bulkTriggerRequest, request.UserId)

	if err != nil {
		impl.logger.Errorw("failed to bulk trigger cd pipelines with error: "+err.Error(),
			"err", err)
	}
	return response, nil
}

// DeleteDeploymentAppsForEnvironment takes in environment id and current deployment app type
// and deletes all the cd pipelines for that deployment type in all apps that belongs to
// that environment.
func (impl PipelineBuilderImpl) DeleteDeploymentAppsForEnvironment(ctx context.Context, environmentId int,
	currentDeploymentAppType bean.DeploymentType, exclusionList []int, includeApps []int, userId int32) (*bean.DeploymentAppTypeChangeResponse, error) {

	// fetch active pipelines from database for the given environment id and current deployment app type
	pipelines, err := impl.pipelineRepository.FindActiveByEnvIdAndDeploymentType(environmentId,
		string(currentDeploymentAppType), exclusionList, includeApps)

	if err != nil {
		impl.logger.Errorw("Error fetching cd pipelines",
			"environmentId", environmentId,
			"currentDeploymentAppType", currentDeploymentAppType,
			"err", err)

		return &bean.DeploymentAppTypeChangeResponse{
			EnvId:               environmentId,
			SuccessfulPipelines: []*bean.DeploymentChangeStatus{},
			FailedPipelines:     []*bean.DeploymentChangeStatus{},
		}, err
	}

	// Currently deleting apps only in argocd is supported
	return impl.DeleteDeploymentApps(ctx, pipelines, userId), nil
}

// DeleteDeploymentApps takes in a list of pipelines and delete the applications
// from argocd / helm.
func (impl PipelineBuilderImpl) DeleteDeploymentApps(ctx context.Context,
	pipelines []*pipelineConfig.Pipeline, userId int32) *bean.DeploymentAppTypeChangeResponse {

	successfulPipelines := make([]*bean.DeploymentChangeStatus, 0)
	failedPipelines := make([]*bean.DeploymentChangeStatus, 0)

	isGitOpsConfigured, gitOpsConfigErr := impl.IsGitopsConfigured()

	// Iterate over all the pipelines in the environment for given deployment app type
	for _, pipeline := range pipelines {

		var isValid bool
		// check if pipeline info like app name and environment is empty or not
		if failedPipelines, isValid = impl.isPipelineInfoValid(pipeline, failedPipelines); !isValid {
			continue
		}

		var healthChkErr error
		// check health of the app if it is argocd deployment type
		if failedPipelines, healthChkErr = impl.handleNotHealthyAppsIfArgoDeploymentType(pipeline, failedPipelines); healthChkErr != nil {

			// cannot delete unhealthy app
			continue
		}

		deploymentAppName := fmt.Sprintf("%s-%s", pipeline.App.AppName, pipeline.Environment.Name)
		var err error

		// delete request
		if pipeline.DeploymentAppType == string(bean.ArgoCd) {
			err = impl.deleteArgoCdApp(ctx, pipeline, deploymentAppName, true)

		} else {

			// For converting from Helm to ArgoCD, GitOps should be configured
			if gitOpsConfigErr != nil || !isGitOpsConfigured {
				err = errors.New("GitOps not configured or unable to fetch GitOps configuration")

			} else {
				// Register app in ACD
				err = impl.RegisterInACD(
					&app2.App{Id: pipeline.AppId, AppName: pipeline.App.AppName},
					&bean.CdPipelines{UserId: userId},
					ctx)
			}
			if err != nil {
				impl.logger.Errorw("error registering app on ACD with error: "+err.Error(),
					"deploymentAppName", deploymentAppName,
					"envId", pipeline.EnvironmentId,
					"appId", pipeline.AppId,
					"err", err)

				// deletion failed, append to the list of failed pipelines
				failedPipelines = impl.handleFailedDeploymentAppChange(pipeline, failedPipelines,
					"failed to register app on ACD with error: "+err.Error())

				continue
			}
			err = impl.deleteHelmApp(ctx, pipeline)
		}

		if err != nil {
			impl.logger.Errorw("error deleting app on "+pipeline.DeploymentAppType,
				"deployment app name", deploymentAppName,
				"err", err)

			// deletion failed, append to the list of failed pipelines
			failedPipelines = impl.handleFailedDeploymentAppChange(pipeline, failedPipelines,
				"error deleting app with error: "+err.Error())

			continue
		}

		// deletion successful, append to the list of successful pipelines
		successfulPipelines = impl.appendToDeploymentChangeStatusList(
			successfulPipelines,
			pipeline,
			"",
			bean.Success)
	}

	return &bean.DeploymentAppTypeChangeResponse{
		SuccessfulPipelines: successfulPipelines,
		FailedPipelines:     failedPipelines,
	}
}

func (impl PipelineBuilderImpl) isGitRepoUrlPresent(appId int) bool {
	fetchedChart, err := impl.chartRepository.FindLatestByAppId(appId)

	if err != nil || len(fetchedChart.GitRepoUrl) == 0 {
		impl.logger.Errorw("error fetching git repo url or it is not present")
		return false
	}
	return true
}

func (impl PipelineBuilderImpl) isPipelineInfoValid(pipeline *pipelineConfig.Pipeline,
	failedPipelines []*bean.DeploymentChangeStatus) ([]*bean.DeploymentChangeStatus, bool) {

	if len(pipeline.App.AppName) == 0 || len(pipeline.Environment.Name) == 0 {
		impl.logger.Errorw("app name or environment name is not present",
			"pipeline id", pipeline.Id)

		failedPipelines = impl.handleFailedDeploymentAppChange(pipeline, failedPipelines,
			"could not fetch app name or environment name")

		return failedPipelines, false
	}
	return failedPipelines, true
}

func (impl PipelineBuilderImpl) handleFailedDeploymentAppChange(pipeline *pipelineConfig.Pipeline,
	failedPipelines []*bean.DeploymentChangeStatus, err string) []*bean.DeploymentChangeStatus {

	return impl.appendToDeploymentChangeStatusList(
		failedPipelines,
		pipeline,
		err,
		bean.Failed)
}

func (impl PipelineBuilderImpl) handleNotHealthyAppsIfArgoDeploymentType(pipeline *pipelineConfig.Pipeline,
	failedPipelines []*bean.DeploymentChangeStatus) ([]*bean.DeploymentChangeStatus, error) {

	if pipeline.DeploymentAppType == string(bean.ArgoCd) {
		// check if app status is Healthy
		status, err := impl.appStatusRepository.Get(pipeline.AppId, pipeline.EnvironmentId)

		// case: missing status row in db
		if len(status.Status) == 0 {
			return failedPipelines, nil
		}

		// cannot delete the app from argocd if app status is Progressing
		if err != nil || status.Status == "Progressing" {

			healthCheckErr := errors.New("unable to fetch app status or app status is progressing")

			impl.logger.Errorw(healthCheckErr.Error(),
				"appId", pipeline.AppId,
				"environmentId", pipeline.EnvironmentId,
				"err", err)

			failedPipelines = impl.handleFailedDeploymentAppChange(pipeline, failedPipelines, healthCheckErr.Error())

			return failedPipelines, healthCheckErr
		}
		return failedPipelines, nil
	}
	return failedPipelines, nil
}

// deleteArgoCdApp takes context and deployment app name used in argo cd and deletes
// the application in argo cd.
func (impl PipelineBuilderImpl) deleteArgoCdApp(ctx context.Context, pipeline *pipelineConfig.Pipeline, deploymentAppName string,
	cascadeDelete bool) error {

	if !pipeline.DeploymentAppCreated {
		return nil
	}

	// building the argocd application delete request
	req := &application2.ApplicationDeleteRequest{
		Name:    &deploymentAppName,
		Cascade: &cascadeDelete,
	}

	_, err := impl.application.Delete(ctx, req)

	if err != nil {
		impl.logger.Errorw("error in deleting argocd application", "err", err)
		// Possible that argocd app got deleted but db updation failed
		if strings.Contains(err.Error(), "code = NotFound") {
			return nil
		}
		return err
	}
	return nil
}

// deleteHelmApp takes in context and pipeline object and deletes the release in helm
func (impl PipelineBuilderImpl) deleteHelmApp(ctx context.Context, pipeline *pipelineConfig.Pipeline) error {

	if !pipeline.DeploymentAppCreated {
		return nil
	}

	// validation
	if !util.IsHelmApp(pipeline.DeploymentAppType) {
		return errors.New("unable to delete pipeline with id: " + strconv.Itoa(pipeline.Id) + ", not a helm app")
	}

	// create app identifier
	appIdentifier := &client.AppIdentifier{
		ClusterId:   pipeline.Environment.ClusterId,
		ReleaseName: pipeline.DeploymentAppName,
		Namespace:   pipeline.Environment.Namespace,
	}

	// call for delete resource
	deleteResponse, err := impl.helmAppService.DeleteApplication(ctx, appIdentifier)

	if err != nil {
		impl.logger.Errorw("error in deleting helm application", "error", err, "appIdentifier", appIdentifier)
		return err
	}

	if deleteResponse == nil || !deleteResponse.GetSuccess() {
		return errors.New("helm delete application response unsuccessful")
	}
	return nil
}

func (impl PipelineBuilderImpl) appendToDeploymentChangeStatusList(pipelines []*bean.DeploymentChangeStatus,
	pipeline *pipelineConfig.Pipeline, error string, status bean.Status) []*bean.DeploymentChangeStatus {

	return append(pipelines, &bean.DeploymentChangeStatus{
		Id:      pipeline.Id,
		AppId:   pipeline.AppId,
		AppName: pipeline.App.AppName,
		EnvId:   pipeline.EnvironmentId,
		EnvName: pipeline.Environment.Name,
		Error:   error,
		Status:  status,
	})
}

type DeploymentType struct {
	Deployment Deployment `json:"deployment"`
}

type Deployment struct {
	Strategy map[string]interface{} `json:"strategy"`
}

func (impl PipelineBuilderImpl) createCdPipeline(ctx context.Context, app *app2.App, pipeline *bean.CDPipelineConfigObject, userId int32) (pipelineRes int, err error) {
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return 0, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	if pipeline.AppWorkflowId == 0 && pipeline.ParentPipelineType == "WEBHOOK" {
		externalCiPipeline := &pipelineConfig.ExternalCiPipeline{
			AppId:       app.Id,
			AccessToken: "",
			Active:      true,
			AuditLog:    sql.AuditLog{CreatedBy: userId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: userId},
		}
		externalCiPipeline, err = impl.ciPipelineRepository.SaveExternalCi(externalCiPipeline, tx)
		wf := &appWorkflow.AppWorkflow{
			Name:     fmt.Sprintf("wf-%d-%s", app.Id, util2.Generate(4)),
			AppId:    app.Id,
			Active:   true,
			AuditLog: sql.AuditLog{CreatedBy: userId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: userId},
		}
		savedAppWf, err := impl.appWorkflowRepository.SaveAppWorkflowWithTx(wf, tx)
		if err != nil {
			impl.logger.Errorw("err", err)
			return 0, err
		}
		appWorkflowMap := &appWorkflow.AppWorkflowMapping{
			AppWorkflowId: savedAppWf.Id,
			ComponentId:   externalCiPipeline.Id,
			Type:          "WEBHOOK",
			Active:        true,
			AuditLog:      sql.AuditLog{CreatedBy: userId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: userId},
		}
		appWorkflowMap, err = impl.appWorkflowRepository.SaveAppWorkflowMapping(appWorkflowMap, tx)
		if err != nil {
			return 0, err
		}
		pipeline.ParentPipelineId = externalCiPipeline.Id
		pipeline.AppWorkflowId = savedAppWf.Id
	}

	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(app.Id)
	if err != nil {
		return 0, err
	}
	envOverride, err := impl.propertiesConfigService.CreateIfRequired(chart, pipeline.EnvironmentId, userId, false, models.CHARTSTATUS_NEW, false, false, pipeline.Namespace, chart.IsBasicViewLocked, chart.CurrentViewEditor, tx)
	if err != nil {
		return 0, err
	}

	// Get pipeline override based on Deployment strategy
	//TODO: mark as created in our db
	pipelineId, err := impl.ciCdPipelineOrchestrator.CreateCDPipelines(pipeline, app.Id, userId, tx, app.AppName)
	if err != nil {
		impl.logger.Errorw("error in ")
		return 0, err
	}

	//adding pipeline to workflow
	_, err = impl.appWorkflowRepository.FindByIdAndAppId(pipeline.AppWorkflowId, app.Id)
	if err != nil && err != pg.ErrNoRows {
		return 0, err
	}
	if pipeline.AppWorkflowId > 0 {
		var parentPipelineId int
		var parentPipelineType string
		if pipeline.ParentPipelineId == 0 {
			parentPipelineId = pipeline.CiPipelineId
			parentPipelineType = "CI_PIPELINE"
		} else {
			parentPipelineId = pipeline.ParentPipelineId
			parentPipelineType = pipeline.ParentPipelineType
		}
		appWorkflowMap := &appWorkflow.AppWorkflowMapping{
			AppWorkflowId: pipeline.AppWorkflowId,
			ParentId:      parentPipelineId,
			ParentType:    parentPipelineType,
			ComponentId:   pipelineId,
			Type:          "CD_PIPELINE",
			Active:        true,
			AuditLog:      sql.AuditLog{CreatedBy: userId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: userId},
		}
		_, err = impl.appWorkflowRepository.SaveAppWorkflowMapping(appWorkflowMap, tx)
		if err != nil {
			return 0, err
		}
	}
	//getting global app metrics for cd pipeline create because env level metrics is not created yet
	appLevelAppMetricsEnabled := false
	appLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(app.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting app level metrics app level", "error", err)
	} else if err == nil {
		appLevelAppMetricsEnabled = appLevelMetrics.AppMetrics
	}
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromEnvOverrideTemplate(envOverride, tx, appLevelAppMetricsEnabled, pipelineId)
	if err != nil {
		impl.logger.Errorw("error in creating entry for env deployment template history", "err", err, "envOverride", envOverride)
		return 0, err
	}
	// strategies for pipeline ids, there is only one is default
	defaultCount := 0
	for _, item := range pipeline.Strategies {
		if item.Default {
			defaultCount = defaultCount + 1
			if defaultCount > 1 {
				impl.logger.Warnw("already have one strategy is default in this pipeline", "strategy", item.DeploymentTemplate)
				item.Default = false
			}
		}
		strategy := &chartConfig.PipelineStrategy{
			PipelineId: pipelineId,
			Strategy:   item.DeploymentTemplate,
			Config:     string(item.Config),
			Default:    item.Default,
			Deleted:    false,
			AuditLog:   sql.AuditLog{UpdatedBy: userId, CreatedBy: userId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
		}
		err = impl.pipelineConfigRepository.Save(strategy, tx)
		if err != nil {
			impl.logger.Errorw("error in saving strategy", "strategy", item.DeploymentTemplate)
			return pipelineId, fmt.Errorf("pipeline created but failed to add strategy")
		}
		//creating history entry for strategy
		_, err = impl.pipelineStrategyHistoryService.CreatePipelineStrategyHistory(strategy, pipeline.TriggerType, tx)
		if err != nil {
			impl.logger.Errorw("error in creating strategy history entry", "err", err)
			return 0, err
		}

	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	impl.logger.Debugw("pipeline created with GitMaterialId ", "id", pipelineId, "pipeline", pipeline)
	return pipelineId, nil
}

func (impl PipelineBuilderImpl) updateCdPipeline(ctx context.Context, pipeline *bean.CDPipelineConfigObject, userID int32) (err error) {

	if len(pipeline.PreStage.Config) > 0 && !strings.Contains(pipeline.PreStage.Config, "beforeStages") {
		err = &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: "invalid yaml config, must include - beforeStages",
			UserMessage:     "invalid yaml config, must include - beforeStages",
		}
		return err
	}
	if len(pipeline.PostStage.Config) > 0 && !strings.Contains(pipeline.PostStage.Config, "afterStages") {
		err = &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: "invalid yaml config, must include - afterStages",
			UserMessage:     "invalid yaml config, must include - afterStages",
		}
		return err
	}
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	err = impl.ciCdPipelineOrchestrator.UpdateCDPipeline(pipeline, userID, tx)
	if err != nil {
		impl.logger.Errorw("error in updating pipeline")
		return err
	}

	// strategies for pipeline ids, there is only one is default
	existingStrategies, err := impl.pipelineConfigRepository.GetAllStrategyByPipelineId(pipeline.Id)
	if err != nil && !errors.IsNotFound(err) {
		impl.logger.Errorw("error in getting pipeline strategies", "err", err)
		return err
	}
	for _, oldItem := range existingStrategies {
		notFound := true
		for _, newItem := range pipeline.Strategies {
			if newItem.DeploymentTemplate == oldItem.Strategy {
				notFound = false
			}
		}

		if notFound {
			//delete from db
			err := impl.pipelineConfigRepository.Delete(oldItem, tx)
			if err != nil {
				impl.logger.Errorw("error in delete pipeline strategies", "err", err)
				return fmt.Errorf("error in delete pipeline strategies")
			}
		}
	}

	defaultCount := 0
	for _, item := range pipeline.Strategies {
		if item.Default {
			defaultCount = defaultCount + 1
			if defaultCount > 1 {
				impl.logger.Warnw("already have one strategy is default in this pipeline, skip this", "strategy", item.DeploymentTemplate)
				continue
			}
		}
		strategy, err := impl.pipelineConfigRepository.FindByStrategyAndPipelineId(item.DeploymentTemplate, pipeline.Id)
		if err != nil && pg.ErrNoRows != err {
			impl.logger.Errorw("error in getting strategy", "err", err)
			return err
		}
		if strategy.Id > 0 {
			strategy.Config = string(item.Config)
			strategy.Default = item.Default
			strategy.UpdatedBy = userID
			strategy.UpdatedOn = time.Now()
			err = impl.pipelineConfigRepository.Update(strategy, tx)
			if err != nil {
				impl.logger.Errorw("error in updating strategy", "strategy", item.DeploymentTemplate)
				return fmt.Errorf("pipeline updated but failed to update one strategy")
			}
			//creating history entry for strategy
			_, err = impl.pipelineStrategyHistoryService.CreatePipelineStrategyHistory(strategy, pipeline.TriggerType, tx)
			if err != nil {
				impl.logger.Errorw("error in creating strategy history entry", "err", err)
				return err
			}
		} else {
			strategy := &chartConfig.PipelineStrategy{
				PipelineId: pipeline.Id,
				Strategy:   item.DeploymentTemplate,
				Config:     string(item.Config),
				Default:    item.Default,
				Deleted:    false,
				AuditLog:   sql.AuditLog{UpdatedBy: userID, CreatedBy: userID, UpdatedOn: time.Now(), CreatedOn: time.Now()},
			}
			err = impl.pipelineConfigRepository.Save(strategy, tx)
			if err != nil {
				impl.logger.Errorw("error in saving strategy", "strategy", item.DeploymentTemplate)
				return fmt.Errorf("pipeline created but failed to add strategy")
			}
			//creating history entry for strategy
			_, err = impl.pipelineStrategyHistoryService.CreatePipelineStrategyHistory(strategy, pipeline.TriggerType, tx)
			if err != nil {
				impl.logger.Errorw("error in creating strategy history entry", "err", err)
				return err
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (impl PipelineBuilderImpl) filterDeploymentTemplate(strategyKey string, pipelineStrategiesJson string) (string, error) {
	var pipelineStrategies DeploymentType
	err := json.Unmarshal([]byte(pipelineStrategiesJson), &pipelineStrategies)
	if err != nil {
		impl.logger.Errorw("error while unmarshal strategies", "err", err)
		return "", err
	}
	if pipelineStrategies.Deployment.Strategy[strategyKey] == nil {
		return "", fmt.Errorf("no deployment strategy found for %s", strategyKey)
	}
	strategy := make(map[string]interface{})
	strategy[strategyKey] = pipelineStrategies.Deployment.Strategy[strategyKey].(map[string]interface{})
	pipelineStrategy := DeploymentType{
		Deployment: Deployment{
			Strategy: strategy,
		},
	}
	pipelineOverrideBytes, err := json.Marshal(pipelineStrategy)
	if err != nil {
		impl.logger.Errorw("error while marshal strategies", "err", err)
		return "", err
	}
	pipelineStrategyJson := string(pipelineOverrideBytes)
	return pipelineStrategyJson, nil
}

func (impl PipelineBuilderImpl) getStrategiesMapping(dbPipelineIds []int) (map[int][]*chartConfig.PipelineStrategy, error) {
	strategiesMapping := make(map[int][]*chartConfig.PipelineStrategy)
	strategiesByPipelineIds, err := impl.pipelineConfigRepository.GetAllStrategyByPipelineIds(dbPipelineIds)
	if err != nil && !errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching strategies by pipelineIds", "PipelineIds", dbPipelineIds, "err", err)
		return strategiesMapping, err
	}
	for _, strategy := range strategiesByPipelineIds {
		strategiesMapping[strategy.PipelineId] = append(strategiesMapping[strategy.PipelineId], strategy)
	}
	return strategiesMapping, nil
}

func (impl PipelineBuilderImpl) GetTriggerViewCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error) {
	triggerViewCdPipelinesResp, err := impl.ciCdPipelineOrchestrator.GetCdPipelinesForApp(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching triggerViewCdPipelinesResp by appId", "err", err, "appId", appId)
		return triggerViewCdPipelinesResp, err
	}
	var dbPipelineIds []int
	for _, dbPipeline := range triggerViewCdPipelinesResp.Pipelines {
		dbPipelineIds = append(dbPipelineIds, dbPipeline.Id)
	}

	//construct strategiesMapping to get all strategies against pipelineId
	strategiesMapping, err := impl.getStrategiesMapping(dbPipelineIds)
	if err != nil {
		return triggerViewCdPipelinesResp, err
	}
	for _, dbPipeline := range triggerViewCdPipelinesResp.Pipelines {
		var strategies []*chartConfig.PipelineStrategy
		var deploymentTemplate chartRepoRepository.DeploymentStrategy
		if len(strategiesMapping[dbPipeline.Id]) != 0 {
			strategies = strategiesMapping[dbPipeline.Id]
		}
		for _, item := range strategies {
			if item.Default {
				deploymentTemplate = item.Strategy
			}
		}
		dbPipeline.DeploymentTemplate = deploymentTemplate
	}

	return triggerViewCdPipelinesResp, err
}

func (impl PipelineBuilderImpl) GetCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error) {
	cdPipelines, err = impl.ciCdPipelineOrchestrator.GetCdPipelinesForApp(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching cd Pipelines for appId", "err", err, "appId", appId)
		return nil, err
	}
	var envIds []*int
	var dbPipelineIds []int
	for _, dbPipeline := range cdPipelines.Pipelines {
		envIds = append(envIds, &dbPipeline.EnvironmentId)
		dbPipelineIds = append(dbPipelineIds, dbPipeline.Id)
	}
	if len(envIds) == 0 || len(dbPipelineIds) == 0 {
		return cdPipelines, nil
	}
	envMapping := make(map[int]*repository2.Environment)
	appWorkflowMapping := make(map[int]*appWorkflow.AppWorkflowMapping)

	envs, err := impl.environmentRepository.FindByIds(envIds)
	if err != nil && errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching environments", "err", err)
		return cdPipelines, err
	}
	//creating map for envId and respective env
	for _, env := range envs {
		envMapping[env.Id] = env
	}
	strategiesMapping, err := impl.getStrategiesMapping(dbPipelineIds)
	if err != nil {
		return cdPipelines, err
	}
	appWorkflowMappings, err := impl.appWorkflowRepository.FindByCDPipelineIds(dbPipelineIds)
	if err != nil {
		impl.logger.Errorw("error in fetching app workflow mappings by pipelineIds", "err", err)
		return nil, err
	}
	for _, appWorkflow := range appWorkflowMappings {
		appWorkflowMapping[appWorkflow.ComponentId] = appWorkflow
	}

	var pipelines []*bean.CDPipelineConfigObject
	for _, dbPipeline := range cdPipelines.Pipelines {
		environment := &repository2.Environment{}
		var strategies []*chartConfig.PipelineStrategy
		appToWorkflowMapping := &appWorkflow.AppWorkflowMapping{}

		if envMapping[dbPipeline.EnvironmentId] != nil {
			environment = envMapping[dbPipeline.EnvironmentId]
		}
		if len(strategiesMapping[dbPipeline.Id]) != 0 {
			strategies = strategiesMapping[dbPipeline.Id]
		}
		if appWorkflowMapping[dbPipeline.Id] != nil {
			appToWorkflowMapping = appWorkflowMapping[dbPipeline.Id]
		}
		var strategiesBean []bean.Strategy
		var deploymentTemplate chartRepoRepository.DeploymentStrategy
		for _, item := range strategies {
			strategiesBean = append(strategiesBean, bean.Strategy{
				Config:             []byte(item.Config),
				DeploymentTemplate: item.Strategy,
				Default:            item.Default,
			})

			if item.Default {
				deploymentTemplate = item.Strategy
			}
		}
		pipeline := &bean.CDPipelineConfigObject{
			Id:                            dbPipeline.Id,
			Name:                          dbPipeline.Name,
			EnvironmentId:                 dbPipeline.EnvironmentId,
			EnvironmentName:               environment.Name,
			Description:                   environment.Description,
			CiPipelineId:                  dbPipeline.CiPipelineId,
			DeploymentTemplate:            deploymentTemplate,
			TriggerType:                   dbPipeline.TriggerType,
			Strategies:                    strategiesBean,
			PreStage:                      dbPipeline.PreStage,
			PostStage:                     dbPipeline.PostStage,
			PreStageConfigMapSecretNames:  dbPipeline.PreStageConfigMapSecretNames,
			PostStageConfigMapSecretNames: dbPipeline.PostStageConfigMapSecretNames,
			RunPreStageInEnv:              dbPipeline.RunPreStageInEnv,
			RunPostStageInEnv:             dbPipeline.RunPostStageInEnv,
			DeploymentAppType:             dbPipeline.DeploymentAppType,
			ParentPipelineType:            appToWorkflowMapping.ParentType,
			ParentPipelineId:              appToWorkflowMapping.ParentId,
			DeploymentAppDeleteRequest:    dbPipeline.DeploymentAppDeleteRequest,
		}
		pipelines = append(pipelines, pipeline)
	}
	cdPipelines.Pipelines = pipelines
	return cdPipelines, err
}

func (impl PipelineBuilderImpl) GetCdPipelinesForAppAndEnv(appId int, envId int) (cdPipelines *bean.CdPipelines, err error) {
	return impl.ciCdPipelineOrchestrator.GetCdPipelinesForAppAndEnv(appId, envId)
}

type ConfigMapSecretsResponse struct {
	Maps    []bean2.Map `json:"maps"`
	Secrets []bean2.Map `json:"secrets"`
}

func (impl PipelineBuilderImpl) FetchConfigmapSecretsForCdStages(appId, envId, cdPipelineId int) (ConfigMapSecretsResponse, error) {
	configMapSecrets, err := impl.appService.GetConfigMapAndSecretJson(appId, envId, cdPipelineId)
	if err != nil {
		impl.logger.Errorw("error while fetching config secrets ", "err", err)
		return ConfigMapSecretsResponse{}, err
	}
	existingConfigMapSecrets := ConfigMapSecretsResponse{}
	err = json.Unmarshal([]byte(configMapSecrets), &existingConfigMapSecrets)
	if err != nil {
		impl.logger.Error(err)
		return ConfigMapSecretsResponse{}, err
	}
	return existingConfigMapSecrets, nil
}

// RetrieveParentDetails returns the parent id and type of the parent
func (impl PipelineBuilderImpl) RetrieveParentDetails(pipelineId int) (parentId int, parentType bean2.WorkflowType, err error) {

	workflow, err := impl.appWorkflowRepository.GetParentDetailsByPipelineId(pipelineId)
	if err != nil {
		impl.logger.Errorw("failed to get parent component details",
			"componentId", pipelineId,
			"err", err)
		return 0, "", err
	}

	if workflow.ParentType == appWorkflow.CDPIPELINE {
		// workflow is of type CD, check for stage
		parentPipeline, err := impl.pipelineRepository.GetPostStageConfigById(workflow.ParentId)
		if err != nil {
			impl.logger.Errorw("failed to get the post_stage_config_yaml",
				"cdPipelineId", workflow.ParentId,
				"err", err)
			return 0, "", err
		}

		if len(parentPipeline.PostStageConfig) == 0 {
			return workflow.ParentId, bean2.CD_WORKFLOW_TYPE_DEPLOY, nil
		}
		return workflow.ParentId, bean2.CD_WORKFLOW_TYPE_POST, nil

	} else if workflow.ParentType == appWorkflow.WEBHOOK {
		// For webhook type
		return workflow.ParentId, bean2.WEBHOOK_WORKFLOW_TYPE, nil
	}

	return workflow.ParentId, bean2.CI_WORKFLOW_TYPE, nil
}

// RetrieveArtifactsByCDPipeline returns all the artifacts for the cd pipeline (pre / deploy / post)
func (impl PipelineBuilderImpl) RetrieveArtifactsByCDPipeline(pipeline *pipelineConfig.Pipeline, stage bean2.WorkflowType) (*bean.CiArtifactResponse, error) {

	// retrieve parent details
	parentId, parentType, err := impl.RetrieveParentDetails(pipeline.Id)
	if err != nil {
		impl.logger.Errorw("failed to retrieve parent details",
			"cdPipelineId", pipeline.Id,
			"err", err)
		return nil, err
	}

	parentCdId := 0
	if parentType == bean2.CD_WORKFLOW_TYPE_POST || (parentType == bean2.CD_WORKFLOW_TYPE_DEPLOY && stage != bean2.CD_WORKFLOW_TYPE_POST) {
		// parentCdId is being set to store the artifact currently deployed on parent cd (if applicable).
		// Parent component is CD only if parent type is POST/DEPLOY
		parentCdId = parentId
	}

	if stage == bean2.CD_WORKFLOW_TYPE_DEPLOY && len(pipeline.PreStageConfig) > 0 {
		// Parent type will be PRE for DEPLOY stage
		parentId = pipeline.Id
		parentType = bean2.CD_WORKFLOW_TYPE_PRE
	}
	if stage == bean2.CD_WORKFLOW_TYPE_POST {
		// Parent type will be DEPLOY for POST stage
		parentId = pipeline.Id
		parentType = bean2.CD_WORKFLOW_TYPE_DEPLOY
	}

	// Build artifacts for cd stages
	var ciArtifacts []bean.CiArtifactBean
	ciArtifactsResponse := &bean.CiArtifactResponse{}

	artifactMap := make(map[int]int)
	limit := 10

	ciArtifacts, artifactMap, latestWfArtifactId, latestWfArtifactStatus, err := impl.
		BuildArtifactsForCdStage(pipeline.Id, stage, ciArtifacts, artifactMap, false, limit, parentCdId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting artifacts for child cd stage", "err", err, "stage", stage)
		return nil, err
	}

	ciArtifacts, err = impl.BuildArtifactsForParentStage(pipeline.Id, parentId, parentType, ciArtifacts, artifactMap, limit, parentCdId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting artifacts for cd", "err", err, "parentStage", parentType, "stage", stage)
		return nil, err
	}

	//sorting ci artifacts on the basis of creation time
	if ciArtifacts != nil {
		sort.SliceStable(ciArtifacts, func(i, j int) bool {
			return ciArtifacts[i].Id > ciArtifacts[j].Id
		})
	}

	artifactIds := make([]int, 0, len(ciArtifacts))
	for _, artifact := range ciArtifacts {
		artifactIds = append(artifactIds, artifact.Id)
	}

	artifacts, err := impl.ciArtifactRepository.GetArtifactParentCiAndWorkflowDetailsByIds(artifactIds)
	if err != nil {
		return ciArtifactsResponse, err
	}

	for i, artifact := range artifacts {
		if artifact.ExternalCiPipelineId != 0 {
			// if external webhook continue
			continue
		}

		var ciWorkflow *pipelineConfig.CiWorkflow
		if artifact.ParentCiArtifact != 0 {
			ciWorkflow, err = impl.ciWorkflowRepository.FindLastTriggeredWorkflowGitTriggersByArtifactId(artifact.ParentCiArtifact)
			if err != nil {
				impl.logger.Errorw("error in getting ci_workflow for artifacts", "err", err, "artifact", artifact, "parentStage", parentType, "stage", stage)
				return ciArtifactsResponse, err
			}

		} else {
			ciWorkflow, err = impl.ciWorkflowRepository.FindCiWorkflowGitTriggersById(*artifact.WorkflowId)
			if err != nil {
				impl.logger.Errorw("error in getting ci_workflow for artifacts", "err", err, "artifact", artifact, "parentStage", parentType, "stage", stage)
				return ciArtifactsResponse, err
			}
		}
		ciArtifacts[i].CiConfigureSourceType = ciWorkflow.GitTriggers[ciWorkflow.CiPipelineId].CiConfigureSourceType
		ciArtifacts[i].CiConfigureSourceValue = ciWorkflow.GitTriggers[ciWorkflow.CiPipelineId].CiConfigureSourceValue
	}

	ciArtifactsResponse.CdPipelineId = pipeline.Id
	ciArtifactsResponse.LatestWfArtifactId = latestWfArtifactId
	ciArtifactsResponse.LatestWfArtifactStatus = latestWfArtifactStatus
	if ciArtifacts == nil {
		ciArtifacts = []bean.CiArtifactBean{}
	}
	ciArtifactsResponse.CiArtifacts = ciArtifacts
	return ciArtifactsResponse, nil
}

func (impl PipelineBuilderImpl) BuildArtifactsForParentStage(cdPipelineId int, parentId int, parentType bean2.WorkflowType, ciArtifacts []bean.CiArtifactBean, artifactMap map[int]int, limit int, parentCdId int) ([]bean.CiArtifactBean, error) {
	var ciArtifactsFinal []bean.CiArtifactBean
	var err error
	if parentType == bean2.CI_WORKFLOW_TYPE {
		ciArtifactsFinal, err = impl.BuildArtifactsForCIParent(cdPipelineId, parentId, parentType, ciArtifacts, artifactMap, limit)
	} else if parentType == bean2.WEBHOOK_WORKFLOW_TYPE {
		ciArtifactsFinal, err = impl.BuildArtifactsForCIParent(cdPipelineId, parentId, parentType, ciArtifacts, artifactMap, limit)
	} else {
		//parent type is PRE, POST or DEPLOY type
		ciArtifactsFinal, _, _, _, err = impl.BuildArtifactsForCdStage(parentId, parentType, ciArtifacts, artifactMap, true, limit, parentCdId)
	}
	return ciArtifactsFinal, err
}

func (impl PipelineBuilderImpl) BuildArtifactsForCdStage(pipelineId int, stageType bean2.WorkflowType, ciArtifacts []bean.CiArtifactBean, artifactMap map[int]int, parent bool, limit int, parentCdId int) ([]bean.CiArtifactBean, map[int]int, int, string, error) {
	//getting running artifact id for parent cd
	parentCdRunningArtifactId := 0
	if parentCdId > 0 && parent {
		parentCdWfrList, err := impl.cdWorkflowRepository.FindArtifactByPipelineIdAndRunnerType(parentCdId, bean2.CD_WORKFLOW_TYPE_DEPLOY, 1)
		if err != nil || len(parentCdWfrList) == 0 {
			impl.logger.Errorw("error in getting artifact for parent cd", "parentCdPipelineId", parentCdId)
			return ciArtifacts, artifactMap, 0, "", err
		}
		parentCdRunningArtifactId = parentCdWfrList[0].CdWorkflow.CiArtifact.Id
	}
	//getting wfr for parent and updating artifacts
	parentWfrList, err := impl.cdWorkflowRepository.FindArtifactByPipelineIdAndRunnerType(pipelineId, stageType, limit)
	if err != nil {
		impl.logger.Errorw("error in getting artifact for deployed items", "cdPipelineId", pipelineId)
		return ciArtifacts, artifactMap, 0, "", err
	}
	deploymentArtifactId := 0
	deploymentArtifactStatus := ""
	for index, wfr := range parentWfrList {
		if !parent && index == 0 {
			deploymentArtifactId = wfr.CdWorkflow.CiArtifact.Id
			deploymentArtifactStatus = wfr.Status
		}
		if wfr.Status == application.Healthy || wfr.Status == application.SUCCEEDED {
			lastSuccessfulTriggerOnParent := parent && index == 0
			latest := !parent && index == 0
			runningOnParentCd := parentCdRunningArtifactId == wfr.CdWorkflow.CiArtifact.Id
			if ciArtifactIndex, ok := artifactMap[wfr.CdWorkflow.CiArtifact.Id]; !ok {
				//entry not present, creating new entry
				mInfo, err := parseMaterialInfo([]byte(wfr.CdWorkflow.CiArtifact.MaterialInfo), wfr.CdWorkflow.CiArtifact.DataSource)
				if err != nil {
					mInfo = []byte("[]")
					impl.logger.Errorw("Error in parsing artifact material info", "err", err)
				}
				ciArtifact := bean.CiArtifactBean{
					Id:                            wfr.CdWorkflow.CiArtifact.Id,
					Image:                         wfr.CdWorkflow.CiArtifact.Image,
					ImageDigest:                   wfr.CdWorkflow.CiArtifact.ImageDigest,
					MaterialInfo:                  mInfo,
					LastSuccessfulTriggerOnParent: lastSuccessfulTriggerOnParent,
					Latest:                        latest,
					Scanned:                       wfr.CdWorkflow.CiArtifact.Scanned,
					ScanEnabled:                   wfr.CdWorkflow.CiArtifact.ScanEnabled,
				}
				if !parent {
					ciArtifact.Deployed = true
					ciArtifact.DeployedTime = formatDate(wfr.StartedOn, bean.LayoutRFC3339)
				}
				if runningOnParentCd {
					ciArtifact.RunningOnParentCd = runningOnParentCd
				}
				ciArtifacts = append(ciArtifacts, ciArtifact)
				//storing index of ci artifact for using when updating old entry
				artifactMap[wfr.CdWorkflow.CiArtifact.Id] = len(ciArtifacts) - 1
			} else {
				//entry already present, updating running on parent
				if parent {
					ciArtifacts[ciArtifactIndex].LastSuccessfulTriggerOnParent = lastSuccessfulTriggerOnParent
				}
				if runningOnParentCd {
					ciArtifacts[ciArtifactIndex].RunningOnParentCd = runningOnParentCd
				}
			}
		}
	}
	return ciArtifacts, artifactMap, deploymentArtifactId, deploymentArtifactStatus, nil
}

// method for building artifacts for parent CI

func (impl PipelineBuilderImpl) BuildArtifactsForCIParent(cdPipelineId int, parentId int, parentType bean2.WorkflowType, ciArtifacts []bean.CiArtifactBean, artifactMap map[int]int, limit int) ([]bean.CiArtifactBean, error) {
	artifacts, err := impl.ciArtifactRepository.GetArtifactsByCDPipeline(cdPipelineId, limit, parentId, parentType)
	if err != nil {
		impl.logger.Errorw("error in getting artifacts for ci", "err", err)
		return ciArtifacts, err
	}
	for _, artifact := range artifacts {
		if _, ok := artifactMap[artifact.Id]; !ok {
			mInfo, err := parseMaterialInfo([]byte(artifact.MaterialInfo), artifact.DataSource)
			if err != nil {
				mInfo = []byte("[]")
				impl.logger.Errorw("Error in parsing artifact material info", "err", err, "artifact", artifact)
			}
			ciArtifacts = append(ciArtifacts, bean.CiArtifactBean{
				Id:           artifact.Id,
				Image:        artifact.Image,
				ImageDigest:  artifact.ImageDigest,
				MaterialInfo: mInfo,
				ScanEnabled:  artifact.ScanEnabled,
				Scanned:      artifact.Scanned,
			})
		}
	}
	return ciArtifacts, nil
}

func (impl PipelineBuilderImpl) FetchArtifactForRollback(cdPipelineId, offset, limit int) (bean.CiArtifactResponse, error) {
	var deployedCiArtifacts []bean.CiArtifactBean
	var deployedCiArtifactsResponse bean.CiArtifactResponse

	cdWfrs, err := impl.cdWorkflowRepository.FetchArtifactsByCdPipelineId(cdPipelineId, bean2.CD_WORKFLOW_TYPE_DEPLOY, offset, limit)
	if err != nil {
		impl.logger.Errorw("error in getting artifacts for rollback by cdPipelineId", "err", err, "cdPipelineId", cdPipelineId)
		return deployedCiArtifactsResponse, err
	}
	var ids []int32
	for _, item := range cdWfrs {
		ids = append(ids, item.TriggeredBy)
	}
	userEmails := make(map[int32]string)
	users, err := impl.userService.GetByIds(ids)
	if err != nil {
		impl.logger.Errorw("unable to fetch users by ids", "err", err, "ids", ids)
	}
	for _, item := range users {
		userEmails[item.Id] = item.EmailId
	}
	for _, cdWfr := range cdWfrs {
		ciArtifact := &repository.CiArtifact{}
		if cdWfr.CdWorkflow != nil && cdWfr.CdWorkflow.CiArtifact != nil {
			ciArtifact = cdWfr.CdWorkflow.CiArtifact
		}
		if ciArtifact == nil {
			continue
		}
		mInfo, err := parseMaterialInfo([]byte(ciArtifact.MaterialInfo), ciArtifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("error in parsing ciArtifact material info", "err", err, "ciArtifact", ciArtifact)
		}
		userEmail := userEmails[cdWfr.TriggeredBy]
		deployedCiArtifacts = append(deployedCiArtifacts, bean.CiArtifactBean{
			Id:           ciArtifact.Id,
			Image:        ciArtifact.Image,
			MaterialInfo: mInfo,
			DeployedTime: formatDate(cdWfr.StartedOn, bean.LayoutRFC3339),
			WfrId:        cdWfr.Id,
			DeployedBy:   userEmail,
		})
	}

	deployedCiArtifactsResponse.CdPipelineId = cdPipelineId
	if deployedCiArtifacts == nil {
		deployedCiArtifacts = []bean.CiArtifactBean{}
	}
	deployedCiArtifactsResponse.CiArtifacts = deployedCiArtifacts

	return deployedCiArtifactsResponse, nil
}

func parseMaterialInfo(materialInfo json.RawMessage, source string) (json.RawMessage, error) {
	if source != "GOCD" && source != "CI-RUNNER" && source != "EXTERNAL" {
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

func (impl PipelineBuilderImpl) FindAppsByTeamId(teamId int) ([]*AppBean, error) {
	var appsRes []*AppBean
	apps, err := impl.appRepo.FindAppsByTeamId(teamId)
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		appsRes = append(appsRes, &AppBean{Id: app.Id, Name: app.AppName})
	}
	return appsRes, err
}

func (impl PipelineBuilderImpl) FindAppsByTeamName(teamName string) ([]AppBean, error) {
	var appsRes []AppBean
	apps, err := impl.appRepo.FindAppsByTeamName(teamName)
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		appsRes = append(appsRes, AppBean{Id: app.Id, Name: app.AppName})
	}
	return appsRes, err
}

func (impl PipelineBuilderImpl) FindPipelineById(cdPipelineId int) (*pipelineConfig.Pipeline, error) {
	return impl.pipelineRepository.FindById(cdPipelineId)
}

func (impl PipelineBuilderImpl) FindAppAndEnvDetailsByPipelineId(cdPipelineId int) (*pipelineConfig.Pipeline, error) {
	return impl.pipelineRepository.FindAppAndEnvDetailsByPipelineId(cdPipelineId)
}

type TeamAppBean struct {
	ProjectId   int        `json:"projectId"`
	ProjectName string     `json:"projectName"`
	AppList     []*AppBean `json:"appList"`
}

type AppBean struct {
	Id     int    `json:"id"`
	Name   string `json:"name,notnull"`
	TeamId int    `json:"teamId,omitempty"`
}

func (impl PipelineBuilderImpl) GetAppList() ([]AppBean, error) {
	var appsRes []AppBean
	apps, err := impl.appRepo.FindAll()
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		appsRes = append(appsRes, AppBean{Id: app.Id, Name: app.AppName})
	}
	return appsRes, err
}

func (impl PipelineBuilderImpl) FetchCDPipelineStrategy(appId int) (PipelineStrategiesResponse, error) {
	pipelineStrategiesResponse := PipelineStrategiesResponse{}
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorf("invalid state", "err", err, "appId", appId)
		return pipelineStrategiesResponse, err
	}
	if chart.Id == 0 {
		return pipelineStrategiesResponse, fmt.Errorf("no chart configured")
	}

	//get global strategy for this chart
	globalStrategies, err := impl.globalStrategyMetadataChartRefMappingRepository.GetByChartRefId(chart.ChartRefId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting global strategies", "err", err)
		return pipelineStrategiesResponse, err
	} else if err == pg.ErrNoRows {
		impl.logger.Infow("no strategies configured for chart", "chartRefId", chart.ChartRefId)
		return pipelineStrategiesResponse, nil
	}
	pipelineOverride := chart.PipelineOverride
	for _, globalStrategy := range globalStrategies {
		pipelineStrategyJson, err := impl.filterDeploymentTemplate(globalStrategy.GlobalStrategyMetadata.Key, pipelineOverride)
		if err != nil {
			return pipelineStrategiesResponse, err
		}
		pipelineStrategy := PipelineStrategy{
			DeploymentTemplate: globalStrategy.GlobalStrategyMetadata.Name,
			Config:             []byte(pipelineStrategyJson),
		}
		pipelineStrategy.Default = globalStrategy.Default
		pipelineStrategiesResponse.PipelineStrategy = append(pipelineStrategiesResponse.PipelineStrategy, pipelineStrategy)
	}
	return pipelineStrategiesResponse, nil
}

type PipelineStrategiesResponse struct {
	PipelineStrategy []PipelineStrategy `json:"pipelineStrategy"`
}
type PipelineStrategy struct {
	DeploymentTemplate chartRepoRepository.DeploymentStrategy `json:"deploymentTemplate,omitempty"` //
	Config             json.RawMessage                        `json:"config"`
	Default            bool                                   `json:"default"`
}

func (impl PipelineBuilderImpl) GetEnvironmentByCdPipelineId(pipelineId int) (int, error) {
	dbPipeline, err := impl.pipelineRepository.FindById(pipelineId)
	if err != nil || dbPipeline == nil {
		impl.logger.Errorw("error in fetching pipeline", "err", err)
		return 0, err
	}
	return dbPipeline.EnvironmentId, err
}

func (impl PipelineBuilderImpl) GetCdPipelineById(pipelineId int) (cdPipeline *bean.CDPipelineConfigObject, err error) {
	dbPipeline, err := impl.pipelineRepository.FindById(pipelineId)
	if err != nil && errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching pipeline", "err", err)
		return cdPipeline, err
	}
	environment, err := impl.environmentRepository.FindById(dbPipeline.EnvironmentId)
	if err != nil && errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching pipeline", "err", err)
		return cdPipeline, err
	}
	strategies, err := impl.pipelineConfigRepository.GetAllStrategyByPipelineId(dbPipeline.Id)
	if err != nil && errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching strategies", "err", err)
		return cdPipeline, err
	}
	var strategiesBean []bean.Strategy
	var deploymentTemplate chartRepoRepository.DeploymentStrategy
	for _, item := range strategies {
		strategiesBean = append(strategiesBean, bean.Strategy{
			Config:             []byte(item.Config),
			DeploymentTemplate: item.Strategy,
			Default:            item.Default,
		})

		if item.Default {
			deploymentTemplate = item.Strategy
		}
	}

	preStage := bean.CdStage{}
	if len(dbPipeline.PreStageConfig) > 0 {
		preStage.Name = "Pre-Deployment"
		preStage.Config = dbPipeline.PreStageConfig
		preStage.TriggerType = dbPipeline.PreTriggerType
	}
	postStage := bean.CdStage{}
	if len(dbPipeline.PostStageConfig) > 0 {
		postStage.Name = "Post-Deployment"
		postStage.Config = dbPipeline.PostStageConfig
		postStage.TriggerType = dbPipeline.PostTriggerType
	}

	preStageConfigmapSecrets := bean.PreStageConfigMapSecretNames{}
	postStageConfigmapSecrets := bean.PostStageConfigMapSecretNames{}

	if dbPipeline.PreStageConfigMapSecretNames != "" {
		err = json.Unmarshal([]byte(dbPipeline.PreStageConfigMapSecretNames), &preStageConfigmapSecrets)
		if err != nil {
			impl.logger.Error(err)
			return nil, err
		}
	}
	if dbPipeline.PostStageConfigMapSecretNames != "" {
		err = json.Unmarshal([]byte(dbPipeline.PostStageConfigMapSecretNames), &postStageConfigmapSecrets)
		if err != nil {
			impl.logger.Error(err)
			return nil, err
		}
	}
	appWorkflowMapping, err := impl.appWorkflowRepository.FindWFCDMappingByCDPipelineId(pipelineId)
	if err != nil {
		return nil, err
	}
	cdPipeline = &bean.CDPipelineConfigObject{
		Id:                            dbPipeline.Id,
		Name:                          dbPipeline.Name,
		EnvironmentId:                 dbPipeline.EnvironmentId,
		EnvironmentName:               environment.Name,
		CiPipelineId:                  dbPipeline.CiPipelineId,
		DeploymentTemplate:            deploymentTemplate,
		TriggerType:                   dbPipeline.TriggerType,
		Strategies:                    strategiesBean,
		PreStage:                      preStage,
		PostStage:                     postStage,
		PreStageConfigMapSecretNames:  preStageConfigmapSecrets,
		PostStageConfigMapSecretNames: postStageConfigmapSecrets,
		RunPreStageInEnv:              dbPipeline.RunPreStageInEnv,
		RunPostStageInEnv:             dbPipeline.RunPostStageInEnv,
		CdArgoSetup:                   environment.Cluster.CdArgoSetup,
		ParentPipelineId:              appWorkflowMapping.ParentId,
		ParentPipelineType:            appWorkflowMapping.ParentType,
		DeploymentAppType:             dbPipeline.DeploymentAppType,
		DeploymentAppCreated:          dbPipeline.DeploymentAppCreated,
	}

	return cdPipeline, err
}

func (impl PipelineBuilderImpl) FindByIds(ids []*int) ([]*AppBean, error) {
	var appsRes []*AppBean
	apps, err := impl.appRepo.FindByIds(ids)
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		appsRes = append(appsRes, &AppBean{Id: app.Id, Name: app.AppName, TeamId: app.TeamId})
	}
	return appsRes, err
}

func (impl PipelineBuilderImpl) GetCiPipelineById(pipelineId int) (ciPipeline *bean.CiPipeline, err error) {
	pipeline, err := impl.ciPipelineRepository.FindById(pipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "pipelineId", pipelineId, "err", err)
		return nil, err
	}
	dockerArgs := make(map[string]string)
	if len(pipeline.DockerArgs) > 0 {
		err := json.Unmarshal([]byte(pipeline.DockerArgs), &dockerArgs)
		if err != nil {
			impl.logger.Warnw("error in unmarshal", "err", err)
		}
	}

	if impl.ciConfig.ExternalCiWebhookUrl == "" {
		hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
		if err != nil {
			impl.logger.Errorw("there is no external ci webhook url configured", "ci pipeline", pipeline)
			return nil, err
		}
		if hostUrl != nil {
			impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, ExternalCiWebhookPath)
		}
	}

	var externalCiConfig bean.ExternalCiConfig

	ciPipelineScripts, err := impl.ciPipelineRepository.FindCiScriptsByCiPipelineId(pipeline.Id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ci scripts")
		return nil, err
	}

	var beforeDockerBuildScripts []*bean.CiScript
	var afterDockerBuildScripts []*bean.CiScript
	for _, ciScript := range ciPipelineScripts {
		ciScriptResp := &bean.CiScript{
			Id:             ciScript.Id,
			Index:          ciScript.Index,
			Name:           ciScript.Name,
			Script:         ciScript.Script,
			OutputLocation: ciScript.OutputLocation,
		}
		if ciScript.Stage == BEFORE_DOCKER_BUILD {
			beforeDockerBuildScripts = append(beforeDockerBuildScripts, ciScriptResp)
		} else if ciScript.Stage == AFTER_DOCKER_BUILD {
			afterDockerBuildScripts = append(afterDockerBuildScripts, ciScriptResp)
		}
	}
	parentCiPipeline, err := impl.ciPipelineRepository.FindById(pipeline.ParentCiPipeline)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("err", err)
		return nil, err
	}
	ciPipeline = &bean.CiPipeline{
		Id:                       pipeline.Id,
		Version:                  pipeline.Version,
		Name:                     pipeline.Name,
		Active:                   pipeline.Active,
		Deleted:                  pipeline.Deleted,
		DockerArgs:               dockerArgs,
		IsManual:                 pipeline.IsManual,
		IsExternal:               pipeline.IsExternal,
		AppId:                    pipeline.AppId,
		ParentCiPipeline:         pipeline.ParentCiPipeline,
		ParentAppId:              parentCiPipeline.AppId,
		ExternalCiConfig:         externalCiConfig,
		BeforeDockerBuildScripts: beforeDockerBuildScripts,
		AfterDockerBuildScripts:  afterDockerBuildScripts,
		ScanEnabled:              pipeline.ScanEnabled,
		IsDockerConfigOverridden: pipeline.IsDockerConfigOverridden,
	}
	if !ciPipeline.IsExternal && ciPipeline.IsDockerConfigOverridden {
		ciTemplateBean, err := impl.ciTemplateService.FindTemplateOverrideByCiPipelineId(ciPipeline.Id)
		if err != nil {
			return nil, err
		}
		templateOverride := ciTemplateBean.CiTemplateOverride
		ciBuildConfig := ciTemplateBean.CiBuildConfig
		ciPipeline.DockerConfigOverride = bean.DockerConfigOverride{
			DockerRegistry:   templateOverride.DockerRegistryId,
			DockerRepository: templateOverride.DockerRepository,
			CiBuildConfig:    ciBuildConfig,
			//DockerBuildConfig: &bean.DockerBuildConfig{
			//	GitMaterialId:  templateOverride.GitMaterialId,
			//	DockerfilePath: templateOverride.DockerfilePath,
			//},
		}
	}
	for _, material := range pipeline.CiPipelineMaterials {
		if material == nil || material.GitMaterial == nil || !material.GitMaterial.Active {
			continue
		}
		ciMaterial := &bean.CiMaterial{
			Id:              material.Id,
			CheckoutPath:    material.CheckoutPath,
			Path:            material.Path,
			ScmId:           material.ScmId,
			GitMaterialId:   material.GitMaterialId,
			GitMaterialName: material.GitMaterial.Name[strings.Index(material.GitMaterial.Name, "-")+1:],
			ScmName:         material.ScmName,
			ScmVersion:      material.ScmVersion,
			IsRegex:         material.Regex != "",
			Source:          &bean.SourceTypeConfig{Type: material.Type, Value: material.Value, Regex: material.Regex},
		}
		ciPipeline.CiMaterial = append(ciPipeline.CiMaterial, ciMaterial)
	}

	linkedCis, err := impl.ciPipelineRepository.FindByParentCiPipelineId(ciPipeline.Id)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	ciPipeline.LinkedCount = len(linkedCis)

	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFCIMappingByCIPipelineId(ciPipeline.Id)
	for _, mapping := range appWorkflowMappings {
		//there will be only one active entry in db always
		ciPipeline.AppWorkflowId = mapping.AppWorkflowId
	}

	//getting pre stage and post stage details
	preStageDetail, postStageDetail, err := impl.pipelineStageService.GetCiPipelineStageData(ciPipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in getting pre & post stage detail by ciPipelineId", "err", err, "ciPipelineId", ciPipeline.Id)
		return nil, err
	}
	ciPipeline.PreBuildStage = preStageDetail
	ciPipeline.PostBuildStage = postStageDetail
	return ciPipeline, err
}

func (impl PipelineBuilderImpl) FindAllMatchesByAppName(appName string, appType helper.AppType) ([]*AppBean, error) {
	var appsRes []*AppBean
	var apps []*app2.App
	var err error
	if len(appName) == 0 {
		apps, err = impl.appRepo.FindAll()
	} else {
		apps, err = impl.appRepo.FindAllMatchesByAppName(appName, appType)
	}
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		name := app.AppName
		if appType == helper.Job {
			name = app.DisplayName
		}
		appsRes = append(appsRes, &AppBean{Id: app.Id, Name: name})
	}
	return appsRes, err
}

func (impl PipelineBuilderImpl) updateGitRepoUrlInCharts(appId int, chartGitAttribute *util.ChartGitAttribute, userId int32) error {
	charts, err := impl.chartRepository.FindActiveChartsByAppId(appId)
	if err != nil && pg.ErrNoRows != err {
		return err
	}
	for _, ch := range charts {
		if len(ch.GitRepoUrl) == 0 {
			ch.GitRepoUrl = chartGitAttribute.RepoUrl
			ch.ChartLocation = chartGitAttribute.ChartLocation
			ch.UpdatedOn = time.Now()
			ch.UpdatedBy = userId
			err = impl.chartRepository.Update(ch)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (impl PipelineBuilderImpl) PerformBulkActionOnCdPipelines(dto *bean.CdBulkActionRequestDto, impactedPipelines []*pipelineConfig.Pipeline, ctx context.Context, dryRun bool, userId int32) ([]*bean.CdBulkActionResponseDto, error) {
	switch dto.Action {
	case bean.CD_BULK_DELETE:
		bulkDeleteResp := impl.BulkDeleteCdPipelines(impactedPipelines, ctx, dryRun, dto.ForceDelete, userId)
		return bulkDeleteResp, nil
	default:
		return nil, &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "this action is not supported"}
	}
}

func (impl PipelineBuilderImpl) BulkDeleteCdPipelines(impactedPipelines []*pipelineConfig.Pipeline, ctx context.Context, dryRun, forceDelete bool, userId int32) []*bean.CdBulkActionResponseDto {
	var respDtos []*bean.CdBulkActionResponseDto
	for _, pipeline := range impactedPipelines {
		respDto := &bean.CdBulkActionResponseDto{
			PipelineName:    pipeline.Name,
			AppName:         pipeline.App.AppName,
			EnvironmentName: pipeline.Environment.Name,
		}
		if !dryRun {
			err := impl.DeleteCdPipeline(pipeline, ctx, forceDelete, true, userId)
			if err != nil {
				impl.logger.Errorw("error in deleting cd pipeline", "err", err, "pipelineId", pipeline.Id)
				respDto.DeletionResult = fmt.Sprintf("Not able to delete pipeline, %v", err)
			} else {
				respDto.DeletionResult = "Pipeline deleted successfully."
			}
		}
		respDtos = append(respDtos, respDto)
	}
	return respDtos

}

func (impl PipelineBuilderImpl) GetBulkActionImpactedPipelines(dto *bean.CdBulkActionRequestDto) ([]*pipelineConfig.Pipeline, error) {
	if len(dto.EnvIds) == 0 || (len(dto.AppIds) == 0 && len(dto.ProjectIds) == 0) {
		//invalid payload, envIds are must and either of appIds or projectIds are must
		return nil, &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "invalid payload, can not get pipelines for this filter"}
	}
	var pipelineIdsByAppLevel []int
	var pipelineIdsByProjectLevel []int
	var err error
	if len(dto.AppIds) > 0 && len(dto.EnvIds) > 0 {
		//getting pipeline IDs for app level deletion request
		pipelineIdsByAppLevel, err = impl.pipelineRepository.FindIdsByAppIdsAndEnvironmentIds(dto.AppIds, dto.EnvIds)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting cd pipelines by appIds and envIds", "err", err)
			return nil, err
		}
	}
	if len(dto.ProjectIds) > 0 && len(dto.EnvIds) > 0 {
		//getting pipeline IDs for project level deletion request
		pipelineIdsByProjectLevel, err = impl.pipelineRepository.FindIdsByProjectIdsAndEnvironmentIds(dto.ProjectIds, dto.EnvIds)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting cd pipelines by projectIds and envIds", "err", err)
			return nil, err
		}
	}
	var pipelineIdsMerged []int
	//it might be possible that pipelineIdsByAppLevel & pipelineIdsByProjectLevel have some same values
	//we are still appending them to save operation cost of checking same ids as we will get pipelines from
	//in clause which gives correct results even if some values are repeating
	pipelineIdsMerged = append(pipelineIdsMerged, pipelineIdsByAppLevel...)
	pipelineIdsMerged = append(pipelineIdsMerged, pipelineIdsByProjectLevel...)
	var pipelines []*pipelineConfig.Pipeline
	if len(pipelineIdsMerged) > 0 {
		pipelines, err = impl.pipelineRepository.FindByIdsIn(pipelineIdsMerged)
		if err != nil {
			impl.logger.Errorw("error in getting cd pipelines by ids", "err", err, "ids", pipelineIdsMerged)
			return nil, err
		}
	}
	return pipelines, nil
}

func (impl PipelineBuilderImpl) buildExternalCiWebhookSchema() map[string]interface{} {
	schema := make(map[string]interface{})
	schema["dockerImage"] = &bean.SchemaObject{Description: "docker image created for your application (Eg. quay.io/devtron/test:da3ba325-161-467)", DataType: "String", Example: "test-docker-repo/test:b150cc81-5-20", Optional: false}
	//schema["digest"] = &bean.SchemaObject{Description: "docker image sha1 digest", DataType: "String", Example: "sha256:94180dead8336237430e848ef8145f060b51", Optional: true}
	//schema["materialType"] = &bean.SchemaObject{Description: "git", DataType: "String", Example: "git", Optional: true}

	ciProjectDetails := make([]map[string]interface{}, 0)
	ciProjectDetail := make(map[string]interface{})
	ciProjectDetail["commitHash"] = &bean.SchemaObject{Description: "Hash of git commit used to build the image (Eg. 4bd84gba5ebdd6b1937ffd6c0734c2ad52ede782)", DataType: "String", Example: "dg46f67559dbsdfdfdfdsfba47901caf47f8b7e", Optional: true}
	ciProjectDetail["commitTime"] = &bean.SchemaObject{Description: "Time at which the code was committed to git (Eg. 2022-11-12T12:12:00)", DataType: "String", Example: "2022-11-12T12:12:00", Optional: true}
	ciProjectDetail["message"] = &bean.SchemaObject{Description: "Message provided during code commit (Eg. This is a sample commit message)", DataType: "String", Example: "commit message", Optional: true}
	ciProjectDetail["author"] = &bean.SchemaObject{Description: "Name or email id of the user who has done git commit (Eg. John Doe, johndoe@company.com)", DataType: "String", Example: "Devtron User", Optional: true}
	ciProjectDetails = append(ciProjectDetails, ciProjectDetail)

	schema["ciProjectDetails"] = &bean.SchemaObject{Description: "Git commit details used to build the image", DataType: "Array", Example: "[{}]", Optional: true, Child: ciProjectDetails}
	return schema
}

func (impl PipelineBuilderImpl) buildPayloadOption() []bean.PayloadOptionObject {
	payloadOption := make([]bean.PayloadOptionObject, 0)
	payloadOption = append(payloadOption, bean.PayloadOptionObject{
		Key:        "dockerImage",
		PayloadKey: []string{"dockerImage"},
		Label:      "Container image tag",
		Mandatory:  true,
	})

	payloadOption = append(payloadOption, bean.PayloadOptionObject{
		Key:        "commitHash",
		PayloadKey: []string{"ciProjectDetails.commitHash"},
		Label:      "Commit hash",
		Mandatory:  false,
	})
	payloadOption = append(payloadOption, bean.PayloadOptionObject{
		Key:        "message",
		PayloadKey: []string{"ciProjectDetails.message"},
		Label:      "Commit message",
		Mandatory:  false,
	})
	payloadOption = append(payloadOption, bean.PayloadOptionObject{
		Key:        "author",
		PayloadKey: []string{"ciProjectDetails.author"},
		Label:      "Author",
		Mandatory:  false,
	})
	payloadOption = append(payloadOption, bean.PayloadOptionObject{
		Key:        "commitTime",
		PayloadKey: []string{"ciProjectDetails.commitTime"},
		Label:      "Date & time of commit",
		Mandatory:  false,
	})
	return payloadOption
}

func (impl PipelineBuilderImpl) buildResponses() []bean.ResponseSchemaObject {
	responseSchemaObjects := make([]bean.ResponseSchemaObject, 0)
	schema := make(map[string]interface{})
	schema["code"] = &bean.SchemaObject{Description: "http status code", DataType: "integer", Example: "200,400,401", Optional: false}
	schema["result"] = &bean.SchemaObject{Description: "api response", DataType: "string", Example: "url", Optional: true}
	schema["status"] = &bean.SchemaObject{Description: "api response status", DataType: "string", Example: "url", Optional: true}

	error := make(map[string]interface{})
	error["code"] = &bean.SchemaObject{Description: "http status code", DataType: "integer", Example: "200,400,401", Optional: true}
	error["userMessage"] = &bean.SchemaObject{Description: "api error user message", DataType: "string", Example: "message", Optional: true}
	schema["error"] = &bean.SchemaObject{Description: "api error", DataType: "object", Example: "{}", Optional: true, Child: error}
	description200 := bean.ResponseDescriptionSchemaObject{
		Description: "success http api response",
		ExampleValue: bean.ExampleValueDto{
			Code:   200,
			Result: "api response result",
		},
		Schema: schema,
	}
	response200 := bean.ResponseSchemaObject{
		Description: description200,
		Code:        "200",
	}
	badReq := bean.ErrorDto{
		Code:        400,
		UserMessage: "Bad request",
	}
	description400 := bean.ResponseDescriptionSchemaObject{
		Description: "bad http request api response",
		ExampleValue: bean.ExampleValueDto{
			Code:   400,
			Errors: []bean.ErrorDto{badReq},
		},
		Schema: schema,
	}

	response400 := bean.ResponseSchemaObject{
		Description: description400,
		Code:        "400",
	}
	description401 := bean.ResponseDescriptionSchemaObject{
		Description: "unauthorized http api response",
		ExampleValue: bean.ExampleValueDto{
			Code:   401,
			Result: "Unauthorized",
		},
		Schema: schema,
	}
	response401 := bean.ResponseSchemaObject{
		Description: description401,
		Code:        "401",
	}
	responseSchemaObjects = append(responseSchemaObjects, response200)
	responseSchemaObjects = append(responseSchemaObjects, response400)
	responseSchemaObjects = append(responseSchemaObjects, response401)
	return responseSchemaObjects
}

func (impl PipelineBuilderImpl) MarkGitOpsDevtronAppsDeletedWhereArgoAppIsDeleted(appId int, envId int, acdToken string, pipeline *pipelineConfig.Pipeline) (bool, error) {

	acdAppFound := false
	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", acdToken)
	acdAppName := pipeline.DeploymentAppName
	_, err := impl.application.Get(ctx, &application2.ApplicationQuery{Name: &acdAppName})
	if err == nil {
		// acd app is not yet deleted so return
		acdAppFound = true
		return acdAppFound, err
	}
	impl.logger.Warnw("app not found in argo, deleting from db ", "err", err)
	//make call to delete it from pipeline DB because it's ACD counterpart is deleted
	err = impl.DeleteCdPipeline(pipeline, context.Background(), true, false, 0)
	if err != nil {
		impl.logger.Errorw("error in deleting cd pipeline", "err", err)
		return acdAppFound, err
	}
	return acdAppFound, nil
}

func (impl PipelineBuilderImpl) GetCiPipelineByEnvironment(request appGroup2.AppGroupingRequest) ([]*bean.CiConfigRequest, error) {
	ciPipelinesConfigByApps := make([]*bean.CiConfigRequest, 0)
	_, span := otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.AppGroupingCiPipelinesAuthorization")
	var cdPipelines []*pipelineConfig.Pipeline
	var err error
	if request.AppGroupId > 0 {
		appIds, err := impl.appGroupService.GetAppIdsByAppGroupId(request.AppGroupId)
		if err != nil {
			return nil, err
		}
		//override appIds if already provided app group id in request.
		request.AppIds = appIds
	}
	if len(request.AppIds) > 0 {
		cdPipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.EnvId, request.AppIds)
	} else {
		cdPipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.EnvId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return nil, err
	}

	var appIds []int
	ciPipelineIds := make([]int, 0)
	cdPipelineIds := make([]int, 0)
	for _, pipeline := range cdPipelines {
		cdPipelineIds = append(cdPipelineIds, pipeline.Id)
	}

	//authorization block starts here
	var appObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByDbPipeline(cdPipelines)
	ciPipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
	}
	appResults, _ := request.CheckAuthBatch(request.EmailId, appObjectArr, []string{})
	for _, pipeline := range cdPipelines {
		appObject := objects[pipeline.Id]
		if !appResults[appObject[0]] {
			//if user unauthorized, skip items
			continue
		}
		appIds = append(appIds, pipeline.AppId)
		ciPipelineIds = append(ciPipelineIds, pipeline.CiPipelineId)
	}
	//authorization block ends here
	span.End()
	if len(appIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching app found"}
		return nil, err
	}
	if impl.ciConfig.ExternalCiWebhookUrl == "" {
		hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
		if err != nil {
			return nil, err
		}
		if hostUrl != nil {
			impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, ExternalCiWebhookPath)
		}
	}

	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.GetCiTemplateVariables")
	defer span.End()
	ciPipelinesConfigMap, err := impl.getCiTemplateVariablesByAppIds(appIds)
	if err != nil {
		impl.logger.Debugw("error in fetching ci pipeline", "appIds", appIds, "err", err)
		return nil, err
	}

	ciPipelineByApp := make(map[int][]*pipelineConfig.CiPipeline)
	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.FindByAppIds")
	ciPipelines, err := impl.ciPipelineRepository.FindByAppIds(appIds)
	span.End()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "appIds", appIds, "err", err)
		return nil, err
	}
	parentCiPipelineIds := make([]int, 0)
	for _, ciPipeline := range ciPipelines {
		ciPipelineByApp[ciPipeline.AppId] = append(ciPipelineByApp[ciPipeline.AppId], ciPipeline)
		if ciPipeline.ParentCiPipeline > 0 && ciPipeline.IsExternal {
			parentCiPipelineIds = append(parentCiPipelineIds, ciPipeline.ParentCiPipeline)
		}
	}
	pipelineIdVsAppId, err := impl.ciPipelineRepository.FindAppIdsForCiPipelineIds(parentCiPipelineIds)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching appIds for pipelineIds", "parentCiPipelineIds", parentCiPipelineIds, "err", err)
		return nil, err
	}

	if len(ciPipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching ci pipeline found"}
		return nil, err
	}
	linkedCiPipelinesMap := make(map[int][]*pipelineConfig.CiPipeline)
	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.FindByParentCiPipelineIds")
	linkedCiPipelines, err := impl.ciPipelineRepository.FindByParentCiPipelineIds(ciPipelineIds)
	span.End()
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	for _, linkedCiPipeline := range linkedCiPipelines {
		linkedCiPipelinesMap[linkedCiPipeline.ParentCiPipeline] = append(linkedCiPipelinesMap[linkedCiPipeline.Id], linkedCiPipeline)
	}

	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.FindTemplateOverrideByCiPipelineIds")
	ciTemplateBeanOverrides, err := impl.ciTemplateService.FindTemplateOverrideByCiPipelineIds(ciPipelineIds)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching templates override", "appIds", appIds, "err", err)
		return nil, err
	}
	ciOverrideTemplateMap := make(map[int]*bean3.CiTemplateBean)
	for _, templateBeanOverride := range ciTemplateBeanOverrides {
		ciTemplateOverride := templateBeanOverride.CiTemplateOverride
		ciOverrideTemplateMap[ciTemplateOverride.CiPipelineId] = templateBeanOverride
	}

	var externalCiConfig bean.ExternalCiConfig
	//var parentCiPipelineIds []int
	for appId, ciPipelinesConfigByApp := range ciPipelinesConfigMap {
		var ciPipelineResp []*bean.CiPipeline

		ciPipelines := ciPipelineByApp[appId]
		for _, pipeline := range ciPipelines {
			dockerArgs := make(map[string]string)
			if len(pipeline.DockerArgs) > 0 {
				err := json.Unmarshal([]byte(pipeline.DockerArgs), &dockerArgs)
				if err != nil {
					impl.logger.Warnw("error in unmarshal", "err", err)
				}
			}
			parentCiPipelineId := pipeline.ParentCiPipeline
			ciPipeline := &bean.CiPipeline{
				Id:                       pipeline.Id,
				Version:                  pipeline.Version,
				Name:                     pipeline.Name,
				Active:                   pipeline.Active,
				Deleted:                  pipeline.Deleted,
				DockerArgs:               dockerArgs,
				IsManual:                 pipeline.IsManual,
				IsExternal:               pipeline.IsExternal,
				ParentCiPipeline:         parentCiPipelineId,
				ExternalCiConfig:         externalCiConfig,
				ScanEnabled:              pipeline.ScanEnabled,
				IsDockerConfigOverridden: pipeline.IsDockerConfigOverridden,
			}
			parentPipelineAppId, ok := pipelineIdVsAppId[parentCiPipelineId]
			if ok {
				ciPipeline.ParentAppId = parentPipelineAppId
			}
			if ciTemplateBean, ok := ciOverrideTemplateMap[pipeline.Id]; ok {
				templateOverride := ciTemplateBean.CiTemplateOverride
				ciPipeline.DockerConfigOverride = bean.DockerConfigOverride{
					DockerRegistry:   templateOverride.DockerRegistryId,
					DockerRepository: templateOverride.DockerRepository,
					CiBuildConfig:    ciTemplateBean.CiBuildConfig,
				}
			}

			//this will build ci materials for each ci pipeline
			for _, material := range pipeline.CiPipelineMaterials {
				// ignore those materials which have inactive git material
				if material == nil || material.GitMaterial == nil || !material.GitMaterial.Active {
					continue
				}
				ciMaterial := &bean.CiMaterial{
					Id:              material.Id,
					CheckoutPath:    material.CheckoutPath,
					Path:            material.Path,
					ScmId:           material.ScmId,
					GitMaterialId:   material.GitMaterialId,
					GitMaterialName: material.GitMaterial.Name[strings.Index(material.GitMaterial.Name, "-")+1:],
					ScmName:         material.ScmName,
					ScmVersion:      material.ScmVersion,
					IsRegex:         material.Regex != "",
					Source:          &bean.SourceTypeConfig{Type: material.Type, Value: material.Value, Regex: material.Regex},
				}
				ciPipeline.CiMaterial = append(ciPipeline.CiMaterial, ciMaterial)
			}

			//this will count the length of child ci pipelines, of each ci pipeline
			linkedCi := linkedCiPipelinesMap[pipeline.Id]
			ciPipeline.LinkedCount = len(linkedCi)
			ciPipelineResp = append(ciPipelineResp, ciPipeline)

			//this will use for fetch the parent ci pipeline app id, of each ci pipeline
			//parentCiPipelineIds = append(parentCiPipelineIds, pipeline.ParentCiPipeline)
		}
		ciPipelinesConfigByApp.CiPipelines = ciPipelineResp
		ciPipelinesConfigByApp.CiGitMaterialId = ciPipelinesConfigByApp.CiBuildConfig.GitMaterialId
		ciPipelinesConfigByApps = append(ciPipelinesConfigByApps, ciPipelinesConfigByApp)
	}

	return ciPipelinesConfigByApps, err
}

func (impl PipelineBuilderImpl) GetCiPipelineByEnvironmentMin(request appGroup2.AppGroupingRequest) ([]*bean.CiPipelineMinResponse, error) {
	results := make([]*bean.CiPipelineMinResponse, 0)
	var cdPipelines []*pipelineConfig.Pipeline
	var err error
	if request.AppGroupId > 0 {
		appIds, err := impl.appGroupService.GetAppIdsByAppGroupId(request.AppGroupId)
		if err != nil {
			return results, err
		}
		//override appIds if already provided app group id in request.
		request.AppIds = appIds
	}
	if len(request.AppIds) > 0 {
		cdPipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.EnvId, request.AppIds)
	} else {
		cdPipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.EnvId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return results, err
	}
	foundAppIds := make([]int, 0)
	for _, pipeline := range cdPipelines {
		foundAppIds = append(foundAppIds, pipeline.AppId)
	}
	if len(foundAppIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching pipeline found"}
		return nil, err
	}
	ciPipelines, err := impl.ciPipelineRepository.FindByAppIds(foundAppIds)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "err", err)
		return nil, err
	}
	ciPipelineByApp := make(map[int]*pipelineConfig.CiPipeline)
	parentCiPipelineIds := make([]int, 0)
	for _, ciPipeline := range ciPipelines {
		ciPipelineByApp[ciPipeline.Id] = ciPipeline
		if ciPipeline.ParentCiPipeline > 0 && ciPipeline.IsExternal {
			parentCiPipelineIds = append(parentCiPipelineIds, ciPipeline.ParentCiPipeline)
		}
	}
	pipelineIdVsAppId, err := impl.ciPipelineRepository.FindAppIdsForCiPipelineIds(parentCiPipelineIds)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching appIds for pipelineIds", "parentCiPipelineIds", parentCiPipelineIds, "err", err)
		return nil, err
	}

	//authorization block starts here
	var appObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByDbPipeline(cdPipelines)
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
	}
	appResults, _ := request.CheckAuthBatch(request.EmailId, appObjectArr, []string{})
	authorizedIds := make([]int, 0)
	for _, pipeline := range cdPipelines {
		appObject := objects[pipeline.Id]
		if !appResults[appObject[0]] {
			//if user unauthorized, skip items
			continue
		}
		if pipeline.CiPipelineId == 0 {
			//skip for external ci
			continue
		}
		ciPipeline := ciPipelineByApp[pipeline.CiPipelineId]
		parentAppId := pipelineIdVsAppId[ciPipeline.ParentCiPipeline]
		result := &bean.CiPipelineMinResponse{
			Id:               pipeline.CiPipelineId,
			AppId:            pipeline.AppId,
			AppName:          pipeline.App.AppName,
			ParentCiPipeline: ciPipeline.ParentCiPipeline,
			ParentAppId:      parentAppId,
		}
		results = append(results, result)
		authorizedIds = append(authorizedIds, pipeline.CiPipelineId)
	}
	//authorization block ends here

	return results, err
}

func (impl PipelineBuilderImpl) GetCdPipelinesByEnvironment(request appGroup2.AppGroupingRequest) (cdPipelines *bean.CdPipelines, err error) {
	_, span := otel.Tracer("orchestrator").Start(request.Ctx, "cdHandler.authorizationCdPipelinesForAppGrouping")
	if request.AppGroupId > 0 {
		appIds, err := impl.appGroupService.GetAppIdsByAppGroupId(request.AppGroupId)
		if err != nil {
			return nil, err
		}
		//override appIds if already provided app group id in request.
		request.AppIds = appIds
	}
	cdPipelines, err = impl.ciCdPipelineOrchestrator.GetCdPipelinesForEnv(request.EnvId, request.AppIds)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline", "err", err)
		return cdPipelines, err
	}
	pipelineIds := make([]int, 0)
	for _, pipeline := range cdPipelines.Pipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
	}
	if len(pipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching pipeline found"}
		return cdPipelines, err
	}
	//authorization block starts here
	var appObjectArr []string
	var envObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByPipeline(cdPipelines.Pipelines)
	pipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}
	appResults, envResults := request.CheckAuthBatch(request.EmailId, appObjectArr, envObjectArr)
	//authorization block ends here
	span.End()
	var pipelines []*bean.CDPipelineConfigObject
	authorizedPipelines := make(map[int]*bean.CDPipelineConfigObject)
	for _, dbPipeline := range cdPipelines.Pipelines {
		appObject := objects[dbPipeline.Id][0]
		envObject := objects[dbPipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
			//if user unauthorized, skip items
			continue
		}
		pipelineIds = append(pipelineIds, dbPipeline.Id)
		authorizedPipelines[dbPipeline.Id] = dbPipeline
	}

	pipelineDeploymentTemplate := make(map[int]chartRepoRepository.DeploymentStrategy)
	pipelineWorkflowMapping := make(map[int]*appWorkflow.AppWorkflowMapping)
	if len(pipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no authorized pipeline found"}
		return cdPipelines, err
	}
	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "cdHandler.GetAllStrategyByPipelineIds")
	strategies, err := impl.pipelineConfigRepository.GetAllStrategyByPipelineIds(pipelineIds)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching strategies", "err", err)
		return cdPipelines, err
	}
	for _, item := range strategies {
		if item.Default {
			pipelineDeploymentTemplate[item.PipelineId] = item.Strategy
		}
	}
	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "cdHandler.FindByCDPipelineIds")
	appWorkflowMappings, err := impl.appWorkflowRepository.FindByCDPipelineIds(pipelineIds)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching workflows", "err", err)
		return nil, err
	}
	for _, item := range appWorkflowMappings {
		pipelineWorkflowMapping[item.ComponentId] = item
	}

	for _, dbPipeline := range authorizedPipelines {
		pipeline := &bean.CDPipelineConfigObject{
			Id:                            dbPipeline.Id,
			Name:                          dbPipeline.Name,
			EnvironmentId:                 dbPipeline.EnvironmentId,
			EnvironmentName:               dbPipeline.EnvironmentName,
			CiPipelineId:                  dbPipeline.CiPipelineId,
			DeploymentTemplate:            pipelineDeploymentTemplate[dbPipeline.Id],
			TriggerType:                   dbPipeline.TriggerType,
			PreStage:                      dbPipeline.PreStage,
			PostStage:                     dbPipeline.PostStage,
			PreStageConfigMapSecretNames:  dbPipeline.PreStageConfigMapSecretNames,
			PostStageConfigMapSecretNames: dbPipeline.PostStageConfigMapSecretNames,
			RunPreStageInEnv:              dbPipeline.RunPreStageInEnv,
			RunPostStageInEnv:             dbPipeline.RunPostStageInEnv,
			DeploymentAppType:             dbPipeline.DeploymentAppType,
			ParentPipelineType:            pipelineWorkflowMapping[dbPipeline.Id].ParentType,
			ParentPipelineId:              pipelineWorkflowMapping[dbPipeline.Id].ParentId,
			AppName:                       dbPipeline.AppName,
			AppId:                         dbPipeline.AppId,
		}
		pipelines = append(pipelines, pipeline)
	}
	cdPipelines.Pipelines = pipelines
	return cdPipelines, err
}

func (impl PipelineBuilderImpl) GetCdPipelinesByEnvironmentMin(request appGroup2.AppGroupingRequest) (cdPipelines []*bean.CDPipelineConfigObject, err error) {
	_, span := otel.Tracer("orchestrator").Start(request.Ctx, "cdHandler.authorizationCdPipelinesForAppGrouping")
	if request.AppGroupId > 0 {
		appIds, err := impl.appGroupService.GetAppIdsByAppGroupId(request.AppGroupId)
		if err != nil {
			return cdPipelines, err
		}
		//override appIds if already provided app group id in request.
		request.AppIds = appIds
	}
	var pipelines []*pipelineConfig.Pipeline
	if len(request.AppIds) > 0 {
		pipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.EnvId, request.AppIds)
	} else {
		pipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.EnvId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return cdPipelines, err
	}
	//authorization block starts here
	var appObjectArr []string
	var envObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByDbPipeline(pipelines)
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}
	appResults, envResults := request.CheckAuthBatch(request.EmailId, appObjectArr, envObjectArr)
	//authorization block ends here
	span.End()
	for _, dbPipeline := range pipelines {
		appObject := objects[dbPipeline.Id][0]
		envObject := objects[dbPipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
			//if user unauthorized, skip items
			continue
		}
		pcObject := &bean.CDPipelineConfigObject{
			AppId:             dbPipeline.AppId,
			AppName:           dbPipeline.App.AppName,
			EnvironmentId:     dbPipeline.EnvironmentId,
			Id:                dbPipeline.Id,
			DeploymentAppType: dbPipeline.DeploymentAppType,
		}
		cdPipelines = append(cdPipelines, pcObject)
	}
	return cdPipelines, err
}

func (impl PipelineBuilderImpl) GetExternalCiByEnvironment(request appGroup2.AppGroupingRequest) (ciConfig []*bean.ExternalCiConfig, err error) {
	_, span := otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.authorizationExternalCiForAppGrouping")
	externalCiConfigs := make([]*bean.ExternalCiConfig, 0)
	var cdPipelines []*pipelineConfig.Pipeline
	if request.AppGroupId > 0 {
		appIds, err := impl.appGroupService.GetAppIdsByAppGroupId(request.AppGroupId)
		if err != nil {
			return nil, err
		}
		//override appIds if already provided app group id in request.
		request.AppIds = appIds
	}
	if len(request.AppIds) > 0 {
		cdPipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.EnvId, request.AppIds)
	} else {
		cdPipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.EnvId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return nil, err
	}

	var appIds []int
	//authorization block starts here
	var appObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByDbPipeline(cdPipelines)
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
	}
	appResults, _ := request.CheckAuthBatch(request.EmailId, appObjectArr, []string{})
	for _, pipeline := range cdPipelines {
		appObject := objects[pipeline.Id]
		if !appResults[appObject[0]] {
			//if user unauthorized, skip items
			continue
		}
		//add only those who have external ci
		if pipeline.CiPipelineId == 0 {
			appIds = append(appIds, pipeline.AppId)
		}
	}

	//authorization block ends here
	span.End()

	if len(appIds) == 0 {
		impl.logger.Warnw("there is no app id found for fetching external ci pipelines", "request", request)
		return externalCiConfigs, nil
	}
	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.FindExternalCiByAppIds")
	externalCiPipelines, err := impl.ciPipelineRepository.FindExternalCiByAppIds(appIds)
	span.End()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching external ci", "request", request, "err", err)
		return nil, err
	}
	hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
	if err != nil {
		impl.logger.Errorw("error in fetching external ci", "request", request, "err", err)
		return nil, err
	}
	if hostUrl != nil {
		impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, ExternalCiWebhookPath)
	}

	var externalCiPipelineIds []int
	appWorkflowMappingsMap := make(map[int][]*appWorkflow.AppWorkflowMapping)

	for _, externalCiPipeline := range externalCiPipelines {
		externalCiPipelineIds = append(externalCiPipelineIds, externalCiPipeline.Id)
	}
	if len(externalCiPipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no external ci pipeline found"}
		return externalCiConfigs, err
	}
	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFCDMappingByExternalCiIdByIdsIn(externalCiPipelineIds)
	if err != nil {
		impl.logger.Errorw("Error in fetching app workflow mapping for CD pipeline by external CI ID", "err", err)
		return nil, err
	}

	CDPipelineMap := make(map[int]*pipelineConfig.Pipeline)
	appIdMap := make(map[int]*app2.App)
	var componentIds []int
	for _, appWorkflowMapping := range appWorkflowMappings {
		appWorkflowMappingsMap[appWorkflowMapping.ParentId] = append(appWorkflowMappingsMap[appWorkflowMapping.ParentId], appWorkflowMapping)
		componentIds = append(componentIds, appWorkflowMapping.ComponentId)
	}
	if len(componentIds) == 0 {
		return nil, err
	}
	cdPipelines, err = impl.pipelineRepository.FindAppAndEnvironmentAndProjectByPipelineIds(componentIds)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching external ci", "request", request, "err", err)
		return nil, err
	}
	for _, pipeline := range cdPipelines {
		CDPipelineMap[pipeline.Id] = pipeline
		appIds = append(appIds, pipeline.AppId)
	}
	if len(appIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching apps found"}
		return nil, err
	}
	apps, err := impl.appRepo.FindAppAndProjectByIdsIn(appIds)
	for _, app := range apps {
		appIdMap[app.Id] = app
	}

	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.FindAppAndEnvironmentAndProjectByPipelineIds")
	for _, externalCiPipeline := range externalCiPipelines {
		externalCiConfig := &bean.ExternalCiConfig{
			Id:         externalCiPipeline.Id,
			WebhookUrl: fmt.Sprintf("%s/%d", impl.ciConfig.ExternalCiWebhookUrl, externalCiPipeline.Id),
			Payload:    impl.ciConfig.ExternalCiPayload,
			AccessKey:  "",
		}

		if _, ok := appWorkflowMappingsMap[externalCiPipeline.Id]; !ok {
			return nil, errors.New("Error in fetching app workflow mapping for cd pipeline by parent id")
		}
		appWorkflowMappings := appWorkflowMappingsMap[externalCiPipeline.Id]
		roleData := make(map[string]interface{})
		for _, appWorkflowMapping := range appWorkflowMappings {
			if _, ok := CDPipelineMap[appWorkflowMapping.ComponentId]; !ok {
				impl.logger.Errorw("error in getting cd pipeline data for workflow", "app workflow id", appWorkflowMapping.ComponentId, "err", err)
				return nil, errors.New("error in getting cd pipeline data for workflow")
			}
			cdPipeline := CDPipelineMap[appWorkflowMapping.ComponentId]
			if _, ok := roleData[teamIdKey]; !ok {
				if _, ok := appIdMap[cdPipeline.AppId]; !ok {
					impl.logger.Errorw("error in getting app data for pipeline", "app id", cdPipeline.AppId, "err", err)
					return nil, errors.New("error in getting app data for pipeline")
				}
				app := appIdMap[cdPipeline.AppId]
				roleData[teamIdKey] = app.TeamId
				roleData[teamNameKey] = app.Team.Name
				roleData[appIdKey] = cdPipeline.AppId
				roleData[appNameKey] = cdPipeline.App.AppName
			}
			if _, ok := roleData[environmentNameKey]; !ok {
				roleData[environmentNameKey] = cdPipeline.Environment.Name
			} else {
				roleData[environmentNameKey] = fmt.Sprintf("%s,%s", roleData[environmentNameKey], cdPipeline.Environment.Name)
			}
			if _, ok := roleData[environmentIdentifierKey]; !ok {
				roleData[environmentIdentifierKey] = cdPipeline.Environment.EnvironmentIdentifier
			} else {
				roleData[environmentIdentifierKey] = fmt.Sprintf("%s,%s", roleData[environmentIdentifierKey], cdPipeline.Environment.EnvironmentIdentifier)
			}
		}

		externalCiConfig.ExternalCiConfigRole = bean.ExternalCiConfigRole{
			ProjectId:             roleData[teamIdKey].(int),
			ProjectName:           roleData[teamNameKey].(string),
			AppId:                 roleData[appIdKey].(int),
			AppName:               roleData[appNameKey].(string),
			EnvironmentName:       roleData[environmentNameKey].(string),
			EnvironmentIdentifier: roleData[environmentIdentifierKey].(string),
			Role:                  "Build and deploy",
		}
		externalCiConfigs = append(externalCiConfigs, externalCiConfig)
	}
	span.End()
	//--------pipeline population end
	return externalCiConfigs, err
}

func (impl PipelineBuilderImpl) GetEnvironmentListForAutocompleteFilter(envName string, clusterIds []int, offset int, size int, emailId string, checkAuthBatch func(emailId string, appObject []string, envObject []string) (map[string]bool, map[string]bool), ctx context.Context) (*cluster.AppGroupingResponse, error) {
	result := &cluster.AppGroupingResponse{}
	var models []*repository2.Environment
	var beans []cluster.EnvironmentBean
	var err error
	if len(envName) > 0 && len(clusterIds) > 0 {
		models, err = impl.environmentRepository.FindByEnvNameAndClusterIds(envName, clusterIds)
	} else if len(clusterIds) > 0 {
		models, err = impl.environmentRepository.FindByClusterIdsWithFilter(clusterIds)
	} else if len(envName) > 0 {
		models, err = impl.environmentRepository.FindByEnvName(envName)
	} else {
		models, err = impl.environmentRepository.FindAllActiveWithFilter()
	}
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching environment", "err", err)
		return result, err
	}
	var envIds []int
	for _, model := range models {
		envIds = append(envIds, model.Id)
	}
	if len(envIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching environment found"}
		return nil, err
	}
	_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineBuilder.FindActiveByEnvIds")
	cdPipelines, err := impl.pipelineRepository.FindActiveByEnvIds(envIds)
	span.End()
	if err != nil && err != pg.ErrNoRows {
		return result, err
	}
	pipelineIds := make([]int, 0)
	for _, pipeline := range cdPipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
	}
	if len(pipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching pipeline found"}
		return nil, err
	}
	//authorization block starts here
	var appObjectArr []string
	var envObjectArr []string
	_, span = otel.Tracer("orchestrator").Start(ctx, "pipelineBuilder.GetAppAndEnvObjectByPipelineIds")
	objects := impl.enforcerUtil.GetAppAndEnvObjectByPipelineIds(pipelineIds)
	span.End()
	pipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}
	_, span = otel.Tracer("orchestrator").Start(ctx, "pipelineBuilder.checkAuthBatch")
	appResults, envResults := checkAuthBatch(emailId, appObjectArr, envObjectArr)
	span.End()
	//authorization block ends here

	pipelinesMap := make(map[int][]*pipelineConfig.Pipeline)
	for _, pipeline := range cdPipelines {
		appObject := objects[pipeline.Id][0]
		envObject := objects[pipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
			//if user unauthorized, skip items
			continue
		}
		pipelinesMap[pipeline.EnvironmentId] = append(pipelinesMap[pipeline.EnvironmentId], pipeline)
	}
	for _, model := range models {
		environment := cluster.EnvironmentBean{
			Id:                    model.Id,
			Environment:           model.Name,
			Namespace:             model.Namespace,
			CdArgoSetup:           model.Cluster.CdArgoSetup,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
			ClusterName:           model.Cluster.ClusterName,
		}

		//authorization block starts here
		appCount := 0
		envPipelines := pipelinesMap[model.Id]
		if _, ok := pipelinesMap[model.Id]; ok {
			appCount = len(envPipelines)
		}
		environment.AppCount = appCount
		beans = append(beans, environment)
	}

	envCount := len(beans)
	// Apply pagination
	if size > 0 {
		if offset+size <= len(beans) {
			beans = beans[offset : offset+size]
		} else {
			beans = beans[offset:]
		}
	}
	result.EnvList = beans
	result.EnvCount = envCount
	return result, nil
}

func (impl PipelineBuilderImpl) GetAppListForEnvironment(request appGroup2.AppGroupingRequest) ([]*AppBean, error) {
	var applicationList []*AppBean
	var cdPipelines []*pipelineConfig.Pipeline
	var err error
	if request.AppGroupId > 0 {
		appIds, err := impl.appGroupService.GetAppIdsByAppGroupId(request.AppGroupId)
		if err != nil {
			return nil, err
		}
		//override appIds if already provided app group id in request.
		request.AppIds = appIds
	}
	if len(request.AppIds) > 0 {
		cdPipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.EnvId, request.AppIds)
	} else {
		cdPipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.EnvId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return nil, err
	}
	if len(cdPipelines) == 0 {
		return applicationList, nil
	}
	var appObjectArr []string
	var envObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByDbPipeline(cdPipelines)
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}
	appResults, envResults := request.CheckAuthBatch(request.EmailId, appObjectArr, envObjectArr)
	for _, pipeline := range cdPipelines {
		appObject := objects[pipeline.Id][0]
		envObject := objects[pipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
			//if user unauthorized, skip items
			continue
		}
		applicationList = append(applicationList, &AppBean{Id: pipeline.AppId, Name: pipeline.App.AppName})
	}
	return applicationList, err
}
