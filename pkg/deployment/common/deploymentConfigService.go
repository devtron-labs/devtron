package common

import (
	"errors"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/client/argocdServer"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	installedAppReader "github.com/devtron-labs/devtron/pkg/appStore/installedApp/read"
	bean3 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/cluster/bean"
	bean4 "github.com/devtron-labs/devtron/pkg/cluster/environment/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/read"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/helper"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeploymentConfigService interface {
	CreateOrUpdateConfig(tx *pg.Tx, config *bean.DeploymentConfig, userId int32) (*bean.DeploymentConfig, error)
	GetConfigForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error)
	GetAndMigrateConfigIfAbsentForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error)
	GetConfigForHelmApps(appId, envId int) (*bean.DeploymentConfig, error)
	GetConfigEvenIfInactive(appId, envId int) (*bean.DeploymentConfig, error)
	GetAndMigrateConfigIfAbsentForHelmApp(appId, envId int) (*bean.DeploymentConfig, error)
	UpdateRepoUrlForAppAndEnvId(repoURL string, appId, envId int) error
	GetAllConfigsWithCustomGitOpsURL() ([]*bean.DeploymentConfig, error)
}

type DeploymentConfigServiceImpl struct {
	deploymentConfigRepository  deploymentConfig.Repository
	logger                      *zap.SugaredLogger
	chartRepository             chartRepoRepository.ChartRepository
	pipelineRepository          pipelineConfig.PipelineRepository
	appRepository               appRepository.AppRepository
	installedAppReadService     installedAppReader.InstalledAppReadServiceEA
	deploymentServiceTypeConfig *util.DeploymentServiceTypeConfig
	envConfigOverrideService    read.EnvConfigOverrideService
	environmentRepository       repository.EnvironmentRepository
}

func NewDeploymentConfigServiceImpl(
	deploymentConfigRepository deploymentConfig.Repository,
	logger *zap.SugaredLogger,
	chartRepository chartRepoRepository.ChartRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	appRepository appRepository.AppRepository,
	installedAppReadService installedAppReader.InstalledAppReadServiceEA,
	envVariables *util.EnvironmentVariables,
	envConfigOverrideService read.EnvConfigOverrideService,
	environmentRepository repository.EnvironmentRepository,
) *DeploymentConfigServiceImpl {

	return &DeploymentConfigServiceImpl{
		deploymentConfigRepository:  deploymentConfigRepository,
		logger:                      logger,
		chartRepository:             chartRepository,
		pipelineRepository:          pipelineRepository,
		appRepository:               appRepository,
		installedAppReadService:     installedAppReadService,
		deploymentServiceTypeConfig: envVariables.DeploymentServiceTypeConfig,
		envConfigOverrideService:    envConfigOverrideService,
		environmentRepository:       environmentRepository,
	}
}

func (impl *DeploymentConfigServiceImpl) CreateOrUpdateConfig(tx *pg.Tx, config *bean.DeploymentConfig, userId int32) (*bean.DeploymentConfig, error) {

	configDbObj, err := impl.GetConfigDBObj(config.AppId, config.EnvironmentId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching deployment config from DB by appId and envId",
			"appId", config.AppId, "envId", config.EnvironmentId, "err", err)
	}

	newDBObj, err := adapter.ConvertDeploymentConfigDTOToDbObj(config)
	if err != nil {
		impl.logger.Errorw("error in converting deployment config DTO to db object", "appId", config.AppId, "envId", config.EnvironmentId)
		return nil, err
	}

	if configDbObj == nil || configDbObj.Id == 0 {
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
	newObj, err := adapter.ConvertDeploymentConfigDbObjToDTO(newDBObj)
	if err != nil {
		impl.logger.Errorw("error in converting deployment config DTO to db object", "appId", config.AppId, "envId", config.EnvironmentId)
		return nil, err
	}
	return newObj, nil
}

