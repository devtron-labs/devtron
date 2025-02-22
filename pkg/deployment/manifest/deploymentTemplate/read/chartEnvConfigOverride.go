package read

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/bean"
	"go.uber.org/zap"
)

type EnvConfigOverrideService interface {
	GetByChartAndEnvironment(chartId, targetEnvironmentId int) (*bean.EnvConfigOverride, error)
	ActiveEnvConfigOverride(appId, environmentId int) (*bean.EnvConfigOverride, error) //successful env config
	GetByIdIncludingInactive(id int) (*bean.EnvConfigOverride, error)
	GetByEnvironment(targetEnvironmentId int) ([]*bean.EnvConfigOverride, error)
	GetEnvConfigByChartId(chartId int) ([]*bean.EnvConfigOverride, error)
	FindLatestChartForAppByAppIdAndEnvId(appId, targetEnvironmentId int) (*bean.EnvConfigOverride, error)
	FindChartRefIdsForLatestChartForAppByAppIdAndEnvIds(appId int, targetEnvironmentIds []int) (map[int]int, error)
	FindChartByAppIdAndEnvIdAndChartRefId(appId, targetEnvironmentId int, chartRefId int) (*bean.EnvConfigOverride, error)
	FindChartForAppByAppIdAndEnvId(appId, targetEnvironmentId int) (*bean.EnvConfigOverride, error)
	GetByAppIdEnvIdAndChartRefId(appId, envId int, chartRefId int) (*bean.EnvConfigOverride, error)
	GetAllOverridesForApp(appId int) ([]*bean.EnvConfigOverride, error)
}

type EnvConfigOverrideReadServiceImpl struct {
	envConfigOverrideRepository chartConfig.EnvConfigOverrideRepository
	logger                      *zap.SugaredLogger
}

func NewEnvConfigOverrideReadServiceImpl(repository chartConfig.EnvConfigOverrideRepository,
	logger *zap.SugaredLogger) *EnvConfigOverrideReadServiceImpl {
	return &EnvConfigOverrideReadServiceImpl{
		envConfigOverrideRepository: repository,
		logger:                      logger,
	}
}

func (impl EnvConfigOverrideReadServiceImpl) GetByChartAndEnvironment(chartId, targetEnvironmentId int) (*bean.EnvConfigOverride, error) {
	overrideDBObj, err := impl.envConfigOverrideRepository.GetByChartAndEnvironment(chartId, targetEnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in getting chart env config override", "chartId", chartId, "targetEnvironmentId", targetEnvironmentId, "err", err)
		return nil, err
	}
	return adapter.EnvOverrideDBToDTO(overrideDBObj), nil
}

func (impl EnvConfigOverrideReadServiceImpl) ActiveEnvConfigOverride(appId, environmentId int) (*bean.EnvConfigOverride, error) {
	overrideDBObj, err := impl.envConfigOverrideRepository.ActiveEnvConfigOverride(appId, environmentId)
	if err != nil {
		impl.logger.Errorw("error in getting chart env config override", "appId", appId, "environmentId", environmentId, "err", err)
		return nil, err
	}
	return adapter.EnvOverrideDBToDTO(overrideDBObj), nil
}

func (impl EnvConfigOverrideReadServiceImpl) GetByIdIncludingInactive(id int) (*bean.EnvConfigOverride, error) {
	overrideDBObj, err := impl.envConfigOverrideRepository.GetByIdIncludingInactive(id)
	if err != nil {
		impl.logger.Errorw("error in getting chart env config override", "id", id, "err", err)
		return nil, err
	}
	return adapter.EnvOverrideDBToDTO(overrideDBObj), nil
}

func (impl EnvConfigOverrideReadServiceImpl) GetByEnvironment(targetEnvironmentId int) ([]*bean.EnvConfigOverride, error) {
	overrideDBObjs, err := impl.envConfigOverrideRepository.GetByEnvironment(targetEnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in getting chart env config override", "targetEnvironmentId", targetEnvironmentId, "err", err)
		return nil, err
	}
	envConfigOverrides := make([]*bean.EnvConfigOverride, len(overrideDBObjs))
	for _, dbObj := range overrideDBObjs {
		envConfigOverrides = append(envConfigOverrides, adapter.EnvOverrideDBToDTO(&dbObj))
	}
	return envConfigOverrides, nil
}

