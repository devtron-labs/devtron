package common

import (
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeploymentConfigService interface {
	CreateOrUpdateConfig(tx *pg.Tx, config *bean.DeploymentConfig, userId int32) (*bean.DeploymentConfig, error)
	CreateOrUpdateConfigsInBulk(tx *pg.Tx, configs []*bean.DeploymentConfig, userId int32) ([]*bean.DeploymentConfig, error)
	UpdateConfigs(tx *pg.Tx, configs []*bean.DeploymentConfig, userId int32) ([]*bean.DeploymentConfig, error)
	GetConfigForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error)
	GetAndMigrateConfigIfAbsentForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error)
	GetConfigForHelmApps(appId, envId int) (*bean.DeploymentConfig, error)
	GetConfigEvenIfInactive(appId, envId int) (*bean.DeploymentConfig, error)
	GetAndMigrateConfigIfAbsentForHelmApp(appId, envId int) (*bean.DeploymentConfig, error)
	GetDeploymentConfigInBulk(configSelector []*bean.DeploymentConfigSelector) (map[bean.UniqueDeploymentConfigIdentifier]*bean.DeploymentConfig, error)
}

type DeploymentConfigServiceImpl struct {
	deploymentConfigRepository deploymentConfig.Repository
	logger                     *zap.SugaredLogger
	chartRepository            chartRepoRepository.ChartRepository
	pipelineRepository         pipelineConfig.PipelineRepository
	appRepository              appRepository.AppRepository
	installedAppRepository     repository.InstalledAppRepository
}

func NewDeploymentConfigServiceImpl(
	deploymentConfigRepository deploymentConfig.Repository,
	logger *zap.SugaredLogger,
	chartRepository chartRepoRepository.ChartRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	appRepository appRepository.AppRepository,
	installedAppRepository repository.InstalledAppRepository,
) *DeploymentConfigServiceImpl {
	return &DeploymentConfigServiceImpl{
		deploymentConfigRepository: deploymentConfigRepository,
		logger:                     logger,
		chartRepository:            chartRepository,
		pipelineRepository:         pipelineRepository,
		appRepository:              appRepository,
		installedAppRepository:     installedAppRepository,
	}
}

func (impl *DeploymentConfigServiceImpl) CreateOrUpdateConfig(tx *pg.Tx, config *bean.DeploymentConfig, userId int32) (*bean.DeploymentConfig, error) {

	configDbObj, err := impl.GetConfigDBObj(config.AppId, config.EnvironmentId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching deployment config from DB by appId and envId",
			"appId", config.AppId, "envId", config.EnvironmentId, "err", err)
	}

	newDBObj := ConvertDeploymentConfigDTOToDbObj(config)

	if configDbObj == nil || (configDbObj != nil && configDbObj.Id == 0) {
		newDBObj.AuditLog.CreateAuditLog(userId)
		newDBObj, err = impl.deploymentConfigRepository.Save(tx, newDBObj)
		if err != nil {
			impl.logger.Errorw("error in saving deploymentConfig", "appId", config.AppId, "envId", config.EnvironmentId, "err", err)
			return nil, err
		}
	} else {
		newDBObj.Id = configDbObj.Id
		newDBObj.AuditLog.UpdateAuditLog(userId)
		newDBObj, err = impl.deploymentConfigRepository.Update(tx, newDBObj)
		if err != nil {
			impl.logger.Errorw("error in updating deploymentConfig", "appId", config.AppId, "envId", config.EnvironmentId, "err", err)
			return nil, err
		}
	}

	return ConvertDeploymentConfigDbObjToDTO(newDBObj), nil
}

