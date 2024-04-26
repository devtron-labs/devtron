package generateManifest

import (
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	yamlUtil "github.com/devtron-labs/common-lib/utils/yaml"
	"github.com/devtron-labs/devtron/api/helm-app/bean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	client "github.com/devtron-labs/devtron/api/helm-app/service"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/chart"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strconv"
	"time"
)

type DeploymentTemplateService interface {
	FetchDeploymentsWithChartRefs(appId int, envId int) ([]*repository.DeploymentTemplateComparisonMetadata, error)
	GetDeploymentTemplate(ctx context.Context, request DeploymentTemplateRequest) (DeploymentTemplateResponse, error)
	GenerateManifest(ctx context.Context, chartRefId int, valuesYaml string) (*openapi2.TemplateChartResponse, error)
	GetRestartWorkloadData(ctx context.Context, appIds []int, envId int) (*RestartPodResponse, error)
}
type DeploymentTemplateServiceImpl struct {
	Logger                           *zap.SugaredLogger
	chartService                     chart.ChartService
	appListingService                app.AppListingService
	deploymentTemplateRepository     repository.DeploymentTemplateRepository
	helmAppService                   client.HelmAppService
	chartTemplateServiceImpl         util.ChartTemplateService
	K8sUtil                          *k8s.K8sServiceImpl
	helmAppClient                    gRPC.HelmAppClient
	propertiesConfigService          pipeline.PropertiesConfigService
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService
	environmentRepository            repository3.EnvironmentRepository
	appRepository                    appRepository.AppRepository
	scopedVariableManager            variables.ScopedVariableManager
	chartRefService                  chartRef.ChartRefService
	pipelineOverrideRepository       chartConfig.PipelineOverrideRepository
	chartRepository                  chartRepoRepository.ChartRepository
}

func NewDeploymentTemplateServiceImpl(Logger *zap.SugaredLogger, chartService chart.ChartService,
	appListingService app.AppListingService,
	deploymentTemplateRepository repository.DeploymentTemplateRepository,
	helmAppService client.HelmAppService,
	chartTemplateServiceImpl util.ChartTemplateService,
	helmAppClient gRPC.HelmAppClient,
	K8sUtil *k8s.K8sServiceImpl,
	propertiesConfigService pipeline.PropertiesConfigService,
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService,
	environmentRepository repository3.EnvironmentRepository,
	appRepository appRepository.AppRepository,
	scopedVariableManager variables.ScopedVariableManager,
	chartRefService chartRef.ChartRefService,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	chartRepository chartRepoRepository.ChartRepository,
) (*DeploymentTemplateServiceImpl, error) {
	deploymentTemplateServiceImpl := &DeploymentTemplateServiceImpl{
		Logger:                           Logger,
		chartService:                     chartService,
		appListingService:                appListingService,
		deploymentTemplateRepository:     deploymentTemplateRepository,
		helmAppService:                   helmAppService,
		chartTemplateServiceImpl:         chartTemplateServiceImpl,
		K8sUtil:                          K8sUtil,
		helmAppClient:                    helmAppClient,
		propertiesConfigService:          propertiesConfigService,
		deploymentTemplateHistoryService: deploymentTemplateHistoryService,
		environmentRepository:            environmentRepository,
		appRepository:                    appRepository,
		scopedVariableManager:            scopedVariableManager,
		chartRefService:                  chartRefService,
		pipelineOverrideRepository:       pipelineOverrideRepository,
		chartRepository:                  chartRepository,
	}
	return deploymentTemplateServiceImpl, nil
}

