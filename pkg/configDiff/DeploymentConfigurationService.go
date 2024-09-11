package configDiff

import (
	"context"
	"encoding/json"
	"errors"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/util"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/configDiff/adaptor"
	bean2 "github.com/devtron-labs/devtron/pkg/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/configDiff/helper"
	"github.com/devtron-labs/devtron/pkg/configDiff/utils"
	"github.com/devtron-labs/devtron/pkg/generateManifest"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type DeploymentConfigurationService interface {
	ConfigAutoComplete(appId int, envId int) (*bean2.ConfigDataResponse, error)
	GetAllConfigData(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams) (*bean2.DeploymentAndCmCsConfigDto, error)
}

type DeploymentConfigurationServiceImpl struct {
	logger                              *zap.SugaredLogger
	configMapService                    pipeline.ConfigMapService
	appRepository                       appRepository.AppRepository
	environmentRepository               repository.EnvironmentRepository
	chartService                        chartService.ChartService
	deploymentTemplateService           generateManifest.DeploymentTemplateService
	deploymentTemplateHistoryRepository repository3.DeploymentTemplateHistoryRepository
	pipelineStrategyHistoryRepository   repository3.PipelineStrategyHistoryRepository
	configMapHistoryRepository          repository3.ConfigMapHistoryRepository
}

func NewDeploymentConfigurationServiceImpl(logger *zap.SugaredLogger,
	configMapService pipeline.ConfigMapService,
	appRepository appRepository.AppRepository,
	environmentRepository repository.EnvironmentRepository,
	chartService chartService.ChartService,
	deploymentTemplateService generateManifest.DeploymentTemplateService,
	deploymentTemplateHistoryRepository repository3.DeploymentTemplateHistoryRepository,
	pipelineStrategyHistoryRepository repository3.PipelineStrategyHistoryRepository,
	configMapHistoryRepository repository3.ConfigMapHistoryRepository,
) (*DeploymentConfigurationServiceImpl, error) {
	deploymentConfigurationService := &DeploymentConfigurationServiceImpl{
		logger:                              logger,
		configMapService:                    configMapService,
		appRepository:                       appRepository,
		environmentRepository:               environmentRepository,
		chartService:                        chartService,
		deploymentTemplateService:           deploymentTemplateService,
		deploymentTemplateHistoryRepository: deploymentTemplateHistoryRepository,
		pipelineStrategyHistoryRepository:   pipelineStrategyHistoryRepository,
		configMapHistoryRepository:          configMapHistoryRepository,
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

func (impl *DeploymentConfigurationServiceImpl) GetAllConfigData(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams) (*bean2.DeploymentAndCmCsConfigDto, error) {
	if !configDataQueryParams.IsValidConfigType() {
		return nil, &util.ApiError{HttpStatusCode: http.StatusBadRequest, Code: strconv.Itoa(http.StatusBadRequest), InternalMessage: bean2.InvalidConfigTypeErr, UserMessage: bean2.InvalidConfigTypeErr}
	}
	var err error
	var envId int
	var appId int
	if configDataQueryParams.IsEnvNameProvided() {
		envId, err = impl.environmentRepository.FindIdByName(configDataQueryParams.EnvName)
		if err != nil {
			impl.logger.Errorw("GetAllConfigData, error in getting environment model by envName", "envName", configDataQueryParams.EnvName, "err", err)
			return nil, err
		}
	}
	appId, err = impl.appRepository.FindAppIdByName(configDataQueryParams.AppName)
	if err != nil {
		impl.logger.Errorw("GetAllConfigData, error in getting app model by appName", "appName", configDataQueryParams.AppName, "err", err)
		return nil, err
	}

	switch configDataQueryParams.ConfigArea {
	case bean2.CdRollback.ToString():
		return impl.getConfigDataForCdRollback(ctx, configDataQueryParams, appId, envId)
	case bean2.DeploymentHistory.ToString():
		return impl.getConfigDataForDeploymentHistory(ctx, configDataQueryParams, appId, envId)
	}
	// this would be the default case
	return impl.getConfigDataForAppConfiguration(ctx, configDataQueryParams, appId, envId)
}

func (impl *DeploymentConfigurationServiceImpl) getConfigDataForCdRollback(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams,
	appId, envId int) (*bean2.DeploymentAndCmCsConfigDto, error) {
	// we would be expecting wfrId in case of getting data for cdRollback

}

func (impl *DeploymentConfigurationServiceImpl) getDeploymentHistoryConfig(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams) (*json.RawMessage, error) {
	deploymentJson := &json.RawMessage{}
	deploymentHistory, err := impl.deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId(configDataQueryParams.PipelineId, configDataQueryParams.WfrId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting deployment template history for pipelineId and wfrId", "pipelineId", configDataQueryParams.PipelineId, "wfrId", configDataQueryParams.WfrId, "err", err)
		return nil, err
	} else if err == pg.ErrNoRows {
		//history not created yet
		return deploymentJson, nil
	}
	err = deploymentJson.UnmarshalJSON([]byte(deploymentHistory.Template))
	if err != nil {
		impl.logger.Errorw("getDeploymentTemplateForEnvLevel, error in unmarshalling string  deploymentTemplateResponse data into json Raw message", "data", deploymentHistory.Template, "err", err)
		return nil, err
	}
	return deploymentJson, nil
}

func (impl *DeploymentConfigurationServiceImpl) getPipelineStrategyConfigHistory(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams) (*json.RawMessage, error) {
	pipelineStrategyJson := &json.RawMessage{}
	pipelineStrategyHistory, err := impl.pipelineStrategyHistoryRepository.GetHistoryByPipelineIdAndWfrId(ctx, configDataQueryParams.PipelineId, configDataQueryParams.WfrId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in checking if history exists for pipelineId and wfrId", "pipelineId", configDataQueryParams.PipelineId, "wfrId", configDataQueryParams.WfrId, "err", err)
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return pipelineStrategyJson, nil
	}

	err = pipelineStrategyJson.UnmarshalJSON([]byte(pipelineStrategyHistory.Config))
	if err != nil {
		impl.logger.Errorw("getDeploymentTemplateForEnvLevel, error in unmarshalling string  pipelineStrategyHistory data into json Raw message", "pipelineStrategyHistoryConfig", pipelineStrategyHistory.Config, "err", err)
		return nil, err
	}
	return pipelineStrategyJson, nil
}

func (impl *DeploymentConfigurationServiceImpl) getConfigDataForDeploymentHistory(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams,
	appId, envId int) (*bean2.DeploymentAndCmCsConfigDto, error) {
	// we would be expecting wfrId in case of getting data for Deployment history
	configDataDto := &bean2.DeploymentAndCmCsConfigDto{}
	var err error
	//fetching history for deployment config starts
	deploymentConfigJson, err := impl.getDeploymentHistoryConfig(ctx, configDataQueryParams)
	if err != nil {
		impl.logger.Errorw("getConfigDataForDeploymentHistory, error in getDeploymentHistoryConfig", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	if deploymentConfigJson != nil {
		deploymentConfig := bean2.NewDeploymentAndCmCsConfig().WithConfigData(*deploymentConfigJson).WithResourceType(bean.DeploymentTemplate)
		configDataDto.WithDeploymentTemplateData(deploymentConfig)
	}
	// fetching for deployment config ends

	// fetching for pipeline strategy config starts
	pipelineConfigJson, err := impl.getPipelineStrategyConfigHistory(ctx, configDataQueryParams)
	if err != nil {
		impl.logger.Errorw("getConfigDataForDeploymentHistory, error in getPipelineStrategyConfigHistory", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	if pipelineConfigJson != nil {
		pipelineConfig := bean2.NewDeploymentAndCmCsConfig().WithConfigData(*pipelineConfigJson).WithResourceType(bean.PipelineStrategy)
		configDataDto.WithPipelineConfigData(pipelineConfig)
	}
	// fetching for pipeline strategy config ends

	// fetching for cm config starts
	cmConfigJson, err := impl.getCmCsConfigHistory(ctx, configDataQueryParams, repository3.CONFIGMAP_TYPE)
	if err != nil {
		impl.logger.Errorw("getConfigDataForDeploymentHistory, error in getCmConfigHistory", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	if cmConfigJson != nil {
		cmConfigData := bean2.NewDeploymentAndCmCsConfig().WithConfigData(*cmConfigJson).WithResourceType(bean.CM)
		configDataDto.WithConfigMapData(cmConfigData)
	}
	// fetching for cm config ends

	// fetching for cs config starts
	secretConfigJson, err := impl.getCmCsConfigHistory(ctx, configDataQueryParams, repository3.SECRET_TYPE)
	if err != nil {
		impl.logger.Errorw("getConfigDataForDeploymentHistory, error in getSecretConfigHistory", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	if secretConfigJson != nil {
		secretConfigData := bean2.NewDeploymentAndCmCsConfig().WithConfigData(*secretConfigJson).WithResourceType(bean.CS)
		configDataDto.WithSecretData(secretConfigData)
	}
	// fetching for cs config ends

	return configDataDto, nil
}

func (impl *DeploymentConfigurationServiceImpl) getCmCsConfigHistory(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams, configType repository3.ConfigType) (*json.RawMessage, error) {
	cmJson := &json.RawMessage{}
	history, err := impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(configDataQueryParams.PipelineId, configDataQueryParams.WfrId, configType)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in checking if cm cs history exists for pipelineId and wfrId", "pipelineId", configDataQueryParams.PipelineId, "wfrId", configDataQueryParams.WfrId, "err", err)
		return nil, err
	} else if err == pg.ErrNoRows {
		return cmJson, nil
	}
	//var configData []*bean3.ConfigData
	if configType == repository3.CONFIGMAP_TYPE {
		configList := bean3.ConfigList{}
		if len(history.Data) > 0 {
			err = json.Unmarshal([]byte(history.Data), &configList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		//configData = configList.ConfigData
	} else if configType == repository3.SECRET_TYPE {
		secretList := bean3.SecretList{}
		if len(history.Data) > 0 {
			err = json.Unmarshal([]byte(history.Data), &secretList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		//configData = secretList.ConfigData
	}

	return nil, nil
}

func (impl *DeploymentConfigurationServiceImpl) getConfigDataForAppConfiguration(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams,
	appId, envId int) (*bean2.DeploymentAndCmCsConfigDto, error) {
	configDataDto := &bean2.DeploymentAndCmCsConfigDto{}
	var err error
	switch configDataQueryParams.ConfigType {
	default: // keeping default as PublishedOnly
		configDataDto, err = impl.getPublishedConfigData(ctx, configDataQueryParams, appId, envId)
		if err != nil {
			impl.logger.Errorw("GetAllConfigData, error in config data for PublishedOnly", "configDataQueryParams", configDataQueryParams, "err", err)
			return nil, err
		}
	}
	return configDataDto, nil
}

func (impl *DeploymentConfigurationServiceImpl) getCmCsEditDataForPublishedOnly(configDataQueryParams *bean2.ConfigDataQueryParams, envId, appId int) (*bean2.DeploymentAndCmCsConfigDto, error) {
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
		impl.logger.Errorw("getCmCsEditDataForPublishedOnly, error in getting config response", "resourceName", configDataQueryParams.ResourceName, "envName", configDataQueryParams.EnvName, "err", err)
		return nil, err
	}

	respJson, err := utils.ConvertToJsonRawMessage(cmcsConfigData)
	if err != nil {
		impl.logger.Errorw("getCmCsEditDataForPublishedOnly, error in converting to json raw message", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}

	cmCsConfig := bean2.NewDeploymentAndCmCsConfig().WithConfigData(respJson).WithResourceType(resourceType)
	if resourceType == bean.CS {
		configDataDto.WithSecretData(cmCsConfig)
	} else if resourceType == bean.CM {
		configDataDto.WithConfigMapData(cmCsConfig)
	}
	return configDataDto, nil
}

func (impl *DeploymentConfigurationServiceImpl) getCmCsPublishedConfigResponse(envId, appId int) (*bean2.DeploymentAndCmCsConfigDto, error) {

	configDataDto := &bean2.DeploymentAndCmCsConfigDto{}
	secretData, err := impl.getSecretConfigResponse("", 0, envId, appId)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in getting secret config response by appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	//iterate on secret configData and then and set draft data from draftResourcesMap if same resourceName found do the same for configMap below
	cmData, err := impl.getConfigMapResponse("", 0, envId, appId)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in getting config map by appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	secretRespJson, err := utils.ConvertToJsonRawMessage(secretData)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in converting secret data to json raw message", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	cmRespJson, err := utils.ConvertToJsonRawMessage(cmData)
	if err != nil {
		impl.logger.Errorw("getCmCsPublishedConfigResponse, error in converting config map data to json raw message", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	cmConfigData := bean2.NewDeploymentAndCmCsConfig().WithConfigData(cmRespJson).WithResourceType(bean.CM)
	secretConfigData := bean2.NewDeploymentAndCmCsConfig().WithConfigData(secretRespJson).WithResourceType(bean.CS)

	configDataDto.WithConfigMapData(cmConfigData).WithSecretData(secretConfigData)
	return configDataDto, nil

}

func (impl *DeploymentConfigurationServiceImpl) getPublishedDeploymentConfig(ctx context.Context, appId, envId int) (json.RawMessage, error) {
	if envId > 0 {
		return impl.getDeploymentTemplateForEnvLevel(ctx, appId, envId)
	}
	return impl.getBaseDeploymentTemplate(appId)
}

func (impl *DeploymentConfigurationServiceImpl) getPublishedConfigData(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams,
	appId, envId int) (*bean2.DeploymentAndCmCsConfigDto, error) {

	if configDataQueryParams.IsRequestMadeForOneResource() {
		return impl.getCmCsEditDataForPublishedOnly(configDataQueryParams, envId, appId)
	}
	//ConfigMapsData and SecretsData are populated here
	configData, err := impl.getCmCsPublishedConfigResponse(envId, appId)
	if err != nil {
		impl.logger.Errorw("getPublishedConfigData, error in getting cm cs for PublishedOnly state", "appName", configDataQueryParams.AppName, "envName", configDataQueryParams.EnvName, "err", err)
		return nil, err
	}
	deploymentTemplateJsonData, err := impl.getPublishedDeploymentConfig(ctx, appId, envId)
	if err != nil {
		impl.logger.Errorw("getPublishedConfigData, error in getting publishedOnly deployment config ", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	deploymentConfig := bean2.NewDeploymentAndCmCsConfig().WithConfigData(deploymentTemplateJsonData).WithResourceType(bean.DeploymentTemplate)

	configData.WithDeploymentTemplateData(deploymentConfig)
	return configData, nil
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

func (impl *DeploymentConfigurationServiceImpl) getSecretConfigResponse(resourceName string, resourceId, envId, appId int) (*bean.ConfigDataRequest, error) {
	if len(resourceName) > 0 {
		if envId > 0 {
			return impl.configMapService.CSEnvironmentFetchForEdit(resourceName, resourceId, appId, envId)
		}
		return impl.configMapService.ConfigGlobalFetchEditUsingAppId(resourceName, appId, bean.CS)
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
		return impl.configMapService.ConfigGlobalFetchEditUsingAppId(resourceName, appId, bean.CM)
	}

	if envId > 0 {
		return impl.configMapService.CMEnvironmentFetch(appId, envId)
	}
	return impl.configMapService.CMGlobalFetch(appId)
}
