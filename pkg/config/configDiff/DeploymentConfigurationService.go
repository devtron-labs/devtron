package configDiff

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	k8sUtil "github.com/devtron-labs/common-lib/utils/k8s"
	bean4 "github.com/devtron-labs/devtron/api/bean"
	bean5 "github.com/devtron-labs/devtron/api/helm-app/bean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	read3 "github.com/devtron-labs/devtron/api/helm-app/service/read"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	repository4 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/config/configDiff/adaptor"
	bean2 "github.com/devtron-labs/devtron/pkg/config/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/config/configDiff/helper"
	"github.com/devtron-labs/devtron/pkg/config/configDiff/utils"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/configMapAndSecret"
	read2 "github.com/devtron-labs/devtron/pkg/deployment/manifest/configMapAndSecret/read"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	chartRefBean "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/read"
	"github.com/devtron-labs/devtron/pkg/generateManifest"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/adapter"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	repository6 "github.com/devtron-labs/devtron/pkg/variables/repository"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/ghodss/yaml"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"go.uber.org/zap"
	chart2 "helm.sh/helm/v3/pkg/chart"
	"net/http"
	"strconv"
)

type DeploymentConfigurationService interface {
	ConfigAutoComplete(appId int, envId int) (*bean2.ConfigDataResponse, error)
	GetAllConfigData(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfigDto, error)
	CompareCategoryWiseConfigData(ctx context.Context, comparisonRequestDto bean2.ComparisonRequestDto, userHasAdminAccess bool) (*bean2.ComparisonResponseDto, error)
	GetManifest(ctx context.Context, manifestRequest *bean2.ManifestRequest) (*bean2.ManifestResponse, error)
}

type DeploymentConfigurationServiceImpl struct {
	logger                               *zap.SugaredLogger
	configMapService                     pipeline.ConfigMapService
	appRepository                        appRepository.AppRepository
	environmentRepository                repository4.EnvironmentRepository
	chartService                         chartService.ChartService
	deploymentTemplateService            generateManifest.DeploymentTemplateService
	deploymentTemplateHistoryRepository  repository3.DeploymentTemplateHistoryRepository
	pipelineStrategyHistoryRepository    repository3.PipelineStrategyHistoryRepository
	configMapHistoryRepository           repository3.ConfigMapHistoryRepository
	scopedVariableManager                variables.ScopedVariableCMCSManager
	configMapRepository                  chartConfig.ConfigMapRepository
	deploymentConfigService              pipeline.PipelineDeploymentConfigService
	chartRefService                      chartRef.ChartRefService
	pipelineRepository                   pipelineConfig.PipelineRepository
	configMapHistoryService              configMapAndSecret.ConfigMapHistoryService
	deploymentTemplateHistoryReadService read.DeploymentTemplateHistoryReadService
	configMapHistoryReadService          read2.ConfigMapHistoryReadService
	cdWorkflowRepository                 pipelineConfig.CdWorkflowRepository
	envConfigOverrideService             read.EnvConfigOverrideService
	chartTemplateService                 util.ChartTemplateService
	helmAppClient                        gRPC.HelmAppClient
	helmAppService                       service.HelmAppService
	k8sUtil                              k8sUtil.K8sService
	mergeUtil                            util.MergeUtil
	HelmAppReadService                   read3.HelmAppReadService
}