func (impl DeploymentTemplateServiceImpl) FetchDeploymentsWithChartRefs(appId int, envId int) ([]*repository.DeploymentTemplateComparisonMetadata, error) {

	var responseList []*repository.DeploymentTemplateComparisonMetadata

	defaultVersions, err := impl.chartService.ChartRefAutocompleteForAppOrEnv(appId, 0)
	if err != nil {
		impl.Logger.Errorw("error in getting defaultVersions", "err", err, "appId", appId, "envId", envId)
		return nil, err
	}

	for _, item := range defaultVersions.ChartRefs {
		res := &repository.DeploymentTemplateComparisonMetadata{
			ChartRefId:   item.Id,
			ChartVersion: item.Version,
			ChartType:    item.Name,
			Type:         repository.DefaultVersions,
		}
		responseList = append(responseList, res)
	}

	publishedOnEnvs, err := impl.appListingService.FetchMinDetailOtherEnvironment(appId)
	if err != nil {
		impl.Logger.Errorw("error in getting publishedOnEnvs", "err", err, "appId", appId, "envId", envId)
		return nil, err
	}

	for _, env := range publishedOnEnvs {
		item := &repository.DeploymentTemplateComparisonMetadata{
			ChartRefId:      env.ChartRefId,
			EnvironmentId:   env.EnvironmentId,
			EnvironmentName: env.EnvironmentName,
			Type:            repository.PublishedOnEnvironments,
		}
		responseList = append(responseList, item)
	}

	deployedOnEnv, err := impl.deploymentTemplateRepository.FetchDeploymentHistoryWithChartRefs(appId, envId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("error in getting deployedOnEnv", "err", err, "appId", appId, "envId", envId)
		return nil, err
	}

	for _, deployedItem := range deployedOnEnv {
		deployedItem.Type = repository.DeployedOnSelfEnvironment
		deployedItem.EnvironmentId = envId
		responseList = append(responseList, deployedItem)
	}

	deployedOnOtherEnvs, err := impl.deploymentTemplateRepository.FetchLatestDeploymentWithChartRefs(appId, envId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("error in getting deployedOnOtherEnvs", "err", err, "appId", appId, "envId", envId)
		return nil, err
	}

	for _, deployedItem := range deployedOnOtherEnvs {
		deployedItem.Type = repository.DeployedOnOtherEnvironment
		responseList = append(responseList, deployedItem)
	}

	return responseList, nil
}

func (impl DeploymentTemplateServiceImpl) GetDeploymentTemplate(ctx context.Context, request DeploymentTemplateRequest) (DeploymentTemplateResponse, error) {
	var result DeploymentTemplateResponse
	var values, resolvedValue string
	var err error
	var variableSnapshot map[string]string

	if request.Values != "" {
		values = request.Values
		resolvedValue, variableSnapshot, err = impl.resolveTemplateVariables(ctx, request.Values, request)
		if err != nil {
			return result, err
		}
	} else {
		switch request.Type {
		case repository.DefaultVersions:
			_, values, err = impl.chartRefService.GetAppOverrideForDefaultTemplate(request.ChartRefId)
			resolvedValue = values
		case repository.PublishedOnEnvironments:
			values, resolvedValue, variableSnapshot, err = impl.fetchResolvedTemplateForPublishedEnvs(ctx, request)
		case repository.DeployedOnSelfEnvironment, repository.DeployedOnOtherEnvironment:
			values, resolvedValue, variableSnapshot, err = impl.fetchTemplateForDeployedEnv(ctx, request)
		}
		if err != nil {
			impl.Logger.Errorw("error in getting values", "err", err)
			return result, err
		}
	}

	if request.RequestDataMode == Values {
		result.Data = values
		result.ResolvedData = resolvedValue
		result.VariableSnapshot = variableSnapshot
		return result, nil
	}

	manifest, err := impl.GenerateManifest(ctx, request.ChartRefId, resolvedValue)
	if err != nil {
		return result, err
	}
	result.Data = *manifest.Manifest
	return result, nil
}

func (impl DeploymentTemplateServiceImpl) fetchResolvedTemplateForPublishedEnvs(ctx context.Context, request DeploymentTemplateRequest) (string, string, map[string]string, error) {
	var values string
	override, err := impl.propertiesConfigService.GetEnvironmentProperties(request.AppId, request.EnvId, request.ChartRefId)
	if err == nil && override.GlobalConfig != nil {
		if override.EnvironmentConfig.EnvOverrideValues != nil {
			values = string(override.EnvironmentConfig.EnvOverrideValues)
		} else {
			values = string(override.GlobalConfig)
		}
	} else {
		impl.Logger.Errorw("error in getting overridden values", "err", err)
		return "", "", nil, err
	}
	resolvedTemplate, variableSnapshot, err := impl.resolveTemplateVariables(ctx, values, request)
	if err != nil {
		return values, values, variableSnapshot, err
	}
	return values, resolvedTemplate, variableSnapshot, nil
}

