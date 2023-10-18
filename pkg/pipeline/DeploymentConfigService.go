package pipeline

import (
	"context"
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	repository2 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	repository6 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	errors2 "github.com/juju/errors"
	"go.uber.org/zap"
)

type DeploymentConfigService interface {
	GetLatestDeploymentConfigurationByPipelineId(ctx context.Context, pipelineId int, userHasAdminAccess bool) (*history.AllDeploymentConfigurationDetail, error)
}

type DeploymentConfigServiceImpl struct {
	logger                       *zap.SugaredLogger
	envConfigOverrideRepository  chartConfig.EnvConfigOverrideRepository
	chartRepository              chartRepoRepository.ChartRepository
	pipelineRepository           pipelineConfig.PipelineRepository
	envLevelAppMetricsRepository repository.EnvLevelAppMetricsRepository
	appLevelMetricsRepository    repository.AppLevelMetricsRepository
	pipelineConfigRepository     chartConfig.PipelineConfigRepository
	configMapRepository          chartConfig.ConfigMapRepository
	configMapHistoryService      history.ConfigMapHistoryService
	chartRefRepository           chartRepoRepository.ChartRefRepository
	variableEntityMappingService variables.VariableEntityMappingService
	scopedVariableService        variables.ScopedVariableService
	variableTemplateParser       parsers.VariableTemplateParser
}

func NewDeploymentConfigServiceImpl(logger *zap.SugaredLogger,
	envConfigOverrideRepository chartConfig.EnvConfigOverrideRepository,
	chartRepository chartRepoRepository.ChartRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	envLevelAppMetricsRepository repository.EnvLevelAppMetricsRepository,
	appLevelMetricsRepository repository.AppLevelMetricsRepository,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	configMapRepository chartConfig.ConfigMapRepository,
	configMapHistoryService history.ConfigMapHistoryService,
	chartRefRepository chartRepoRepository.ChartRefRepository,
	variableEntityMappingService variables.VariableEntityMappingService,
	scopedVariableService variables.ScopedVariableService,
	variableTemplateParser parsers.VariableTemplateParser,
) *DeploymentConfigServiceImpl {
	return &DeploymentConfigServiceImpl{
		logger:                       logger,
		envConfigOverrideRepository:  envConfigOverrideRepository,
		chartRepository:              chartRepository,
		pipelineRepository:           pipelineRepository,
		envLevelAppMetricsRepository: envLevelAppMetricsRepository,
		appLevelMetricsRepository:    appLevelMetricsRepository,
		pipelineConfigRepository:     pipelineConfigRepository,
		configMapRepository:          configMapRepository,
		configMapHistoryService:      configMapHistoryService,
		chartRefRepository:           chartRefRepository,
		variableEntityMappingService: variableEntityMappingService,
		scopedVariableService:        scopedVariableService,
		variableTemplateParser:       variableTemplateParser,
	}
}

func (impl *DeploymentConfigServiceImpl) GetLatestDeploymentConfigurationByPipelineId(ctx context.Context, pipelineId int, userHasAdminAccess bool) (*history.AllDeploymentConfigurationDetail, error) {
	configResp := &history.AllDeploymentConfigurationDetail{}
	pipeline, err := impl.pipelineRepository.FindById(pipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting pipeline by id", "err", err, "id", pipelineId)
		return nil, err
	}

	deploymentTemplateConfig, err := impl.GetLatestDeploymentTemplateConfig(ctx, pipeline)
	if err != nil {
		impl.logger.Errorw("error in getting latest deploymentTemplate", "err", err)
		return nil, err
	}
	configResp.DeploymentTemplateConfig = deploymentTemplateConfig

	pipelineStrategyConfig, err := impl.GetLatestPipelineStrategyConfig(pipeline)
	if err != nil && errors2.IsNotFound(err) == false {
		impl.logger.Errorw("error in getting latest pipelineStrategyConfig", "err", err)
		return nil, err
	}
	configResp.StrategyConfig = pipelineStrategyConfig

	configMapConfig, secretConfig, err := impl.GetLatestCMCSConfig(pipeline, userHasAdminAccess)
	if err != nil {
		impl.logger.Errorw("error in getting latest CM/CS config", "err", err)
		return nil, err
	}
	configResp.ConfigMapConfig = configMapConfig
	configResp.SecretConfig = secretConfig
	return configResp, nil
}

