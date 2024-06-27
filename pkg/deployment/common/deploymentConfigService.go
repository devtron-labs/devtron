package common

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean3 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeploymentConfigService interface {
	CreateOrUpdateConfig(config *bean.DeploymentConfig, userId int32) (*bean.DeploymentConfig, error)
	GetDeploymentConfig(appId, envId int) (*bean.DeploymentConfig, error)
}

type DeploymentConfigServiceImpl struct {
	deploymentConfigRepository deploymentConfig.Repository
	logger                     *zap.SugaredLogger
	chartRepository            chartRepoRepository.ChartRepository
	pipelineRepository         pipelineConfig.PipelineRepository
	gitOpsConfigReadService    config.GitOpsConfigReadService
}

func NewDeploymentConfigServiceImpl(
	deploymentConfigRepository deploymentConfig.Repository,
	logger *zap.SugaredLogger,
	chartRepository chartRepoRepository.ChartRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	gitOpsConfigReadService config.GitOpsConfigReadService,
) *DeploymentConfigServiceImpl {
	return &DeploymentConfigServiceImpl{
		deploymentConfigRepository: deploymentConfigRepository,
		logger:                     logger,
		chartRepository:            chartRepository,
		pipelineRepository:         pipelineRepository,
		gitOpsConfigReadService:    gitOpsConfigReadService,
	}
}

func (impl *DeploymentConfigServiceImpl) CreateOrUpdateConfig(config *bean.DeploymentConfig, userId int32) (*bean.DeploymentConfig, error) {

	configDbObj, err := impl.GetConfigDBObj(config.AppId, config.EnvironmentId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching deployment config from DB by appId and envId",
			"appId", config.AppId, "envId", config.EnvironmentId, "err", err)
	}

	newDBObj := ConvertDeploymentConfigDTOToDbObj(config)

	if configDbObj == nil || (configDbObj != nil && configDbObj.Id == 0) {
		newDBObj.AuditLog.CreateAuditLog(userId)
		newDBObj, err = impl.deploymentConfigRepository.Save(newDBObj)
		if err != nil {
			impl.logger.Errorw("error in saving deploymentConfig", "appId", config.AppId, "envId", config.EnvironmentId, "err", err)
			return nil, err
		}
	} else {
		newDBObj.AuditLog.UpdateAuditLog(userId)
		newDBObj, err = impl.deploymentConfigRepository.Update(newDBObj)
		if err != nil {
			impl.logger.Errorw("error in updating deploymentConfig", "appId", config.AppId, "envId", config.EnvironmentId, "err", err)
			return nil, err
		}
	}

	return ConvertDeploymentConfigDbObjToDTO(newDBObj), nil
}

func (impl *DeploymentConfigServiceImpl) GetDeploymentConfig(appId, envId int) (*bean.DeploymentConfig, error) {

	appLevelConfigDbObj, err := impl.deploymentConfigRepository.GetAppLevelConfig(appId)
	if err != nil {
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
		if err != nil {
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
	ConfigDbObj, err = impl.deploymentConfigRepository.Save(ConfigDbObj)
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
			return configDbObj, nil
		}
		configDbObj.CredentialIdInt = gitOpsConfig.Id
	}

	configDbObj.AuditLog.CreateAuditLog(bean3.SYSTEM_USER_ID)
	configDbObj, err = impl.deploymentConfigRepository.Save(configDbObj)
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