func (impl EnvConfigOverrideReadServiceImpl) GetEnvConfigByChartId(chartId int) ([]*bean.EnvConfigOverride, error) {
	overrideDBObjs, err := impl.envConfigOverrideRepository.GetEnvConfigByChartId(chartId)
	if err != nil {
		impl.logger.Errorw("error in getting chart env config override", "chartId", chartId, "err", err)
		return nil, err
	}
	envConfigOverrides := make([]*bean.EnvConfigOverride, len(overrideDBObjs))
	for _, dbObj := range overrideDBObjs {
		envConfigOverrides = append(envConfigOverrides, adapter.EnvOverrideDBToDTO(&dbObj))
	}
	return envConfigOverrides, nil
}

func (impl EnvConfigOverrideReadServiceImpl) FindLatestChartForAppByAppIdAndEnvId(appId, targetEnvironmentId int) (*bean.EnvConfigOverride, error) {
	overrideDBObj, err := impl.envConfigOverrideRepository.FindLatestChartForAppByAppIdAndEnvId(appId, targetEnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in getting chart env config override", "appId", appId, "targetEnvironmentId", targetEnvironmentId, "err", err)
		return nil, err
	}
	return adapter.EnvOverrideDBToDTO(overrideDBObj), nil
}

func (impl EnvConfigOverrideReadServiceImpl) FindChartRefIdsForLatestChartForAppByAppIdAndEnvIds(appId int, targetEnvironmentIds []int) (map[int]int, error) {
	envChartMap, err := impl.envConfigOverrideRepository.FindChartRefIdsForLatestChartForAppByAppIdAndEnvIds(appId, targetEnvironmentIds)
	if err != nil {
		impl.logger.Errorw("error in getting chart env config override", "appId", appId, "targetEnvironmentIds", targetEnvironmentIds, "err", err)
		return nil, err
	}
	return envChartMap, nil
}

func (impl EnvConfigOverrideReadServiceImpl) FindChartByAppIdAndEnvIdAndChartRefId(appId, targetEnvironmentId int, chartRefId int) (*bean.EnvConfigOverride, error) {
	overrideDBObj, err := impl.envConfigOverrideRepository.FindChartByAppIdAndEnvIdAndChartRefId(appId, targetEnvironmentId, chartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting chart env config override", "appId", appId, "targetEnvironmentIds", targetEnvironmentId, "chartRefId", chartRefId, "err", err)
		return nil, err
	}
	return adapter.EnvOverrideDBToDTO(overrideDBObj), nil
}

func (impl EnvConfigOverrideReadServiceImpl) FindChartForAppByAppIdAndEnvId(appId, targetEnvironmentId int) (*bean.EnvConfigOverride, error) {
	overrideDBObj, err := impl.envConfigOverrideRepository.FindChartForAppByAppIdAndEnvId(appId, targetEnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in getting chart env config override", "appId", appId, "targetEnvironmentId", targetEnvironmentId, "err", err)
		return nil, err
	}
	return adapter.EnvOverrideDBToDTO(overrideDBObj), nil
}

func (impl EnvConfigOverrideReadServiceImpl) GetByAppIdEnvIdAndChartRefId(appId, envId int, chartRefId int) (*bean.EnvConfigOverride, error) {
	overrideDBObj, err := impl.envConfigOverrideRepository.GetByAppIdEnvIdAndChartRefId(appId, envId, chartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting chart env config override", "appId", appId, "envId", envId, "chartRefId", chartRefId, "err", err)
		return nil, err
	}
	return adapter.EnvOverrideDBToDTO(overrideDBObj), nil
}

func (impl EnvConfigOverrideReadServiceImpl) GetAllOverridesForApp(appId int) ([]*bean.EnvConfigOverride, error) {
	overrideDBObjs, err := impl.envConfigOverrideRepository.GetAllOverridesForApp(appId)
	if err != nil {
		impl.logger.Errorw("error in getting chart env config override", "appId", appId, "envId", "err", err)
		return nil, err
	}
	envConfigOverrides := make([]*bean.EnvConfigOverride, len(overrideDBObjs))
	for _, dbObj := range overrideDBObjs {
		envConfigOverrides = append(envConfigOverrides, adapter.EnvOverrideDBToDTO(&dbObj))
	}
	return envConfigOverrides, nil
}