func (impl *DeploymentConfigServiceImpl) extractVariablesAndGetScopedVariables(template string, scope resourceQualifiers.Scope, entity repository6.Entity, isSuperAdmin bool) (string, map[string]string, error) {

	variableMap := make(map[string]string)
	entityToVariables, err := impl.variableEntityMappingService.GetAllMappingsForEntities([]repository6.Entity{entity})
	if err != nil {
		return template, variableMap, err
	}
	scopedVariables := make([]*models.ScopedVariableData, 0)
	if _, ok := entityToVariables[entity]; ok && len(entityToVariables[entity]) > 0 {
		scopedVariables, err = impl.scopedVariableService.GetScopedVariables(scope, entityToVariables[entity], isSuperAdmin)
		if err != nil {
			return template, variableMap, err
		}
	}

	for _, variable := range scopedVariables {
		variableMap[variable.VariableName] = variable.VariableValue.StringValue()
	}

	if len(variableMap) == 0 {
		return template, variableMap, nil
	}

	parserRequest := parsers.VariableParserRequest{Template: template, Variables: scopedVariables, TemplateType: parsers.JsonVariableTemplate}
	parserResponse := impl.variableTemplateParser.ParseTemplate(parserRequest)
	err = parserResponse.Error
	if err != nil {
		return template, variableMap, err
	}
	resolvedTemplate := parserResponse.ResolvedTemplate

	return resolvedTemplate, variableMap, nil
}

func (impl *DeploymentConfigServiceImpl) GetLatestDeploymentTemplateConfig(ctx context.Context, pipeline *pipelineConfig.Pipeline) (*history.HistoryDetailDto, error) {
	isAppMetricsEnabled := false
	envLevelAppMetrics, err := impl.envLevelAppMetricsRepository.FindByAppIdAndEnvId(pipeline.AppId, pipeline.EnvironmentId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting env level app metrics", "err", err, "appId", pipeline.AppId, "envId", pipeline.EnvironmentId)
	} else if err == pg.ErrNoRows {
		appLevelAppMetrics, err := impl.appLevelMetricsRepository.FindByAppId(pipeline.AppId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting app level app metrics", "err", err, "appId", pipeline.AppId)
		} else if err == nil {
			isAppMetricsEnabled = appLevelAppMetrics.AppMetrics
		}
	} else {
		isAppMetricsEnabled = *envLevelAppMetrics.AppMetrics
	}
	envOverride, err := impl.envConfigOverrideRepository.ActiveEnvConfigOverride(pipeline.AppId, pipeline.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("not able to get envConfigOverride", "err", err, "appId", pipeline.AppId, "envId", pipeline.EnvironmentId)
		return nil, err
	}
	impl.logger.Infow("received override chart", "envConfigOverride", envOverride)
	deploymentTemplateConfig := &history.HistoryDetailDto{}
	if envOverride != nil && envOverride.Id > 0 && envOverride.IsOverride {
		if envOverride.Chart != nil {
			chartRef, err := impl.chartRefRepository.FindById(envOverride.Chart.ChartRefId)
			if err != nil {
				impl.logger.Errorw("error in getting chartRef by id", "err", err, "chartRefId", envOverride.Chart.ChartRefId)
				return nil, err
			}
			scope := resourceQualifiers.Scope{
				AppId:     pipeline.AppId,
				EnvId:     pipeline.EnvironmentId,
				ClusterId: pipeline.Environment.ClusterId,
			}
			entity := repository6.Entity{
				EntityType: repository6.EntityTypeDeploymentTemplateEnvLevel,
				EntityId:   envOverride.Id,
			}
			isSuperAdmin, err := util.GetIsSuperAdminFromContext(ctx)
			if err != nil {
				return nil, err
			}
			resolvedTemplate, scopedVariablesMap, err := impl.extractVariablesAndGetScopedVariables(envOverride.EnvOverrideValues, scope, entity, isSuperAdmin)
			if err != nil {
				impl.logger.Errorw("could not resolve template", "err", err, "envOverrideId", envOverride.Id, "scope", scope, "pipelineId", pipeline.Id)
			}

			deploymentTemplateConfig = &history.HistoryDetailDto{
				TemplateName:        envOverride.Chart.ChartName,
				TemplateVersion:     chartRef.Version,
				IsAppMetricsEnabled: &isAppMetricsEnabled,
				CodeEditorValue: &history.HistoryDetailConfig{
					DisplayName: "values.yaml",
					Value:       envOverride.EnvOverrideValues,
				},
				VariableSnapshot:     scopedVariablesMap,
				ResolvedTemplateData: resolvedTemplate,
			}
		}
	} else {
		chart, err := impl.chartRepository.FindLatestChartForAppByAppId(pipeline.AppId)
		if err != nil {
			impl.logger.Errorw("error in getting chart by appId", "err", err, "appId", pipeline.AppId)
			return nil, err
		}
		chartRef, err := impl.chartRefRepository.FindById(chart.ChartRefId)
		if err != nil {
			impl.logger.Errorw("error in getting chartRef by id", "err", err, "chartRefId", envOverride.Chart.ChartRefId)
			return nil, err
		}
		//Scope contains env and cluster ID because a pipeline will always have those even if inheriting base template
		scope := resourceQualifiers.Scope{
			AppId:     pipeline.AppId,
			EnvId:     pipeline.EnvironmentId,
			ClusterId: pipeline.Environment.ClusterId,
		}
		entity := repository6.Entity{
			EntityType: repository6.EntityTypeDeploymentTemplateAppLevel,
			EntityId:   chart.Id,
		}
		isSuperAdmin, err := util.GetIsSuperAdminFromContext(ctx)
		if err != nil {
			return nil, err
		}
		resolvedTemplate, scopedVariablesMap, err := impl.extractVariablesAndGetScopedVariables(chart.GlobalOverride, scope, entity, isSuperAdmin)
		if err != nil {
			impl.logger.Errorw("could not resolve template", "err", err, "chartId", chart.Id, "scope", scope, "pipelineId", pipeline.Id)
		}
		deploymentTemplateConfig = &history.HistoryDetailDto{
			TemplateName:        chart.ChartName,
			TemplateVersion:     chartRef.Version,
			IsAppMetricsEnabled: &isAppMetricsEnabled,
			CodeEditorValue: &history.HistoryDetailConfig{
				DisplayName: "values.yaml",
				Value:       chart.GlobalOverride,
			},
			VariableSnapshot:     scopedVariablesMap,
			ResolvedTemplateData: resolvedTemplate,
		}
	}
	return deploymentTemplateConfig, nil
}