func (impl DeploymentTemplateServiceImpl) fetchTemplateForDeployedEnv(ctx context.Context, request DeploymentTemplateRequest) (string, string, map[string]string, error) {
	historyObject, err := impl.deploymentTemplateHistoryService.GetHistoryForDeployedTemplateById(ctx, request.DeploymentTemplateHistoryId, request.PipelineId)
	if err != nil {
		impl.Logger.Errorw("error in getting deployment template history", "err", err, "id", request.DeploymentTemplateHistoryId, "pipelineId", request.PipelineId)
		return "", "", nil, err
	}

	//todo Subhashish solve variable leak
	return historyObject.CodeEditorValue.Value, historyObject.CodeEditorValue.ResolvedValue, historyObject.CodeEditorValue.VariableSnapshot, nil
}

func (impl DeploymentTemplateServiceImpl) resolveTemplateVariables(ctx context.Context, values string, request DeploymentTemplateRequest) (string, map[string]string, error) {

	isSuperAdmin, err := util2.GetIsSuperAdminFromContext(ctx)
	if err != nil {
		return values, nil, err
	}
	scope, err := impl.extractScopeData(request)
	if err != nil {
		return values, nil, err
	}
	maskUnknownVariableForHelmGenerate := request.RequestDataMode == Manifest
	resolvedTemplate, variableSnapshot, err := impl.scopedVariableManager.ExtractVariablesAndResolveTemplate(scope, values, parsers.StringVariableTemplate, isSuperAdmin, maskUnknownVariableForHelmGenerate)
	if err != nil {
		return values, variableSnapshot, err
	}
	return resolvedTemplate, variableSnapshot, nil
}

func (impl DeploymentTemplateServiceImpl) extractScopeData(request DeploymentTemplateRequest) (resourceQualifiers.Scope, error) {
	app, err := impl.appRepository.FindById(request.AppId)
	scope := resourceQualifiers.Scope{}
	if err != nil {
		return scope, err
	}
	scope.AppId = request.AppId
	scope.EnvId = request.EnvId
	scope.SystemMetadata = &resourceQualifiers.SystemMetadata{AppName: app.AppName}

	if request.EnvId != 0 {
		environment, err := impl.environmentRepository.FindById(request.EnvId)
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("error in getting system metadata", "err", err)
			return scope, err
		}
		if environment != nil {
			scope.ClusterId = environment.ClusterId
			scope.SystemMetadata.EnvironmentName = environment.Name
			scope.SystemMetadata.ClusterName = environment.Cluster.ClusterName
			scope.SystemMetadata.Namespace = environment.Namespace
		}
	}
	return scope, nil
}