func (impl *DeploymentConfigServiceImpl) GetConfigForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error) {

	appLevelConfig, err := impl.getAppLevelConfigForDevtronApps(appId, envId, false)
	if err != nil {
		impl.logger.Errorw("error in getting app level Config for devtron apps", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	if envId > 0 {
		// if envId>0 then only env level config will be returned,
		//for getting app level config envId should be zero
		appAndEnvLevelConfig, err := impl.getEnvLevelDataForDevtronApps(appId, envId, appLevelConfig, false)
		if err != nil {
			impl.logger.Errorw("error in getting env level data for devtron apps", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}

		appAndEnvLevelConfig, err = impl.ConfigureEnvURLByAppURLIfNotConfigured(appAndEnvLevelConfig, appLevelConfig.GetRepoURL(), false)
		if err != nil {
			impl.logger.Errorw("error in configuring env level url with app url", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}

		return appAndEnvLevelConfig, nil
	}
	return appLevelConfig, nil
}

func (impl *DeploymentConfigServiceImpl) getAppLevelConfigForDevtronApps(appId int, envId int, migrateDataIfAbsent bool) (*bean.DeploymentConfig, error) {

	var (
		appLevelConfig    *bean.DeploymentConfig
		isMigrationNeeded bool
	)
	appLevelConfigDbObj, err := impl.deploymentConfigRepository.GetAppLevelConfigForDevtronApps(appId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in getting deployment config db object by appId", "appId", appId, "err", err)
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		isMigrationNeeded = true
		appLevelConfig, err = impl.parseAppLevelMigrationDataForDevtronApps(appId)
		if err != nil {
			impl.logger.Errorw("error in migrating app level config to deployment config", "appId", appId, "err", err)
			return nil, err
		}
	} else {
		appLevelConfig, err = adapter.ConvertDeploymentConfigDbObjToDTO(appLevelConfigDbObj)
		if err != nil {
			impl.logger.Errorw("error in converting deployment config db object", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		if appLevelConfig.ReleaseConfiguration == nil || len(appLevelConfig.ReleaseConfiguration.Version) == 0 {
			isMigrationNeeded = true
			releaseConfig, err := impl.parseAppLevelReleaseConfigForDevtronApp(appId, appLevelConfig)
			if err != nil {
				impl.logger.Errorw("error in parsing release configuration for app", "appId", appId, "err", err)
				return nil, err
			}
			appLevelConfig.ReleaseConfiguration = releaseConfig
		}
	}
	if migrateDataIfAbsent && isMigrationNeeded {
		_, err := impl.CreateOrUpdateConfig(nil, appLevelConfig, bean3.SYSTEM_USER_ID)
		if err != nil {
			impl.logger.Errorw("error in migrating app level config to deployment config", "appId", appId, "err", err)
			return nil, err
		}
	}
	return appLevelConfig, nil

}

func (impl *DeploymentConfigServiceImpl) parseAppLevelReleaseConfigForDevtronApp(appId int, appLevelConfig *bean.DeploymentConfig) (*bean.ReleaseConfiguration, error) {

	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		return nil, err
	}

	repoURL := chart.GitRepoUrl
	if len(appLevelConfig.RepoURL) > 0 {
		repoURL = appLevelConfig.RepoURL
	}

	releaseConfig := newAppLevelReleaseConfigFromChart(repoURL, chart.ChartLocation)
	return releaseConfig, nil
}

func (impl *DeploymentConfigServiceImpl) getEnvLevelDataForDevtronApps(appId, envId int, appLevelConfig *bean.DeploymentConfig, migrateDataIfAbsent bool) (*bean.DeploymentConfig, error) {
	var (
		appAndEnvLevelConfig *bean.DeploymentConfig
		isMigrationNeeded    bool
	)
	appAndEnvLevelConfigDBObj, err := impl.deploymentConfigRepository.GetByAppIdAndEnvId(appId, envId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in getting deployment config db object by appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		// case: deployment config data is not yet migrated
		appAndEnvLevelConfig, err = impl.parseEnvLevelMigrationDataForDevtronApps(appLevelConfig, appId, envId)
		if err != nil {
			impl.logger.Errorw("error in parsing env level config to deployment config", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		isMigrationNeeded = true

	} else {
		// case: deployment config is migrated but release config is absent
		appAndEnvLevelConfig, err = adapter.ConvertDeploymentConfigDbObjToDTO(appAndEnvLevelConfigDBObj)
		if err != nil {
			impl.logger.Errorw("error in converting deployment config db object", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}

		if appAndEnvLevelConfig.ReleaseConfiguration == nil || len(appAndEnvLevelConfig.ReleaseConfiguration.Version) == 0 {
			isMigrationNeeded = true
			releaseConfig, err := impl.parseEnvLevelReleaseConfigForDevtronApp(appAndEnvLevelConfig, appId, envId)
			if err != nil {
				impl.logger.Errorw("error in parsing env level release config", "appId", appId, "envId", envId, "err", err)
				return nil, err
			}
			appAndEnvLevelConfig.ReleaseConfiguration = releaseConfig
		}
	}
	if migrateDataIfAbsent && isMigrationNeeded {
		_, err := impl.CreateOrUpdateConfig(nil, appAndEnvLevelConfig, bean3.SYSTEM_USER_ID)
		if err != nil {
			impl.logger.Errorw("error in migrating env level config to deployment config", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
	}
	return appAndEnvLevelConfig, nil
}

func (impl *DeploymentConfigServiceImpl) ConfigureEnvURLByAppURLIfNotConfigured(appAndEnvLevelConfig *bean.DeploymentConfig, appLevelURL string, migrateDataIfAbsent bool) (*bean.DeploymentConfig, error) {

	/*
		if custom gitOps is configured in repo
		and app is cloned then cloned pipelines repo URL=NOT_CONFIGURED .
		In this case User manually configures repoURL. The configured repo_url is saved in app level config but is absent
		in env level config.
	*/

	if gitOps.IsGitOpsRepoNotConfigured(appAndEnvLevelConfig.GetRepoURL()) &&
		gitOps.IsGitOpsRepoConfigured(appLevelURL) {
		// if url is present at app level and not at env level then copy app level url to env level config
		appAndEnvLevelConfig.SetRepoURL(appAndEnvLevelConfig.GetRepoURL())
	}

	if migrateDataIfAbsent {
		_, err := impl.CreateOrUpdateConfig(nil, appAndEnvLevelConfig, bean3.SYSTEM_USER_ID)
		if err != nil {
			return nil, err
		}
	}

	return appAndEnvLevelConfig, nil
}

func (impl *DeploymentConfigServiceImpl) GetAndMigrateConfigIfAbsentForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error) {

	migrateDeploymentConfigData := impl.deploymentServiceTypeConfig.MigrateDeploymentConfigData

	appLevelConfig, err := impl.getAppLevelConfigForDevtronApps(appId, envId, migrateDeploymentConfigData)
	if err != nil {
		impl.logger.Errorw("error in getting app level Config for devtron apps", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	var envLevelConfig *bean.DeploymentConfig
	if envId > 0 {
		envLevelConfig, err = impl.getEnvLevelDataForDevtronApps(appId, envId, appLevelConfig, migrateDeploymentConfigData)
		if err != nil {
			impl.logger.Errorw("error in getting env level data for devtron apps", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		envLevelConfig, err = impl.ConfigureEnvURLByAppURLIfNotConfigured(envLevelConfig, appLevelConfig.GetRepoURL(), migrateDeploymentConfigData)
		if err != nil {
			impl.logger.Errorw("error in getting env level data for devtron apps", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		return envLevelConfig, nil
	}

	return envLevelConfig, nil
}

func newAppLevelReleaseConfigFromChart(gitRepoURL, chartLocation string) *bean.ReleaseConfiguration {
	return &bean.ReleaseConfiguration{
		Version: bean.Version,
		ArgoCDSpec: bean.ArgoCDSpec{
			Spec: bean.ApplicationSpec{
				Source: &bean.ApplicationSource{
					RepoURL: gitRepoURL,
					Path:    chartLocation,
				},
			},
		}}
}

func (impl *DeploymentConfigServiceImpl) parseAppLevelMigrationDataForDevtronApps(appId int) (*bean.DeploymentConfig, error) {
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		return nil, err
	}
	releaseConfig := newAppLevelReleaseConfigFromChart(chart.GitRepoUrl, chart.ChartLocation)
	config := &bean.DeploymentConfig{
		AppId:                appId,
		ConfigType:           GetDeploymentConfigType(chart.IsCustomGitRepository),
		Active:               true,
		ReleaseConfiguration: releaseConfig,
	}
	return config, nil
}

func (impl *DeploymentConfigServiceImpl) parseEnvLevelMigrationDataForDevtronApps(appLevelConfig *bean.DeploymentConfig, appId, envId int) (*bean.DeploymentConfig, error) {

	/*
		We can safely assume that no link argoCD pipeline is created if migration is happening
		migration case, default values for below fields will be =>
		1) repoUrl => same as app level url
		2) chartLocation => we should fetch active envConfigOverride and use chart path from that
		3) valuesFile => _<environmentId>-values.yaml
		4) branch => master
		5) releaseMode => create
		6) Default ClusterId for application object => 1
		7) Default Namespace for application object => devtroncd
	*/

	config := &bean.DeploymentConfig{
		AppId:         appId,
		EnvironmentId: envId,
		ConfigType:    appLevelConfig.ConfigType,
		ReleaseMode:   util2.PIPELINE_RELEASE_MODE_CREATE,
		Active:        true,
	}

	deploymentAppType, err := impl.pipelineRepository.FindDeploymentAppTypeByAppIdAndEnvId(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment app type by appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	config.DeploymentAppType = deploymentAppType

	releaseConfig, err := impl.parseEnvLevelReleaseConfigForDevtronApp(config, appId, envId)
	if err != nil {
		impl.logger.Errorw("error in parsing env level release config", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	config.ReleaseConfiguration = releaseConfig

	return config, nil
}

func (impl *DeploymentConfigServiceImpl) parseEnvLevelReleaseConfigForDevtronApp(config *bean.DeploymentConfig, appId int, envId int) (*bean.ReleaseConfiguration, error) {
	releaseConfig := &bean.ReleaseConfiguration{}
	if config.DeploymentAppType == util2.PIPELINE_DEPLOYMENT_TYPE_ACD {

		releaseConfig.Version = bean.Version

		envOverride, err := impl.envConfigOverrideService.ActiveEnvConfigOverride(appId, envId)
		if err != nil {
			return nil, err
		}

		var latestChart *chartRepoRepository.Chart
		if (envOverride.Id == 0) || (envOverride.Id > 0 && !envOverride.IsOverride) {
			latestChart, err = impl.chartRepository.FindLatestChartForAppByAppId(appId)
			if err != nil {
				return nil, err
			}
		} else {
			//if chart is overrides in env, it means it may have different version than app level.
			latestChart = envOverride.Chart
		}

		env, err := impl.environmentRepository.FindById(envId)
		if err != nil {
			impl.logger.Errorw("error in finding environment by id", "envId", envId, "err", err)
			return nil, err
		}

		gitRepoUrl := latestChart.GitRepoUrl
		if len(config.RepoURL) > 0 {
			gitRepoUrl = config.RepoURL
		}
		releaseConfig.ArgoCDSpec = bean.ArgoCDSpec{
			Metadata: bean.ApplicationMetadata{
				ClusterId: bean2.DefaultClusterId,
				Namespace: argocdServer.DevtronInstalationNs,
			},
			Spec: bean.ApplicationSpec{
				Source: &bean.ApplicationSource{
					RepoURL: gitRepoUrl,
					Path:    latestChart.ChartLocation,
					Helm: &bean.ApplicationSourceHelm{
						ValueFiles: []string{helper.GetValuesFileForEnv(env.Id)},
					},
					TargetRevision: util.GetDefaultTargetRevision(),
				},
				Destination: &bean.Destination{
					Namespace: env.Namespace,
					Server:    commonBean.DefaultClusterUrl,
				},
			},
		}
	}
	return releaseConfig, nil
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
	helmDeploymentConfig, err := impl.getConfigForHelmApps(appId, envId, false)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config for helm app", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	return helmDeploymentConfig, nil
}

func (impl *DeploymentConfigServiceImpl) getConfigForHelmApps(appId int, envId int, migrateIfAbsent bool) (*bean.DeploymentConfig, error) {
	var (
		helmDeploymentConfig *bean.DeploymentConfig
		isMigrationNeeded    bool
	)
	config, err := impl.deploymentConfigRepository.GetByAppIdAndEnvId(appId, envId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching deployment config by by appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		isMigrationNeeded = true
		helmDeploymentConfig, err = impl.parseDeploymentConfigForHelmApps(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in parsing helm deployment config", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
	} else {
		helmDeploymentConfig, err = adapter.ConvertDeploymentConfigDbObjToDTO(config)
		if err != nil {
			impl.logger.Errorw("error in converting helm deployment config dbObj to DTO", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		if helmDeploymentConfig.ReleaseConfiguration == nil || len(helmDeploymentConfig.ReleaseConfiguration.Version) == 0 {
			isMigrationNeeded = true
			releaseConfig, err := impl.parseReleaseConfigForHelmApps(appId, envId, helmDeploymentConfig)
			if err != nil {
				impl.logger.Errorw("error in parsing release config", "appId", appId, "envId", envId, "err", err)
				return nil, err
			}
			helmDeploymentConfig.ReleaseConfiguration = releaseConfig
		}
	}
	if migrateIfAbsent && isMigrationNeeded {
		_, err = impl.CreateOrUpdateConfig(nil, helmDeploymentConfig, bean3.SYSTEM_USER_ID)
		if err != nil {
			impl.logger.Errorw("error in creating helm deployment config ", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
	}
	return helmDeploymentConfig, err
}

func (impl *DeploymentConfigServiceImpl) GetConfigEvenIfInactive(appId, envId int) (*bean.DeploymentConfig, error) {
	dbConfig, err := impl.deploymentConfigRepository.GetByAppIdAndEnvIdEvenIfInactive(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config by appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	config, err := adapter.ConvertDeploymentConfigDbObjToDTO(dbConfig)
	if err != nil {
		impl.logger.Errorw("error in converting deployment config db obj to dto", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	return config, nil
}

func (impl *DeploymentConfigServiceImpl) GetAndMigrateConfigIfAbsentForHelmApp(appId, envId int) (*bean.DeploymentConfig, error) {
	migrateDataIfAbsent := impl.deploymentServiceTypeConfig.MigrateDeploymentConfigData
	helmDeploymentConfig, err := impl.getConfigForHelmApps(appId, envId, migrateDataIfAbsent)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config for helm app", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	return helmDeploymentConfig, nil
}

func (impl *DeploymentConfigServiceImpl) parseDeploymentConfigForHelmApps(appId int, envId int) (*bean.DeploymentConfig, error) {
	installedApp, err := impl.installedAppReadService.GetInstalledAppsByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in getting installed app by appId", "appId", appId, "err", err)
		return nil, err
	}
	if installedApp.EnvironmentId != envId {
		return nil, pg.ErrNoRows
	}
	helmDeploymentConfig := &bean.DeploymentConfig{
		AppId:             appId,
		EnvironmentId:     envId,
		DeploymentAppType: installedApp.DeploymentAppType,
		ConfigType:        GetDeploymentConfigType(installedApp.IsCustomRepository),
		RepoURL:           installedApp.GitOpsRepoUrl,
		Active:            true,
	}
	releaseConfig, err := impl.parseReleaseConfigForHelmApps(appId, envId, helmDeploymentConfig)
	if err != nil {
		return nil, err
	}
	helmDeploymentConfig.ReleaseConfiguration = releaseConfig
	return helmDeploymentConfig, nil
}

func (impl *DeploymentConfigServiceImpl) parseReleaseConfigForHelmApps(appId int, envId int, config *bean.DeploymentConfig) (*bean.ReleaseConfiguration, error) {
	releaseConfig := &bean.ReleaseConfiguration{}
	if config.DeploymentAppType == bean4.PIPELINE_DEPLOYMENT_TYPE_ACD {
		releaseConfig.Version = bean.Version
		app, err := impl.appRepository.FindById(appId)
		if err != nil {
			impl.logger.Errorw("error in getting app by id", "appId", appId, "err", err)
			return nil, err
		}
		env, err := impl.environmentRepository.FindById(envId)
		if err != nil {
			impl.logger.Errorw("error in getting installed app by environmentId", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}

		var gitRepoURL string
		if len(config.RepoURL) > 0 {
			gitRepoURL = config.RepoURL
		} else {
			installedApp, err := impl.installedAppReadService.GetInstalledAppsByAppId(appId)
			if err != nil {
				impl.logger.Errorw("error in getting installed app by appId", "appId", appId, "err", err)
				return nil, err
			}
			gitRepoURL = installedApp.GitOpsRepoUrl
		}

		releaseConfig = &bean.ReleaseConfiguration{
			Version: bean.Version,
			ArgoCDSpec: bean.ArgoCDSpec{
				Metadata: bean.ApplicationMetadata{
					ClusterId: bean2.DefaultClusterId,
					Namespace: argocdServer.DevtronInstalationNs,
				},
				Spec: bean.ApplicationSpec{
					Destination: &bean.Destination{
						Namespace: env.Namespace,
						Server:    commonBean.DefaultClusterUrl,
					},
					Source: &bean.ApplicationSource{
						RepoURL: gitRepoURL,
						Path:    util.BuildDeployedAppName(app.AppName, env.Name),
						Helm: &bean.ApplicationSourceHelm{
							ValueFiles: []string{"values.yaml"},
						},
						TargetRevision: util.GetDefaultTargetRevision(),
					},
				},
			},
		}
	}
	return releaseConfig, nil
}

func (impl *DeploymentConfigServiceImpl) UpdateRepoUrlForAppAndEnvId(repoURL string, appId, envId int) error {

	dbObj, err := impl.deploymentConfigRepository.GetByAppIdAndEnvId(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config by appId", "appId", appId, "envId", envId, "err", err)
		return err
	}

	config, err := adapter.ConvertDeploymentConfigDbObjToDTO(dbObj)
	if err != nil {
		impl.logger.Errorw("error in converting deployment config to DTO", "appId", appId, "envId", envId, "err", err)
		return err
	}

	config.SetRepoURL(repoURL)

	dbObj, err = impl.deploymentConfigRepository.Update(nil, dbObj)
	if err != nil {
		impl.logger.Errorw("error in updating deployment config", appId, "envId", envId, "err", err)
		return err
	}

	return nil
}

func (impl *DeploymentConfigServiceImpl) GetAllConfigsWithCustomGitOpsURL() ([]*bean.DeploymentConfig, error) {
	dbConfigs, err := impl.deploymentConfigRepository.GetAllConfigsWithCustomGitOpsURL()
	if err != nil {
		impl.logger.Errorw("error in getting all configs with custom gitops url", "err", err)
		return nil, err
	}
	var configs []*bean.DeploymentConfig
	for _, dbConfig := range dbConfigs {
		config, err := adapter.ConvertDeploymentConfigDbObjToDTO(dbConfig)
		if err != nil {
			impl.logger.Error("error in converting dbObj to dto", "err", err)
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}