func (impl *DeploymentConfigServiceImpl) GetLatestPipelineStrategyConfig(pipeline *pipelineConfig.Pipeline) (*history.HistoryDetailDto, error) {

	pipelineStrategy, err := impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(pipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in getting default pipelineStrategy by pipelineId", "err", err, "pipelineId", pipeline.Id)
		return nil, err
	}
	pipelineStrategyConfig := &history.HistoryDetailDto{
		Strategy:            string(pipelineStrategy.Strategy),
		PipelineTriggerType: pipeline.TriggerType,
		CodeEditorValue: &history.HistoryDetailConfig{
			DisplayName: "Strategy configuration",
			Value:       pipelineStrategy.Config,
		},
	}
	return pipelineStrategyConfig, nil
}

func (impl *DeploymentConfigServiceImpl) GetLatestCMCSConfig(pipeline *pipelineConfig.Pipeline, userHasAdminAccess bool) ([]*history.ComponentLevelHistoryDetailDto, []*history.ComponentLevelHistoryDetailDto, error) {

	configAppLevel, err := impl.configMapRepository.GetByAppIdAppLevel(pipeline.AppId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Errorw("error in getting CM/CS app level data", "err", err, "appId", pipeline.AppId)
		return nil, nil, err
	}
	var configMapAppLevel string
	var secretAppLevel string
	if configAppLevel != nil && configAppLevel.Id > 0 {
		configMapAppLevel = configAppLevel.ConfigMapData
		secretAppLevel = configAppLevel.SecretData
	}
	configEnvLevel, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(pipeline.AppId, pipeline.EnvironmentId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Errorw("error in getting CM/CS env level data", "err", err, "appId", pipeline.AppId)
		return nil, nil, err
	}
	var configMapEnvLevel string
	var secretEnvLevel string
	if configEnvLevel != nil && configEnvLevel.Id > 0 {
		configMapEnvLevel = configEnvLevel.ConfigMapData
		secretEnvLevel = configEnvLevel.SecretData
	}
	mergedConfigMap, err := impl.GetMergedCMCSConfigMap(configMapAppLevel, configMapEnvLevel, repository2.CONFIGMAP_TYPE)
	if err != nil {
		impl.logger.Errorw("error in merging app level and env level CM configs", "err", err)
		return nil, nil, err
	}
	mergedSecret, err := impl.GetMergedCMCSConfigMap(secretAppLevel, secretEnvLevel, repository2.SECRET_TYPE)
	if err != nil {
		impl.logger.Errorw("error in merging app level and env level CM configs", "err", err)
		return nil, nil, err
	}

	var cmConfigsDto []*history.ComponentLevelHistoryDetailDto
	for _, data := range mergedConfigMap {
		convertedData, err := impl.configMapHistoryService.ConvertConfigDataToComponentLevelDto(data, repository2.CONFIGMAP_TYPE, userHasAdminAccess, nil, "")
		if err != nil {
			impl.logger.Errorw("error in converting cmConfig to componentLevelData", "err", err)
			return nil, nil, err
		}
		cmConfigsDto = append(cmConfigsDto, convertedData)
	}

	var secretConfigsDto []*history.ComponentLevelHistoryDetailDto
	for _, data := range mergedSecret {
		convertedData, err := impl.configMapHistoryService.ConvertConfigDataToComponentLevelDto(data, repository2.SECRET_TYPE, userHasAdminAccess, nil, "")
		if err != nil {
			impl.logger.Errorw("error in converting secretConfig to componentLevelData", "err", err)
			return nil, nil, err
		}
		secretConfigsDto = append(secretConfigsDto, convertedData)
	}
	return cmConfigsDto, secretConfigsDto, nil
}

