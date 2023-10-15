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
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	"net/url"
	"strings"
	"time"

	"github.com/caarlos0/env"
	util4 "github.com/devtron-labs/common-lib/utils/k8s"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/chart"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	resourceGroup2 "github.com/devtron-labs/devtron/pkg/resourceGroup"
	"github.com/devtron-labs/devtron/pkg/user"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
)

const DashboardConfigMap = "dashboard-cm"
const SECURITY_SCANNING = "FORCE_SECURITY_SCANNING"

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

type SecurityConfig struct {
	//FORCE_SECURITY_SCANNING flag is being maintained in both dashboard and orchestrator CM's
	//TODO: rishabh will remove FORCE_SECURITY_SCANNING from dashboard's CM.
	ForceSecurityScanning bool `env:"FORCE_SECURITY_SCANNING" envDefault:"false"`
}

func GetDeploymentServiceTypeConfig() (*DeploymentServiceTypeConfig, error) {
	cfg := &DeploymentServiceTypeConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

type DevtronAppConfigService interface {
	//CreateApp : This function creates applications of type Job as well as Devtronapps
	// In case of error response object is nil
	CreateApp(request *bean.CreateAppDTO) (*bean.CreateAppDTO, error)
	//DeleteApp : This function deletes applications of type Job as well as DevtronApps
	DeleteApp(appId int, userId int32) error
	//GetApp : Gets Application along with Git materials for given appId.
	//If the application type is a 'Chart Store App', it doesnt provide any detail.
	//For application types like Jobs and DevtronApps, it retrieves Git materials associated with the application.
	//In case of error response object is nil
	GetApp(appId int) (application *bean.CreateAppDTO, err error)
	//FindByIds : Find applications by given IDs, delegating the request to the appRepository.
	// It queries the repository for applications corresponding to the given IDs and constructs
	//a list of AppBean objects containing ID, name, and team ID.
	//It returns the list of AppBean instances.
	//In case of error,AppBean is returned as nil.
	FindByIds(ids []*int) ([]*AppBean, error)
	//GetAppList : Retrieve and return a list of applications after converting in proper bean object.
	//In case of any error , []AppBean is returned as nil.
	GetAppList() ([]AppBean, error)
	//FindAllMatchesByAppName : Find and return applications matching the given name and type.
	//Internally,It performs a case-insensitive search based on the applicationName("%"+appName+"%") and type.
	//In case of error,[]*AppBean is returned as nil.
	FindAllMatchesByAppName(appName string, appType helper.AppType) ([]*AppBean, error)
	//GetAppListForEnvironment : Retrieves a list of applications (AppBean) based on the provided ResourceGroupingRequest.
	// It first determines the relevant application and environment objects based on the active pipelines fetched from the repository.
	//The function then performs authorization checks on these objects for the given user.
	//Finally , the corresponding AppBean objects are added to the applicationList and then returned.
	//In case of error,[]*AppBean is returned as nil.
	GetAppListForEnvironment(request resourceGroup2.ResourceGroupingRequest) ([]*AppBean, error)
	//FindAppsByTeamId : Retrieves applications (AppBean) associated with the provided teamId
	//It queries the repository for applications belonging to the specified team(project) and
	//constructs a list of AppBean instances containing ID and name.
	//The function returns the list of applications in valid case.
	//In case of error,[]*AppBean is returned as nil.
	FindAppsByTeamId(teamId int) ([]*AppBean, error)
	//FindAppsByTeamName : Retrieves applications (AppBean) associated with the provided teamName
	// It queries the repository for applications belonging to the specified team(project) and
	// constructs a list of AppBean instances containing ID and name.
	// The function returns the list of applications in valid case.
	// In case of error,[]*AppBean is returned as nil.
	FindAppsByTeamName(teamName string) ([]AppBean, error)
}

type PipelineBuilder interface {
	DevtronAppConfigService
	CiPipelineConfigService
	CiMaterialConfigService
	AppArtifactManager
	CdPipelineConfigService
	DevtronAppCMCSService
	DevtronAppStrategyService
	AppDeploymentTypeChangeManager
	CdPipelineConfigService
	//TODO: remove this once this is uncommented in AppArtifactManager
	//RetrieveArtifactsByCDPipeline(pipeline *pipelineConfig.Pipeline, stage bean2.WorkflowType) (*bean.CiArtifactResponse, error)
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
	clusterRepository                repository2.ClusterRepository
	pipelineConfigRepository         chartConfig.PipelineConfigRepository
	mergeUtil                        util.MergeUtil
	appWorkflowRepository            appWorkflow.AppWorkflowRepository
	ciConfig                         *CiCdConfig
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
	resourceGroupService                            resourceGroup2.ResourceGroupService
	chartDeploymentService                          util.ChartDeploymentService
	K8sUtil                                         *util4.K8sUtil
	attributesRepository                            repository.AttributesRepository
	securityConfig                                  *SecurityConfig
	imageTaggingService                             ImageTaggingService
	variableEntityMappingService                    variables.VariableEntityMappingService
	variableTemplateParser                          parsers.VariableTemplateParser
	CiPipelineConfigService
	CiMaterialConfigService
	AppArtifactManager
	DevtronAppCMCSService
	DevtronAppStrategyService
	AppDeploymentTypeChangeManager
	CdPipelineConfigService
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
	clusterRepository repository2.ClusterRepository,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	mergeUtil util.MergeUtil,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	ciConfig *CiCdConfig,
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
	resourceGroupService resourceGroup2.ResourceGroupService,
	chartDeploymentService util.ChartDeploymentService,
	K8sUtil *util4.K8sUtil,
	attributesRepository repository.AttributesRepository,
	imageTaggingService ImageTaggingService,
	variableEntityMappingService variables.VariableEntityMappingService,
	variableTemplateParser parsers.VariableTemplateParser,
	ciPipelineConfigService CiPipelineConfigService,
	ciMaterialConfigService CiMaterialConfigService,
	appArtifactManager AppArtifactManager,
	devtronAppCMCSService DevtronAppCMCSService,
	devtronAppStrategyService DevtronAppStrategyService,
	appDeploymentTypeChangeManager AppDeploymentTypeChangeManager,
	cdPipelineConfigService CdPipelineConfigService) *PipelineBuilderImpl {

	securityConfig := &SecurityConfig{}
	err := env.Parse(securityConfig)
	if err != nil {
		logger.Errorw("error in parsing securityConfig,setting  ForceSecurityScanning to default value", "defaultValue", securityConfig.ForceSecurityScanning, "err", err)
	}
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
		clusterRepository:                clusterRepository,
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
		resourceGroupService:                            resourceGroupService,
		chartDeploymentService:                          chartDeploymentService,
		K8sUtil:                                         K8sUtil,
		attributesRepository:                            attributesRepository,
		securityConfig:                                  securityConfig,
		imageTaggingService:                             imageTaggingService,
		variableEntityMappingService:                    variableEntityMappingService,
		variableTemplateParser:                          variableTemplateParser,
		CiPipelineConfigService:                         ciPipelineConfigService,
		CiMaterialConfigService:                         ciMaterialConfigService,
		AppArtifactManager:                              appArtifactManager,
		DevtronAppCMCSService:                           devtronAppCMCSService,
		DevtronAppStrategyService:                       devtronAppStrategyService,
		AppDeploymentTypeChangeManager:                  appDeploymentTypeChangeManager,
		CdPipelineConfigService:                         cdPipelineConfigService,
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

/*
   1. create pipelineGroup
   2. save material (add credential provider support)

*/

func (impl *PipelineBuilderImpl) CreateApp(request *bean.CreateAppDTO) (*bean.CreateAppDTO, error) {
	impl.logger.Debugw("app create request received", "req", request)

	res, err := impl.ciCdPipelineOrchestrator.CreateApp(request)
	if err != nil {
		impl.logger.Errorw("error in saving create app req", "req", request, "err", err)
	}
	return res, err
}

func (impl *PipelineBuilderImpl) DeleteApp(appId int, userId int32) error {
	impl.logger.Debugw("app delete request received", "app", appId)
	err := impl.ciCdPipelineOrchestrator.DeleteApp(appId, userId)
	return err
}

func (impl *PipelineBuilderImpl) GetApp(appId int) (application *bean.CreateAppDTO, err error) {
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

func (impl *PipelineBuilderImpl) FindByIds(ids []*int) ([]*AppBean, error) {
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

func (impl *PipelineBuilderImpl) GetAppList() ([]AppBean, error) {
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

func (impl *PipelineBuilderImpl) FindAllMatchesByAppName(appName string, appType helper.AppType) ([]*AppBean, error) {
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

func (impl PipelineBuilderImpl) GetAppListForEnvironment(request resourceGroup2.ResourceGroupingRequest) ([]*AppBean, error) {
	var applicationList []*AppBean
	var cdPipelines []*pipelineConfig.Pipeline
	var err error
	if request.ResourceGroupId > 0 {
		appIds, err := impl.resourceGroupService.GetResourceIdsByResourceGroupId(request.ResourceGroupId)
		if err != nil {
			return nil, err
		}
		//override appIds if already provided app group id in request.
		request.ResourceIds = appIds
	}
	if len(request.ResourceIds) > 0 {
		cdPipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.ParentResourceId, request.ResourceIds)
	} else {
		cdPipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.ParentResourceId)
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

func (impl *PipelineBuilderImpl) FindAppsByTeamId(teamId int) ([]*AppBean, error) {
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

func (impl *PipelineBuilderImpl) FindAppsByTeamName(teamName string) ([]AppBean, error) {
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

func (impl *PipelineBuilderImpl) getGitMaterialsForApp(appId int) ([]*bean.GitMaterial, error) {
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

func getPatchStatus(err error) bean.CiPatchStatus {
	if err != nil {
		if err.Error() == string(bean.CI_PATCH_NOT_AUTHORIZED_MESSAGE) {
			return bean.CI_PATCH_NOT_AUTHORIZED
		}
		return bean.CI_PATCH_FAILED
	}
	return bean.CI_PATCH_SUCCESS
}

func getPatchMessage(err error) bean.CiPatchMessage {
	if err != nil {
		return bean.CiPatchMessage(err.Error())
	}
	return ""
}

func allDeploymentConfigTrue(deploymentConfig map[string]bool) bool {
	for _, value := range deploymentConfig {
		if !value {
			return false
		}
	}
	return true
}

func validDeploymentConfigReceived(deploymentConfig map[string]bool, deploymentTypeSent string) bool {
	for key, value := range deploymentConfig {
		if value && key == deploymentTypeSent {
			return true
		}
	}
	return false
}

func (impl *PipelineBuilderImpl) isGitRepoUrlPresent(appId int) bool {
	fetchedChart, err := impl.chartRepository.FindLatestByAppId(appId)

	if err != nil || len(fetchedChart.GitRepoUrl) == 0 {
		impl.logger.Errorw("error fetching git repo url or it is not present")
		return false
	}
	return true
}

type DeploymentType struct {
	Deployment Deployment `json:"deployment"`
}

type Deployment struct {
	Strategy map[string]interface{} `json:"strategy"`
}

type ConfigMapSecretsResponse struct {
	Maps    []bean2.ConfigSecretMap `json:"maps"`
	Secrets []bean2.ConfigSecretMap `json:"secrets"`
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

type PipelineStrategiesResponse struct {
	PipelineStrategy []PipelineStrategy `json:"pipelineStrategy"`
}
type PipelineStrategy struct {
	DeploymentTemplate chartRepoRepository.DeploymentStrategy `json:"deploymentTemplate,omitempty"` //
	Config             json.RawMessage                        `json:"config"`
	Default            bool                                   `json:"default"`
}

func checkAppReleaseNotExist(err error) bool {
	// RELEASE_NOT_EXIST check for helm App and NOT_FOUND check for argo app
	return strings.Contains(err.Error(), bean.NOT_FOUND) || strings.Contains(err.Error(), bean.RELEASE_NOT_EXIST)
}
