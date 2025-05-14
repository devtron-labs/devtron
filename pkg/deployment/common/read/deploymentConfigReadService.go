/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package read

import (
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	interalUtil "github.com/devtron-labs/devtron/internal/util"
	serviceBean "github.com/devtron-labs/devtron/pkg/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	clusterBean "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/read"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/helper"
	"github.com/devtron-labs/devtron/util"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"path/filepath"
)

type DeploymentConfigReadService interface {
	GetDeploymentConfigMinForAppAndEnv(appId, envId int) (*bean.DeploymentConfigMin, error)
	GetDeploymentAppTypeForCDInBulk(pipelines []*serviceBean.CDPipelineMinConfig, appIdToGitOpsConfiguredMap map[int]bool) (map[int]*bean.DeploymentConfigMin, error)

	GetDeploymentConfigForApp(appId int) (*bean.DeploymentConfig, bool, error)
	GetDeploymentConfigForAppAndEnv(appLevelConfig *bean.DeploymentConfig, appId, envId int) (*bean.DeploymentConfig, bool, error)
	ParseEnvLevelReleaseConfigForDevtronApp(config *bean.DeploymentConfig, appId int, envId int) (*bean.ReleaseConfiguration, error)
}

type DeploymentConfigReadServiceImpl struct {
	logger                      *zap.SugaredLogger
	deploymentConfigRepository  deploymentConfig.Repository
	deploymentServiceTypeConfig *util.DeploymentServiceTypeConfig
	chartRepository             chartRepoRepository.ChartRepository
	pipelineRepository          pipelineConfig.PipelineRepository
	appRepository               app.AppRepository
	environmentRepository       repository.EnvironmentRepository
	envConfigOverrideService    read.EnvConfigOverrideService
}

func NewDeploymentConfigReadServiceImpl(logger *zap.SugaredLogger,
	deploymentConfigRepository deploymentConfig.Repository,
	envVariables *util.EnvironmentVariables,
	chartRepository chartRepoRepository.ChartRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	appRepository app.AppRepository,
	environmentRepository repository.EnvironmentRepository,
	envConfigOverrideService read.EnvConfigOverrideService,
) *DeploymentConfigReadServiceImpl {
	return &DeploymentConfigReadServiceImpl{
		logger:                      logger,
		deploymentConfigRepository:  deploymentConfigRepository,
		deploymentServiceTypeConfig: envVariables.DeploymentServiceTypeConfig,
		chartRepository:             chartRepository,
		pipelineRepository:          pipelineRepository,
		appRepository:               appRepository,
		environmentRepository:       environmentRepository,
		envConfigOverrideService:    envConfigOverrideService,
	}
}

func (impl *DeploymentConfigReadServiceImpl) GetDeploymentConfigMinForAppAndEnv(appId, envId int) (*bean.DeploymentConfigMin, error) {
	deploymentDetail := &bean.DeploymentConfigMin{}
	configBean, err := impl.getDeploymentConfigMinForAppAndEnv(appId, envId)
	if err != nil && !interalUtil.IsErrNoRows(err) {
		impl.logger.Errorw("error in getting deployment config by appId and envId", "appId", appId, "envId", envId, "err", err)
		return deploymentDetail, err
	} else if interalUtil.IsErrNoRows(err) {
		// case: deployment config data not found in app level or env level
		// this means deployment template is not yet created for this app and env
		deploymentDetail.ReleaseMode = interalUtil.PIPELINE_RELEASE_MODE_CREATE
		return deploymentDetail, nil
	}
	if configBean != nil {
		deploymentDetail.DeploymentAppType = configBean.DeploymentAppType
		deploymentDetail.ReleaseMode = configBean.ReleaseMode
		deploymentDetail.GitRepoUrl = configBean.GetRepoURL()
		deploymentDetail.IsGitOpsRepoConfigured = !apiGitOpsBean.IsGitOpsRepoNotConfigured(configBean.GetRepoURL())
	}
	return deploymentDetail, nil
}

