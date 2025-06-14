/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/client/argocdServer"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	installedAppReader "github.com/devtron-labs/devtron/pkg/appStore/installedApp/read"
	bean3 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	clusterBean "github.com/devtron-labs/devtron/pkg/cluster/bean"
	bean4 "github.com/devtron-labs/devtron/pkg/cluster/environment/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	commonErr "github.com/devtron-labs/devtron/pkg/deployment/common/errors"
	read2 "github.com/devtron-labs/devtron/pkg/deployment/common/read"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/read"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"path/filepath"
)

type DeploymentConfigService interface {
	CreateOrUpdateConfig(tx *pg.Tx, config *bean.DeploymentConfig, userId int32) (*bean.DeploymentConfig, error)
	CreateOrUpdateConfigInBulk(tx *pg.Tx, configToBeCreated, configToBeUpdated []*bean.DeploymentConfig, userId int32) error
	GetConfigForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error)
	GetAndMigrateConfigIfAbsentForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error)
	GetConfigForHelmApps(appId, envId int) (*bean.DeploymentConfig, error)
	IsChartStoreAppManagedByArgoCd(appId int) (bool, error)
	GetConfigEvenIfInactive(appId, envId int) (*bean.DeploymentConfig, error)
	GetAndMigrateConfigIfAbsentForHelmApp(appId, envId int) (*bean.DeploymentConfig, error)
	UpdateRepoUrlForAppAndEnvId(repoURL string, appId, envId int) error
	GetConfigsByAppIds(appIds []int) ([]*bean.DeploymentConfig, error)
	UpdateChartLocationInDeploymentConfig(appId, envId, chartRefId int, userId int32, chartVersion string) error
	GetAllArgoAppInfosByDeploymentAppNames(deploymentAppNames []string) ([]*bean.DevtronArgoCdAppInfo, error)
	GetExternalReleaseType(appId, environmentId int) (bean.ExternalReleaseType, error)
	CheckIfURLAlreadyPresent(repoURL string) (bool, error)
	FilterPipelinesByApplicationClusterIdAndNamespace(pipelines []pipelineConfig.Pipeline, applicationObjectClusterId int, applicationObjectNamespace string) (pipelineConfig.Pipeline, error)
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
	chartRefRepository          chartRepoRepository.ChartRefRepository
	deploymentConfigReadService read2.DeploymentConfigReadService
	acdAuthConfig               *util3.ACDAuthConfig
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
	chartRefRepository chartRepoRepository.ChartRefRepository,
	deploymentConfigReadService read2.DeploymentConfigReadService,
	acdAuthConfig *util3.ACDAuthConfig,
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
		chartRefRepository:          chartRefRepository,
		deploymentConfigReadService: deploymentConfigReadService,
		acdAuthConfig:               acdAuthConfig,
	}
}

