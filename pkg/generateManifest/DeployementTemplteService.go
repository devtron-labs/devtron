package generateManifest

import (
	"context"
	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/chart"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/util/k8s"
	"go.uber.org/zap"
)

type ValuesAndManifestRequest struct {
	AppId                    int                               `json:"appId"`
	EnvId                    int                               `json:"envId,omitempty"`
	ChartRefId               int                               `json:"chartRefId"`
	GetValues                bool                              `json:"getValues"`
	Values                   string                            `json:"values"`
	Type                     repository.DeploymentTemplateType `json:"type"`
	PipelineConfigOverrideId int                               `json:"pipelineConfigOverrideId,omitempty"`
	ResourceName             string                            `json:"resourceName"`
}

type ValuesAndManifestResponse struct {
	Data string `json:"data"`
}

type DeploymentTemplateService interface {
	FetchDeploymentsWithChartRefs(appId int, envId int) ([]*repository.FetchTemplateComparisonList, error)
	GetValuesAndManifest(ctx context.Context, request ValuesAndManifestRequest) (ValuesAndManifestResponse, error)
	GetManifest(ctx context.Context, chartRefId int, valuesYaml string) (*openapi2.TemplateChartResponse, error)
}
type DeploymentTemplateServiceImpl struct {
	Logger                       *zap.SugaredLogger
	chartService                 chart.ChartService
	appListingService            app.AppListingService
	appListingRepository         repository.AppListingRepository
	deploymentTemplateRepository repository.DeploymentTemplateRepository
	helmAppService               client.HelmAppService
	chartRepository              chartRepoRepository.ChartRepository
	chartTemplateServiceImpl     util.ChartTemplateService
	K8sUtil                      *k8s.K8sUtil
	helmAppClient                client.HelmAppClient
}

func NewDeploymentTemplateServiceImpl(Logger *zap.SugaredLogger, chartService chart.ChartService,
	appListingService app.AppListingService,
	appListingRepository repository.AppListingRepository,
	deploymentTemplateRepository repository.DeploymentTemplateRepository,
	helmAppService client.HelmAppService,
	chartRepository chartRepoRepository.ChartRepository,
	chartTemplateServiceImpl util.ChartTemplateService,
	helmAppClient client.HelmAppClient,
	K8sUtil *k8s.K8sUtil,
) *DeploymentTemplateServiceImpl {
	return &DeploymentTemplateServiceImpl{
		Logger:                       Logger,
		chartService:                 chartService,
		appListingService:            appListingService,
		appListingRepository:         appListingRepository,
		deploymentTemplateRepository: deploymentTemplateRepository,
		helmAppService:               helmAppService,
		chartRepository:              chartRepository,
		chartTemplateServiceImpl:     chartTemplateServiceImpl,
		K8sUtil:                      K8sUtil,
		helmAppClient:                helmAppClient,
	}
}

func (impl DeploymentTemplateServiceImpl) FetchDeploymentsWithChartRefs(appId int, envId int) ([]*repository.FetchTemplateComparisonList, error) {

	var responseList []*repository.FetchTemplateComparisonList

	defaultVersions, err := impl.chartService.ChartRefAutocompleteForAppOrEnv(appId, 0)
	if err != nil {
		impl.Logger.Errorw("error in getting defaultVersions", "err", err, "appId", appId, "envId", envId)
		return nil, err
	}

	for _, item := range defaultVersions.ChartRefs {
		res := &repository.FetchTemplateComparisonList{
			ChartId:      item.Id,
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
		item := &repository.FetchTemplateComparisonList{
			ChartId:         env.ChartRefId,
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

func (impl DeploymentTemplateServiceImpl) GetValuesAndManifest(ctx context.Context, request ValuesAndManifestRequest) (ValuesAndManifestResponse, error) {
	var result ValuesAndManifestResponse
	var values string
	var err error

	if request.Values != "" {
		values = request.Values
	} else {
		switch request.Type {
		case repository.DefaultVersions:
			_, values, err = impl.chartService.GetAppOverrideForDefaultTemplate(request.ChartRefId)
			if err != nil {
				impl.Logger.Errorw("err", err)
				return result, err
			}
		case repository.PublishedOnEnvironments:
			chart, err := impl.chartRepository.FindLatestChartForAppByAppId(request.AppId)
			if chart != nil && chart.Id > 0 {
				values = chart.GlobalOverride
			}
			if err != nil {
				impl.Logger.Errorw("err", err)
				return result, err
			}
		case repository.DeployedOnSelfEnvironment, repository.DeployedOnOtherEnvironment:
			values, err = impl.deploymentTemplateRepository.FetchPipelineOverrideValues(request.PipelineConfigOverrideId)
			if err != nil {
				impl.Logger.Errorw("err", err)
				return result, err
			}
		}
	}

	if request.GetValues {
		result.Data = values
		return result, nil
	}
	manifest, err := impl.GetManifest(ctx, request.ChartRefId, values)
	if err != nil {
		return result, err
	}
	result.Data = *manifest.Manifest
	return result, nil
}

func (impl DeploymentTemplateServiceImpl) GetManifest(ctx context.Context, chartRefId int, valuesYaml string) (*openapi2.TemplateChartResponse, error) {
	refChart, template, err, version, _ := impl.chartService.GetRefChart(chart.TemplateRequest{ChartRefId: chartRefId})
	if err != nil {
		impl.Logger.Errorw("error in getting refChart", "err", err, "chartRefId", chartRefId)
		return nil, err
	}

	chartBytes, chartZipPath, err := impl.chartTemplateServiceImpl.LoadChartInBytes(refChart, false, "", "")
	if err != nil {
		impl.Logger.Errorw("error in getting chart", "err", err)
		return nil, err
	}
	defer impl.chartTemplateServiceImpl.CleanDir(chartZipPath)

	k8sServerVersion, err := impl.K8sUtil.GetKubeVersion()
	if err != nil {
		impl.Logger.Errorw("exception caught in getting k8sServerVersion", "err", err)
		return nil, err
	}
	installReleaseRequest := &client.InstallReleaseRequest{
		ChartName:    template,
		ChartVersion: version,
		ValuesYaml:   valuesYaml,
		K8SVersion:   k8sServerVersion.String(),
		ChartRepository: &client.ChartRepository{
			Name:     "repo",
			Url:      "http://localhost:8080/",
			Username: "admin",
			Password: "password",
		},
		ReleaseIdentifier: &client.ReleaseIdentifier{
			ReleaseNamespace: "devtron-demo",
			ReleaseName:      "release-name",
		},
		ChartContent: &client.ChartContent{
			Content: chartBytes,
		},
	}
	config, err := impl.helmAppService.GetClusterConf(client.DEFAULT_CLUSTER_ID)
	if err != nil {
		impl.Logger.Errorw("error in fetching cluster detail", "clusterId", 1, "err", err)
		return nil, err
	}

	installReleaseRequest.ReleaseIdentifier.ClusterConfig = config

	templateChartResponse, err := impl.helmAppClient.TemplateChart(ctx, installReleaseRequest)
	if err != nil {
		impl.Logger.Errorw("error in templating chart", "err", err)
		return nil, err
	}
	response := &openapi2.TemplateChartResponse{
		Manifest: &templateChartResponse.GeneratedManifest,
	}

	return response, nil
}