func (impl *DeploymentConfigReadServiceImpl) GetDeploymentAppTypeForCDInBulk(pipelines []*serviceBean.CDPipelineMinConfig, appIdToGitOpsConfiguredMap map[int]bool) (map[int]*bean.DeploymentConfigMin, error) {
	resp := make(map[int]*bean.DeploymentConfigMin, len(pipelines)) //map of pipelineId and deploymentAppType
	appIdEnvIdMapping := make(map[int][]int, len(pipelines))
	appIdEnvIdKeyPipelineIdMap := make(map[string]int, len(pipelines))
	for _, pipeline := range pipelines {
		appIdEnvIdMapping[pipeline.AppId] = append(appIdEnvIdMapping[pipeline.AppId], pipeline.EnvironmentId)
		appIdEnvIdKeyPipelineIdMap[fmt.Sprintf("%d-%d", pipeline.AppId, pipeline.EnvironmentId)] = pipeline.Id
	}
	configs, err := impl.deploymentConfigRepository.GetAppAndEnvLevelConfigsInBulk(appIdEnvIdMapping)
	if err != nil {
		impl.logger.Errorw("error, GetAppAndEnvLevelConfigsInBulk", "appIdEnvIdMapping", appIdEnvIdMapping, "err", err)
		return nil, err
	}
	for _, config := range configs {
		configBean, err := adapter.ConvertDeploymentConfigDbObjToDTO(config)
		if err != nil {
			impl.logger.Errorw("error, ConvertDeploymentConfigDbObjToDTO", "config", config, "err", err)
			return nil, err
		}
		pipelineId := appIdEnvIdKeyPipelineIdMap[fmt.Sprintf("%d-%d", configBean.AppId, configBean.EnvironmentId)]
		isGitOpsRepoConfigured := configBean.IsPipelineGitOpsRepoConfigured(appIdToGitOpsConfiguredMap[configBean.AppId])
		resp[pipelineId] = adapter.NewDeploymentConfigMin(configBean.DeploymentAppType, configBean.ReleaseMode, isGitOpsRepoConfigured)
	}
	for _, pipeline := range pipelines {
		if _, ok := resp[pipeline.Id]; !ok {
			isGitOpsRepoConfigured := appIdToGitOpsConfiguredMap[pipeline.AppId]
			// not found in map, either flag is disabled or config not migrated yet. Getting from old data
			resp[pipeline.Id] = adapter.NewDeploymentConfigMin(pipeline.DeploymentAppType, interalUtil.PIPELINE_RELEASE_MODE_CREATE, isGitOpsRepoConfigured)
		}
	}
	return resp, nil
}

func (impl *DeploymentConfigReadServiceImpl) GetDeploymentConfigForApp(appId int) (*bean.DeploymentConfig, bool, error) {
	var (
		appLevelConfig    *bean.DeploymentConfig
		isMigrationNeeded bool
	)
	appLevelConfigDbObj, err := impl.deploymentConfigRepository.GetAppLevelConfigForDevtronApps(appId)
	if err != nil && !interalUtil.IsErrNoRows(err) {
		impl.logger.Errorw("error in getting deployment config db object by appId", "appId", appId, "err", err)
		return appLevelConfig, isMigrationNeeded, err
	} else if interalUtil.IsErrNoRows(err) {
		isMigrationNeeded = true
		appLevelConfig, err = impl.parseAppLevelMigrationDataForDevtronApps(appId)
		if err != nil {
			impl.logger.Errorw("error in migrating app level config to deployment config", "appId", appId, "err", err)
			return appLevelConfig, isMigrationNeeded, err
		}
	} else {
		appLevelConfig, err = adapter.ConvertDeploymentConfigDbObjToDTO(appLevelConfigDbObj)
		if err != nil {
			impl.logger.Errorw("error in converting deployment config db object", "appId", appId, "err", err)
			return appLevelConfig, isMigrationNeeded, err
		}
		if appLevelConfig.ReleaseConfiguration == nil || len(appLevelConfig.ReleaseConfiguration.Version) == 0 {
			isMigrationNeeded = true
			releaseConfig, err := impl.parseAppLevelReleaseConfigForDevtronApp(appId, appLevelConfig)
			if err != nil {
				impl.logger.Errorw("error in parsing release configuration for app", "appId", appId, "err", err)
				return appLevelConfig, isMigrationNeeded, err
			}
			appLevelConfig.ReleaseConfiguration = releaseConfig
		}
	}
	return appLevelConfig, isMigrationNeeded, nil
}