func (impl *DeploymentConfigServiceImpl) CreateOrUpdateConfigsInBulk(tx *pg.Tx, configs []*bean.DeploymentConfig, userId int32) ([]*bean.DeploymentConfig, error) {

	appIdToEnvIdsMap := GetAppIdToEnvIsMapping(
		configs,
		func(config *bean.DeploymentConfig) int { return config.AppId },
		func(config *bean.DeploymentConfig) int { return config.EnvironmentId })

	deploymentConfigDbObj, err := impl.deploymentConfigRepository.GetAppAndEnvLevelConfigsInBulk(appIdToEnvIdsMap)
	if err != nil {
		impl.logger.Errorw("error in getting deployment configs by appIdToEnvIds", "appIdToEnvIdsMap", appIdToEnvIdsMap, "err", err)
		return nil, err
	}

	savedDeploymentConfigsMap := make(map[bean.UniqueDeploymentConfigIdentifier]*deploymentConfig.DeploymentConfig)
	for _, savedConfig := range deploymentConfigDbObj {
		savedDeploymentConfigsMap[bean.GetConfigUniqueIdentifier(savedConfig.AppId, savedConfig.EnvironmentId)] = savedConfig
	}

	oldDeploymentConfigs := make([]*deploymentConfig.DeploymentConfig, 0)
	newDeploymentConfigs := make([]*deploymentConfig.DeploymentConfig, 0)

	for _, requestConfig := range configs {

		if oldConfig, isConfigPresent := savedDeploymentConfigsMap[bean.GetConfigUniqueIdentifier(requestConfig.AppId, requestConfig.EnvironmentId)]; isConfigPresent {

			oldConfigDbObject := ConvertDeploymentConfigDTOToDbObj(requestConfig)
			oldConfigDbObject.Id = oldConfig.Id
			oldConfigDbObject.AuditLog.UpdateAuditLog(userId)
			oldDeploymentConfigs = append(oldDeploymentConfigs, oldConfigDbObject)

		} else if !isConfigPresent {

			newConfigDbObject := ConvertDeploymentConfigDTOToDbObj(requestConfig)
			newConfigDbObject.AuditLog.CreateAuditLog(userId)
			newDeploymentConfigs = append(newDeploymentConfigs, ConvertDeploymentConfigDTOToDbObj(requestConfig))

		}
	}

	if len(newDeploymentConfigs) > 0 {
		newDeploymentConfigs, err = impl.deploymentConfigRepository.SaveAll(tx, newDeploymentConfigs)
		if err != nil {
			impl.logger.Errorw("error in saving deployment config in db", "err", err)
			return nil, err
		}
	}
	if len(oldDeploymentConfigs) > 0 {
		oldDeploymentConfigs, err = impl.deploymentConfigRepository.UpdateAll(tx, oldDeploymentConfigs)
		if err != nil {
			impl.logger.Errorw("error in saving deployment config in db", "err", err)
			return nil, err
		}
	}

	allDeploymentConfigs := make([]*bean.DeploymentConfig, 0, len(oldDeploymentConfigs)+len(newDeploymentConfigs))

	for _, c := range oldDeploymentConfigs {
		allDeploymentConfigs = append(allDeploymentConfigs, ConvertDeploymentConfigDbObjToDTO(c))
	}
	for _, c := range newDeploymentConfigs {
		allDeploymentConfigs = append(allDeploymentConfigs, ConvertDeploymentConfigDbObjToDTO(c))
	}
	return allDeploymentConfigs, nil
}

func (impl *DeploymentConfigServiceImpl) UpdateConfigs(tx *pg.Tx, configs []*bean.DeploymentConfig, userId int32) ([]*bean.DeploymentConfig, error) {

	configDbObject := make([]*deploymentConfig.DeploymentConfig, 0)
	for _, c := range configs {
		dbObj := ConvertDeploymentConfigDTOToDbObj(c)
		dbObj.AuditLog.UpdateAuditLog(userId)
		configDbObject = append(configDbObject)
	}

	var err error
	configDbObject, err = impl.deploymentConfigRepository.UpdateAll(tx, configDbObject)
	if err != nil {
		impl.logger.Errorw("error in saving deployment config in db", "err", err)
		return nil, err
	}

	updatedConfigs := make([]*bean.DeploymentConfig, 0)
	for _, c := range configDbObject {
		updatedConfigs = append(updatedConfigs, ConvertDeploymentConfigDbObjToDTO(c))
	}

	return updatedConfigs, nil
}

