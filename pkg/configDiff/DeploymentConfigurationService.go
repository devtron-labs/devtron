package configDiff

import (
	"context"
	"encoding/json"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/configDiff/adaptor"
	bean2 "github.com/devtron-labs/devtron/pkg/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/configDiff/helper"
	"github.com/devtron-labs/devtron/pkg/configDiff/utils"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
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
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"net/http"
)

type DeploymentConfigurationService interface {
	ConfigAutoComplete(appId int, envId int) (*bean2.ConfigDataResponse, error)
	GetAllConfigData(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfigDto, error)
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
	scopedVariableManager               variables.ScopedVariableCMCSManager
	configMapRepository                 chartConfig.ConfigMapRepository
	deploymentConfigService             pipeline.PipelineDeploymentConfigService
	chartRefService                     chartRef.ChartRefService
	pipelineRepository                  pipelineConfig.PipelineRepository
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
	scopedVariableManager variables.ScopedVariableCMCSManager,
	configMapRepository chartConfig.ConfigMapRepository,
	deploymentConfigService pipeline.PipelineDeploymentConfigService,
	chartRefService chartRef.ChartRefService,
	pipelineRepository pipelineConfig.PipelineRepository,
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
		scopedVariableManager:               scopedVariableManager,
		configMapRepository:                 configMapRepository,
		deploymentConfigService:             deploymentConfigService,
		chartRefService:                     chartRefService,
		pipelineRepository:                  pipelineRepository,
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
	var envId int
	var appId int
	var clusterId int
	if configDataQueryParams.IsEnvNameProvided() {
		env, err := impl.environmentRepository.FindByName(configDataQueryParams.EnvName)
		if err != nil {
			impl.logger.Errorw("GetAllConfigData, error in getting environment model by envName", "envName", configDataQueryParams.EnvName, "err", err)
			return nil, err
		}
		envId = env.Id
		clusterId = env.ClusterId
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
	return impl.getConfigDataForAppConfiguration(ctx, configDataQueryParams, appId, envId, clusterId, userHasAdminAccess)
}

func (impl *DeploymentConfigurationServiceImpl) getConfigDataForCdRollback(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfigDto, error) {
	// wfrId is expected in this case to return the expected data
	return impl.getConfigDataForDeploymentHistory(ctx, configDataQueryParams, userHasAdminAccess)
}

func (impl *DeploymentConfigurationServiceImpl) getDeploymentHistoryConfig(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams) (*bean2.DeploymentAndCmCsConfig, error) {
	deploymentJson := json.RawMessage{}
	deploymentHistory, err := impl.deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId(configDataQueryParams.PipelineId, configDataQueryParams.WfrId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in getting deployment template history for pipelineId and wfrId", "pipelineId", configDataQueryParams.PipelineId, "wfrId", configDataQueryParams.WfrId, "err", err)
		return nil, err
	} else if util.IsErrNoRows(err) {
		return nil, util.GetApiError(http.StatusNotFound, bean2.NoDeploymentDoneForSelectedImage, bean2.NoDeploymentDoneForSelectedImage)
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
	configDataDto.WithPipelineConfigData(pipelineConfig)
	// fetching for pipeline strategy config ends

	// fetching for cm config starts
	cmConfigData, err := impl.getCmCsConfigHistory(ctx, configDataQueryParams, repository3.CONFIGMAP_TYPE, userHasAdminAccess)
	if err != nil {
		impl.logger.Errorw("getConfigDataForDeploymentHistory, error in getCmConfigHistory", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	configDataDto.WithConfigMapData(cmConfigData)
	// fetching for cm config ends

	// fetching for cs config starts
	secretConfigDto, err := impl.getCmCsConfigHistory(ctx, configDataQueryParams, repository3.SECRET_TYPE, userHasAdminAccess)
	if err != nil {
		impl.logger.Errorw("getConfigDataForDeploymentHistory, error in getSecretConfigHistory", "configDataQueryParams", configDataQueryParams, "err", err)
		return nil, err
	}
	configDataDto.WithSecretData(secretConfigDto)
	// fetching for cs config ends

	return configDataDto, nil
}

func (impl *DeploymentConfigurationServiceImpl) getCmCsConfigHistory(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams, configType repository3.ConfigType, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfig, error) {
	var resourceType bean.ResourceType
	history, err := impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(configDataQueryParams.PipelineId, configDataQueryParams.WfrId, configType)
	if err != nil {
		impl.logger.Errorw("error in checking if cm cs history exists for pipelineId and wfrId", "pipelineId", configDataQueryParams.PipelineId, "wfrId", configDataQueryParams.WfrId, "err", err)
		return nil, err
	}
	var configData []*bean.ConfigData
	configList := pipeline.ConfigsList{}
	secretList := bean.SecretsList{}
	if configType == repository3.CONFIGMAP_TYPE {
		if len(history.Data) > 0 {
			err = json.Unmarshal([]byte(history.Data), &configList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "err", err)
				return nil, err
			}
		}
		resourceType = bean.CM
		configData = configList.ConfigData
	} else if configType == repository3.SECRET_TYPE {
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
		impl.encodeSecretDataFromNonAdminUsers(configData, userHasAdminAccess)
		impl.encodeSecretDataFromNonAdminUsers(resolvedConfigDataList, userHasAdminAccess)

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

func (impl *DeploymentConfigurationServiceImpl) encodeSecretDataFromNonAdminUsers(configDataList []*bean.ConfigData, userHasAdminAccess bool) {
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
					resultMapFinal[key] = "*****"
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

func (impl *DeploymentConfigurationServiceImpl) getConfigDataForAppConfiguration(ctx context.Context, configDataQueryParams *bean2.ConfigDataQueryParams,
	appId, envId, clusterId int, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfigDto, error) {
	configDataDto := &bean2.DeploymentAndCmCsConfigDto{}
	var err error
	switch configDataQueryParams.ConfigType {
	default: // keeping default as PublishedOnly
		configDataDto, err = impl.getPublishedConfigData(ctx, configDataQueryParams, appId, envId, clusterId, userHasAdminAccess)
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

func (impl *DeploymentConfigurationServiceImpl) getCmCsPublishedConfigResponse(ctx context.Context, envId, appId, clusterId int, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfigDto, error) {

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

	resolvedCmCsMetadataDto, err := impl.ResolveCmCs(ctx, envId, appId, clusterId, userHasAdminAccess)
	if err != nil {
		impl.logger.Errorw("error in resolving cm and cs for published only config only response", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	resolvedConfigMapDataStringJson, err := utils.ConvertToJsonRawMessage(resolvedCmCsMetadataDto.ResolvedConfigMapData)
	if err != nil {
		impl.logger.Errorw("error in ConvertToJsonRawMessage for resolvedConfigMapDataStringJson", "resolvedCmData", resolvedCmCsMetadataDto.ResolvedConfigMapData, "err", err)
		return nil, err
	}
	resolvedSecretDataStringJson, err := utils.ConvertToJsonRawMessage(resolvedCmCsMetadataDto.ResolvedSecretData)
	if err != nil {
		impl.logger.Errorw(" error in ConvertToJsonRawMessage for resolvedConfigDataString", "err", err)
		return nil, err
	}

	cmConfigData := bean2.NewDeploymentAndCmCsConfig().WithConfigData(cmRespJson).WithResourceType(bean.CM).
		WithResolvedValue(resolvedConfigMapDataStringJson).WithVariableSnapshot(resolvedCmCsMetadataDto.VariableMapCM)

	secretConfigData := bean2.NewDeploymentAndCmCsConfig().WithConfigData(secretRespJson).WithResourceType(bean.CS).
		WithResolvedValue(resolvedSecretDataStringJson).WithVariableSnapshot(resolvedCmCsMetadataDto.VariableMapCS)

	configDataDto.WithConfigMapData(cmConfigData).WithSecretData(secretConfigData)
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

func (impl *DeploymentConfigurationServiceImpl) ResolveCmCs(ctx context.Context, envId, appId, clusterId int, userHasAdminAccess bool) (*bean2.ResolvedCmCsMetadataDto, error) {
	scope := resourceQualifiers.Scope{
		AppId:     appId,
		EnvId:     envId,
		ClusterId: clusterId,
	}
	cmcsMetadataDto, err := impl.getMergedCmCs(envId, appId)
	if err != nil {
		impl.logger.Errorw("error in getting merged cm cs", "appId", appId, "envId", envId, "err", err)
		return nil, err
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
		impl.encodeSecretDataFromNonAdminUsers(resolvedSecretDataList, userHasAdminAccess)
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
	appId, envId, clusterId int, userHasAdminAccess bool) (*bean2.DeploymentAndCmCsConfigDto, error) {

	if configDataQueryParams.IsRequestMadeForOneResource() {
		return impl.getCmCsEditDataForPublishedOnly(configDataQueryParams, envId, appId)
	}
	//ConfigMapsData and SecretsData are populated here
	configData, err := impl.getCmCsPublishedConfigResponse(ctx, envId, appId, clusterId, userHasAdminAccess)
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
	configData.WithPipelineConfigData(pipelineConfigData)
	return configData, nil
}

func (impl *DeploymentConfigurationServiceImpl) getPublishedPipelineStrategyConfig(ctx context.Context, appId int, envId int) (*bean2.DeploymentAndCmCsConfig, error) {
	pipelineStrategyJson := json.RawMessage{}
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