func (impl *DeploymentConfigReadServiceImpl) GetDeploymentConfigForAppAndEnv(appLevelConfig *bean.DeploymentConfig, appId, envId int) (*bean.DeploymentConfig, bool, error) {
	var (
		envLevelConfig    *bean.DeploymentConfig
		isMigrationNeeded bool
	)
	appAndEnvLevelConfigDBObj, err := impl.deploymentConfigRepository.GetByAppIdAndEnvId(appId, envId)
	if err != nil && !interalUtil.IsErrNoRows(err) {
		impl.logger.Errorw("error in getting deployment config db object by appId and envId", "appId", appId, "envId", envId, "err", err)
		return envLevelConfig, isMigrationNeeded, err
	} else if interalUtil.IsErrNoRows(err) {
		// case: deployment config data is not yet migrated
		envLevelConfig, err = impl.parseEnvLevelMigrationDataForDevtronApps(appLevelConfig, appId, envId)
		if err != nil {
			impl.logger.Errorw("error in parsing env level config to deployment config", "appId", appId, "envId", envId, "err", err)
			return envLevelConfig, isMigrationNeeded, err
		}
		isMigrationNeeded = true
	} else {
		envLevelConfig, err = adapter.ConvertDeploymentConfigDbObjToDTO(appAndEnvLevelConfigDBObj)
		if err != nil {
			impl.logger.Errorw("error in converting deployment config db object", "appId", appId, "envId", envId, "err", err)
			return envLevelConfig, isMigrationNeeded, err
		}
		// case: deployment config is migrated; but release config is absent.
		if envLevelConfig.ReleaseConfiguration == nil || len(envLevelConfig.ReleaseConfiguration.Version) == 0 {
			isMigrationNeeded = true
			releaseConfig, err := impl.ParseEnvLevelReleaseConfigForDevtronApp(envLevelConfig, appId, envId)
			if err != nil {
				impl.logger.Errorw("error in parsing env level release config", "appId", appId, "envId", envId, "err", err)
				return envLevelConfig, isMigrationNeeded, err
			}
			envLevelConfig.ReleaseConfiguration = releaseConfig
		}
	}
	var isRepoUrlUpdated bool
	envLevelConfig, isRepoUrlUpdated, err = impl.configureEnvURLByAppURLIfNotConfigured(envLevelConfig, appLevelConfig.GetRepoURL())
	if err != nil {
		impl.logger.Errorw("error in configuring env level url with app url", "appId", appId, "envId", envId, "err", err)
		return envLevelConfig, isMigrationNeeded, err
	}
	if isRepoUrlUpdated {
		isMigrationNeeded = true
	}
	return envLevelConfig, isMigrationNeeded, nil
}