func (impl *DeploymentConfigServiceImpl) GetConfigForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error) {
	appLevelConfigDbObj, err := impl.deploymentConfigRepository.GetAppLevelConfigForDevtronApps(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting deployment config db object by appId", "appId", appId, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		appLevelConfigDbObj, err = impl.parseAppLevelConfigForDevtronApps(appId)
		if err != nil {
			impl.logger.Errorw("error in migrating app level config to deployment config", "appId", appId, "err", err)
			return nil, err
		}
	}
	if envId > 0 {
		// if envId>0 then only env level config will be returned, for getting app level config envId should be zero
		appAndEnvLevelConfig, err := impl.deploymentConfigRepository.GetByAppIdAndEnvId(appId, envId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting deployment config db object by appId and envId", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		if err == pg.ErrNoRows {
			appAndEnvLevelConfig, err = impl.parseEnvLevelConfigForDevtronApps(appLevelConfigDbObj, appId, envId)
			if err != nil {
				impl.logger.Errorw("error in migrating app level config to deployment config", "appId", appId, "err", err)
				return nil, err
			}
		}
		return ConvertDeploymentConfigDbObjToDTO(appAndEnvLevelConfig), nil
	}
	return ConvertDeploymentConfigDbObjToDTO(appLevelConfigDbObj), nil
}

func (impl *DeploymentConfigServiceImpl) GetAndMigrateConfigIfAbsentForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error) {

	appLevelConfigDbObj, err := impl.deploymentConfigRepository.GetAppLevelConfigForDevtronApps(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting deployment config db object by appId", "appId", appId, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		impl.logger.Infow("app level deployment config not found, migrating data from charts to deployment_config", "appId", appId, "err", err)
		appLevelConfigDbObj, err = impl.migrateChartsDataToDeploymentConfig(appId)
		if err != nil {
			impl.logger.Errorw("error in migrating app level config to deployment config", "appId", appId, "err", err)
			return nil, err
		}
	}

	if envId > 0 {
		appAndEnvLevelConfig, err := impl.deploymentConfigRepository.GetByAppIdAndEnvId(appId, envId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting deployment config db object by appId and envId", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		if err == pg.ErrNoRows {
			impl.logger.Infow("env level deployment config not found, migrating data from pipeline to deployment_config", "appId", appId, "envId", envId, "err", err)
			appAndEnvLevelConfig, err = impl.migrateDevtronAppsPipelineDataToDeploymentConfig(appLevelConfigDbObj, appId, envId)
			if err != nil {
				impl.logger.Errorw("error in migrating app level config to deployment config", "appId", appId, "err", err)
				return nil, err
			}
		}
		return ConvertDeploymentConfigDbObjToDTO(appAndEnvLevelConfig), nil
	}

	return ConvertDeploymentConfigDbObjToDTO(appLevelConfigDbObj), nil
}

func (impl *DeploymentConfigServiceImpl) migrateChartsDataToDeploymentConfig(appId int) (*deploymentConfig.DeploymentConfig, error) {

	configDbObj, err := impl.parseAppLevelConfigForDevtronApps(appId)
	if err != nil {
		impl.logger.Errorw("error in parsing charts data for devtron apps", "appId", appId, "err", err)
		return nil, err
	}
	configDbObj.AuditLog.CreateAuditLog(1)
	configDbObj, err = impl.deploymentConfigRepository.Save(nil, configDbObj)
	if err != nil {
		impl.logger.Errorw("error in saving deployment config in DB", "appId", appId, "err", err)
		return nil, err
	}
	return configDbObj, nil
}

func (impl *DeploymentConfigServiceImpl) parseAppLevelConfigForDevtronApps(appId int) (*deploymentConfig.DeploymentConfig, error) {
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetch chart for git repo migration by appId", "appId", appId, "err", err)
		return nil, err
	}
	ConfigDbObj := &deploymentConfig.DeploymentConfig{
		ConfigType:    GetDeploymentConfigType(chart.IsCustomGitRepository),
		AppId:         appId,
		Active:        true,
		RepoUrl:       chart.GitRepoUrl,
		ChartLocation: chart.ChartLocation,
	}
	return ConfigDbObj, nil
}