func (impl *DeploymentConfigServiceImpl) GetMergedCMCSConfigMap(appLevelConfig, envLevelConfig string, configType repository2.ConfigType) (map[string]*history.ConfigData, error) {
	envLevelMap := make(map[string]*history.ConfigData, 0)
	finalMap := make(map[string]*history.ConfigData, 0)
	if configType == repository2.CONFIGMAP_TYPE {
		appLevelConfigMap := &history.ConfigList{}
		envLevelConfigMap := &history.ConfigList{}
		if len(appLevelConfig) > 0 {
			err := json.Unmarshal([]byte(appLevelConfig), appLevelConfigMap)
			if err != nil {
				impl.logger.Errorw("error in un-marshaling CM app level config", "err", err)
				return nil, err
			}
		}
		if len(envLevelConfig) > 0 {
			err := json.Unmarshal([]byte(envLevelConfig), envLevelConfigMap)
			if err != nil {
				impl.logger.Errorw("error in un-marshaling CM env level config", "err", err)
				return nil, err
			}
		}
		for _, data := range envLevelConfigMap.ConfigData {
			envLevelMap[data.Name] = data
			finalMap[data.Name] = data
		}
		for _, data := range appLevelConfigMap.ConfigData {
			if _, ok := envLevelMap[data.Name]; !ok {
				finalMap[data.Name] = data
			}
		}
	} else if configType == repository2.SECRET_TYPE {
		appLevelSecret := &history.SecretList{}
		envLevelSecret := &history.SecretList{}
		if len(appLevelConfig) > 0 {
			err := json.Unmarshal([]byte(appLevelConfig), appLevelSecret)
			if err != nil {
				impl.logger.Errorw("error in un-marshaling CS app level config", "err", err)
				return nil, err
			}
		}
		if len(envLevelConfig) > 0 {
			err := json.Unmarshal([]byte(envLevelConfig), envLevelSecret)
			if err != nil {
				impl.logger.Errorw("error in un-marshaling CS env level config", "err", err)
				return nil, err
			}
		}
		for _, data := range envLevelSecret.ConfigData {
			envLevelMap[data.Name] = data
			finalMap[data.Name] = data
		}
		for _, data := range appLevelSecret.ConfigData {
			if _, ok := envLevelMap[data.Name]; !ok {
				finalMap[data.Name] = data
			}
		}
	}
	return finalMap, nil
}