func (impl *DeploymentConfigServiceImpl) CreateOrUpdateConfig(tx *pg.Tx, config *bean.DeploymentConfig, userId int32) (*bean.DeploymentConfig, error) {
	newDBObj, err := adapter.ConvertDeploymentConfigDTOToDbObj(config)
	if err != nil {
		impl.logger.Errorw("error in converting deployment config DTO to db object", "appId", config.AppId, "envId", config.EnvironmentId)
		return nil, err
	}

	configDbObj, err := impl.GetConfigDBObj(tx, config.AppId, config.EnvironmentId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching deployment config from DB by appId and envId",
			"appId", config.AppId, "envId", config.EnvironmentId, "err", err)
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

func (impl *DeploymentConfigServiceImpl) CreateOrUpdateConfigInBulk(tx *pg.Tx, configToBeCreated, configToBeUpdated []*bean.DeploymentConfig, userId int32) error {

	dbObjCreate := make([]*deploymentConfig.DeploymentConfig, 0, len(configToBeCreated))
	for i := range configToBeCreated {
		dbObj, err := adapter.ConvertDeploymentConfigDTOToDbObj(configToBeCreated[i])
		if err != nil {
			impl.logger.Errorw("error in converting deployment config DTO to db object", "appId", configToBeCreated[i].AppId, "envId", configToBeCreated[i].EnvironmentId)
			return err
		}
		dbObj.AuditLog.CreateAuditLog(userId)
		dbObjCreate = append(dbObjCreate, dbObj)
	}

	dbObjUpdate := make([]*deploymentConfig.DeploymentConfig, 0, len(configToBeUpdated))
	for i := range configToBeUpdated {
		dbObj, err := adapter.ConvertDeploymentConfigDTOToDbObj(configToBeUpdated[i])
		if err != nil {
			impl.logger.Errorw("error in converting deployment config DTO to db object", "appId", configToBeUpdated[i].AppId, "envId", configToBeUpdated[i].EnvironmentId)
			return err
		}
		dbObj.AuditLog.UpdateAuditLog(userId)
		dbObjUpdate = append(dbObjUpdate, dbObj)
	}

	if len(dbObjCreate) > 0 {
		_, err := impl.deploymentConfigRepository.SaveAll(tx, dbObjCreate)
		if err != nil {
			impl.logger.Errorw("error in saving deploymentConfig", "dbObjCreate", dbObjCreate, "err", err)
			return err
		}
	}

	if len(dbObjUpdate) > 0 {
		_, err := impl.deploymentConfigRepository.UpdateAll(tx, dbObjUpdate)
		if err != nil {
			impl.logger.Errorw("error in updating deploymentConfig", "dbObjUpdate", dbObjUpdate, "err", err)
			return err
		}
	}

	return nil
}

func (impl *DeploymentConfigServiceImpl) GetConfigForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error) {

	appLevelConfig, err := impl.getAppLevelConfigForDevtronApps(appId, false)
	if err != nil {
		impl.logger.Errorw("error in getting app level Config for devtron apps", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	if envId > 0 {
		// if envId > 0 then only env level config will be returned,
		// for getting app level config envId should be zero
		envLevelConfig, err := impl.getEnvLevelDataForDevtronApps(appId, envId, appLevelConfig, false)
		if err != nil {
			impl.logger.Errorw("error in getting env level data for devtron apps", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		return envLevelConfig, nil
	}
	return appLevelConfig, nil
}

func (impl *DeploymentConfigServiceImpl) GetAndMigrateConfigIfAbsentForDevtronApps(appId, envId int) (*bean.DeploymentConfig, error) {
	migrateDeploymentConfigData := impl.deploymentServiceTypeConfig.MigrateDeploymentConfigData
	appLevelConfig, err := impl.getAppLevelConfigForDevtronApps(appId, migrateDeploymentConfigData)
	if err != nil {
		impl.logger.Errorw("error in getting app level Config for devtron apps", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	var envLevelConfig *bean.DeploymentConfig
	if envId > 0 {
		// if envId > 0 then only env level config will be returned,
		// for getting app level config envId should be zero
		envLevelConfig, err = impl.getEnvLevelDataForDevtronApps(appId, envId, appLevelConfig, migrateDeploymentConfigData)
		if err != nil {
			impl.logger.Errorw("error in getting env level data for devtron apps", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		return envLevelConfig, nil
	}
	return appLevelConfig, nil
}

func (impl *DeploymentConfigServiceImpl) GetConfigForHelmApps(appId, envId int) (*bean.DeploymentConfig, error) {
	helmDeploymentConfig, err := impl.getConfigForHelmApps(appId, envId, false)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config for helm app", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	return helmDeploymentConfig, nil
}

func (impl *DeploymentConfigServiceImpl) IsChartStoreAppManagedByArgoCd(appId int) (bool, error) {
	deploymentAppType, err := impl.deploymentConfigRepository.GetDeploymentAppTypeForChartStoreAppByAppId(appId)
	if err != nil && !util2.IsErrNoRows(err) {
		impl.logger.Errorw("error in GetDeploymentAppTypeForChartStoreAppByAppId", "appId", appId, "err", err)
		return false, err
	} else if util2.IsErrNoRows(err) {
		return impl.installedAppReadService.IsChartStoreAppManagedByArgoCd(appId)
	}
	return util2.IsAcdApp(deploymentAppType), nil
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

func (impl *DeploymentConfigServiceImpl) UpdateRepoUrlForAppAndEnvId(repoURL string, appId, envId int) error {

	dbObj, err := impl.deploymentConfigRepository.GetByAppIdAndEnvId(nil, appId, envId)
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

func (impl *DeploymentConfigServiceImpl) GetConfigsByAppIds(appIds []int) ([]*bean.DeploymentConfig, error) {
	if len(appIds) == 0 {
		return nil, nil
	}
	configs, err := impl.deploymentConfigRepository.GetConfigByAppIds(appIds)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config db object by appIds", "appIds", appIds, "err", err)
		return nil, err
	}
	resp := make([]*bean.DeploymentConfig, 0, len(configs))
	for _, config := range configs {
		newObj, err := adapter.ConvertDeploymentConfigDbObjToDTO(config)
		if err != nil {
			impl.logger.Errorw("error in converting deployment config DTO to db object", "appId", config.AppId, "envId", config.EnvironmentId)
			return nil, err
		}
		resp = append(resp, newObj)
	}
	return resp, nil
}

func (impl *DeploymentConfigServiceImpl) UpdateChartLocationInDeploymentConfig(appId, envId, chartRefId int, userId int32, chartVersion string) error {

	pipeline, err := impl.pipelineRepository.FindOneByAppIdAndEnvId(appId, envId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in finding pipeline by app id and env id", "appId", appId, "envId", envId, "err", err)
		return err
	}
	// no need to update deployment config if pipeline is not present
	if errors.Is(err, pg.ErrNoRows) || (pipeline != nil && pipeline.Id == 0) {
		return nil
	}

	config, err := impl.GetConfigForDevtronApps(appId, envId)
	if err != nil {
		impl.logger.Errorw("error, GetConfigForDevtronApps", "appId", appId, "envId", envId, "err", err)
		return err
	}
	if config.ReleaseMode == util2.PIPELINE_RELEASE_MODE_CREATE && (config.IsAcdRelease() || config.IsFluxRelease()) {
		chartRef, err := impl.chartRefRepository.FindById(chartRefId)
		if err != nil {
			impl.logger.Errorw("error in chartRefRepository.FindById", "chartRefId", chartRefId, "err", err)
			return err
		}
		chartLocation := filepath.Join(chartRef.Location, chartVersion)
		config.SetChartLocation(chartLocation)
		config, err = impl.CreateOrUpdateConfig(nil, config, userId)
		if err != nil {
			impl.logger.Errorw("error in CreateOrUpdateConfig", "appId", appId, "envId", envId, "err", err)
			return err
		}
	}
	return nil
}

func (impl *DeploymentConfigServiceImpl) GetAllArgoAppInfosByDeploymentAppNames(deploymentAppNames []string) ([]*bean.DevtronArgoCdAppInfo, error) {
	allDevtronManagedArgoAppsInfo := make([]*bean.DevtronArgoCdAppInfo, 0)
	linkedReleaseConfig, err := impl.getAllEnvLevelConfigsForLinkedReleases()
	if err != nil {
		impl.logger.Errorw("error while fetching linked release configs", "deploymentAppNames", deploymentAppNames, "error", err)
		return allDevtronManagedArgoAppsInfo, err
	}
	linkedReleaseConfigMap := make(map[string]*bean.DeploymentConfig)
	for _, config := range linkedReleaseConfig {
		uniqueKey := fmt.Sprintf("%d-%d", config.AppId, config.EnvironmentId)
		linkedReleaseConfigMap[uniqueKey] = config
	}
	devtronArgoAppsInfo, err := impl.pipelineRepository.GetAllArgoAppInfoByDeploymentAppNames(deploymentAppNames)
	if err != nil {
		impl.logger.Errorw("error while fetching argo app names", "deploymentAppNames", deploymentAppNames, "error", err)
		return allDevtronManagedArgoAppsInfo, err
	}
	for _, acdAppInfo := range devtronArgoAppsInfo {
		uniqueKey := fmt.Sprintf("%d-%d", acdAppInfo.AppId, acdAppInfo.EnvironmentId)
		var devtronArgoCdAppInfo *bean.DevtronArgoCdAppInfo
		if config, ok := linkedReleaseConfigMap[uniqueKey]; ok &&
			config.IsAcdRelease() && config.IsLinkedRelease() {
			acdAppClusterId := config.GetApplicationObjectClusterId()
			acdDefaultNamespace := config.GetApplicationObjectNamespace()
			devtronArgoCdAppInfo = adapter.GetDevtronArgoCdAppInfo(acdAppInfo.DeploymentAppName, acdAppClusterId, acdDefaultNamespace)
		} else {
			devtronArgoCdAppInfo = adapter.GetDevtronArgoCdAppInfo(acdAppInfo.DeploymentAppName, clusterBean.DefaultClusterId, impl.acdAuthConfig.ACDConfigMapNamespace)
		}
		allDevtronManagedArgoAppsInfo = append(allDevtronManagedArgoAppsInfo, devtronArgoCdAppInfo)
	}
	chartStoreArgoAppNames, err := impl.installedAppReadService.GetAllArgoAppNamesByDeploymentAppNames(deploymentAppNames)
	if err != nil {
		impl.logger.Errorw("error while fetching argo app names from chart store", "deploymentAppNames", deploymentAppNames, "error", err)
		return allDevtronManagedArgoAppsInfo, err
	}
	for _, chartStoreArgoAppName := range chartStoreArgoAppNames {
		// NOTE: Chart Store doesn't support linked releases
		chartStoreArgoCdAppInfo := adapter.GetDevtronArgoCdAppInfo(chartStoreArgoAppName, clusterBean.DefaultClusterId, impl.acdAuthConfig.ACDConfigMapNamespace)
		allDevtronManagedArgoAppsInfo = append(allDevtronManagedArgoAppsInfo, chartStoreArgoCdAppInfo)
	}
	return allDevtronManagedArgoAppsInfo, nil
}

func (impl *DeploymentConfigServiceImpl) GetExternalReleaseType(appId, environmentId int) (bean.ExternalReleaseType, error) {
	config, err := impl.GetConfigForDevtronApps(appId, environmentId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in getting deployment config by appId and envId", "appId", appId, "envId", environmentId, "err", err)
		return bean.Undefined, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return bean.Undefined, nil
	}
	externalHelmReleaseType, _ := config.GetMigratedFrom()
	return externalHelmReleaseType, nil
}

func (impl *DeploymentConfigServiceImpl) CheckIfURLAlreadyPresent(repoURL string) (bool, error) {
	//TODO: optimisation
	configs, err := impl.getAllAppLevelConfigsWithCustomGitOpsURL()
	if err != nil {
		impl.logger.Errorw("error in getting all configs", "err", err)
		return false, err
	}
	for _, dc := range configs {
		if dc.GetRepoURL() == repoURL {
			impl.logger.Warnw("repository is already in use for helm app", "repoUrl", repoURL)
			return true, nil
		}
	}
	return false, nil
}

func (impl *DeploymentConfigServiceImpl) FilterPipelinesByApplicationClusterIdAndNamespace(pipelines []pipelineConfig.Pipeline, applicationObjectClusterId int, applicationObjectNamespace string) (pipelineConfig.Pipeline, error) {
	pipeline := pipelineConfig.Pipeline{}
	for _, p := range pipelines {
		dc, err := impl.GetConfigForDevtronApps(p.AppId, p.EnvironmentId)
		if err != nil {
			impl.logger.Errorw("error, GetConfigForDevtronApps", "appId", p.AppId, "environmentId", p.EnvironmentId, "err", err)
			return pipeline, err
		}
		if dc.GetApplicationObjectClusterId() == applicationObjectClusterId &&
			dc.GetApplicationObjectNamespace() == applicationObjectNamespace {
			return p, nil
		}
	}
	return pipeline, commonErr.PipelineNotFoundError
}

func (impl *DeploymentConfigServiceImpl) getConfigForHelmApps(appId int, envId int, migrateIfAbsent bool) (*bean.DeploymentConfig, error) {
	var (
		helmDeploymentConfig *bean.DeploymentConfig
		isMigrationNeeded    bool
	)
	config, err := impl.deploymentConfigRepository.GetByAppIdAndEnvId(nil, appId, envId)
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
		ConfigType:        adapter.GetDeploymentConfigType(installedApp.IsCustomRepository),
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
					ClusterId: clusterBean.DefaultClusterId,
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

func (impl *DeploymentConfigServiceImpl) getAllAppLevelConfigsWithCustomGitOpsURL() ([]*bean.DeploymentConfig, error) {
	dbConfigs, err := impl.deploymentConfigRepository.GetAllConfigsForActiveApps()
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

func (impl *DeploymentConfigServiceImpl) getAllEnvLevelConfigsForLinkedReleases() ([]*bean.DeploymentConfig, error) {
	dbConfigs, err := impl.deploymentConfigRepository.GetAllEnvLevelConfigsWithReleaseMode(util2.PIPELINE_RELEASE_MODE_LINK)
	if err != nil {
		impl.logger.Errorw("error in getting all env level configs with custom gitops url", "err", err)
		return nil, err
	}
	configs := make([]*bean.DeploymentConfig, 0)
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
func (impl *DeploymentConfigServiceImpl) GetConfigDBObj(tx *pg.Tx, appId, envId int) (*deploymentConfig.DeploymentConfig, error) {
	var configDbObj *deploymentConfig.DeploymentConfig
	var err error
	if envId == 0 {
		configDbObj, err = impl.deploymentConfigRepository.GetAppLevelConfigForDevtronApps(tx, appId)
		if err != nil {
			impl.logger.Errorw("error in getting deployment config db object by appId", "appId", appId, "err", err)
			return nil, err
		}
	} else {
		configDbObj, err = impl.deploymentConfigRepository.GetByAppIdAndEnvId(tx, appId, envId)
		if err != nil {
			impl.logger.Errorw("error in getting deployment config db object by appId and envId", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
	}
	return configDbObj, nil
}

func (impl *DeploymentConfigServiceImpl) getAppLevelConfigForDevtronApps(appId int, migrateDataIfAbsent bool) (*bean.DeploymentConfig, error) {
	appLevelConfig, isMigrationNeeded, err := impl.deploymentConfigReadService.GetDeploymentConfigForApp(appId)
	if err != nil {
		impl.logger.Errorw("error in getting app level Config for devtron apps", "appId", appId, "err", err)
		return nil, err
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

func (impl *DeploymentConfigServiceImpl) getEnvLevelDataForDevtronApps(appId, envId int, appLevelConfig *bean.DeploymentConfig, migrateDataIfAbsent bool) (*bean.DeploymentConfig, error) {
	envLevelConfig, isMigrationNeeded, err := impl.deploymentConfigReadService.GetDeploymentConfigForAppAndEnv(appLevelConfig, appId, envId)
	if err != nil {
		impl.logger.Errorw("error in getting env level data for devtron apps", "appId", appId, "envId", envId, "appLevelConfig", appLevelConfig, "err", err)
		return nil, err
	}
	if migrateDataIfAbsent && isMigrationNeeded {
		_, err := impl.CreateOrUpdateConfig(nil, envLevelConfig, bean3.SYSTEM_USER_ID)
		if err != nil {
			impl.logger.Errorw("error in migrating env level config to deployment config", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
	}
	return envLevelConfig, nil
}