func (impl *DeploymentConfigServiceImpl) migrateDevtronAppsPipelineDataToDeploymentConfig(appLevelConfig *deploymentConfig.DeploymentConfig, appId int, envId int) (*deploymentConfig.DeploymentConfig, error) {

	configDbObj, err := impl.parseEnvLevelConfigForDevtronApps(appLevelConfig, appId, envId)
	if err != nil {
		impl.logger.Errorw("error in parsing config for cd pipeline from appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	configDbObj.AuditLog.CreateAuditLog(bean3.SYSTEM_USER_ID)
	configDbObj, err = impl.deploymentConfigRepository.Save(nil, configDbObj)
	if err != nil {
		impl.logger.Errorw("error in saving deployment config in DB", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	return configDbObj, nil
}

func (impl *DeploymentConfigServiceImpl) parseEnvLevelConfigForDevtronApps(appLevelConfig *deploymentConfig.DeploymentConfig, appId int, envId int) (*deploymentConfig.DeploymentConfig, error) {

	configDbObj := &deploymentConfig.DeploymentConfig{
		AppId:         appId,
		EnvironmentId: envId,
		ConfigType:    appLevelConfig.ConfigType,
		RepoUrl:       appLevelConfig.RepoUrl,
		ChartLocation: appLevelConfig.ChartLocation,
		Active:        true,
	}

	deploymentAppType, err := impl.pipelineRepository.FindDeploymentAppTypeByAppIdAndEnvId(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment app type by appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	configDbObj.DeploymentAppType = deploymentAppType

	return configDbObj, nil
}

func (impl *DeploymentConfigServiceImpl) GetConfigDBObj(appId, envId int) (*deploymentConfig.DeploymentConfig, error) {
	var configDbObj *deploymentConfig.DeploymentConfig
	var err error
	if envId == 0 {
		configDbObj, err = impl.deploymentConfigRepository.GetAppLevelConfigForDevtronApps(appId)
		if err != nil {
			impl.logger.Errorw("error in getting deployment config db object by appId", "appId", configDbObj.AppId, "err", err)
			return nil, err
		}
	} else {
		configDbObj, err = impl.deploymentConfigRepository.GetByAppIdAndEnvId(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in getting deployment config db object by appId and envId", "appId", configDbObj.AppId, "envId", configDbObj.EnvironmentId, "err", err)
			return nil, err
		}
	}
	return configDbObj, nil
}

func (impl *DeploymentConfigServiceImpl) GetConfigForHelmApps(appId, envId int) (*bean.DeploymentConfig, error) {

	helmDeploymentConfig, err := impl.deploymentConfigRepository.GetByAppIdAndEnvId(appId, envId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching deployment config by by appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	if err == pg.ErrNoRows {
		helmDeploymentConfig, err = impl.parseConfigForHelmApps(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in migrating helm deployment config", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
	}
	return ConvertDeploymentConfigDbObjToDTO(helmDeploymentConfig), nil
}

func (impl *DeploymentConfigServiceImpl) GetConfigEvenIfInactive(appId, envId int) (*bean.DeploymentConfig, error) {
	config, err := impl.deploymentConfigRepository.GetByAppIdAndEnvIdEvenIfInactive(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config by appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	return ConvertDeploymentConfigDbObjToDTO(config), nil
}

func (impl *DeploymentConfigServiceImpl) GetAndMigrateConfigIfAbsentForHelmApp(appId, envId int) (*bean.DeploymentConfig, error) {

	helmDeploymentConfig, err := impl.deploymentConfigRepository.GetByAppIdAndEnvId(appId, envId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching deployment config by by appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	if err == pg.ErrNoRows {
		helmDeploymentConfig, err = impl.migrateHelmAppDataToDeploymentConfig(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in migrating helm deployment config", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
	}
	return ConvertDeploymentConfigDbObjToDTO(helmDeploymentConfig), nil
}

func (impl *DeploymentConfigServiceImpl) migrateHelmAppDataToDeploymentConfig(appId, envId int) (*deploymentConfig.DeploymentConfig, error) {

	helmDeploymentConfig, err := impl.parseConfigForHelmApps(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in parsing deployment config for helm app", "appId", appId, "envId", envId, "err", err)
		return helmDeploymentConfig, err
	}

	helmDeploymentConfig.CreateAuditLog(bean3.SYSTEM_USER_ID)
	helmDeploymentConfig, err = impl.deploymentConfigRepository.Save(nil, helmDeploymentConfig)
	if err != nil {
		impl.logger.Errorw("error in saving deployment config for helm app", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	return helmDeploymentConfig, nil
}

func (impl *DeploymentConfigServiceImpl) parseConfigForHelmApps(appId int, envId int) (*deploymentConfig.DeploymentConfig, error) {
	installedApp, err := impl.installedAppRepository.GetInstalledAppsByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in getting installed app by appId", "appId", appId, "err", err)
		return nil, err
	}
	helmDeploymentConfig := &deploymentConfig.DeploymentConfig{
		AppId:             appId,
		EnvironmentId:     envId,
		DeploymentAppType: installedApp.DeploymentAppType,
		ConfigType:        GetDeploymentConfigType(installedApp.IsCustomRepository),
		RepoUrl:           installedApp.GitOpsRepoUrl,
		RepoName:          installedApp.GitOpsRepoName,
		Active:            true,
	}
	return helmDeploymentConfig, nil
}

func (impl *DeploymentConfigServiceImpl) GetDeploymentConfigInBulk(configSelector []*bean.DeploymentConfigSelector) (map[bean.UniqueDeploymentConfigIdentifier]*bean.DeploymentConfig, error) {

	appIds := make([]*int, len(configSelector))
	for _, s := range configSelector {
		appIds = append(appIds, &s.AppId)
	}

	apps, err := impl.appRepository.FindByIds(appIds)
	if err != nil {
		impl.logger.Errorw("error in fetching apps by ids", "ids", appIds, "err", err)
		return nil, err
	}

	devtronAppIds := make([]int, len(appIds))
	helmAppIds := make([]int, len(appIds))
	devtronAppIdMap := make(map[int]bool)
	helmAppIdMap := make(map[int]bool)

	for _, a := range apps {
		switch a.AppType {
		case helper.CustomApp:
			devtronAppIds = append(devtronAppIds, a.Id)
			devtronAppIdMap[a.Id] = true
		case helper.ChartStoreApp:
			helmAppIds = append(helmAppIds, a.Id)
			helmAppIdMap[a.Id] = true
		}
	}

	devtronAppSelectors := make([]*bean.DeploymentConfigSelector, len(appIds))
	helmAppSelectors := make([]*bean.DeploymentConfigSelector, len(appIds))

	for _, c := range configSelector {
		if _, ok := devtronAppIdMap[c.AppId]; ok {
			devtronAppSelectors = append(devtronAppSelectors, c)
		}
		if _, ok := helmAppIdMap[c.AppId]; ok {
			helmAppSelectors = append(helmAppSelectors, c)
		}
	}

	devtronAppConfigs, err := impl.GetDevtronAppConfigInBulk(devtronAppIds, devtronAppSelectors)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config for devtron apps", "devtronAppSelectors", devtronAppSelectors, "err", err)
		return nil, err
	}

	helmAppConfigs, err := impl.GetHelmAppConfigInBulk(helmAppIds, helmAppSelectors)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config for helm apps", "helmAppSelectors", helmAppSelectors, "err", err)
		return nil, err
	}

	allConfigs := make(map[bean.UniqueDeploymentConfigIdentifier]*bean.DeploymentConfig)

	for _, c := range devtronAppConfigs {
		allConfigs[bean.GetConfigUniqueIdentifier(c.AppId, c.EnvironmentId)] = c
	}

	for _, c := range helmAppConfigs {
		allConfigs[bean.GetConfigUniqueIdentifier(c.AppId, c.EnvironmentId)] = c
	}

	return allConfigs, nil
}

func (impl *DeploymentConfigServiceImpl) GetDevtronAppConfigInBulk(appIds []int, configSelector []*bean.DeploymentConfigSelector) ([]*bean.DeploymentConfig, error) {

	appLevelDeploymentConfigs, err := impl.deploymentConfigRepository.GetAppLevelConfigByAppIds(appIds)
	if err != nil {
		impl.logger.Errorw("error in fetching deployment configs by appIds", "appIds", appIds, "err", err)
		return nil, err
	}

	allAppLevelConfigs := make([]*deploymentConfig.DeploymentConfig, 0, len(appIds))
	allAppLevelConfigs = append(allAppLevelConfigs, appLevelDeploymentConfigs...)

	if len(appLevelDeploymentConfigs) < len(appIds) {
		presentAppIds := make(map[int]bool)
		for _, c := range appLevelDeploymentConfigs {
			presentAppIds[c.AppId] = true
		}
		notFoundAppIds := make([]int, 0, len(appIds))
		for _, id := range appIds {
			if _, ok := presentAppIds[id]; !ok {
				notFoundAppIds = append(notFoundAppIds, id)
			}
		}
		migratedAppLevelDeploymentConfigs, err := impl.migrateAppLevelDataTODeploymentConfigInBulk(notFoundAppIds)
		if err != nil {
			impl.logger.Errorw("error in migrating all level deployment configs", "appIds", notFoundAppIds, "err", err)
			return nil, err
		}
		allAppLevelConfigs = append(allAppLevelConfigs, migratedAppLevelDeploymentConfigs...)
	}

	appIdToAppLevelConfigMapping := make(map[int]*deploymentConfig.DeploymentConfig, len(appIds))
	for _, appLevelConfig := range allAppLevelConfigs {
		appIdToAppLevelConfigMapping[appLevelConfig.AppId] = appLevelConfig
	}

	allEnvLevelConfigs := make([]*deploymentConfig.DeploymentConfig, 0, len(configSelector))

	if len(configSelector) > 0 {

		appIdToEnvIdsMap := GetAppIdToEnvIsMapping(
			configSelector,
			func(config *bean.DeploymentConfigSelector) int { return config.AppId },
			func(config *bean.DeploymentConfigSelector) int { return config.EnvironmentId })

		envLevelConfig, err := impl.deploymentConfigRepository.GetAppAndEnvLevelConfigsInBulk(appIdToEnvIdsMap)
		if err != nil {
			impl.logger.Errorw("error in getting and and env level config in bulk", "appIdToEnvIdsMap", appIdToEnvIdsMap, "err", err)
			return nil, err
		}
		allEnvLevelConfigs = append(allEnvLevelConfigs, envLevelConfig...)
		if len(envLevelConfig) < len(configSelector) {

			notFoundDeploymentConfigMap := make(map[bean.UniqueDeploymentConfigIdentifier]bool)
			presentDeploymentConfigMap := make(map[bean.UniqueDeploymentConfigIdentifier]bool)

			for _, c := range envLevelConfig {
				presentDeploymentConfigMap[bean.GetConfigUniqueIdentifier(c.AppId, c.EnvironmentId)] = true
			}

			for _, c := range configSelector {
				key := bean.GetConfigUniqueIdentifier(c.AppId, c.EnvironmentId)
				if _, ok := presentDeploymentConfigMap[key]; !ok {
					notFoundDeploymentConfigMap[key] = true
				}
			}

			migratedEnvLevelDeploymentConfigs, err := impl.migrateEnvLevelDataToDeploymentConfigInBulk(notFoundDeploymentConfigMap, appIdToAppLevelConfigMapping)
			if err != nil {
				impl.logger.Errorw("error in migrating env level configs", "notFoundDeploymentConfigMap", notFoundDeploymentConfigMap, "err", err)
				return nil, err
			}
			allEnvLevelConfigs = append(allEnvLevelConfigs, migratedEnvLevelDeploymentConfigs...)
		}
	}

	envLevelConfigMapping := make(map[bean.UniqueDeploymentConfigIdentifier]*bean.DeploymentConfig)
	for _, c := range envLevelConfigMapping {
		envLevelConfigMapping[bean.GetConfigUniqueIdentifier(c.AppId, c.EnvironmentId)] = c
	}

	allConfigs := make([]*bean.DeploymentConfig, 0)
	for _, c := range configSelector {
		if c.AppId > 0 && c.EnvironmentId > 0 {
			allConfigs = append(allConfigs, envLevelConfigMapping[bean.GetConfigUniqueIdentifier(c.AppId, c.EnvironmentId)])
		} else if c.AppId > 0 {
			// if user has sent only appId in config selector then only app level
			allConfigs = append(allConfigs, ConvertDeploymentConfigDbObjToDTO(appIdToAppLevelConfigMapping[c.AppId]))
		}
	}

	return allConfigs, err
}

func (impl *DeploymentConfigServiceImpl) migrateAppLevelDataTODeploymentConfigInBulk(appIds []int) ([]*deploymentConfig.DeploymentConfig, error) {

	charts, err := impl.chartRepository.FindLatestChartByAppIds(appIds)
	if err != nil {
		impl.logger.Errorw("error in fetching latest chart by appIds", "appIds", appIds, "err", err)
		return nil, err
	}

	allRepoUrls := make([]string, 0)
	for _, c := range charts {
		allRepoUrls = append(allRepoUrls, c.GitRepoUrl)
	}

	configDBObjects := make([]*deploymentConfig.DeploymentConfig, 0, len(appIds))
	for _, c := range charts {
		dbObj := &deploymentConfig.DeploymentConfig{
			ConfigType:    GetDeploymentConfigType(c.IsCustomGitRepository),
			AppId:         c.AppId,
			Active:        true,
			RepoUrl:       c.GitRepoUrl,
			ChartLocation: c.ChartLocation,
		}
		dbObj.AuditLog.CreateAuditLog(bean3.SYSTEM_USER_ID)
		configDBObjects = append(configDBObjects, dbObj)
	}
	configDBObjects, err = impl.deploymentConfigRepository.SaveAll(nil, configDBObjects)
	if err != nil {
		impl.logger.Errorw("error in saving deployment config in DB", "appIds", appIds, "err", err)
		return nil, err
	}

	return configDBObjects, nil
}

func (impl *DeploymentConfigServiceImpl) migrateEnvLevelDataToDeploymentConfigInBulk(notFoundDeploymentConfigMap map[bean.UniqueDeploymentConfigIdentifier]bool, appIdToAppLevelConfigMapping map[int]*deploymentConfig.DeploymentConfig) ([]*deploymentConfig.DeploymentConfig, error) {

	notFoundAppIdEnvIdsMap := make(map[int][]int)
	for uniqueKey, _ := range notFoundDeploymentConfigMap {
		appId, envId := uniqueKey.GetAppAndEnvId()
		if _, ok := notFoundAppIdEnvIdsMap[appId]; !ok {
			notFoundAppIdEnvIdsMap[appId] = make([]int, 0)
		}
		notFoundAppIdEnvIdsMap[appId] = append(notFoundAppIdEnvIdsMap[appId], envId)
	}

	pipelines, err := impl.pipelineRepository.FindByAppIdToEnvIdsMapping(notFoundAppIdEnvIdsMap)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline by envToAppId mapping", "notFoundAppIdEnvIdsMap", notFoundAppIdEnvIdsMap, "err", err)
		return nil, err
	}
	if len(pipelines) == 0 {
		// pipelines are deleted
		return nil, err
	}

	pipelineMap := make(map[bean.UniqueDeploymentConfigIdentifier]*pipelineConfig.Pipeline)
	for _, p := range pipelines {
		pipelineMap[bean.GetConfigUniqueIdentifier(p.AppId, p.EnvironmentId)] = p
	}

	configDBObjects := make([]*deploymentConfig.DeploymentConfig, 0, len(notFoundDeploymentConfigMap))
	for uniqueKey, _ := range notFoundDeploymentConfigMap {
		if _, ok := pipelineMap[uniqueKey]; !ok {
			//skipping for deleted pipelines
			continue
		}
		pipeline := pipelineMap[uniqueKey]
		deploymentAppType := pipelineMap[uniqueKey].DeploymentAppType
		appLevelConfig := appIdToAppLevelConfigMapping[pipeline.AppId]
		configDbObj := &deploymentConfig.DeploymentConfig{
			AppId:             pipeline.AppId,
			EnvironmentId:     pipeline.EnvironmentId,
			ConfigType:        appLevelConfig.ConfigType,
			RepoUrl:           appLevelConfig.RepoUrl,
			ChartLocation:     appLevelConfig.ChartLocation,
			Active:            true,
			DeploymentAppType: deploymentAppType,
		}
		configDbObj.AuditLog.CreateAuditLog(bean3.SYSTEM_USER_ID)
		configDBObjects = append(configDBObjects, configDbObj)
	}

	if len(configDBObjects) > 0 {
		configDBObjects, err = impl.deploymentConfigRepository.SaveAll(nil, configDBObjects)
		if err != nil {
			impl.logger.Errorw("error in saving deployment config in DB", "notFoundDeploymentConfigMap", notFoundDeploymentConfigMap, "err", err)
			return nil, err
		}
	}

	return configDBObjects, nil
}

func (impl *DeploymentConfigServiceImpl) GetHelmAppConfigInBulk(appIds []int, configSelector []*bean.DeploymentConfigSelector) ([]*bean.DeploymentConfig, error) {

	allDeploymentConfigs := make([]*deploymentConfig.DeploymentConfig, 0, len(appIds))

	appIdToEnvIdsMap := GetAppIdToEnvIsMapping(
		configSelector,
		func(config *bean.DeploymentConfigSelector) int { return config.AppId },
		func(config *bean.DeploymentConfigSelector) int { return config.EnvironmentId })

	helmDeploymentConfig, err := impl.deploymentConfigRepository.GetAppAndEnvLevelConfigsInBulk(appIdToEnvIdsMap)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching deployment config by by appId and envId", "envIdToAppIdMapping", appIdToEnvIdsMap, "err", err)
		return nil, err
	}

	allDeploymentConfigs = append(allDeploymentConfigs, helmDeploymentConfig...)

	if len(helmDeploymentConfig) < len(configSelector) {

		notFoundDeploymentConfigMap := make(map[bean.UniqueDeploymentConfigIdentifier]bool)
		presentDeploymentConfigMap := make(map[bean.UniqueDeploymentConfigIdentifier]bool)

		for _, c := range helmDeploymentConfig {
			presentDeploymentConfigMap[bean.GetConfigUniqueIdentifier(c.AppId, c.EnvironmentId)] = true
		}

		for _, c := range configSelector {
			key := bean.GetConfigUniqueIdentifier(c.AppId, c.EnvironmentId)
			if _, ok := presentDeploymentConfigMap[key]; !ok {
				notFoundDeploymentConfigMap[key] = true
			}
		}

		migratedConfigs, err := impl.migrateDeploymentConfigInBulkForHelmApps(appIds, notFoundDeploymentConfigMap)
		if err != nil {
			impl.logger.Errorw("error in migrating helm apps config data in bult to deployment config", "notFoundDeploymentConfigMap", notFoundDeploymentConfigMap, "err", err)
			return nil, err
		}
		allDeploymentConfigs = append(allDeploymentConfigs, migratedConfigs...)
	}

	deploymentConfigsResult := make([]*bean.DeploymentConfig, 0, len(appIds))
	for _, c := range allDeploymentConfigs {
		deploymentConfigsResult = append(deploymentConfigsResult, ConvertDeploymentConfigDbObjToDTO(c))
	}
	return deploymentConfigsResult, nil
}

func (impl *DeploymentConfigServiceImpl) migrateDeploymentConfigInBulkForHelmApps(appIds []int, notFoundDeploymentConfigMap map[bean.UniqueDeploymentConfigIdentifier]bool) ([]*deploymentConfig.DeploymentConfig, error) {
	installedApps, err := impl.installedAppRepository.FindInstalledAppByIds(appIds)
	if err != nil {
		impl.logger.Errorw("error in getting installed app by appId", "appIds", appIds, "err", err)
		return nil, err
	}

	appIdToInstalledAppMapping := make(map[int]*repository.InstalledApps)

	allRepoURLS := make([]string, 0, len(installedApps))

	for _, ia := range installedApps {
		appIdToInstalledAppMapping[ia.AppId] = ia
		allRepoURLS = append(allRepoURLS, ia.GitOpsRepoUrl)
	}

	configDBObjects := make([]*deploymentConfig.DeploymentConfig, 0, len(notFoundDeploymentConfigMap))

	for appEnvUniqueKey, _ := range notFoundDeploymentConfigMap {

		appId, envId := appEnvUniqueKey.GetAppAndEnvId()
		installedApp := appIdToInstalledAppMapping[appId]
		helmDeploymentConfig := &deploymentConfig.DeploymentConfig{
			AppId:             appId,
			EnvironmentId:     envId,
			DeploymentAppType: installedApp.DeploymentAppType,
			ConfigType:        GetDeploymentConfigType(installedApp.IsCustomRepository),
			RepoUrl:           installedApp.GitOpsRepoUrl,
			RepoName:          installedApp.GitOpsRepoName,
			Active:            true,
		}
		helmDeploymentConfig.AuditLog.CreateAuditLog(bean3.SYSTEM_USER_ID)
		configDBObjects = append(configDBObjects, helmDeploymentConfig)
	}

	configDBObjects, err = impl.deploymentConfigRepository.SaveAll(nil, configDBObjects)
	if err != nil {
		impl.logger.Errorw("error in saving deployment config in DB", "notFoundDeploymentConfigMap", notFoundDeploymentConfigMap, "err", err)
		return nil, err
	}
	return configDBObjects, nil
}
