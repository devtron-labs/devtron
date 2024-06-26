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
	configDbObj, err := impl.GetConfigDBObj(appId, envId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching deployment config from DB by appId and envId",
			"appId", appId, "envId", envId, "err", err)
	}
	if err == pg.ErrNoRows || (configDbObj == nil || (configDbObj != nil && configDbObj.Id == 0)) {
		configDbObj, err = impl.migrateOldDataToDeploymentConfig(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in migrating old data to deployment config", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
	}
	return ConvertDeploymentConfigDbObjToDTO(configDbObj), nil
}

func (impl *DeploymentConfigServiceImpl) migrateOldDataToDeploymentConfig(appId int, envId int) (*deploymentConfig.DeploymentConfig, error) {

	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetch chart for git repo migration by appId", "appId", appId, "err", err)
		return nil, err
	}
	var deploymentAppType string
	if envId > 0 {
		deploymentAppType, err = impl.pipelineRepository.FindDeploymentAppTypeByAppIdAndEnvId(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in getting deployment app type by appId and envId", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
	}

	ConfigDbObj := &deploymentConfig.DeploymentConfig{
		AppId:             appId,
		EnvironmentId:     envId,
		DeploymentAppType: deploymentAppType,
		Active:            true,
	}

	switch deploymentAppType {
	//TODO: handling for other deployment app type in future
	case bean2.ArgoCd:
		ConfigDbObj.ConfigType = GetDeploymentConfigType(chart.IsCustomGitRepository)
		ConfigDbObj.RepoUrl = chart.GitRepoUrl
		ConfigDbObj.ChartLocation = chart.ChartLocation
		ConfigDbObj.CredentialType = string(bean.GitOps)
		gitOpsConfig, err := impl.gitOpsConfigReadService.GetGitOpsProviderByRepoURL(ConfigDbObj.RepoUrl)
		if err != nil {
			impl.logger.Infow("error in fetching gitOps config by repoUrl, skipping migration to deployment config", "repoURL", ConfigDbObj.RepoUrl)
			return ConfigDbObj, nil
		}
		ConfigDbObj.CredentialIdInt = gitOpsConfig.Id
	}

	ConfigDbObj.AuditLog.CreateAuditLog(bean3.SYSTEM_USER_ID)
	ConfigDbObj, err = impl.deploymentConfigRepository.Save(ConfigDbObj)
	if err != nil {
		impl.logger.Errorw("error in saving deployment config in DB", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	return ConfigDbObj, nil
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
