package common

import (
	"github.com/devtron-labs/devtron/api/bean/gitOps"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeploymentConfigService interface {
	CreateOrUpdateConfig(tx *pg.Tx, config *bean.DeploymentConfig, userId int32) (*bean.DeploymentConfig, error)
	IsDeploymentConfigUsed() bool
	GetConfigForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error)
	GetAndMigrateConfigIfAbsentForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error)
	GetConfigForHelmApps(appId, envId int) (*bean.DeploymentConfig, error)
	GetConfigEvenIfInactive(appId, envId int) (*bean.DeploymentConfig, error)
	GetAndMigrateConfigIfAbsentForHelmApp(appId, envId int) (*bean.DeploymentConfig, error)
	GetAppLevelConfigForDevtronApp(appId int) (*bean.DeploymentConfig, error)
	UpdateRepoUrlForAppAndEnvId(repoURL string, appId, envId int) error
}

type DeploymentConfigServiceImpl struct {
	deploymentConfigRepository  deploymentConfig.Repository
	logger                      *zap.SugaredLogger
	chartRepository             chartRepoRepository.ChartRepository
	pipelineRepository          pipelineConfig.PipelineRepository
	appRepository               appRepository.AppRepository
	installedAppRepository      repository.InstalledAppRepository
	deploymentServiceTypeConfig *util.DeploymentServiceTypeConfig
}

func NewDeploymentConfigServiceImpl(
	deploymentConfigRepository deploymentConfig.Repository,
	logger *zap.SugaredLogger,
	chartRepository chartRepoRepository.ChartRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	appRepository appRepository.AppRepository,
	installedAppRepository repository.InstalledAppRepository,
	envVariables *util.EnvironmentVariables,
) *DeploymentConfigServiceImpl {
	return &DeploymentConfigServiceImpl{
		deploymentConfigRepository:  deploymentConfigRepository,
		logger:                      logger,
		chartRepository:             chartRepository,
		pipelineRepository:          pipelineRepository,
		appRepository:               appRepository,
		installedAppRepository:      installedAppRepository,
		deploymentServiceTypeConfig: envVariables.DeploymentServiceTypeConfig,
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
		newDBObj.AuditLog = configDbObj.AuditLog
		newDBObj.AuditLog.UpdateAuditLog(userId)
		newDBObj, err = impl.deploymentConfigRepository.Update(tx, newDBObj)
		if err != nil {
			impl.logger.Errorw("error in updating deploymentConfig", "appId", config.AppId, "envId", config.EnvironmentId, "err", err)
			return nil, err
		}
	}

	return ConvertDeploymentConfigDbObjToDTO(newDBObj), nil
}

func (impl *DeploymentConfigServiceImpl) IsDeploymentConfigUsed() bool {
	return impl.deploymentServiceTypeConfig.UseDeploymentConfigData
}

func (impl *DeploymentConfigServiceImpl) GetConfigForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error) {

	if !impl.deploymentServiceTypeConfig.UseDeploymentConfigData {
		configFromOldData, err := impl.parseFromOldTablesForDevtronApps(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in parsing config from charts and pipeline repository", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		return configFromOldData, nil
	}

	// if USE_DEPLOYMENT_CONFIG_DATA is true, first try to fetch data from deployment_config table and if not found use charts and pipeline respectively

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
		} else if gitOps.IsGitOpsRepoNotConfigured(appAndEnvLevelConfig.RepoUrl) && gitOps.IsGitOpsRepoConfigured(appLevelConfigDbObj.RepoUrl) {
			// if url is present at app level and not at env level then copy app level url to env level config
			appAndEnvLevelConfig.RepoUrl = appLevelConfigDbObj.RepoUrl
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
	var envLevelConfig *bean.DeploymentConfig
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
		} else if gitOps.IsGitOpsRepoNotConfigured(appAndEnvLevelConfig.RepoUrl) && gitOps.IsGitOpsRepoConfigured(appLevelConfigDbObj.RepoUrl) {
			// if url is present at app level and not at env level then copy app level url to env level config
			// will happen when custom gitOps is enabled and app is cloned. In this case when user configure app level gitOps , env level gitOps will not be updated
			appAndEnvLevelConfig.RepoUrl = appLevelConfigDbObj.RepoUrl
			appAndEnvLevelConfig.AuditLog.UpdateAuditLog(1)
			appAndEnvLevelConfig, err = impl.deploymentConfigRepository.Update(nil, appAndEnvLevelConfig)
			if err != nil {
				impl.logger.Errorw("error in updating deploymentConfig", "appId", appAndEnvLevelConfig.AppId, "envId", appAndEnvLevelConfig.EnvironmentId, "err", err)
				return nil, err
			}
		}
		envLevelConfig = ConvertDeploymentConfigDbObjToDTO(appAndEnvLevelConfig)
	}

	if !impl.deploymentServiceTypeConfig.UseDeploymentConfigData {
		configFromOldData, err := impl.parseFromOldTablesForDevtronApps(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in parsing config from charts and pipeline repository", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		return configFromOldData, nil
	}

	if envId > 0 {
		return envLevelConfig, nil
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
		ConfigType: GetDeploymentConfigType(chart.IsCustomGitRepository),
		AppId:      appId,
		Active:     true,
		RepoUrl:    chart.GitRepoUrl,
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

	if !impl.deploymentServiceTypeConfig.UseDeploymentConfigData {
		configFromOldData, err := impl.parseConfigForHelmApps(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in parsing config from charts and pipeline repository", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		return ConvertDeploymentConfigDbObjToDTO(configFromOldData), nil
	}

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

	if !impl.deploymentServiceTypeConfig.UseDeploymentConfigData {
		configFromOldData, err := impl.parseConfigForHelmApps(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in parsing config from charts and pipeline repository", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		return ConvertDeploymentConfigDbObjToDTO(configFromOldData), nil
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
	if installedApp.EnvironmentId != envId {
		return nil, pg.ErrNoRows
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

func (impl *DeploymentConfigServiceImpl) parseFromOldTablesForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error) {
	appLevelConfig, err := impl.parseAppLevelConfigForDevtronApps(appId)
	if err != nil {
		impl.logger.Errorw("error in parsing charts data to deployment config", "appId", appId, "err", err)
		return nil, err
	}
	if envId > 0 {
		appAndEnvLevelConfig, err := impl.parseEnvLevelConfigForDevtronApps(appLevelConfig, appId, envId)
		if err != nil {
			impl.logger.Errorw("error in parsing env level config to deployment config", "appId", appId, "err", err)
			return nil, err
		}
		return ConvertDeploymentConfigDbObjToDTO(appAndEnvLevelConfig), nil
	}
	return ConvertDeploymentConfigDbObjToDTO(appLevelConfig), nil
}

func (impl *DeploymentConfigServiceImpl) GetAppLevelConfigForDevtronApp(appId int) (*bean.DeploymentConfig, error) {
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
	return ConvertDeploymentConfigDbObjToDTO(appLevelConfigDbObj), nil
}

func (impl *DeploymentConfigServiceImpl) UpdateRepoUrlForAppAndEnvId(repoURL string, appId, envId int) error {
	err := impl.deploymentConfigRepository.UpdateRepoUrlByAppIdAndEnvId(repoURL, appId, envId)
	if err != nil {
		impl.logger.Errorw("error in updating repoUrl by app-id and env-id", "appId", appId, "envId", envId, "err", err)
		return err
	}
	return nil
}