func (impl DeploymentTemplateServiceImpl) GenerateManifest(ctx context.Context, chartRefId int, valuesYaml string) (*openapi2.TemplateChartResponse, error) {
	refChart, template, version, _, err := impl.chartRefService.GetRefChart(chartRefId)
	if err != nil {
		impl.Logger.Errorw("error in getting refChart", "err", err, "chartRefId", chartRefId)
		return nil, err
	}

	outputChartPathDir := fmt.Sprintf("%s-%v", refChart, strconv.FormatInt(time.Now().UnixNano(), 16))
	if _, err := os.Stat(outputChartPathDir); os.IsNotExist(err) {
		err = os.Mkdir(outputChartPathDir, 0755)
		if err != nil {
			impl.Logger.Errorw("error in creating temp outputChartPathDir", "err", err, "outputChartPathDir", outputChartPathDir, "chartRefId", chartRefId)
			return nil, err
		}
	}
	//load chart from given refChart
	chart, err := impl.chartTemplateServiceImpl.LoadChartFromDir(refChart)
	if err != nil {
		impl.Logger.Errorw("error in LoadChartFromDir", "err", err, "chartRefId", chartRefId)
		return nil, err
	}

	//create the .tgz file in temp location
	chartBytes, err := impl.chartTemplateServiceImpl.CreateZipFileForChart(chart, outputChartPathDir)
	if err != nil {
		impl.Logger.Errorw("error in CreateZipFileForChart", "err", err, "chartRefId", chartRefId)
		return nil, err
	}

	//deleted the .tgz temp file after reading chart bytes
	defer impl.chartTemplateServiceImpl.CleanDir(outputChartPathDir)

	k8sServerVersion, err := impl.K8sUtil.GetKubeVersion()
	if err != nil {
		impl.Logger.Errorw("exception caught in getting k8sServerVersion", "err", err)
		return nil, err
	}
	installReleaseRequest := &gRPC.InstallReleaseRequest{
		ChartName:         template,
		ChartVersion:      version,
		ValuesYaml:        valuesYaml,
		K8SVersion:        k8sServerVersion.String(), //done
		ChartRepository:   ChartRepository,
		ReleaseIdentifier: ReleaseIdentifier,
		ChartContent: &gRPC.ChartContent{
			Content: chartBytes,
		},
	}
	config, err := impl.helmAppService.GetClusterConf(bean.DEFAULT_CLUSTER_ID)
	if err != nil {
		impl.Logger.Errorw("error in fetching cluster detail", "clusterId", 1, "err", err)
		return nil, err
	}

	installReleaseRequest.ReleaseIdentifier.ClusterConfig = config //done

	templateChartResponse, err := impl.helmAppClient.TemplateChart(ctx, installReleaseRequest)
	if err != nil {
		impl.Logger.Errorw("error in templating chart", "err", err)
		clientErrCode, errMsg := util.GetClientDetailedError(err)
		if clientErrCode.IsInvalidArgumentCode() {
			return nil, &util.ApiError{HttpStatusCode: http.StatusConflict, Code: strconv.Itoa(http.StatusConflict), InternalMessage: errMsg, UserMessage: errMsg}
		}
		return nil, err
	}
	response := &openapi2.TemplateChartResponse{
		Manifest: &templateChartResponse.GeneratedManifest,
	}

	return response, nil
}
func (impl DeploymentTemplateServiceImpl) GetRestartWorkloadData(ctx context.Context, appIds []int, envId int) (*RestartPodResponse, error) {
	podResp := &RestartPodResponse{}
	appIdToInstallReleaseRequest := make(map[int]*gRPC.InstallReleaseRequest)
	err := impl.setChartContent(appIds, appIdToInstallReleaseRequest)
	if err != nil {
		impl.Logger.Errorw("error in setting chart content", "appIds", appIds, "err", err)
		return nil, err
	}
	err = impl.setValuesYaml(appIds, envId, appIdToInstallReleaseRequest)
	if err != nil {
		impl.Logger.Errorw("error in setting values yaml", "appIds", appIds, "err", err)
		return nil, err
	}
	apps, err := impl.appRepository.FindByIds(util2.GetReferencedArray(appIds))
	if err != nil {
		impl.Logger.Errorw("error in fetching app", "err", err)
		return nil, err
	}
	appNameToId := make(map[string]int)

	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		impl.Logger.Errorw("error in fetching environment", "err", err)
		return nil, err
	}
	k8sServerVersion, err := impl.K8sUtil.GetKubeVersion()
	if err != nil {
		impl.Logger.Errorw("exception caught in getting k8sServerVersion", "err", err)
		return nil, err
	}
	config, err := impl.helmAppService.GetClusterConf(bean.DEFAULT_CLUSTER_ID)
	if err != nil {
		impl.Logger.Errorw("error in fetching cluster detail", "clusterId", 1, "err", err)
		return nil, err
	}
	for _, app := range apps {
		appNameToId[app.AppName] = app.Id
		appIdToInstallReleaseRequest[app.Id] = &gRPC.InstallReleaseRequest{ReleaseIdentifier: impl.getReleaseIdentifier(config, app, env),
			K8SVersion: k8sServerVersion.String(),
		}
	}
	installReleaseRequest := make([]*gRPC.InstallReleaseRequest, 0)
	for _, req := range appIdToInstallReleaseRequest {
		installReleaseRequest = append(installReleaseRequest, req)
	}
	templateChartResponse, err := impl.helmAppClient.TemplateChartBulk(ctx, installReleaseRequest)
	appIdToResourceIdentifier := make(map[int]ResourceIdentifierResponse)
	for _, tcResp := range templateChartResponse {
		manifests, err := yamlUtil.SplitYAMLs([]byte(tcResp.GeneratedManifest))
		if err != nil {
			return nil, err
		}
		appName := tcResp.AppName

		resourceMeta := make([]ResourceMetadata, 0)
		for _, manifest := range manifests {
			gvk := manifest.GroupVersionKind()
			name := manifest.GetName()
			switch gvk.Kind {
			case "Deployment", "StatefulSet", "DemonSet", "Rollout":
				resourceMeta = append(resourceMeta, ResourceMetadata{
					Name:             name,
					GroupVersionKind: gvk,
				})
			}
		}
		appIdToResourceIdentifier[appNameToId[tcResp.AppName]] = ResourceIdentifierResponse{
			ResourceMetaData: resourceMeta,
			AppName:          appName,
		}

	}
	podResp = &RestartPodResponse{
		EnvironmentId: envId,
		Namespace:     env.Namespace,
		RestartPodMap: appIdToResourceIdentifier,
	}

	return podResp, nil
}

