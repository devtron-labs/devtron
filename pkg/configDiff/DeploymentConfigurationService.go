package configDiff

import (
	"context"
	"encoding/json"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/util"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/configDiff/helper"
	"github.com/devtron-labs/devtron/pkg/configDiff/utils"
	"github.com/devtron-labs/devtron/pkg/generateManifest"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type DeploymentConfigurationService interface {
	ConfigAutoComplete(appId int, envId int) (*bean2.ConfigDataResponse, error)
	GetAllConfigData(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams) (*bean2.DeploymentAndCmCsConfigDto, error)
}

type DeploymentConfigurationServiceImpl struct {
	logger                    *zap.SugaredLogger
	configMapService          pipeline.ConfigMapService
	appRepository             appRepository.AppRepository
	environmentRepository     repository.EnvironmentRepository
	chartService              chartService.ChartService
	deploymentTemplateService generateManifest.DeploymentTemplateService
}

func NewDeploymentConfigurationServiceImpl(logger *zap.SugaredLogger,
	configMapService pipeline.ConfigMapService,
	appRepository appRepository.AppRepository,
	environmentRepository repository.EnvironmentRepository,
	chartService chartService.ChartService,
	deploymentTemplateService generateManifest.DeploymentTemplateService,
) (*DeploymentConfigurationServiceImpl, error) {
	deploymentConfigurationService := &DeploymentConfigurationServiceImpl{
		logger:                    logger,
		configMapService:          configMapService,
		appRepository:             appRepository,
		environmentRepository:     environmentRepository,
		chartService:              chartService,
		deploymentTemplateService: deploymentTemplateService,
	}

	return deploymentConfigurationService, nil
}
func (impl *DeploymentConfigurationServiceImpl) ConfigAutoComplete(appId int, envId int) (*bean2.ConfigDataResponse, error) {
	cMCSNamesAppLevel, cMCSNamesEnvLevel, err := impl.configMapService.FetchCmCsNamesAppAndEnvLevel(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in fetching CM and CS names at app or env level", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	cmcsKeyPropertyAppLevelMap, cmcsKeyPropertyEnvLevelMap := helper.GetCmCsAppAndEnvLevelMap(cMCSNamesAppLevel, cMCSNamesEnvLevel)
	for key, configProperty := range cmcsKeyPropertyAppLevelMap {
		if _, ok := cmcsKeyPropertyEnvLevelMap[key]; !ok {
			configProperty.Global = true
		}
	}
	for key, configProperty := range cmcsKeyPropertyEnvLevelMap {
		if _, ok := cmcsKeyPropertyAppLevelMap[key]; ok {
			configProperty.Global = true
			configProperty.Overridden = true
		}
	}
	combinedProperties := helper.GetCombinedPropertiesMap(cmcsKeyPropertyAppLevelMap, cmcsKeyPropertyEnvLevelMap)
	combinedProperties = append(combinedProperties, helper.GetConfigProperty(0, "", bean.DeploymentTemplate, bean2.PublishedConfigState))

	configDataResp := bean2.NewConfigDataResponse().WithResourceConfig(combinedProperties)
	return configDataResp, nil
}

func (impl *DeploymentConfigurationServiceImpl) GetAllConfigData(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams) (*bean2.DeploymentAndCmCsConfigDto, error) {
	if !configDataQueryParams.IsValidConfigType() {
		return nil, &util.ApiError{HttpStatusCode: http.StatusBadRequest, Code: strconv.Itoa(http.StatusBadRequest), InternalMessage: bean2.InvalidConfigTypeErr, UserMessage: bean2.InvalidConfigTypeErr}
	}
	var err error
	var envId int
	var appId int
	if configDataQueryParams.IsEnvNameProvided() {
		envModel, err := impl.environmentRepository.FindByName(configDataQueryParams.EnvName)
		if err != nil {
			impl.logger.Errorw("GetAllConfigData, error in getting environment model by envName", "envName", configDataQueryParams.EnvName, "err", err)
			return nil, err
		}
		envId = envModel.Id
	}
	appModel, err := impl.appRepository.FindActiveByName(configDataQueryParams.AppName)
	if err != nil {
		impl.logger.Errorw("GetAllConfigData, error in getting app model by appName", "appName", configDataQueryParams.AppName, "err", err)
		return nil, err
	}
	appId = appModel.Id

	configData := make([]*bean2.DeploymentAndCmCsConfig, 0)
	switch configDataQueryParams.ConfigType {
	default: // keeping default as PublishedOnly
		cmcsConfigData, err := impl.getCmCsConfigResponse(configDataQueryParams, envId, appId)
		if err != nil {
			impl.logger.Errorw("GetAllConfigData, error in getting cm cs", "appName", configDataQueryParams.AppName, "envName", configDataQueryParams.EnvName, "err", err)
			return nil, err
		}
		if configDataQueryParams.IsRequestMadeForOneResource() {
			return bean2.NewDeploymentAndCmCsConfigDto().WithConfigData(configData).WithAppAndEnvIdId(appId, envId), nil
		}
		configData = append(configData, cmcsConfigData...)

		deploymentTemplateJsonData, err := impl.getDeploymentConfig(ctx, appId, envId)
		if err != nil {
			impl.logger.Errorw("GetAllConfigData, error in getting deployment config", "appName", configDataQueryParams.AppName, "envName", configDataQueryParams.EnvName, "err", err)
			return nil, err
		}
		deploymentConfig := bean2.NewDeploymentAndCmCsConfig().WithConfigState(bean2.PublishedConfigState).WithConfigData(deploymentTemplateJsonData).WithResourceType(bean.DeploymentTemplate)
		configData = append(configData, deploymentConfig)
	}

	return bean2.NewDeploymentAndCmCsConfigDto().WithConfigData(configData).WithAppAndEnvIdId(appId, envId), nil
}

func (impl *DeploymentConfigurationServiceImpl) getBaseDeploymentTemplate(appId int) (json.RawMessage, error) {
	deploymentTemplateData, err := impl.chartService.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in getting base deployment template for appId", "appId", appId, "err", err)
		return nil, err
	}
	return deploymentTemplateData.DefaultAppOverride, nil
}

func (impl *DeploymentConfigurationServiceImpl) getDeploymentTemplateForEnvLevel(ctx context.Context, appId, envId int) (json.RawMessage, error) {
	deploymentTemplateRequest := generateManifest.DeploymentTemplateRequest{
		AppId:           appId,
		EnvId:           envId,
		RequestDataMode: generateManifest.Values,
		Type:            repository2.PublishedOnEnvironments,
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

func (impl *DeploymentConfigurationServiceImpl) getDeploymentConfig(ctx context.Context, appId, envId int) (json.RawMessage, error) {
	if envId > 0 {
		return impl.getDeploymentTemplateForEnvLevel(ctx, appId, envId)
	}
	return impl.getBaseDeploymentTemplate(appId)
}

func (impl *DeploymentConfigurationServiceImpl) getSecretConfigResponse(resourceName string, envId, appId int) (*bean.ConfigDataRequest, error) {
	if len(resourceName) > 0 {
		if envId > 0 {
			return impl.configMapService.CSEnvironmentFetchForEdit(resourceName, 0, appId, envId)
		}
		return impl.configMapService.CSGlobalFetchForEditUsingAppId(resourceName, appId)
	}

	if envId > 0 {
		return impl.configMapService.CSEnvironmentFetch(appId, envId)
	}
	return impl.configMapService.CSGlobalFetch(appId)
}

func (impl *DeploymentConfigurationServiceImpl) getConfigMapResponse(resourceName string, envId, appId int) (*bean.ConfigDataRequest, error) {
	if len(resourceName) > 0 {
		if envId > 0 {
			return impl.configMapService.CMEnvironmentFetchForEdit(resourceName, 0, appId, envId)
		}
		return impl.configMapService.CMGlobalFetchForEditUsingAppId(resourceName, appId)
	}

	if envId > 0 {
		return impl.configMapService.CMEnvironmentFetch(appId, envId)
	}
	return impl.configMapService.CMGlobalFetch(appId)
}

func (impl *DeploymentConfigurationServiceImpl) getCmCsConfigResponse(configDataQueryParams *bean2.ConfigDataQueryParams, envId, appId int) ([]*bean2.DeploymentAndCmCsConfig, error) {

	configData := make([]*bean2.DeploymentAndCmCsConfig, 0)
	var cmcsRespJson json.RawMessage
	var resourceType bean.ResourceType
	cmcsConfig := bean2.NewDeploymentAndCmCsConfig().WithConfigState(bean2.PublishedConfigState)

	if configDataQueryParams.IsResourceTypeSecret() {
		secretData, err := impl.getSecretConfigResponse(configDataQueryParams.ResourceName, envId, appId)
		if err != nil {
			impl.logger.Errorw("getCmCsConfigResponse, error in getting secret config response for appName and envName", "appName", configDataQueryParams.AppName, "envName", configDataQueryParams.EnvName, "err", err)
			return nil, err
		}
		resourceType = bean.CS
		cmcsRespJson, err = utils.ConvertToJsonRawMessage(secretData)
		if err != nil {
			impl.logger.Errorw("getCmCsConfigResponse, error in getting secret json raw message", "err", err)
			return nil, err
		}
	} else if configDataQueryParams.IsResourceTypeConfigMap() {
		cmData, err := impl.getConfigMapResponse(configDataQueryParams.ResourceName, envId, appId)
		if err != nil {
			impl.logger.Errorw("getCmCsConfigResponse, error in getting config map for appName and envName", "appName", configDataQueryParams.AppName, "envName", configDataQueryParams.EnvName, "err", err)
			return nil, err
		}
		resourceType = bean.CM
		cmcsRespJson, err = utils.ConvertToJsonRawMessage(cmData)
		if err != nil {
			impl.logger.Errorw("getCmCsConfigResponse, error in getting cm json raw message", "cmData", cmData, "err", err)
			return nil, err
		}
	}
	cmcsConfig.WithConfigData(cmcsRespJson).WithResourceType(resourceType)
	configData = append(configData, cmcsConfig)

	return configData, nil
}
