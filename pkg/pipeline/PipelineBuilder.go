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
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	repository6 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/caarlos0/env"
	util4 "github.com/devtron-labs/common-lib/utils/k8s"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/sql/models"
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
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
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
	variableTemplateParser parsers.VariableTemplateParser) *PipelineBuilderImpl {

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

func (impl *PipelineBuilderImpl) getDefaultArtifactStore(id string) (store *dockerRegistryRepository.DockerArtifactStore, err error) {
	if id == "" {
		impl.logger.Debugw("docker repo is empty adding default repo")
		store, err = impl.dockerArtifactStoreRepository.FindActiveDefaultStore()

	} else {
		store, err = impl.dockerArtifactStoreRepository.FindOne(id)
	}
	return
}

func (impl *PipelineBuilderImpl) getCiTemplateVariables(appId int) (ciConfig *bean.CiConfigRequest, err error) {
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

func (impl *PipelineBuilderImpl) getCiTemplateVariablesByAppIds(appIds []int) (map[int]*bean.CiConfigRequest, error) {
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

func (impl *PipelineBuilderImpl) addpipelineToTemplate(createRequest *bean.CiConfigRequest) (resp *bean.CiConfigRequest, err error) {

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

func (impl *PipelineBuilderImpl) patchCiPipelineUpdateSource(baseCiConfig *bean.CiConfigRequest, modifiedCiPipeline *bean.CiPipeline) (ciConfig *bean.CiConfigRequest, err error) {

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

func (impl *PipelineBuilderImpl) ValidateCDPipelineRequest(pipelineCreateRequest *bean.CdPipelines, isGitOpsConfigured, haveAtleastOneGitOps bool) (bool, error) {

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

func (impl *PipelineBuilderImpl) RegisterInACD(gitOpsRepoName string, chartGitAttr *util.ChartGitAttribute, userId int32, ctx context.Context) error {

	err := impl.chartDeploymentService.RegisterInArgo(chartGitAttr, ctx)
	if err != nil {
		impl.logger.Errorw("error while register git repo in argo", "err", err)
		emptyRepoErrorMessage := []string{"failed to get index: 404 Not Found", "remote repository is empty"}
		if strings.Contains(err.Error(), emptyRepoErrorMessage[0]) || strings.Contains(err.Error(), emptyRepoErrorMessage[1]) {
			// - found empty repository, create some file in repository
			err := impl.chartTemplateService.CreateReadmeInGitRepo(gitOpsRepoName, userId)
			if err != nil {
				impl.logger.Errorw("error in creating file in git repo", "err", err)
				return err
			}
			// - retry register in argo
			err = impl.chartDeploymentService.RegisterInArgo(chartGitAttr, ctx)
			if err != nil {
				impl.logger.Errorw("error in re-try register in argo", "err", err)
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (impl *PipelineBuilderImpl) validateDeploymentAppType(pipeline *bean.CDPipelineConfigObject, deploymentConfig map[string]bool) error {

	// Config value doesn't exist in attribute table
	if deploymentConfig == nil {
		return nil
	}
	//Config value found to be true for ArgoCD and Helm both
	if allDeploymentConfigTrue(deploymentConfig) {
		return nil
	}
	//Case : {ArgoCD : false, Helm: true, HGF : true}
	if validDeploymentConfigReceived(deploymentConfig, pipeline.DeploymentAppType) {
		return nil
	}

	err := &util.ApiError{
		HttpStatusCode:  http.StatusBadRequest,
		InternalMessage: "Received deployment app type doesn't match with the allowed deployment app type for this environment.",
		UserMessage:     "Received deployment app type doesn't match with the allowed deployment app type for this environment.",
	}
	return err
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

func (impl *PipelineBuilderImpl) DeleteCdPipelinePartial(pipeline *pipelineConfig.Pipeline, ctx context.Context, deleteAction int, userId int32) (*bean.AppDeleteResponseDTO, error) {
	cascadeDelete := true
	forceDelete := false
	deleteResponse := &bean.AppDeleteResponseDTO{
		DeleteInitiated:  false,
		ClusterReachable: true,
	}
	if deleteAction == bean.FORCE_DELETE {
		forceDelete = true
		cascadeDelete = false
	} else if deleteAction == bean.NON_CASCADE_DELETE {
		cascadeDelete = false
	}
	//Updating clusterReachable flag
	clusterBean, err := impl.clusterRepository.FindById(pipeline.Environment.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster details", "err", err, "clusterId", pipeline.Environment.ClusterId)
	}
	deleteResponse.ClusterName = clusterBean.ClusterName
	if len(clusterBean.ErrorInConnecting) > 0 {
		deleteResponse.ClusterReachable = false
	}
	//getting children CD pipeline details
	childNodes, err := impl.appWorkflowRepository.FindWFCDMappingByParentCDPipelineId(pipeline.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting children cd details", "err", err)
		return deleteResponse, err
	} else if len(childNodes) > 0 {
		impl.logger.Debugw("cannot delete cd pipeline, contains children cd")
		return deleteResponse, fmt.Errorf("Please delete children CD pipelines before deleting this pipeline.")
	}
	//getting deployment group for this pipeline
	deploymentGroupNames, err := impl.deploymentGroupRepository.GetNamesByAppIdAndEnvId(pipeline.EnvironmentId, pipeline.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting deployment group names by appId and envId", "err", err)
		return deleteResponse, err
	} else if len(deploymentGroupNames) > 0 {
		groupNamesByte, err := json.Marshal(deploymentGroupNames)
		if err != nil {
			impl.logger.Errorw("error in marshaling deployment group names", "err", err, "deploymentGroupNames", deploymentGroupNames)
		}
		impl.logger.Debugw("cannot delete cd pipeline, is being used in deployment group")
		return deleteResponse, fmt.Errorf("Please remove this CD pipeline from deployment groups : %s", string(groupNamesByte))
	}
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return deleteResponse, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	//delete app from argo cd, if created
	if pipeline.DeploymentAppCreated && !pipeline.DeploymentAppDeleteRequest {
		deploymentAppName := fmt.Sprintf("%s-%s", pipeline.App.AppName, pipeline.Environment.Name)
		if util.IsAcdApp(pipeline.DeploymentAppType) {
			if !deleteResponse.ClusterReachable {
				impl.logger.Errorw("cluster connection error", "err", clusterBean.ErrorInConnecting)
				if cascadeDelete {
					return deleteResponse, nil
				}
			}
			impl.logger.Debugw("acd app is already deleted for this pipeline", "pipeline", pipeline)
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
					if cascadeDelete && strings.Contains(err.Error(), "code = NotFound") {
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
					return deleteResponse, err
				}
			}
			impl.logger.Infow("app deleted from argocd", "id", pipeline.Id, "pipelineName", pipeline.Name, "app", deploymentAppName)
			pipeline.DeploymentAppDeleteRequest = true
			pipeline.UpdatedOn = time.Now()
			pipeline.UpdatedBy = userId
			err = impl.pipelineRepository.Update(pipeline, tx)
			if err != nil {
				impl.logger.Errorw("error in partially delete cd pipeline", "err", err)
				return deleteResponse, err
			}
		}
		deleteResponse.DeleteInitiated = true
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing db transaction", "err", err)
		return deleteResponse, err
	}
	return deleteResponse, nil
}

func (impl *PipelineBuilderImpl) isGitRepoUrlPresent(appId int) bool {
	fetchedChart, err := impl.chartRepository.FindLatestByAppId(appId)

	if err != nil || len(fetchedChart.GitRepoUrl) == 0 {
		impl.logger.Errorw("error fetching git repo url or it is not present")
		return false
	}
	return true
}

func (impl *PipelineBuilderImpl) isPipelineInfoValid(pipeline *pipelineConfig.Pipeline,
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

func (impl *PipelineBuilderImpl) handleFailedDeploymentAppChange(pipeline *pipelineConfig.Pipeline,
	failedPipelines []*bean.DeploymentChangeStatus, err string) []*bean.DeploymentChangeStatus {

	return impl.appendToDeploymentChangeStatusList(
		failedPipelines,
		pipeline,
		err,
		bean.Failed)
}

func (impl *PipelineBuilderImpl) handleNotHealthyAppsIfArgoDeploymentType(pipeline *pipelineConfig.Pipeline,
	failedPipelines []*bean.DeploymentChangeStatus) ([]*bean.DeploymentChangeStatus, error) {

	if pipeline.DeploymentAppType == bean.ArgoCd {
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

func (impl *PipelineBuilderImpl) handleNotDeployedAppsIfArgoDeploymentType(pipeline *pipelineConfig.Pipeline,
	failedPipelines []*bean.DeploymentChangeStatus) ([]*bean.DeploymentChangeStatus, error) {

	if pipeline.DeploymentAppType == string(bean.ArgoCd) {
		// check if app status is Healthy
		status, err := impl.appStatusRepository.Get(pipeline.AppId, pipeline.EnvironmentId)

		// case: missing status row in db
		if len(status.Status) == 0 {
			return failedPipelines, nil
		}

		// cannot delete the app from argocd if app status is Progressing
		if err != nil {

			healthCheckErr := errors.New("unable to fetch app status")

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

func (impl *PipelineBuilderImpl) FetchDeletedApp(ctx context.Context,
	pipelines []*pipelineConfig.Pipeline) *bean.DeploymentAppTypeChangeResponse {

	successfulPipelines := make([]*bean.DeploymentChangeStatus, 0)
	failedPipelines := make([]*bean.DeploymentChangeStatus, 0)
	// Iterate over all the pipelines in the environment for given deployment app type
	for _, pipeline := range pipelines {

		deploymentAppName := fmt.Sprintf("%s-%s", pipeline.App.AppName, pipeline.Environment.Name)
		var err error
		if pipeline.DeploymentAppType == string(bean.ArgoCd) {
			appIdentifier := &client.AppIdentifier{
				ClusterId:   pipeline.Environment.ClusterId,
				ReleaseName: pipeline.DeploymentAppName,
				Namespace:   pipeline.Environment.Namespace,
			}
			_, err = impl.helmAppService.GetApplicationDetail(ctx, appIdentifier)
		} else {
			req := &application2.ApplicationQuery{
				Name: &deploymentAppName,
			}
			_, err = impl.application.Get(ctx, req)
		}
		if err != nil {
			impl.logger.Errorw("error in getting application detail", "err", err, "deploymentAppName", deploymentAppName)
		}

		if err != nil && checkAppReleaseNotExist(err) {
			successfulPipelines = impl.appendToDeploymentChangeStatusList(
				successfulPipelines,
				pipeline,
				"",
				bean.Success)
		} else {
			failedPipelines = impl.appendToDeploymentChangeStatusList(
				failedPipelines,
				pipeline,
				"App Not Yet Deleted.",
				bean.NOT_YET_DELETED)
		}
	}

	return &bean.DeploymentAppTypeChangeResponse{
		SuccessfulPipelines: successfulPipelines,
		FailedPipelines:     failedPipelines,
	}
}

// deleteArgoCdApp takes context and deployment app name used in argo cd and deletes
// the application in argo cd.
func (impl *PipelineBuilderImpl) deleteArgoCdApp(ctx context.Context, pipeline *pipelineConfig.Pipeline, deploymentAppName string,
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
func (impl *PipelineBuilderImpl) deleteHelmApp(ctx context.Context, pipeline *pipelineConfig.Pipeline) error {

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

func (impl *PipelineBuilderImpl) appendToDeploymentChangeStatusList(pipelines []*bean.DeploymentChangeStatus,
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

func (impl *PipelineBuilderImpl) createCdPipeline(ctx context.Context, app *app2.App, pipeline *bean.CDPipelineConfigObject, userId int32) (pipelineRes int, err error) {
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
	//VARIABLE_MAPPING_UPDATE
	err = impl.extractAndMapVariables(envOverride.EnvOverrideValues, envOverride.Id, repository6.EntityTypeDeploymentTemplateEnvLevel, envOverride.UpdatedBy, tx)
	if err != nil {
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

func (impl PipelineBuilderImpl) extractAndMapVariables(template string, entityId int, entityType repository6.EntityType, userId int32, tx *pg.Tx) error {
	usedVariables, err := impl.variableTemplateParser.ExtractVariables(template)
	if err != nil {
		return err
	}
	err = impl.variableEntityMappingService.UpdateVariablesForEntity(usedVariables, repository6.Entity{
		EntityType: entityType,
		EntityId:   entityId,
	}, userId, tx)
	if err != nil {
		return err
	}
	return nil
}

func (impl *PipelineBuilderImpl) updateCdPipeline(ctx context.Context, pipeline *bean.CDPipelineConfigObject, userID int32) (err error) {

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

func (impl *PipelineBuilderImpl) filterDeploymentTemplate(strategyKey string, pipelineStrategiesJson string) (string, error) {
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

func (impl *PipelineBuilderImpl) getStrategiesMapping(dbPipelineIds []int) (map[int][]*chartConfig.PipelineStrategy, error) {
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

type ConfigMapSecretsResponse struct {
	Maps    []bean2.ConfigSecretMap `json:"maps"`
	Secrets []bean2.ConfigSecretMap `json:"secrets"`
}

func (impl *PipelineBuilderImpl) BuildArtifactsForParentStage(cdPipelineId int, parentId int, parentType bean2.WorkflowType, ciArtifacts []bean.CiArtifactBean, artifactMap map[int]int, limit int, parentCdId int) ([]bean.CiArtifactBean, error) {
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

func (impl *PipelineBuilderImpl) BuildArtifactsForCdStage(pipelineId int, stageType bean2.WorkflowType, ciArtifacts []bean.CiArtifactBean, artifactMap map[int]int, parent bool, limit int, parentCdId int) ([]bean.CiArtifactBean, map[int]int, int, string, error) {
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

func (impl *PipelineBuilderImpl) BuildArtifactsForCIParent(cdPipelineId int, parentId int, parentType bean2.WorkflowType, ciArtifacts []bean.CiArtifactBean, artifactMap map[int]int, limit int) ([]bean.CiArtifactBean, error) {
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

func (impl *PipelineBuilderImpl) updateGitRepoUrlInCharts(appId int, chartGitAttribute *util.ChartGitAttribute, userId int32) error {
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

func (impl *PipelineBuilderImpl) BulkDeleteCdPipelines(impactedPipelines []*pipelineConfig.Pipeline, ctx context.Context, dryRun bool, deleteAction int, userId int32) []*bean.CdBulkActionResponseDto {
	var respDtos []*bean.CdBulkActionResponseDto
	for _, pipeline := range impactedPipelines {
		respDto := &bean.CdBulkActionResponseDto{
			PipelineName:    pipeline.Name,
			AppName:         pipeline.App.AppName,
			EnvironmentName: pipeline.Environment.Name,
		}
		if !dryRun {
			deleteResponse, err := impl.DeleteCdPipeline(pipeline, ctx, deleteAction, true, userId)
			if err != nil {
				impl.logger.Errorw("error in deleting cd pipeline", "err", err, "pipelineId", pipeline.Id)
				respDto.DeletionResult = fmt.Sprintf("Not able to delete pipeline, %v", err)
			} else if !(deleteResponse.DeleteInitiated || deleteResponse.ClusterReachable) {
				respDto.DeletionResult = fmt.Sprintf("Not able to delete pipeline, cluster connection error")
			} else {
				respDto.DeletionResult = "Pipeline deleted successfully."
			}
		}
		respDtos = append(respDtos, respDto)
	}
	return respDtos

}

func (impl *PipelineBuilderImpl) buildExternalCiWebhookSchema() map[string]interface{} {
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

func (impl *PipelineBuilderImpl) buildPayloadOption() []bean.PayloadOptionObject {
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

func (impl *PipelineBuilderImpl) buildResponses() []bean.ResponseSchemaObject {
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

func checkAppReleaseNotExist(err error) bool {
	// RELEASE_NOT_EXIST check for helm App and NOT_FOUND check for argo app
	return strings.Contains(err.Error(), bean.NOT_FOUND) || strings.Contains(err.Error(), bean.RELEASE_NOT_EXIST)
}