func (impl *DeploymentConfigReadServiceImpl) getDeploymentConfigMinForAppAndEnv(appId, envId int) (*bean.DeploymentConfig, error) {
	appLevelConfig, err := impl.getAppLevelConfigForDevtronApps(appId)
	if err != nil {
		impl.logger.Errorw("error in getting app level config for devtron apps", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	if envId > 0 {
		// if envId > 0 then only env level config will be returned,
		// for getting app level config envId should be zero
		envLevelConfig, err := impl.getEnvLevelDataForDevtronApps(appId, envId, appLevelConfig)
		if err != nil {
			impl.logger.Errorw("error in getting env level data for devtron apps", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		return envLevelConfig, nil
	}
	return appLevelConfig, nil
}

func (impl *DeploymentConfigReadServiceImpl) getAppLevelConfigForDevtronApps(appId int) (*bean.DeploymentConfig, error) {
	appLevelConfig, _, err := impl.GetDeploymentConfigForApp(appId)
	if err != nil {
		impl.logger.Errorw("error in getting app level Config for devtron apps", "appId", appId, "err", err)
		return nil, err
	}
	return appLevelConfig, nil
}

func (impl *DeploymentConfigReadServiceImpl) getEnvLevelDataForDevtronApps(appId, envId int, appLevelConfig *bean.DeploymentConfig) (*bean.DeploymentConfig, error) {
	envLevelConfig, _, err := impl.GetDeploymentConfigForAppAndEnv(appLevelConfig, appId, envId)
	if err != nil {
		impl.logger.Errorw("error in getting env level data for devtron apps", "appId", appId, "envId", envId, "appLevelConfig", appLevelConfig, "err", err)
		return nil, err
	}
	return envLevelConfig, nil
}

func (impl *DeploymentConfigReadServiceImpl) configureEnvURLByAppURLIfNotConfigured(appAndEnvLevelConfig *bean.DeploymentConfig, appLevelURL string) (*bean.DeploymentConfig, bool, error) {
	/*
		if custom gitOps is configured in repo
		and app is cloned then cloned pipelines repo URL=NOT_CONFIGURED .
		In this case User manually configures repoURL. The configured repo_url is saved in app level config but is absent
		in env level config.
	*/
	var isRepoUrlUpdated bool
	if apiGitOpsBean.IsGitOpsRepoNotConfigured(appAndEnvLevelConfig.GetRepoURL()) &&
		apiGitOpsBean.IsGitOpsRepoConfigured(appLevelURL) {
		// if url is present at app level and not at env level then copy app level url to env level config
		appAndEnvLevelConfig.SetRepoURL(appLevelURL)
		// if url is updated then set isRepoUrlUpdated = true
		isRepoUrlUpdated = true
		return appAndEnvLevelConfig, isRepoUrlUpdated, nil
	}
	return appAndEnvLevelConfig, isRepoUrlUpdated, nil
}

func (impl *DeploymentConfigReadServiceImpl) parseEnvLevelMigrationDataForDevtronApps(appLevelConfig *bean.DeploymentConfig, appId, envId int) (*bean.DeploymentConfig, error) {
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
		ReleaseMode:   interalUtil.PIPELINE_RELEASE_MODE_CREATE,
		Active:        true,
	}

	deploymentAppType, err := impl.pipelineRepository.FindDeploymentAppTypeByAppIdAndEnvId(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment app type by appId and envId", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	config.DeploymentAppType = deploymentAppType

	releaseConfig, err := impl.ParseEnvLevelReleaseConfigForDevtronApp(config, appId, envId)
	if err != nil {
		impl.logger.Errorw("error in parsing env level release config", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	config.ReleaseConfiguration = releaseConfig

	config.RepoURL = config.GetRepoURL() //for backward compatibility

	return config, nil
}

func (impl *DeploymentConfigReadServiceImpl) getConfigMetaDataForAppAndEnv(appId int, envId int) (environmentId int, deploymentAppName, namespace string, err error) {
	pipelineModels, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment app type by appId and envId", "appId", appId, "envId", envId, "err", err)
		return environmentId, deploymentAppName, namespace, err
	} else if len(pipelineModels) > 1 {
		impl.logger.Errorw("error, multiple pipelines found for app and env", "appId", appId, "envId", envId, "pipelineModels", pipelineModels)
		return environmentId, deploymentAppName, namespace, errors.New("multiple pipelines found for app and env")
	} else if len(pipelineModels) == 0 {
		appModel, err := impl.appRepository.FindById(appId)
		if err != nil {
			impl.logger.Errorw("error in fetch app", "appId", appId, "err", err)
			return environmentId, deploymentAppName, namespace, err
		}
		envModel, err := impl.environmentRepository.FindById(envId)
		if err != nil {
			impl.logger.Errorw("error in finding environment by id", "envId", envId, "err", err)
			return environmentId, deploymentAppName, namespace, err
		}
		deploymentAppName = util.BuildDeployedAppName(appModel.AppName, envModel.Name)
		environmentId = envModel.Id
		namespace = envModel.Namespace
	} else {
		pipelineModel := pipelineModels[0]
		deploymentAppName = pipelineModel.DeploymentAppName
		environmentId = pipelineModel.EnvironmentId
		namespace = pipelineModel.Environment.Namespace
	}
	return environmentId, deploymentAppName, namespace, nil
}

func (impl *DeploymentConfigReadServiceImpl) ParseEnvLevelReleaseConfigForDevtronApp(config *bean.DeploymentConfig, appId int, envId int) (*bean.ReleaseConfiguration, error) {
	releaseConfig := &bean.ReleaseConfiguration{}
	if config.DeploymentAppType == interalUtil.PIPELINE_DEPLOYMENT_TYPE_ACD {
		releaseConfig.Version = bean.Version
		envOverride, err := impl.envConfigOverrideService.FindLatestChartForAppByAppIdAndEnvId(appId, envId)
		if err != nil && !errors.IsNotFound(err) {
			impl.logger.Errorw("error in fetch")
			return nil, err
		}
		var latestChart *chartRepoRepository.Chart
		if !envOverride.IsOverridden() {
			latestChart, err = impl.chartRepository.FindLatestChartForAppByAppId(appId)
			if err != nil {
				return nil, err
			}
		} else {
			// if chart is overrides in env, it means it may have different version than app level.
			latestChart = envOverride.Chart
		}
		gitRepoUrl := latestChart.GitRepoUrl
		if len(config.RepoURL) > 0 {
			gitRepoUrl = config.RepoURL
		}
		environmentId, deploymentAppName, namespace, err := impl.getConfigMetaDataForAppAndEnv(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in getting config meta data for app and env", "appId", appId, "envId", envId, "err", err)
			return nil, err
		}
		releaseConfig.ArgoCDSpec = bean.ArgoCDSpec{
			Metadata: bean.ApplicationMetadata{
				Name:      deploymentAppName,
				ClusterId: clusterBean.DefaultClusterId,
				Namespace: argocdServer.DevtronInstalationNs,
			},
			Spec: bean.ApplicationSpec{
				Source: &bean.ApplicationSource{
					RepoURL: gitRepoUrl,
					Path:    latestChart.ChartLocation,
					Helm: &bean.ApplicationSourceHelm{
						ValueFiles: []string{helper.GetValuesFileForEnv(environmentId)},
					},
					TargetRevision: util.GetDefaultTargetRevision(),
				},
				Destination: &bean.Destination{
					Namespace: namespace,
					Server:    commonBean.DefaultClusterUrl,
				},
			},
		}
	}
	return releaseConfig, nil
}

func (impl *DeploymentConfigReadServiceImpl) parseAppLevelMigrationDataForDevtronApps(appId int) (*bean.DeploymentConfig, error) {
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		return nil, err
	}

	chartLocation := filepath.Join(chart.ReferenceTemplate, chart.ChartVersion)
	releaseConfig := adapter.NewAppLevelReleaseConfigFromChart(chart.GitRepoUrl, chartLocation)
	config := &bean.DeploymentConfig{
		AppId:                appId,
		ConfigType:           adapter.GetDeploymentConfigType(chart.IsCustomGitRepository),
		Active:               true,
		RepoURL:              chart.GitRepoUrl, //for backward compatibility
		ReleaseConfiguration: releaseConfig,
	}
	return config, nil
}

func (impl *DeploymentConfigReadServiceImpl) parseAppLevelReleaseConfigForDevtronApp(appId int, appLevelConfig *bean.DeploymentConfig) (*bean.ReleaseConfiguration, error) {
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		return nil, err
	}

	repoURL := chart.GitRepoUrl
	if len(appLevelConfig.RepoURL) > 0 {
		repoURL = appLevelConfig.RepoURL
	}
	chartLocation := filepath.Join(chart.ReferenceTemplate, chart.ChartVersion)
	releaseConfig := adapter.NewAppLevelReleaseConfigFromChart(repoURL, chartLocation)
	return releaseConfig, nil
}
