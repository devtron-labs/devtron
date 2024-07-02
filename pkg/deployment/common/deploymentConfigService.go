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
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeploymentConfigService interface {
	CreateOrUpdateConfig(tx *pg.Tx, config *bean.DeploymentConfig, userId int32) (*bean.DeploymentConfig, error)
	GetDeploymentConfig(appId, envId int) (*bean.DeploymentConfig, error)
	GetDeploymentConfigForHelmApp(appId, envId int) (*bean.DeploymentConfig, error)
	GetDeploymentConfigInBulk(configSelector []*bean.DeploymentConfigSelector) ([]*bean.DeploymentConfig, error)
}

type DeploymentConfigServiceImpl struct {
	deploymentConfigRepository deploymentConfig.Repository
	logger                     *zap.SugaredLogger
	chartRepository            chartRepoRepository.ChartRepository
	pipelineRepository         pipelineConfig.PipelineRepository
	gitOpsConfigReadService    config.GitOpsConfigReadService
	appRepository              appRepository.AppRepository
	installedAppRepository     repository.InstalledAppRepository
}

func NewDeploymentConfigServiceImpl(
	deploymentConfigRepository deploymentConfig.Repository,
	logger *zap.SugaredLogger,
	chartRepository chartRepoRepository.ChartRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	appRepository appRepository.AppRepository,
	installedAppRepository repository.InstalledAppRepository,
) *DeploymentConfigServiceImpl {
	return &DeploymentConfigServiceImpl{
		deploymentConfigRepository: deploymentConfigRepository,
		logger:                     logger,
		chartRepository:            chartRepository,
		pipelineRepository:         pipelineRepository,
		gitOpsConfigReadService:    gitOpsConfigReadService,
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

func (impl *DeploymentConfigServiceImpl) GetDeploymentConfig(appId, envId int) (*bean.DeploymentConfig, error) {

	appLevelConfigDbObj, err := impl.deploymentConfigRepository.GetAppLevelConfig(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getiting deployment config db object by appId", "appId", appId, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		appLevelConfigDbObj, err = impl.migrateAppLevelDataTODeploymentConfig(appId)
		if err != nil {
			impl.logger.Errorw("error in migrating app level config to deployment config", "appId", appId, "err", err)
			return nil, err
		}
	}

	if envId > 0 {
		appAndEnvLevelConfig, err := impl.deploymentConfigRepository.GetByAppIdAndEnvId(appId, envId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getiting deployment config db object by appId and envId", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		if err == pg.ErrNoRows {
			appAndEnvLevelConfig, err = impl.migrateAppAndEnvLevelDataToDeploymentConfig(appLevelConfigDbObj, appId, envId)
			if err != nil {
				impl.logger.Errorw("error in migrating app level config to deployment config", "appId", appId, "err", err)
				return nil, err
			}
		}
		return ConvertDeploymentConfigDbObjToDTO(appAndEnvLevelConfig), nil
	}

	return ConvertDeploymentConfigDbObjToDTO(appLevelConfigDbObj), nil
}

func (impl *DeploymentConfigServiceImpl) migrateAppLevelDataTODeploymentConfig(appId int) (*deploymentConfig.DeploymentConfig, error) {
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
	ConfigDbObj.AuditLog.CreateAuditLog(1)
	ConfigDbObj, err = impl.deploymentConfigRepository.Save(nil, ConfigDbObj)
	if err != nil {
		impl.logger.Errorw("error in saving deployment config in DB", "appId", appId, "err", err)
		return nil, err
	}
	return ConfigDbObj, nil
}

func (impl *DeploymentConfigServiceImpl) migrateAppAndEnvLevelDataToDeploymentConfig(appLevelConfig *deploymentConfig.DeploymentConfig, appId int, envId int) (*deploymentConfig.DeploymentConfig, error) {

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

	switch configDbObj.DeploymentAppType {
	//TODO: handling for other deployment app type in future
	case bean2.ArgoCd:
		configDbObj.CredentialType = bean.GitOps.String()
		gitOpsConfig, err := impl.gitOpsConfigReadService.GetGitOpsProviderByRepoURL(configDbObj.RepoUrl)
		if err != nil {
			impl.logger.Infow("error in fetching gitOps config by repoUrl, skipping migration to deployment config", "repoURL", configDbObj.RepoUrl)
			return configDbObj, err
		}
		configDbObj.CredentialIdInt = gitOpsConfig.Id
	}

	configDbObj.AuditLog.CreateAuditLog(bean3.SYSTEM_USER_ID)
	configDbObj, err = impl.deploymentConfigRepository.Save(nil, configDbObj)
	if err != nil {
		impl.logger.Errorw("error in saving deployment config in DB", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	return configDbObj, nil
}

func (impl *DeploymentConfigServiceImpl) GetConfigDBObj(appId, envId int) (*deploymentConfig.DeploymentConfig, error) {
	var configDbObj *deploymentConfig.DeploymentConfig
	var err error
	if envId == 0 {
		configDbObj, err = impl.deploymentConfigRepository.GetAppLevelConfig(appId)
		if err != nil {
			impl.logger.Errorw("error in getiting deployment config db object by appId", "appId", configDbObj.AppId, "err", err)
			return nil, err
		}
	} else {
		configDbObj, err = impl.deploymentConfigRepository.GetByAppIdAndEnvId(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in getiting deployment config db object by appId and envId", "appId", configDbObj.AppId, "envId", configDbObj.EnvironmentId, "err", err)
			return nil, err
		}
	}
	return configDbObj, nil
}

func (impl *DeploymentConfigServiceImpl) GetDeploymentConfigForHelmApp(appId, envId int) (*bean.DeploymentConfig, error) {

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

	switch helmDeploymentConfig.DeploymentAppType {
	case bean2.ArgoCd:
		gitOpsConfig, err := impl.gitOpsConfigReadService.GetGitOpsProviderByRepoURL(installedApp.GitOpsRepoUrl)
		if err != nil {
			impl.logger.Infow("error in fetching gitOps config by repoUrl, skipping migration to deployment config", "repoURL", installedApp.GitOpsRepoUrl)
			return nil, err
		}
		helmDeploymentConfig.ConfigType = bean.SYSTEM_GENERATED.String()
		helmDeploymentConfig.CredentialIdInt = gitOpsConfig.Id
	}
	helmDeploymentConfig.CreateAuditLog(bean3.SYSTEM_USER_ID)
	helmDeploymentConfig, err = impl.deploymentConfigRepository.Save(nil, helmDeploymentConfig)
	if err != nil {
		impl.logger.Errorw("error in saving deployment config for helm app", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	return helmDeploymentConfig, nil
}

func (impl *DeploymentConfigServiceImpl) GetDeploymentConfigInBulk(configSelector []*bean.DeploymentConfigSelector) ([]*bean.DeploymentConfig, error) {

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

	allConfigs := make([]*bean.DeploymentConfig, len(devtronAppConfigs)+len(helmAppConfigs))
	allConfigs = append(allConfigs, devtronAppConfigs...)
	allConfigs = append(allConfigs, helmAppConfigs...)

	return allConfigs, nil
}

func (impl *DeploymentConfigServiceImpl) GetDevtronAppConfigInBulk(appIds []int, configSelector []*bean.DeploymentConfigSelector) ([]*bean.DeploymentConfig, error) {

	AppLevelDeploymentConfigs, err := impl.deploymentConfigRepository.GetAppLevelConfigByAppIds(appIds)
	if err != nil {
		impl.logger.Errorw("error in fetching deployment configs by appIds", "appIds", appIds, "err", err)
		return nil, err
	}

	AllAppLevelConfigs := make([]*deploymentConfig.DeploymentConfig, 0, len(appIds))
	AllAppLevelConfigs = append(AllAppLevelConfigs, AppLevelDeploymentConfigs...)

	if len(AppLevelDeploymentConfigs) < len(appIds) {
		presentAppIds := make(map[int]bool)
		for _, c := range AppLevelDeploymentConfigs {
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
		AllAppLevelConfigs = append(AllAppLevelConfigs, migratedAppLevelDeploymentConfigs...)
	}

	appIdToAppLevelConfigMapping := make(map[int]*deploymentConfig.DeploymentConfig, len(appIds))
	for _, appLevelConfig := range AllAppLevelConfigs {
		appIdToAppLevelConfigMapping[appLevelConfig.AppId] = appLevelConfig
	}

	AllEnvLevelConfigs := make([]*deploymentConfig.DeploymentConfig, 0, len(configSelector))

	if len(configSelector) > 0 {

		appIdToEnvIdsMap := GetAppIdToEnvIsMappingFromConfigSelectors(configSelector)

		envLevelConfig, err := impl.deploymentConfigRepository.GetAppAndEnvLevelConfigsInBulk(appIdToEnvIdsMap)
		if err != nil {
			impl.logger.Errorw("error in getting and and env level config in bulk", "appIdToEnvIdsMap", appIdToEnvIdsMap, "err", err)
			return nil, err
		}
		AllEnvLevelConfigs = append(AllEnvLevelConfigs, envLevelConfig...)
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

			migratedEnvLevelDeploymentConfigs, err := impl.migrateEnvLevelDataTODeploymentConfigInBulk(notFoundDeploymentConfigMap, appIdToAppLevelConfigMapping)
			if err != nil {
				impl.logger.Errorw("error in migrating env level configs", "notFoundDeploymentConfigMap", notFoundDeploymentConfigMap, "err", err)
				return nil, err
			}
			AllEnvLevelConfigs = append(AllEnvLevelConfigs, migratedEnvLevelDeploymentConfigs...)
		}
	}

	allConfigs := make([]*bean.DeploymentConfig, 0)
	for _, c := range AllAppLevelConfigs {
		allConfigs = append(allConfigs, ConvertDeploymentConfigDbObjToDTO(c))
	}
	for _, c := range AllEnvLevelConfigs {
		allConfigs = append(allConfigs, ConvertDeploymentConfigDbObjToDTO(c))
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

	repoUrlToConfigMapping, err := impl.gitOpsConfigReadService.GetGitOpsProviderMapByRepoURL(allRepoUrls)
	if err != nil {
		impl.logger.Errorw("error in fetching repoUrl to config mapping", "err", err)
		return nil, err
	}

	configDBObjects := make([]*deploymentConfig.DeploymentConfig, 0, len(appIds))
	for _, c := range charts {
		dbObj := &deploymentConfig.DeploymentConfig{
			ConfigType:      GetDeploymentConfigType(c.IsCustomGitRepository),
			AppId:           c.AppId,
			Active:          true,
			RepoUrl:         c.GitRepoUrl,
			ChartLocation:   c.ChartLocation,
			CredentialType:  bean.GitOps.String(),
			CredentialIdInt: repoUrlToConfigMapping[c.GitRepoUrl].Id,
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

func (impl *DeploymentConfigServiceImpl) migrateEnvLevelDataTODeploymentConfigInBulk(notFoundDeploymentConfigMap map[bean.UniqueDeploymentConfigIdentifier]bool, appIdToAppLevelConfigMapping map[int]*deploymentConfig.DeploymentConfig) ([]*deploymentConfig.DeploymentConfig, error) {

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
		switch deploymentAppType {
		case bean2.ArgoCd:
			configDbObj.CredentialType = appLevelConfig.CredentialType
			configDbObj.CredentialIdInt = appLevelConfig.CredentialIdInt
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

	appIdToEnvIdsMap := GetAppIdToEnvIsMappingFromConfigSelectors(configSelector)

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

	repoUrlToConfigMapping, err := impl.gitOpsConfigReadService.GetGitOpsProviderMapByRepoURL(allRepoURLS)
	if err != nil {
		impl.logger.Errorw("error in fetching repoUrl to config mapping", "err", err)
		return nil, err
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
		switch installedApp.DeploymentAppType {
		case bean2.ArgoCd:
			helmDeploymentConfig.CredentialType = bean.GitOps.String()
			helmDeploymentConfig.CredentialIdInt = repoUrlToConfigMapping[installedApp.GitOpsRepoUrl].Id
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