func (impl DeploymentTemplateServiceImpl) setValuesYaml(appIds []int, envId int, appIdToInstallReleaseRequest map[int]*gRPC.InstallReleaseRequest) error {
	pipelineOverrides, err := impl.pipelineOverrideRepository.GetLatestReleaseForAppIds(appIds, envId)
	if err != nil {
		impl.Logger.Errorw("error in fetching pipelineOverrides for appIds", "appIds", appIds, "err", err)
	}
	for _, pco := range pipelineOverrides {
		appIdToInstallReleaseRequest[pco.Pipeline.AppId] = &gRPC.InstallReleaseRequest{ValuesYaml: pco.PipelineOverrideValues}
	}
	return err
}

func (impl DeploymentTemplateServiceImpl) setChartContent(appIds []int, appIdToInstallReleaseRequest map[int]*gRPC.InstallReleaseRequest) error {
	charts, err := impl.chartRepository.FindLatestChartByAppIds(appIds)
	if err != nil {
		impl.Logger.Errorw("error in fetching chart", "err", err, "appIds", appIds)
		return err
	}
	appIdToChartRefId := make(map[int]int)
	var chartRefIds []int
	for _, chart := range charts {
		appIdToChartRefId[chart.AppId] = chart.ChartRefId
		chartRefIds = append(chartRefIds, chart.ChartRefId)
	}
	chartRefIdToBytes, err := impl.chartRefService.GetChartBytesInBulk(chartRefIds, true)
	if err != nil {
		impl.Logger.Errorw("error in fetching chartRefBean", "err", err, "chartRefIds", chartRefIds)
		return err
	}
	for appId, chartRefId := range appIdToChartRefId {
		if bytes, ok := chartRefIdToBytes[chartRefId]; ok {
			appIdToInstallReleaseRequest[appId] = &gRPC.InstallReleaseRequest{ChartContent: &gRPC.ChartContent{Content: bytes}}
		}
	}
	return err
}

func (impl DeploymentTemplateServiceImpl) getReleaseIdentifier(config *gRPC.ClusterConfig, app *appRepository.App, env *repository3.Environment) *gRPC.ReleaseIdentifier {
	return &gRPC.ReleaseIdentifier{
		ClusterConfig:    config,
		ReleaseName:      fmt.Sprintf("%s-%s", app.AppName, env.Name),
		ReleaseNamespace: env.Namespace,
	}
}