func NewDeploymentConfigurationServiceImpl(logger *zap.SugaredLogger,
	configMapService pipeline.ConfigMapService,
	appRepository appRepository.AppRepository,
	environmentRepository repository4.EnvironmentRepository,
	chartService chartService.ChartService,
	deploymentTemplateService generateManifest.DeploymentTemplateService,
	deploymentTemplateHistoryRepository repository3.DeploymentTemplateHistoryRepository,
	pipelineStrategyHistoryRepository repository3.PipelineStrategyHistoryRepository,
	configMapHistoryRepository repository3.ConfigMapHistoryRepository,
	scopedVariableManager variables.ScopedVariableCMCSManager,
	configMapRepository chartConfig.ConfigMapRepository,
	deploymentConfigService pipeline.PipelineDeploymentConfigService,
	chartRefService chartRef.ChartRefService,
	pipelineRepository pipelineConfig.PipelineRepository,
	configMapHistoryService configMapAndSecret.ConfigMapHistoryService,
	deploymentTemplateHistoryReadService read.DeploymentTemplateHistoryReadService,
	configMapHistoryReadService read2.ConfigMapHistoryReadService,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	envConfigOverrideService read.EnvConfigOverrideService,
	chartTemplateService util.ChartTemplateService,
	helmAppClient gRPC.HelmAppClient,
	helmAppService service.HelmAppService,
	k8sUtil k8sUtil.K8sService,
	mergeUtil util.MergeUtil,
	HelmAppReadService read3.HelmAppReadService,
) (*DeploymentConfigurationServiceImpl, error) {
	deploymentConfigurationService := &DeploymentConfigurationServiceImpl{
		logger:                               logger,
		configMapService:                     configMapService,
		appRepository:                        appRepository,
		environmentRepository:                environmentRepository,
		chartService:                         chartService,
		deploymentTemplateService:            deploymentTemplateService,
		deploymentTemplateHistoryRepository:  deploymentTemplateHistoryRepository,
		pipelineStrategyHistoryRepository:    pipelineStrategyHistoryRepository,
		configMapHistoryRepository:           configMapHistoryRepository,
		scopedVariableManager:                scopedVariableManager,
		configMapRepository:                  configMapRepository,
		deploymentConfigService:              deploymentConfigService,
		chartRefService:                      chartRefService,
		pipelineRepository:                   pipelineRepository,
		configMapHistoryService:              configMapHistoryService,
		deploymentTemplateHistoryReadService: deploymentTemplateHistoryReadService,
		configMapHistoryReadService:          configMapHistoryReadService,
		cdWorkflowRepository:                 cdWorkflowRepository,
		envConfigOverrideService:             envConfigOverrideService,
		chartTemplateService:                 chartTemplateService,
		helmAppClient:                        helmAppClient,
		helmAppService:                       helmAppService,
		k8sUtil:                              k8sUtil,
		mergeUtil:                            mergeUtil,
		HelmAppReadService:                   HelmAppReadService,
	}

	return deploymentConfigurationService, nil
}
func (impl *DeploymentConfigurationServiceImpl) ConfigAutoComplete(appId int, envId int) (*bean2.ConfigDataResponse, error) {
	cMCSNamesAppLevel, cMCSNamesEnvLevel, err := impl.configMapService.FetchCmCsNamesAppAndEnvLevel(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in fetching CM and CS names at app or env level", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	cmcsKeyPropertyAppLevelMap, cmcsKeyPropertyEnvLevelMap := adaptor.GetCmCsAppAndEnvLevelMap(cMCSNamesAppLevel, cMCSNamesEnvLevel)
	for key, configProperty := range cmcsKeyPropertyAppLevelMap {
		if _, ok := cmcsKeyPropertyEnvLevelMap[key]; !ok {
			if envId > 0 {
				configProperty.ConfigStage = bean2.Inheriting
				configProperty.Id = 0
			}

		}
	}
	for key, configProperty := range cmcsKeyPropertyEnvLevelMap {
		if _, ok := cmcsKeyPropertyAppLevelMap[key]; ok {
			configProperty.ConfigStage = bean2.Overridden
		} else {
			configProperty.ConfigStage = bean2.Env
		}
	}
	combinedProperties := helper.GetCombinedPropertiesMap(cmcsKeyPropertyAppLevelMap, cmcsKeyPropertyEnvLevelMap)
	combinedProperties = append(combinedProperties, adaptor.GetConfigProperty(0, "", bean.DeploymentTemplate, bean2.PublishedConfigState))

	configDataResp := bean2.NewConfigDataResponse().WithResourceConfig(combinedProperties)
	return configDataResp, nil
}

func (impl *DeploymentConfigurationServiceImpl) GetAllConfigData(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfigDto, error) {
	var err error
	var envId, appId, clusterId int
	systemMetadata := &resourceQualifiers.SystemMetadata{
		AppName: configDataQueryParams.AppName,
	}
	if configDataQueryParams.IsEnvNameProvided() {
		env, err := impl.environmentRepository.FindEnvByNameWithClusterDetails(configDataQueryParams.EnvName)
		if err != nil {
			impl.logger.Errorw("GetAllConfigData, error in getting environment model by envName", "envName", configDataQueryParams.EnvName, "err", err)
			return nil, err
		}
		envId = env.Id
		clusterId = env.ClusterId
		systemMetadata.EnvironmentName = env.Name
		systemMetadata.Namespace = env.Namespace
		systemMetadata.ClusterName = env.Cluster.ClusterName
	}
	appId, err = impl.appRepository.FindAppIdByName(configDataQueryParams.AppName)
	if err != nil {
		impl.logger.Errorw("GetAllConfigData, error in getting app model by appName", "appName", configDataQueryParams.AppName, "err", err)
		return nil, err
	}

	switch configDataQueryParams.ConfigArea {
	case bean2.CdRollback.ToString():
		return impl.getConfigDataForCdRollback(ctx, configDataQueryParams, userHasAdminAccess)
	case bean2.DeploymentHistory.ToString():
		return impl.getConfigDataForDeploymentHistory(ctx, configDataQueryParams, userHasAdminAccess)
	}
	// this would be the default case
	return impl.getConfigDataForAppConfiguration(ctx, configDataQueryParams, appId, envId, clusterId, userHasAdminAccess, systemMetadata)
}

func (impl *DeploymentConfigurationServiceImpl) GetManifest(ctx context.Context, manifestRequest *bean2.ManifestRequest) (*bean2.ManifestResponse, error) {

	appId := manifestRequest.AppId
	envId := manifestRequest.EnvironmentId

	app, err := impl.appRepository.FindById(manifestRequest.AppId)
	if err != nil {
		impl.logger.Errorw("error in finding app by id", "appId", appId, "err", err)
		return nil, err
	}

	refChart, chartInBytes, err := impl.getRefChartBytes(ctx, envId, appId, app)
	if err != nil {
		impl.logger.Errorw("error in getting ref chart bytes", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	values := manifestRequest.Values

	scope := resourceQualifiers.Scope{
		AppId: appId,
		EnvId: envId,
		SystemMetadata: &resourceQualifiers.SystemMetadata{
			AppName: app.AppName,
		},
	}
	var namespace, envName string
	if manifestRequest.EnvironmentId > 0 {
		environment, err := impl.environmentRepository.FindById(envId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting environment", "envId", envId, "err", err)
			return nil, err
		}
		envName = environment.Name
		scope.ClusterId = environment.ClusterId
		scope.SystemMetadata.EnvironmentName = envName
		scope.SystemMetadata.ClusterName = environment.Cluster.ClusterName
		namespace = environment.Namespace
		scope.SystemMetadata.Namespace = namespace
	}

	isSuperAdmin, err := util2.GetIsSuperAdminFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var decodedValuesByte []byte
	if manifestRequest.ResourceType == bean.CS {
		decodedValuesByte, err = util2.GetDecodedAndEncodedData(values, util2.DecodeSecret)
		if err != nil {
			impl.logger.Errorw("error in decoding secret", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
	} else {
		decodedValuesByte = []byte(values)
	}

	resolvedTemplate, _, err := impl.scopedVariableManager.ExtractVariablesAndResolveTemplate(scope, string(decodedValuesByte), parsers.JsonVariableTemplate, isSuperAdmin, true)
	if err != nil {
		return nil, err
	}

	k8sServerVersion, err := impl.k8sUtil.GetKubeVersion()
	if err != nil {
		impl.logger.Errorw("exception caught in getting k8sServerVersion", "err", err)
		return nil, err
	}

	sanitizedK8sVersion := k8sServerVersion.String()

	releaseName := app.AppName
	if len(envName) > 0 {
		releaseName = fmt.Sprintf("%s-%s", app.AppName, envName)
	}
	installReleaseRequest := &gRPC.InstallReleaseRequest{
		AppName:         app.AppName,
		ChartName:       refChart.Name,
		ChartVersion:    refChart.Version,
		K8SVersion:      sanitizedK8sVersion,
		ChartRepository: generateManifest.ChartRepository,
		ChartContent: &gRPC.ChartContent{
			Content: chartInBytes,
		},
		ReleaseIdentifier: &gRPC.ReleaseIdentifier{
			ReleaseName: releaseName,
		},
	}
	config, err := impl.HelmAppReadService.GetClusterConf(bean5.DEFAULT_CLUSTER_ID)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", 1, "err", err)
		return nil, err
	}
	installReleaseRequest.ReleaseIdentifier.ClusterConfig = config

	mergedValuesYAML, err := impl.getMergedValuesForCMCSHelmTemplate(manifestRequest, resolvedTemplate, app, envId)
	if err != nil {
		impl.logger.Errorw("error in merging values for cm cs ", "err", err)
		return nil, err
	}

	installReleaseRequest.ValuesYaml = string(mergedValuesYAML)

	templateChartResponse, err := impl.helmAppClient.TemplateChart(ctx, installReleaseRequest)
	if err != nil {
		impl.logger.Errorw("error in templating chart", "err", err)
		clientErrCode, errMsg := util.GetClientDetailedError(err)
		if clientErrCode.IsFailedPreconditionCode() {
			return nil, &util.ApiError{HttpStatusCode: http.StatusUnprocessableEntity, Code: strconv.Itoa(http.StatusUnprocessableEntity), InternalMessage: errMsg, UserMessage: errMsg}
		} else if clientErrCode.IsInvalidArgumentCode() {
			return nil, &util.ApiError{HttpStatusCode: http.StatusConflict, Code: strconv.Itoa(http.StatusConflict), InternalMessage: errMsg, UserMessage: errMsg}
		}
		return nil, err
	}

	yamlSplits, err := kube.SplitYAML([]byte(templateChartResponse.GeneratedManifest))
	for _, yaml := range yamlSplits {
		if (manifestRequest.ResourceType == bean.CM && yaml.GetKind() == "ConfigMap") || (manifestRequest.ResourceType == bean.CS && yaml.GetKind() == "Secret") {
			name := yaml.GetName()
			if name == manifestRequest.ResourceName || name == fmt.Sprintf("%s-%d", manifestRequest.ResourceName, appId) {
				yamlJSON, err := yaml.MarshalJSON()
				if err != nil {
					return nil, err
				}
				return &bean2.ManifestResponse{Manifest: string(yamlJSON)}, nil
			}
		}
	}

	return &bean2.ManifestResponse{Manifest: ""}, nil
}

func (impl *DeploymentConfigurationServiceImpl) getMergedValuesForCMCSHelmTemplate(manifestRequest *bean2.ManifestRequest, resolvedTemplate string, app *appRepository.App, envId int) ([]byte, error) {
	var (
		CMCSValues []byte
		err        error
	)
	switch manifestRequest.ResourceType {
	case bean.CM:
		ConfigMapRoot := bean4.ConfigMapRootJson{
			ConfigMapJson: bean4.ConfigMapJson{
				Enabled: true,
				Maps: []bean4.ConfigSecretMap{
					{
						Name: manifestRequest.ResourceName,
						Data: json.RawMessage(resolvedTemplate),
					},
				},
			},
		}
		CMCSValues, err = json.Marshal(ConfigMapRoot)
		if err != nil {
			impl.logger.Errorw("error in marshalling config map obj", "appId", manifestRequest.AppId, "envId", manifestRequest.EnvironmentId, "err", err)
			return nil, err
		}
	case bean.CS:

		secretMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(resolvedTemplate), &secretMap)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling secret draft data ", "err", err)
			return nil, err
		}
		for key, _ := range secretMap {
			if !manifestRequest.UserHasAdminAccess {
				secretMap[key] = bean2.SecretMaskedValue
			}
		}
		decodedSecretData, err := json.Marshal(secretMap)
		if err != nil {
			impl.logger.Errorw("error in decoding secret data", "err", err)
			return nil, err
		}
		resolvedTemplate = string(decodedSecretData)

		ConfigMapRoot := bean4.ConfigSecretRootJson{
			ConfigSecretJson: bean4.ConfigSecretJson{
				Enabled: true,
				Secrets: []*bean4.ConfigSecretMap{
					{
						Name: manifestRequest.ResourceName,
						Data: json.RawMessage(resolvedTemplate),
					},
				},
			},
		}
		CMCSValues, err = json.Marshal(ConfigMapRoot)
		if err != nil {
			impl.logger.Errorw("error in marshalling config map obj", "appId", manifestRequest.AppId, "envId", manifestRequest.EnvironmentId, "err", err)
			return nil, err
		}
	}
	labelValues := struct {
		App int `json:"app"`
		Env int `json:"env"`
	}{
		App: app.Id,
		Env: envId,
	}

	labelValuesJSON, err := json.Marshal(labelValues)
	if err != nil {
		return nil, err
	}

	mergedValues, err := impl.mergeUtil.JsonPatch(CMCSValues, labelValuesJSON)
	if err != nil {
		return nil, err
	}

	mergedValuesYAML, err := yaml.JSONToYAML(mergedValues)
	if err != nil {
		return nil, err
	}
	return mergedValuesYAML, nil
}

func (impl *DeploymentConfigurationServiceImpl) getRefChartBytes(ctx context.Context, envId int, appId int, app *appRepository.App) (*chartRefBean.ChartRefDto, []byte, error) {

	chartRefId, err := impl.getConfiguredChartRef(envId, appId)
	if err != nil {
		impl.logger.Errorw("error in getting configured chart ref", "appId", appId, "envId", envId, "err", err)
		return nil, nil, err
	}

	refChart, err := impl.chartRefService.FindById(chartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting refChart", "err", err, "chartRefId", chartRefId)
		return nil, nil, err
	}

	chartMetaData := &chart2.Metadata{
		Name:    app.AppName,
		Version: refChart.Version,
	}

	refChartPath, err := impl.chartRefService.GetChartLocation(refChart.Location, refChart.ChartData)
	if err != nil {
		impl.logger.Errorw("error in getting chart location", "chartMetaData", chartMetaData, "refChartLocation", refChart.Location)
		return nil, nil, err
	}

	tempReferenceTemplateDir, err := impl.chartTemplateService.BuildChart(ctx, chartMetaData, refChartPath)
	if err != nil {
		impl.logger.Errorw("error in building chart", "chartMetaData", chartMetaData, "refChartPath", refChartPath)
		return nil, nil, err
	}

	chartInBytes, err := impl.chartTemplateService.LoadChartInBytes(tempReferenceTemplateDir, true)
	if err != nil {
		impl.logger.Errorw("error in loading chart bytes from dir", "dir", tempReferenceTemplateDir, "chartMetadata", chartMetaData, "err", err)
		return nil, nil, err
	}

	return refChart, chartInBytes, nil
}

func (impl *DeploymentConfigurationServiceImpl) getConfiguredChartRef(envId int, appId int) (int, error) {
	var chartRefId int
	if envId > 0 {
		envOverride, err := impl.envConfigOverrideService.FindLatestChartForAppByAppIdAndEnvId(appId, envId)
		if err != nil && !errors.IsNotFound(err) {
			impl.logger.Errorw("error in fetching latest chart", "err", err)
			return 0, nil
		}
		if envOverride != nil && envOverride.Chart != nil {
			chartRefId = envOverride.Chart.ChartRefId
		} else {
			chart, err := impl.chartService.FindLatestChartForAppByAppId(appId)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error in fetching latest chart", "err", err)
				return 0, nil
			}

			chartRefId = chart.ChartRefId
		}
	} else {
		chart, err := impl.chartService.FindLatestChartForAppByAppId(appId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in fetching latest chart", "err", err)
			return 0, nil
		}
		chartRefId = chart.ChartRefId
	}
	return chartRefId, nil
}

func (impl *DeploymentConfigurationServiceImpl) getConfigDataForCdRollback(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfigDto, error) {
	// wfrId is expected in this case to return the expected data
	if configDataQueryParams.WfrId == 0 {
		return nil, &util.ApiError{HttpStatusCode: http.StatusNotFound, Code: strconv.Itoa(http.StatusNotFound), InternalMessage: bean2.ExpectedWfrIdNotPassedInQueryParamErr, UserMessage: bean2.ExpectedWfrIdNotPassedInQueryParamErr}
	}
	return impl.getConfigDataForDeploymentHistory(ctx, configDataQueryParams, userHasAdminAccess)
}

func (impl *DeploymentConfigurationServiceImpl) getDeploymentHistoryConfig(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams) (*bean2.DeploymentAndCmCsConfig, error) {
	deploymentJson := json.RawMessage{}
	deploymentHistory, err := impl.deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId(configDataQueryParams.PipelineId, configDataQueryParams.WfrId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in getting deployment template history for pipelineId and wfrId", "pipelineId", configDataQueryParams.PipelineId, "wfrId", configDataQueryParams.WfrId, "err", err)
		return nil, err
	} else if util.IsErrNoRows(err) {
		return nil, util.NewApiError(http.StatusNotFound, bean2.NoDeploymentDoneForSelectedImage, bean2.NoDeploymentDoneForSelectedImage)
	}
	err = deploymentJson.UnmarshalJSON([]byte(deploymentHistory.Template))
	if err != nil {
		impl.logger.Errorw("getDeploymentTemplateForEnvLevel, error in unmarshalling string  deploymentTemplateResponse data into json Raw message", "data", deploymentHistory.Template, "err", err)
		return nil, err
	}
	isSuperAdmin, err := util2.GetIsSuperAdminFromContext(ctx)
	if err != nil {
		return nil, err
	}
	reference := repository6.HistoryReference{
		HistoryReferenceId:   deploymentHistory.Id,
		HistoryReferenceType: repository6.HistoryReferenceTypeDeploymentTemplate,
	}
	variableSnapshotMap, resolvedTemplate, err := impl.scopedVariableManager.GetVariableSnapshotAndResolveTemplate(deploymentHistory.Template, parsers.JsonVariableTemplate, reference, isSuperAdmin, false)
	if err != nil {
		impl.logger.Errorw("error while resolving template from history", "deploymentHistoryId", deploymentHistory.Id, "pipelineId", configDataQueryParams.PipelineId, "err", err)
	}

	deploymentConfig := bean2.NewDeploymentAndCmCsConfig().
		WithConfigData(deploymentJson).
		WithResourceType(bean.DeploymentTemplate).
		WithVariableSnapshot(map[string]map[string]string{bean.DeploymentTemplate.ToString(): variableSnapshotMap}).
		WithResolvedValue(json.RawMessage(resolvedTemplate)).
		WithDeploymentConfigMetadata(deploymentHistory.TemplateVersion, deploymentHistory.IsAppMetricsEnabled)
	return deploymentConfig, nil
}

func (impl *DeploymentConfigurationServiceImpl) getPipelineStrategyConfigHistory(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams) (*bean2.DeploymentAndCmCsConfig, error) {
	pipelineStrategyJson := json.RawMessage{}
	pipelineConfig := bean2.NewDeploymentAndCmCsConfig()
	pipelineStrategyHistory, err := impl.pipelineStrategyHistoryRepository.GetHistoryByPipelineIdAndWfrId(ctx, configDataQueryParams.PipelineId, configDataQueryParams.WfrId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in checking if history exists for pipelineId and wfrId", "pipelineId", configDataQueryParams.PipelineId, "wfrId", configDataQueryParams.WfrId, "err", err)
		return nil, err
	} else if util.IsErrNoRows(err) {
		return pipelineConfig, nil
	}
	err = pipelineStrategyJson.UnmarshalJSON([]byte(pipelineStrategyHistory.Config))
	if err != nil {
		impl.logger.Errorw("getDeploymentTemplateForEnvLevel, error in unmarshalling string  pipelineStrategyHistory data into json Raw message", "pipelineStrategyHistoryConfig", pipelineStrategyHistory.Config, "err", err)
		return nil, err
	}
	pipelineConfig.WithConfigData(pipelineStrategyJson).
		WithResourceType(bean.PipelineStrategy).
		WithPipelineStrategyMetadata(pipelineStrategyHistory.PipelineTriggerType, string(pipelineStrategyHistory.Strategy))
	return pipelineConfig, nil
}

func (impl *DeploymentConfigurationServiceImpl) getConfigDataForDeploymentHistory(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfigDto, error) {
	// we would be expecting wfrId in case of getting data for Deployment history
	configDataDto := &bean2.DeploymentAndCmCsConfigDto{}
	var err error
	//fetching history for deployment config starts
	deploymentConfig, err := impl.getDeploymentHistoryConfig(ctx, configDataQueryParams)
	if err != nil {
		impl.logger.Errorw("getConfigDataForDeploymentHistory, error in getDeploymentHistoryConfig", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	configDataDto.WithDeploymentTemplateData(deploymentConfig)
	// fetching for deployment config ends

	// fetching for pipeline strategy config starts
	pipelineConfig, err := impl.getPipelineStrategyConfigHistory(ctx, configDataQueryParams)
	if err != nil {
		impl.logger.Errorw("getConfigDataForDeploymentHistory, error in getPipelineStrategyConfigHistory", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	if len(pipelineConfig.Data) > 0 {
		configDataDto.WithPipelineConfigData(pipelineConfig)
	}

	// fetching for pipeline strategy config ends

	// fetching for cm config starts
	cmConfigData, err := impl.getCmCsConfigHistory(ctx, configDataQueryParams, repository3.CONFIGMAP_TYPE, userHasAdminAccess)
	if err != nil {
		impl.logger.Errorw("getConfigDataForDeploymentHistory, error in getCmConfigHistory", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	configDataDto.WithConfigMapData(cmConfigData)
	// fetching for cm config ends
	isWfrFirstDeployment, err := impl.IsFirstDeployment(configDataQueryParams.PipelineId, configDataQueryParams.WfrId)
	if err != nil {
		impl.logger.Errorw("error in checking if a single deployment history is present for pipelineId or not", "pipelineId", configDataQueryParams.PipelineId, "err", err)
		return nil, err
	}
	if userHasAdminAccess || isWfrFirstDeployment {
		// fetching for cs config starts
		secretConfigDto, err := impl.getCmCsConfigHistory(ctx, configDataQueryParams, repository3.SECRET_TYPE, userHasAdminAccess)
		if err != nil {
			impl.logger.Errorw("getConfigDataForDeploymentHistory, error in getSecretConfigHistory", "configDataQueryParams", configDataQueryParams, "err", err)
			return nil, err
		}
		configDataDto.WithSecretData(secretConfigDto)
		// fetching for cs config ends
	}

	return configDataDto, nil
}

func (impl *DeploymentConfigurationServiceImpl) IsFirstDeployment(pipelineId, wfrId int) (bool, error) {
	wfrs, err := impl.cdWorkflowRepository.FindDeployedCdWorkflowRunnersByPipelineId(pipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting all cd workflow runners for a pipeline id", "pipelineId", pipelineId, "err", err)
		return false, err
	}
	if len(wfrs) > 0 && wfrs[0].Id == wfrId {
		return true, nil
	}
	return false, nil
}

func (impl *DeploymentConfigurationServiceImpl) getCmCsConfigHistory(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams, configType repository3.ConfigType, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfig, error) {
	var resourceType bean.ResourceType
	history, err := impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(configDataQueryParams.PipelineId, configDataQueryParams.WfrId, configType)
	if err != nil {
		impl.logger.Errorw("error in checking if cm cs history exists for pipelineId and wfrId", "pipelineId", configDataQueryParams.PipelineId, "wfrId", configDataQueryParams.WfrId, "err", err)
		return nil, err
	}
	var configData []*bean.ConfigData
	configList := bean.ConfigsList{}
	secretList := bean.SecretsList{}
	switch configType {
	case repository3.CONFIGMAP_TYPE:
		if len(history.Data) > 0 {
			err = json.Unmarshal([]byte(history.Data), &configList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		resourceType = bean.CM
		configData = configList.ConfigData
	case repository3.SECRET_TYPE:
		if len(history.Data) > 0 {
			err = json.Unmarshal([]byte(history.Data), &secretList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		resourceType = bean.CS
		configData = secretList.ConfigData

	}

	resolvedDataMap, variableSnapshotMap, err := impl.scopedVariableManager.GetResolvedCMCSHistoryDtos(ctx, configType, adaptor.ReverseConfigListConvertor(configList), history, adaptor.ReverseSecretListConvertor(secretList))
	if err != nil {
		return nil, err
	}
	resolvedConfigDataList := make([]*bean.ConfigData, 0, len(resolvedDataMap))
	for _, resolvedConfigData := range resolvedDataMap {
		resolvedConfigDataList = append(resolvedConfigDataList, adapter.ConvertConfigDataToPipelineConfigData(&resolvedConfigData))
	}

	if configType == repository3.SECRET_TYPE {
		impl.encodeSecretDataFromNonAdminUsers(configData, userHasAdminAccess, bean2.SecretMaskedValue)
		impl.encodeSecretDataFromNonAdminUsers(resolvedConfigDataList, userHasAdminAccess, bean2.SecretMaskedValue)
	}

	configDataReq := &bean.ConfigDataRequest{ConfigData: configData}
	configDataJson, err := utils.ConvertToJsonRawMessage(configDataReq)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in converting config data to json raw message", "pipelineId", configDataQueryParams.PipelineId, "wfrId", configDataQueryParams.WfrId, "err", err)
		return nil, err
	}
	resolvedConfigDataReq := &bean.ConfigDataRequest{ConfigData: resolvedConfigDataList}
	resolvedConfigDataString, err := utils.ConvertToString(resolvedConfigDataReq)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in converting config data to json raw message", "pipelineId", configDataQueryParams.PipelineId, "wfrId", configDataQueryParams.WfrId, "err", err)
		return nil, err
	}
	resolvedConfigDataStringJson, err := utils.ConvertToJsonRawMessage(resolvedConfigDataString)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in ConvertToJsonRawMessage for resolvedConfigDataString", "pipelineId", configDataQueryParams.PipelineId, "wfrId", configDataQueryParams.WfrId, "err", err)
		return nil, err
	}
	cmConfigData := bean2.NewDeploymentAndCmCsConfig().
		WithConfigData(configDataJson).
		WithResourceType(resourceType).
		WithVariableSnapshot(variableSnapshotMap).
		WithResolvedValue(resolvedConfigDataStringJson)
	return cmConfigData, nil
}

func (impl *DeploymentConfigurationServiceImpl) encodeSecretDataFromNonAdminUsers(configDataList []*bean.ConfigData, userHasAdminAccess bool, secretMaskedValue string) {
	for _, config := range configDataList {
		if config.Data != nil {
			if !userHasAdminAccess {
				//removing keys and sending
				resultMap := make(map[string]string)
				resultMapFinal := make(map[string]string)
				err := json.Unmarshal(config.Data, &resultMap)
				if err != nil {
					impl.logger.Errorw("unmarshal failed", "error", err)
					return
				}
				for key, _ := range resultMap {
					//hard-coding values to show them as hidden to user
					resultMapFinal[key] = secretMaskedValue
				}
				config.Data, err = utils.ConvertToJsonRawMessage(resultMapFinal)
				if err != nil {
					impl.logger.Errorw("error while marshaling request", "err", err)
					return
				}
			}
		}
	}
}

func (impl *DeploymentConfigurationServiceImpl) getCmCsDataForPreviousDeployments(ctx context.Context, wfrId, pipelineId int, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfigDto, error) {

	configDataDto := &bean2.DeploymentAndCmCsConfigDto{}

	cmConfigData, err := impl.configMapHistoryReadService.GetCmCsHistoryByWfrIdAndPipelineId(ctx, pipelineId, wfrId, repository3.CONFIGMAP_TYPE, userHasAdminAccess)
	if err != nil {
		impl.logger.Errorw("error in getting secretData and cmData", "err", err, "wfrId", wfrId, "pipelineId", pipelineId)
		return nil, err
	}
	configDataDto.WithConfigMapData(cmConfigData)
	if userHasAdminAccess {
		secretConfigData, err := impl.configMapHistoryReadService.GetCmCsHistoryByWfrIdAndPipelineId(ctx, pipelineId, wfrId, repository3.SECRET_TYPE, userHasAdminAccess)
		if err != nil {
			impl.logger.Errorw("error in getting secretData and cmData", "err", err, "wfrId", wfrId, "pipelineId", pipelineId)
			return nil, err
		}
		configDataDto.WithSecretData(secretConfigData)
	}

	return configDataDto, nil

}
func (impl *DeploymentConfigurationServiceImpl) getPipelineStrategyForPreviousDeployments(ctx context.Context, wfrId, pipelineId int) (*bean2.DeploymentAndCmCsConfig, error) {
	pipelineStrategyJson := json.RawMessage{}
	pipelineConfig := bean2.NewDeploymentAndCmCsConfig()
	pipelineStrategyHistory, err := impl.pipelineStrategyHistoryRepository.GetHistoryByPipelineIdAndWfrId(ctx, pipelineId, wfrId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in FindPipelineStrategyForDeployedOnAndPipelineId", "wfrId", wfrId, "pipelineId", pipelineId, "err", err)
		return nil, err
	} else if util.IsErrNoRows(err) {
		return pipelineConfig, nil
	}
	if pipelineStrategyHistory != nil {
		err = pipelineStrategyJson.UnmarshalJSON([]byte(pipelineStrategyHistory.Config))
		if err != nil {
			impl.logger.Errorw("getDeploymentTemplateForEnvLevel, error in unmarshalling string  pipelineStrategyHistory data into json Raw message", "err", err)
			return nil, err
		}
		pipelineConfig.WithConfigData(pipelineStrategyJson).
			WithResourceType(bean.PipelineStrategy).
			WithPipelineStrategyMetadata(pipelineStrategyHistory.PipelineTriggerType, string(pipelineStrategyHistory.Strategy))
	}
	return pipelineConfig, nil
}

func (impl *DeploymentConfigurationServiceImpl) getDeploymentsConfigForPreviousDeployments(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams) (generateManifest.DeploymentTemplateResponse, error) {

	var deploymentTemplateResponse generateManifest.DeploymentTemplateResponse
	deplTemplate, err := impl.deploymentTemplateHistoryReadService.GetDeployedHistoryByPipelineIdAndWfrId(ctx, configDataQueryParams.PipelineId, configDataQueryParams.WfrId)
	if err != nil {
		impl.logger.Errorw("getDeploymentsConfigForPreviousDeployments, error in getting deployment template  by pipelineId and wfrId ", "pipelineId", configDataQueryParams.PipelineId, "wfrId", configDataQueryParams.WfrId, "err", err)
		return deploymentTemplateResponse, err
	}
	deploymentTemplateResponse = generateManifest.DeploymentTemplateResponse{
		Data:                deplTemplate.CodeEditorValue.Value,
		ResolvedData:        deplTemplate.CodeEditorValue.ResolvedValue,
		VariableSnapshot:    deplTemplate.CodeEditorValue.VariableSnapshot,
		TemplateVersion:     deplTemplate.TemplateVersion,
		IsAppMetricsEnabled: *deplTemplate.IsAppMetricsEnabled,
	}
	return deploymentTemplateResponse, nil
}

func (impl *DeploymentConfigurationServiceImpl) getDeploymentAndCmCsConfigDataForPreviousDeployments(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams,
	appId, envId int, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfigDto, error) {

	// getting DeploymentAndCmCsConfigDto obj with cm and cs data populated
	configDataDto, err := impl.getCmCsDataForPreviousDeployments(ctx, configDataQueryParams.WfrId, configDataQueryParams.PipelineId, userHasAdminAccess)
	if err != nil {
		impl.logger.Errorw("error in getting cm cs for PreviousDeployments state", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	pipelineStrategy, err := impl.getPipelineStrategyForPreviousDeployments(ctx, configDataQueryParams.WfrId, configDataQueryParams.PipelineId)
	if err != nil {
		impl.logger.Errorw(" error in getting cm cs for PreviousDeployments state", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	if len(pipelineStrategy.Data) > 0 {
		configDataDto.WithPipelineConfigData(pipelineStrategy)
	}

	deploymentTemplateData, err := impl.getDeploymentsConfigForPreviousDeployments(ctx, configDataQueryParams)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	deploymentJson := json.RawMessage{}
	err = deploymentJson.UnmarshalJSON([]byte(deploymentTemplateData.Data))
	if err != nil {
		impl.logger.Errorw("error in unmarshalling string  deploymentTemplateResponse data into json Raw message", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	variableSnapShotMap := map[string]map[string]string{bean.DeploymentTemplate.ToString(): deploymentTemplateData.VariableSnapshot}

	deploymentConfig := bean2.NewDeploymentAndCmCsConfig().
		WithDeploymentConfigMetadata(deploymentTemplateData.TemplateVersion, deploymentTemplateData.IsAppMetricsEnabled).
		WithConfigData(deploymentJson).
		WithResourceType(bean.DeploymentTemplate).
		WithResolvedValue(json.RawMessage(deploymentTemplateData.ResolvedData)).
		WithVariableSnapshot(variableSnapShotMap)

	configDataDto.WithDeploymentTemplateData(deploymentConfig)

	return configDataDto, nil
}

func (impl *DeploymentConfigurationServiceImpl) getConfigDataForAppConfiguration(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams,
	appId, envId, clusterId int, userHasAdminAccess bool, systemMetadata *resourceQualifiers.SystemMetadata) (*bean2.DeploymentAndCmCsConfigDto, error) {
	configDataDto := &bean2.DeploymentAndCmCsConfigDto{}
	var err error
	switch configDataQueryParams.ConfigType {
	case bean2.DefaultVersion.ToString():
		configDataDto, err = impl.getDeploymentAndCmCsConfigDataForDefaultVersion(ctx, configDataQueryParams)
		if err != nil {
			impl.logger.Errorw("GetAllConfigData, error in config data for Default version", "configDataQueryParams", configDataQueryParams, "err", err)
			return nil, err
		}
		//no cm or cs to send for default versions
	case bean2.PreviousDeployments.ToString():
		configDataDto, err = impl.getDeploymentAndCmCsConfigDataForPreviousDeployments(ctx, configDataQueryParams, appId, envId, userHasAdminAccess)
		if err != nil {
			impl.logger.Errorw("GetAllConfigData, error in config data for Previous Deployments", "configDataQueryParams", configDataQueryParams, "err", err)
			return nil, err
		}
	default: // keeping default as PublishedOnly
		configDataDto, err = impl.getPublishedConfigData(ctx, configDataQueryParams, appId, envId, clusterId, userHasAdminAccess, systemMetadata)
		if err != nil {
			impl.logger.Errorw("GetAllConfigData, error in config data for PublishedOnly", "configDataQueryParams", configDataQueryParams, "err", err)
			return nil, err
		}
	}
	return configDataDto, nil
}

func (impl *DeploymentConfigurationServiceImpl) getDeploymentsConfigForDefaultVersion(ctx context.Context, chartRefId int) (json.RawMessage, error) {
	deploymentTemplateRequest := generateManifest.DeploymentTemplateRequest{
		ChartRefId:      chartRefId,
		RequestDataMode: generateManifest.Values,
		Type:            repository2.DefaultVersions,
	}
	deploymentTemplateResponse, err := impl.deploymentTemplateService.GetDeploymentTemplate(ctx, deploymentTemplateRequest)
	if err != nil {
		impl.logger.Errorw("getDeploymentTemplateForEnvLevel, error in getting deployment template for ", "deploymentTemplateRequest", deploymentTemplateRequest, "err", err)
		return nil, err
	}
	deploymentJson := json.RawMessage{}
	err = deploymentJson.UnmarshalJSON([]byte(deploymentTemplateResponse.Data))
	if err != nil {
		impl.logger.Errorw("getDeploymentTemplateForEnvLevel, error in unmarshalling string  deploymentTemplateResponse data into json Raw message", "data", deploymentTemplateResponse.Data, "err", err)
		return nil, err
	}
	return deploymentJson, nil
}

func (impl *DeploymentConfigurationServiceImpl) getDeploymentAndCmCsConfigDataForDefaultVersion(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams) (*bean2.DeploymentAndCmCsConfigDto, error) {
	configData := &bean2.DeploymentAndCmCsConfigDto{}
	deploymentTemplateJsonData, err := impl.getDeploymentsConfigForDefaultVersion(ctx, configDataQueryParams.IdentifierId)
	if err != nil {
		impl.logger.Errorw("GetAllConfigData, error in getting deployment config for default version", "chartRefId", configDataQueryParams.IdentifierId, "err", err)
		return nil, err
	}
	deploymentConfig := bean2.NewDeploymentAndCmCsConfig().WithConfigData(deploymentTemplateJsonData).WithResourceType(bean.DeploymentTemplate)
	configData.WithDeploymentTemplateData(deploymentConfig)
	return configData, nil
}

func (impl *DeploymentConfigurationServiceImpl) getCmCsEditDataForPublishedOnly(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams, envId,
	appId int, clusterId int, userHasAdminAccess bool, systemMetadata *resourceQualifiers.SystemMetadata) (*bean2.DeploymentAndCmCsConfigDto, error) {
	configDataDto := &bean2.DeploymentAndCmCsConfigDto{}

	var resourceType bean.ResourceType
	var fetchConfigFunc func(string, int, int, int) (*bean.ConfigDataRequest, error)

	if configDataQueryParams.IsResourceTypeSecret() {
		//handles for single resource when resource type is secret and for a given resource name
		resourceType = bean.CS
		fetchConfigFunc = impl.getSecretConfigResponse
	} else if configDataQueryParams.IsResourceTypeConfigMap() {
		//handles for single resource when resource type is configMap and for a given resource name
		resourceType = bean.CM
		fetchConfigFunc = impl.getConfigMapResponse
	}
	cmcsConfigData, err := fetchConfigFunc(configDataQueryParams.ResourceName, configDataQueryParams.ResourceId, envId, appId)
	if err != nil {
		impl.logger.Errorw("error in getting config response", "resourceName", configDataQueryParams.ResourceName, "envName", configDataQueryParams.EnvName, "err", err)
		return nil, err
	}
	if configDataQueryParams.IsResourceTypeSecret() && !userHasAdminAccess {
		_, err := utils.GetKeyValMapForSecretConfigDataAndMaskData(cmcsConfigData.ConfigData)
		if err != nil {
			impl.logger.Errorw("error in getting config response", "resourceName", configDataQueryParams.ResourceName, "envName", configDataQueryParams.EnvName, "err", err)
			return nil, err
		}
	}

	respJson, err := utils.ConvertToJsonRawMessage(cmcsConfigData)
	if err != nil {
		impl.logger.Errorw("getCmCsEditDataForPublishedOnly, error in converting to json raw message", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	resolvedCmCsMetadataDto, err := impl.ResolveCmCs(ctx, envId, appId, clusterId, userHasAdminAccess, configDataQueryParams.ResourceName, resourceType, systemMetadata)
	if err != nil {
		impl.logger.Errorw("error in resolving cm and cs for published only config only response", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	cmCsConfig := bean2.NewDeploymentAndCmCsConfig().WithConfigData(respJson).WithResourceType(resourceType)

	if resourceType == bean.CS {
		resolvedConfigDataStringJson, err := utils.ConvertToJsonRawMessage(resolvedCmCsMetadataDto.ResolvedSecretData)
		if err != nil {
			impl.logger.Errorw("getCmCsPublishedConfigResponse, error in ConvertToJsonRawMessage ", "err", err)
			return nil, err
		}
		cmCsConfig.WithResolvedValue(resolvedConfigDataStringJson).WithVariableSnapshot(resolvedCmCsMetadataDto.VariableMapCS)
		configDataDto.WithSecretData(cmCsConfig)
	} else if resourceType == bean.CM {
		resolvedConfigDataStringJson, err := utils.ConvertToJsonRawMessage(resolvedCmCsMetadataDto.ResolvedConfigMapData)
		if err != nil {
			impl.logger.Errorw("getCmCsPublishedConfigResponse, error in ConvertToJsonRawMessage for resolvedJson", "ResolvedConfigMapData", resolvedCmCsMetadataDto.ResolvedConfigMapData, "err", err)
			return nil, err
		}
		cmCsConfig.WithResolvedValue(resolvedConfigDataStringJson).WithVariableSnapshot(resolvedCmCsMetadataDto.VariableMapCM)
		configDataDto.WithConfigMapData(cmCsConfig)
	}
	return configDataDto, nil
}

func (impl *DeploymentConfigurationServiceImpl) getCmCsPublishedConfigResponse(ctx context.Context, envId, appId, clusterId int, userHasAdminAccess bool, systemMetadata *resourceQualifiers.SystemMetadata) (*bean2.DeploymentAndCmCsConfigDto, error) {
	configDataDto := &bean2.DeploymentAndCmCsConfigDto{}
	//iterate on secret configData and then and set draft data from draftResourcesMap if same resourceName found do the same for configMap below
	cmData, err := impl.getConfigMapResponse("", 0, envId, appId)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in getting config map by appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	cmRespJson, err := utils.ConvertToJsonRawMessage(cmData)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in converting config map data to json raw message", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	resolvedCmCsMetadataDto, err := impl.ResolveCmCs(ctx, envId, appId, clusterId, userHasAdminAccess, "", "", systemMetadata)
	if err != nil {
		impl.logger.Errorw("error in resolving cm and cs for published only config only response", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	resolvedConfigMapDataStringJson, err := utils.ConvertToJsonRawMessage(resolvedCmCsMetadataDto.ResolvedConfigMapData)
	if err != nil {
		impl.logger.Errorw("error in ConvertToJsonRawMessage for resolvedConfigMapDataStringJson", "resolvedCmData", resolvedCmCsMetadataDto.ResolvedConfigMapData, "err", err)
		return nil, err
	}
	cmConfigData := bean2.NewDeploymentAndCmCsConfig().WithConfigData(cmRespJson).WithResourceType(bean.CM).
		WithResolvedValue(resolvedConfigMapDataStringJson).WithVariableSnapshot(resolvedCmCsMetadataDto.VariableMapCM)
	configDataDto.WithConfigMapData(cmConfigData)

	if userHasAdminAccess {
		secretData, err := impl.getSecretConfigResponse("", 0, envId, appId)
		if err != nil {
			impl.logger.Errorw("getCmCsPublishedConfigResponse, error in getting secret config response by appId and envId", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		secretRespJson, err := utils.ConvertToJsonRawMessage(secretData)
		if err != nil {
			impl.logger.Errorw("getCmCsPublishedConfigResponse, error in converting secret data to json raw message", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		resolvedSecretDataStringJson, err := utils.ConvertToJsonRawMessage(resolvedCmCsMetadataDto.ResolvedSecretData)
		if err != nil {
			impl.logger.Errorw(" error in ConvertToJsonRawMessage for resolvedConfigDataString", "err", err)
			return nil, err
		}
		secretConfigData := bean2.NewDeploymentAndCmCsConfig().WithConfigData(secretRespJson).WithResourceType(bean.CS).
			WithResolvedValue(resolvedSecretDataStringJson).WithVariableSnapshot(resolvedCmCsMetadataDto.VariableMapCS)
		configDataDto.WithSecretData(secretConfigData)
	}
	return configDataDto, nil

}

func (impl *DeploymentConfigurationServiceImpl) getMergedCmCs(envId, appId int) (*bean2.CmCsMetadataDto, error) {
	configAppLevel, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Errorw("error in getting CM/CS app level data", "appId", appId, "err", err)
		return nil, err
	}
	var configMapAppLevel string
	var secretAppLevel string
	if configAppLevel != nil && configAppLevel.Id > 0 {
		configMapAppLevel = configAppLevel.ConfigMapData
		secretAppLevel = configAppLevel.SecretData
	}
	configEnvLevel, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Errorw("error in getting CM/CS env level data", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	var configMapEnvLevel string
	var secretEnvLevel string
	if configEnvLevel != nil && configEnvLevel.Id > 0 {
		configMapEnvLevel = configEnvLevel.ConfigMapData
		secretEnvLevel = configEnvLevel.SecretData
	}
	mergedConfigMap, err := impl.deploymentConfigService.GetMergedCMCSConfigMap(configMapAppLevel, configMapEnvLevel, repository3.CONFIGMAP_TYPE)
	if err != nil {
		impl.logger.Errorw("error in merging app level and env level CM configs", "err", err)
		return nil, err
	}

	mergedSecret, err := impl.deploymentConfigService.GetMergedCMCSConfigMap(secretAppLevel, secretEnvLevel, repository3.SECRET_TYPE)
	if err != nil {
		impl.logger.Errorw("error in merging app level and env level CM configs", "err", err)
		return nil, err
	}
	return &bean2.CmCsMetadataDto{
		CmMap:            mergedConfigMap,
		SecretMap:        mergedSecret,
		ConfigAppLevelId: configAppLevel.Id,
		ConfigEnvLevelId: configEnvLevel.Id,
	}, nil
}

func (impl *DeploymentConfigurationServiceImpl) ResolveCmCs(ctx context.Context, envId, appId, clusterId int, userHasAdminAccess bool,
	resourceName string, resourceType bean.ResourceType, systemMetadata *resourceQualifiers.SystemMetadata) (*bean2.ResolvedCmCsMetadataDto, error) {
	scope := resourceQualifiers.Scope{
		AppId:          appId,
		EnvId:          envId,
		ClusterId:      clusterId,
		SystemMetadata: systemMetadata,
	}
	cmcsMetadataDto, err := impl.getMergedCmCs(envId, appId)
	if err != nil {
		impl.logger.Errorw("error in getting merged cm cs", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	// if resourceName is provided then, resolve cmcs request is for single resource, then remove other data from merged cmCs
	if len(resourceName) > 0 {
		helper.FilterOutMergedCmCsForResourceName(cmcsMetadataDto, resourceName, resourceType)
	}
	resolvedConfigList, resolvedSecretList, variableMapCM, variableMapCS, err := impl.scopedVariableManager.ResolveCMCS(ctx, scope, cmcsMetadataDto.ConfigAppLevelId, cmcsMetadataDto.ConfigEnvLevelId, cmcsMetadataDto.CmMap, cmcsMetadataDto.SecretMap)
	if err != nil {
		impl.logger.Errorw("error in resolving CM/CS", "scope", scope, "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	resolvedConfigString, resolvedSecretString, err := impl.getStringifiedCmCs(resolvedConfigList, resolvedSecretList, userHasAdminAccess)
	if err != nil {
		impl.logger.Errorw("error in getStringifiedCmCs", "resolvedConfigList", resolvedConfigList, "err", err)
		return nil, err
	}
	resolvedData := &bean2.ResolvedCmCsMetadataDto{
		VariableMapCM:         variableMapCM,
		VariableMapCS:         variableMapCS,
		ResolvedSecretData:    resolvedSecretString,
		ResolvedConfigMapData: resolvedConfigString,
	}

	return resolvedData, nil
}

func (impl *DeploymentConfigurationServiceImpl) getStringifiedCmCs(resolvedCmMap map[string]*bean3.ConfigData, resolvedSecretMap map[string]*bean3.ConfigData,
	userHasAdminAccess bool) (string, string, error) {

	resolvedConfigDataList := make([]*bean.ConfigData, 0, len(resolvedCmMap))
	resolvedSecretDataList := make([]*bean.ConfigData, 0, len(resolvedSecretMap))

	for _, resolvedConfigData := range resolvedCmMap {
		resolvedConfigDataList = append(resolvedConfigDataList, adapter.ConvertConfigDataToPipelineConfigData(resolvedConfigData))
	}

	for _, resolvedSecretData := range resolvedSecretMap {
		resolvedSecretDataList = append(resolvedSecretDataList, adapter.ConvertConfigDataToPipelineConfigData(resolvedSecretData))
	}
	if len(resolvedSecretMap) > 0 {
		impl.encodeSecretDataFromNonAdminUsers(resolvedSecretDataList, userHasAdminAccess, bean2.SecretMaskedValue)
	}
	resolvedConfigDataReq := &bean.ConfigDataRequest{ConfigData: resolvedConfigDataList}
	resolvedConfigDataString, err := utils.ConvertToString(resolvedConfigDataReq)
	if err != nil {
		impl.logger.Errorw(" error in converting resolved config data to string", "resolvedConfigDataReq", resolvedConfigDataReq, "err", err)
		return "", "", err
	}
	resolvedSecretDataReq := &bean.ConfigDataRequest{ConfigData: resolvedSecretDataList}
	resolvedSecretDataString, err := utils.ConvertToString(resolvedSecretDataReq)
	if err != nil {
		impl.logger.Errorw(" error in converting resolved config data to string", "err", err)
		return "", "", err
	}
	return resolvedConfigDataString, resolvedSecretDataString, nil
}
func (impl *DeploymentConfigurationServiceImpl) getPublishedDeploymentConfig(ctx context.Context, appId, envId int) (*bean2.DeploymentAndCmCsConfig, error) {
	if envId > 0 {
		deplTemplateResp, err := impl.getDeploymentTemplateForEnvLevel(ctx, appId, envId)
		if err != nil {
			impl.logger.Errorw("error in getting deployment template env level", "err", err)
			return nil, err
		}
		deploymentJson := json.RawMessage{}
		err = deploymentJson.UnmarshalJSON([]byte(deplTemplateResp.Data))
		if err != nil {
			impl.logger.Errorw("getDeploymentTemplateForEnvLevel, error in unmarshalling string  deploymentTemplateResponse data into json Raw message", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}

		variableSnapShotMap := make(map[string]map[string]string, len(deplTemplateResp.VariableSnapshot))
		variableSnapShotMap[bean.DeploymentTemplate.ToString()] = deplTemplateResp.VariableSnapshot

		return bean2.NewDeploymentAndCmCsConfig().WithConfigData(deploymentJson).WithResourceType(bean.DeploymentTemplate).
			WithResolvedValue(json.RawMessage(deplTemplateResp.ResolvedData)).WithVariableSnapshot(variableSnapShotMap).
			WithDeploymentConfigMetadata(deplTemplateResp.TemplateVersion, deplTemplateResp.IsAppMetricsEnabled), nil
	}
	deplMetadata, err := impl.getBaseDeploymentTemplate(appId)
	if err != nil {
		impl.logger.Errorw("getting base depl. template", "appid", appId, "err", err)
		return nil, err
	}
	deploymentTemplateRequest := generateManifest.DeploymentTemplateRequest{
		AppId:           appId,
		RequestDataMode: generateManifest.Values,
	}
	resolvedTemplate, variableSnapshot, err := impl.deploymentTemplateService.ResolveTemplateVariables(ctx, string(deplMetadata.DeploymentTemplateJson), deploymentTemplateRequest)
	if err != nil {
		impl.logger.Errorw("error in getting resolved data for base deployment template", "appid", appId, "err", err)
		return nil, err
	}

	variableSnapShotMap := map[string]map[string]string{bean.DeploymentTemplate.ToString(): variableSnapshot}
	return bean2.NewDeploymentAndCmCsConfig().WithConfigData(deplMetadata.DeploymentTemplateJson).WithResourceType(bean.DeploymentTemplate).
		WithResolvedValue(json.RawMessage(resolvedTemplate)).WithVariableSnapshot(variableSnapShotMap).
		WithDeploymentConfigMetadata(deplMetadata.TemplateVersion, deplMetadata.IsAppMetricsEnabled), nil
}

func (impl *DeploymentConfigurationServiceImpl) getPublishedConfigData(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams,
	appId, envId, clusterId int, userHasAdminAccess bool, systemMetadata *resourceQualifiers.SystemMetadata) (*bean2.DeploymentAndCmCsConfigDto, error) {

	if configDataQueryParams.IsRequestMadeForOneResource() {
		return impl.getCmCsEditDataForPublishedOnly(ctx, configDataQueryParams, envId, appId, clusterId, userHasAdminAccess, systemMetadata)
	}
	//ConfigMapsData and SecretsData are populated here
	configData, err := impl.getCmCsPublishedConfigResponse(ctx, envId, appId, clusterId, userHasAdminAccess, systemMetadata)
	if err != nil {
		impl.logger.Errorw("getPublishedConfigData, error in getting cm cs for PublishedOnly state", "appName", configDataQueryParams.AppName, "envName", configDataQueryParams.EnvName, "err", err)
		return nil, err
	}
	deploymentTemplateData, err := impl.getPublishedDeploymentConfig(ctx, appId, envId)
	if err != nil {
		impl.logger.Errorw("getPublishedConfigData, error in getting publishedOnly deployment config ", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	configData.WithDeploymentTemplateData(deploymentTemplateData)

	pipelineConfigData, err := impl.getPublishedPipelineStrategyConfig(ctx, appId, envId)
	if err != nil {
		impl.logger.Errorw("getPublishedConfigData, error in getting publishedOnly pipeline strategy ", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	if len(pipelineConfigData.Data) > 0 {
		configData.WithPipelineConfigData(pipelineConfigData)
	}
	return configData, nil
}

func (impl *DeploymentConfigurationServiceImpl) getPublishedPipelineStrategyConfig(ctx context.Context, appId int, envId int) (*bean2.DeploymentAndCmCsConfig, error) {
	pipelineConfig := bean2.NewDeploymentAndCmCsConfig()
	if envId == 0 {
		return pipelineConfig, nil
	}
	pipeline, err := impl.pipelineRepository.FindActiveByAppIdAndEnvId(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in FindActiveByAppIdAndEnvId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	pipelineStrategy, err := impl.deploymentConfigService.GetLatestPipelineStrategyConfig(pipeline)
	if err != nil && !errors.IsNotFound(err) {
		impl.logger.Errorw("error in GetLatestPipelineStrategyConfig", "pipelineId", pipeline.Id, "err", err)
		return nil, err
	} else if errors.IsNotFound(err) {
		return pipelineConfig, nil
	}
	pipelineStrategyJson := json.RawMessage{}
	err = pipelineStrategyJson.UnmarshalJSON([]byte(pipelineStrategy.CodeEditorValue.Value))
	if err != nil {
		impl.logger.Errorw("getDeploymentTemplateForEnvLevel, error in unmarshalling string  pipelineStrategyHistory data into json Raw message", "err", err)
		return nil, err
	}
	pipelineConfig.WithConfigData(pipelineStrategyJson).
		WithResourceType(bean.PipelineStrategy).
		WithPipelineStrategyMetadata(pipelineStrategy.PipelineTriggerType, string(pipelineStrategy.Strategy))
	return pipelineConfig, nil
}

func (impl *DeploymentConfigurationServiceImpl) getBaseDeploymentTemplate(appId int) (*bean2.DeploymentTemplateMetadata, error) {
	deploymentTemplateData, err := impl.chartService.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in getting base deployment template for appId", "appId", appId, "err", err)
		return nil, err
	}
	_, _, version, _, err := impl.chartRefService.GetRefChart(deploymentTemplateData.ChartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting chart ref by chartRefId ", "chartRefId", deploymentTemplateData.ChartRefId, "err", err)
		return nil, err
	}
	return &bean2.DeploymentTemplateMetadata{
		DeploymentTemplateJson: deploymentTemplateData.DefaultAppOverride,
		IsAppMetricsEnabled:    deploymentTemplateData.IsAppMetricsEnabled,
		TemplateVersion:        version,
	}, nil
}

func (impl *DeploymentConfigurationServiceImpl) getDeploymentTemplateForEnvLevel(ctx context.Context, appId, envId int) (generateManifest.DeploymentTemplateResponse, error) {
	deploymentTemplateRequest := generateManifest.DeploymentTemplateRequest{
		AppId:           appId,
		EnvId:           envId,
		RequestDataMode: generateManifest.Values,
		Type:            repository2.PublishedOnEnvironments,
	}
	var deploymentTemplateResponse generateManifest.DeploymentTemplateResponse
	var err error
	deploymentTemplateResponse, err = impl.deploymentTemplateService.GetDeploymentTemplate(ctx, deploymentTemplateRequest)
	if err != nil {
		impl.logger.Errorw("getDeploymentTemplateForEnvLevel, error in getting deployment template for ", "deploymentTemplateRequest", deploymentTemplateRequest, "err", err)
		return deploymentTemplateResponse, err
	}
	return deploymentTemplateResponse, nil
}

func (impl *DeploymentConfigurationServiceImpl) getSecretConfigResponse(resourceName string, resourceId, envId, appId int) (*bean.ConfigDataRequest, error) {
	if len(resourceName) > 0 {
		if envId > 0 {
			return impl.configMapService.CSEnvironmentFetchForEdit(resourceName, resourceId, appId, envId)
		}
		return impl.configMapService.CmCsConfigGlobalFetchUsingAppId(resourceName, appId, bean.CS)
	}

	if envId > 0 {
		return impl.configMapService.CSEnvironmentFetch(appId, envId)
	}
	return impl.configMapService.CSGlobalFetch(appId)
}

func (impl *DeploymentConfigurationServiceImpl) getConfigMapResponse(resourceName string, resourceId, envId, appId int) (*bean.ConfigDataRequest, error) {
	if len(resourceName) > 0 {
		if envId > 0 {
			return impl.configMapService.CMEnvironmentFetchForEdit(resourceName, resourceId, appId, envId)
		}
		return impl.configMapService.CmCsConfigGlobalFetchUsingAppId(resourceName, appId, bean.CM)
	}

	if envId > 0 {
		return impl.configMapService.CMEnvironmentFetch(appId, envId)
	}
	return impl.configMapService.CMGlobalFetch(appId)
}

func (impl *DeploymentConfigurationServiceImpl) CompareCategoryWiseConfigData(ctx context.Context, comparisonRequestDto bean2.ComparisonRequestDto, userHasAdminAccess bool) (*bean2.ComparisonResponseDto, error) {

	indexVsSecretConfigMetadata := make(map[int]*bean2.SecretConfigMetadata, len(comparisonRequestDto.ComparisonItems))
	for _, comparisonItem := range comparisonRequestDto.ComparisonItems {
		secretConfigMetadata, err := impl.getSingleSecretConfigForComparison(ctx, comparisonItem)
		if err != nil {
			impl.logger.Errorw("error in getting single secret config for comparison", "comparisonItem", comparisonItem, "err", err)
			return nil, err
		}
		if _, ok := indexVsSecretConfigMetadata[comparisonItem.Index]; !ok {
			indexVsSecretConfigMetadata[comparisonItem.Index] = secretConfigMetadata
		}
	}
	if !userHasAdminAccess {
		// compare secrets data and mask if necessary
		err := impl.CompareSecretDataAndMaskIfNecessary(indexVsSecretConfigMetadata)
		if err != nil {
			impl.logger.Errorw("error in comparing secret and masking if necessary", "err", err)
			return nil, err
		}
	}

	//prepare final response with index
	allSecretConfigDto, err := impl.getAllComparableSecretResponseDto(indexVsSecretConfigMetadata)
	if err != nil {
		impl.logger.Errorw("error in getting all comparable secrets response dto", "err", err)
		return nil, err
	}
	return bean2.DefaultComparisonResponseDto().WithComparisonItemResponse(allSecretConfigDto), nil
}

func (impl *DeploymentConfigurationServiceImpl) getAllComparableSecretResponseDto(indexVsSecretConfigMetadata map[int]*bean2.SecretConfigMetadata) ([]*bean2.DeploymentAndCmCsConfigDto, error) {
	allSecretConfigDto := make([]*bean2.DeploymentAndCmCsConfigDto, 0, len(indexVsSecretConfigMetadata))
	for index, secretConfigMetadata := range indexVsSecretConfigMetadata {
		// prepare secrets list data part for response
		unresolvedSecretConfigData, err := helper.GetConfigDataRequestJsonRawMessage(secretConfigMetadata.SecretsList.ConfigData)
		if err != nil {
			impl.logger.Errorw("error in converting secrets list config data to json raw message", "err", err)
			return nil, err
		}
		// prepare resolved data part for response
		resolvedSecretConfigData, err := helper.GetConfigDataRequestJsonRawMessage(secretConfigMetadata.SecretScopeVariableMetadata.ResolvedConfigData)
		if err != nil {
			impl.logger.Errorw("error in converting resolved secret config data to json raw message", "err", err)
			return nil, err
		}
		secretConfigDto := bean2.NewDeploymentAndCmCsConfig().
			WithConfigData(unresolvedSecretConfigData).
			WithResourceType(bean.CS).
			WithVariableSnapshot(secretConfigMetadata.SecretScopeVariableMetadata.VariableSnapShot).
			WithResolvedValue(resolvedSecretConfigData)

		allSecretConfigDto = append(allSecretConfigDto, bean2.NewDeploymentAndCmCsConfigDto().WithSecretData(secretConfigDto).WithIndex(index))
	}
	return allSecretConfigDto, nil
}

func (impl *DeploymentConfigurationServiceImpl) prepareSecretNameWithKeyValMapAndMaskValue(secretMetadata *bean2.SecretConfigMetadata) (map[string]map[string]string, map[string]map[string]string, error) {

	unresolvedSecretMapping, err := utils.GetKeyValMapForSecretConfigDataAndMaskData(secretMetadata.SecretsList.ConfigData)
	if err != nil {
		impl.logger.Errorw("error in getting key val map for SecretsList's config data with masking", "err", err)
		return nil, nil, err
	}
	resolvedSecretMapping := make(map[string]map[string]string)
	if len(secretMetadata.SecretScopeVariableMetadata.VariableSnapShot) > 0 {
		//scope variable is used so mask scope variable secret data also
		resolvedSecretMapping, err = utils.GetKeyValMapForSecretConfigDataAndMaskData(secretMetadata.SecretScopeVariableMetadata.ResolvedConfigData)
		if err != nil {
			impl.logger.Errorw("error in getting key val map for  resolved config data with masking", "err", err)
			return nil, nil, err
		}
	}

	return unresolvedSecretMapping, resolvedSecretMapping, nil
}

func (impl *DeploymentConfigurationServiceImpl) compareAndMaskOtherComparableSecretValues(secretMetadata2 *bean2.SecretConfigMetadata, unresolvedSecretMapping1 map[string]map[string]string,
	resolvedSecretMapping1 map[string]map[string]string) error {

	err := utils.CompareAndMaskSecretValuesInConfigData(secretMetadata2.SecretsList.ConfigData, unresolvedSecretMapping1)
	if err != nil {
		impl.logger.Errorw("error in comparing and masking secretsList's secret values", "err", err)
		return err
	}
	if len(secretMetadata2.SecretScopeVariableMetadata.VariableSnapShot) > 0 {
		err = utils.CompareAndMaskSecretValuesInConfigData(secretMetadata2.SecretScopeVariableMetadata.ResolvedConfigData, resolvedSecretMapping1)
		if err != nil {
			impl.logger.Errorw("error in comparing and masking resolvedConfigData's secret values", "err", err)
			return err
		}
	}
	return nil
}

func (impl *DeploymentConfigurationServiceImpl) CompareSecretDataAndMaskIfNecessary(indexVsComparisonItems map[int]*bean2.SecretConfigMetadata) error {
	secretComparisonItem1, secretComparisonItem2 := indexVsComparisonItems[0], indexVsComparisonItems[1]
	unresolvedSecretMapping1, resolvedSecretMapping1, err := impl.prepareSecretNameWithKeyValMapAndMaskValue(secretComparisonItem1)
	if err != nil {
		impl.logger.Errorw("error in preparing key val map for secret and mask the values", "err", err)
		return err
	}
	err = impl.compareAndMaskOtherComparableSecretValues(secretComparisonItem2, unresolvedSecretMapping1, resolvedSecretMapping1)
	if err != nil {
		impl.logger.Errorw("error in comparing and masking other secret's value", "err", err)
		return err
	}
	return nil
}

func (impl *DeploymentConfigurationServiceImpl) getAppEnvClusterAndSystemMetadata(comparisonItem *bean2.ComparisonItemRequestDto) (*bean2.AppEnvAndClusterMetadata, *resourceQualifiers.SystemMetadata, error) {
	var err error
	var envId, appId, clusterId int
	systemMetadata := &resourceQualifiers.SystemMetadata{
		AppName: comparisonItem.AppName,
	}
	if len(comparisonItem.EnvName) > 0 {
		env, err := impl.environmentRepository.FindEnvByNameWithClusterDetails(comparisonItem.EnvName)
		if err != nil {
			impl.logger.Errorw("error in getting environment model by envName", "envName", comparisonItem.EnvName, "err", err)
			return nil, nil, err
		}
		envId = env.Id
		clusterId = env.ClusterId
		systemMetadata.EnvironmentName = env.Name
		systemMetadata.Namespace = env.Namespace
		systemMetadata.ClusterName = env.Cluster.ClusterName
	}
	appId, err = impl.appRepository.FindAppIdByName(comparisonItem.AppName)
	if err != nil {
		impl.logger.Errorw("error in getting app model by appName", "appName", comparisonItem.AppName, "err", err)
		return nil, nil, err
	}
	return &bean2.AppEnvAndClusterMetadata{AppId: appId, EnvId: envId, ClusterId: clusterId}, systemMetadata, nil
}

func (impl *DeploymentConfigurationServiceImpl) getSingleSecretConfigForComparison(ctx context.Context, comparisonItemRequest *bean2.ComparisonItemRequestDto) (*bean2.SecretConfigMetadata, error) {
	appEnvAndClusterMetadata, systemMetadata, err := impl.getAppEnvClusterAndSystemMetadata(comparisonItemRequest)
	if err != nil {
		impl.logger.Errorw("error in getting app, env, cluster and systemMetadata", "comparisonItemRequest", comparisonItemRequest, "err", err)
		return nil, err
	}
	switch comparisonItemRequest.ConfigArea {
	case bean2.CdRollback.ToString(), bean2.DeploymentHistory.ToString():
		return impl.getHistoricalSecretData(ctx, comparisonItemRequest)
	}
	// this would be the default case
	return impl.getSingleSecretDataForAppConfiguration(ctx, comparisonItemRequest, appEnvAndClusterMetadata, systemMetadata)
}

func (impl *DeploymentConfigurationServiceImpl) getHistoricalSecretData(ctx context.Context, comparisonItem *bean2.ComparisonItemRequestDto) (*bean2.SecretConfigMetadata, error) {
	// wfrId is expected in this case to return the expected data
	if comparisonItem.WfrId == 0 {
		return nil, &util.ApiError{HttpStatusCode: http.StatusNotFound, Code: strconv.Itoa(http.StatusNotFound), InternalMessage: bean2.ExpectedWfrIdNotPassedInQueryParamErr, UserMessage: bean2.ExpectedWfrIdNotPassedInQueryParamErr}
	}
	secretsList, resolvedSecretsData, err := impl.getSingleSecretDataForPreviousDeployments(ctx, comparisonItem.ConfigDataQueryParams)
	if err != nil {
		impl.logger.Errorw("error in getting historical data for secret", "comparisonDataPayload", comparisonItem.ConfigDataQueryParams, "err", err)
		return nil, err
	}
	secretConfigMetadata := &bean2.SecretConfigMetadata{
		SecretsList:                 secretsList,
		SecretScopeVariableMetadata: resolvedSecretsData,
	}
	return secretConfigMetadata, nil
}

func (impl *DeploymentConfigurationServiceImpl) getSingleSecretDataForAppConfiguration(ctx context.Context, comparisonItem *bean2.ComparisonItemRequestDto, appEnvAndClusterMetadata *bean2.AppEnvAndClusterMetadata,
	systemMetadata *resourceQualifiers.SystemMetadata) (*bean2.SecretConfigMetadata, error) {
	var secretConfigMetadata *bean2.SecretConfigMetadata
	var err error
	switch comparisonItem.ConfigType {
	case bean2.DraftOnly.ToString():
		secretConfigMetadata, err = impl.getSecretDataForDraftOnly(ctx, appEnvAndClusterMetadata, comparisonItem.UserId)
		if err != nil {
			impl.logger.Errorw("error in getting single secret data for draft only", "appEnvAndClusterMetadata", appEnvAndClusterMetadata, "err", err)
			return nil, err
		}
	case bean2.PublishedWithDraft.ToString():
		secretConfigMetadata, err = impl.getSecretDataForPublishedWithDraft(ctx, appEnvAndClusterMetadata, systemMetadata, comparisonItem.UserId)
		if err != nil {
			impl.logger.Errorw("error in getting single secret data for published with draft ", "appEnvAndClusterMetadata", appEnvAndClusterMetadata, "err", err)
			return nil, err
		}

	case bean2.PreviousDeployments.ToString():
		secretConfigMetadata, err = impl.getHistoricalSecretData(ctx, comparisonItem)
		if err != nil {
			impl.logger.Errorw("error in config data for Previous Deployments", "comparisonDataPayload", comparisonItem.ConfigDataQueryParams, "err", err)
			return nil, err
		}
	default: // keeping default as PublishedOnly
		secretConfigMetadata, err = impl.getSecretDataForPublishedOnly(ctx, appEnvAndClusterMetadata, systemMetadata)
		if err != nil {
			impl.logger.Errorw("error in config data for PublishedOnly", "comparisonDataPayload", comparisonItem.ConfigDataQueryParams, "err", err)
			return nil, err
		}
	}
	return secretConfigMetadata, nil
}

func (impl *DeploymentConfigurationServiceImpl) getHistoryAndSecretsListForPreviousDeployments(wfrId, pipelineId int) (*bean.SecretsList, *repository3.ConfigmapAndSecretHistory, error) {
	history, err := impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(pipelineId, wfrId, repository3.SECRET_TYPE)
	if err != nil {
		impl.logger.Errorw("error in getting if cs history exists for pipelineId and wfrId", "pipelineId", pipelineId, "wfrId", wfrId, "err", err)
		return nil, nil, err
	}
	_, secretsList, err := impl.configMapHistoryReadService.GetCmCsListObjectFromHistory(history)
	if err != nil {
		impl.logger.Errorw("error in getting config data request for history", "err", err)
		return nil, nil, err
	}
	return secretsList, history, nil
}

func (impl *DeploymentConfigurationServiceImpl) getResolvedSecretDataForPreviousDeployments(ctx context.Context, secretsList *bean.SecretsList, history *repository3.ConfigmapAndSecretHistory) (*bean2.CmCsScopeVariableMetadata, error) {
	resolvedDataMap, variableSnapshotMap, err := impl.scopedVariableManager.GetResolvedCMCSHistoryDtos(ctx, repository3.SECRET_TYPE, bean3.ConfigList{}, history, adaptor.ReverseSecretListConvertor(*secretsList))
	if err != nil {
		impl.logger.Errorw("error in GetResolvedCMCSHistoryDtos, resolving cm cs historical data", "err", err)
		return nil, err
	}
	resolvedConfigDataList := make([]*bean.ConfigData, 0, len(resolvedDataMap))
	for _, resolvedConfigData := range resolvedDataMap {
		resolvedConfigDataList = append(resolvedConfigDataList, adapter.ConvertConfigDataToPipelineConfigData(&resolvedConfigData))
	}
	return &bean2.CmCsScopeVariableMetadata{ResolvedConfigData: resolvedConfigDataList, VariableSnapShot: variableSnapshotMap}, nil
}

func (impl *DeploymentConfigurationServiceImpl) getSingleSecretDataForPreviousDeployments(ctx context.Context, comparisonItemRequest *bean2.ConfigDataQueryParams) (*bean.SecretsList, *bean2.CmCsScopeVariableMetadata, error) {
	secretsList, history, err := impl.getHistoryAndSecretsListForPreviousDeployments(comparisonItemRequest.WfrId, comparisonItemRequest.PipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting history object and secret list for history", "wfrId", comparisonItemRequest.WfrId, "pipelineId", comparisonItemRequest.PipelineId, "err", err)
		return nil, nil, err
	}
	resolvedSecretData, err := impl.getResolvedSecretDataForPreviousDeployments(ctx, secretsList, history)
	if err != nil {
		impl.logger.Errorw("error in getResolvedSecretDataPreviousDeployments, resolving cm cs historical data", "err", err)
		return nil, nil, err
	}
	return secretsList, resolvedSecretData, nil
}

func (impl *DeploymentConfigurationServiceImpl) getSecretDataForPublishedOnly(ctx context.Context, appEnvAndClusterMetadata *bean2.AppEnvAndClusterMetadata,
	systemMetadata *resourceQualifiers.SystemMetadata) (*bean2.SecretConfigMetadata, error) {
	secretData, err := impl.getSecretConfigResponse("", 0, appEnvAndClusterMetadata.EnvId, appEnvAndClusterMetadata.AppId)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in getting secret config response by appId and envId", "appId", appEnvAndClusterMetadata.AppId, "envId", appEnvAndClusterMetadata.EnvId, "err", err)
		return nil, err
	}
	resolvedSecretData, err := impl.getResolvedSecretDataForPublishedOnly(ctx, appEnvAndClusterMetadata, systemMetadata)
	if err != nil {
		impl.logger.Errorw("error in getResolvedSecretDataPreviousDeployments, resolving cm cs historical data", "err", err)
		return nil, err
	}
	secretConfigMetadata := &bean2.SecretConfigMetadata{
		SecretsList:                 &bean.SecretsList{ConfigData: secretData.ConfigData},
		SecretScopeVariableMetadata: resolvedSecretData,
	}
	return secretConfigMetadata, nil
}

func (impl *DeploymentConfigurationServiceImpl) getResolvedSecretDataForPublishedOnly(ctx context.Context, appEnvAndClusterMetadata *bean2.AppEnvAndClusterMetadata,
	systemMetadata *resourceQualifiers.SystemMetadata) (*bean2.CmCsScopeVariableMetadata, error) {
	scope := resourceQualifiers.Scope{
		AppId:          appEnvAndClusterMetadata.AppId,
		EnvId:          appEnvAndClusterMetadata.EnvId,
		ClusterId:      appEnvAndClusterMetadata.ClusterId,
		SystemMetadata: systemMetadata,
	}
	cmcsMetadataDto, err := impl.getMergedCmCs(appEnvAndClusterMetadata.EnvId, appEnvAndClusterMetadata.AppId)
	if err != nil {
		impl.logger.Errorw("error in getting merged cm cs", "appId", appEnvAndClusterMetadata.AppId, "envId", appEnvAndClusterMetadata.EnvId, "err", err)
		return nil, err
	}
	_, resolvedSecretList, _, variableMapCS, err := impl.scopedVariableManager.ResolveCMCS(ctx, scope, cmcsMetadataDto.ConfigAppLevelId, cmcsMetadataDto.ConfigEnvLevelId, cmcsMetadataDto.CmMap, cmcsMetadataDto.SecretMap)
	if err != nil {
		impl.logger.Errorw("error in resolving CM/CS", "scope", scope, "appId", appEnvAndClusterMetadata.AppId, "envId", appEnvAndClusterMetadata.EnvId, "err", err)
		return nil, err
	}

	resolvedSecretDataList := make([]*bean.ConfigData, 0, len(resolvedSecretList))
	for _, resolvedSecret := range resolvedSecretList {
		resolvedSecretDataList = append(resolvedSecretDataList, adapter.ConvertConfigDataToPipelineConfigData(resolvedSecret))
	}
	return &bean2.CmCsScopeVariableMetadata{ResolvedConfigData: resolvedSecretDataList, VariableSnapShot: variableMapCS}, nil
}
