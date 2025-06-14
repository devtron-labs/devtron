package adapter

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
)

func NewDeploymentConfigMin(deploymentAppType, releaseMode string, isGitOpsRepoConfigured bool) *bean.DeploymentConfigMin {
	return &bean.DeploymentConfigMin{
		DeploymentAppType:      deploymentAppType,
		ReleaseMode:            releaseMode,
		IsGitOpsRepoConfigured: isGitOpsRepoConfigured,
	}
}

func ConvertDeploymentConfigDTOToDbObj(config *bean.DeploymentConfig) (*deploymentConfig.DeploymentConfig, error) {
	releaseConfigJson, err := json.Marshal(config.ReleaseConfiguration)
	if err != nil {
		return nil, err
	}
	return &deploymentConfig.DeploymentConfig{
		Id:                config.Id,
		AppId:             config.AppId,
		EnvironmentId:     config.EnvironmentId,
		DeploymentAppType: config.DeploymentAppType,
		RepoUrl:           config.RepoURL,
		ConfigType:        config.ConfigType,
		Active:            config.Active,
		ReleaseMode:       config.ReleaseMode,
		ReleaseConfig:     string(releaseConfigJson),
	}, nil
}

func ConvertDeploymentConfigDbObjToDTO(dbObj *deploymentConfig.DeploymentConfig) (*bean.DeploymentConfig, error) {

	if dbObj == nil {
		return nil, nil
	}

	var releaseConfig bean.ReleaseConfiguration

	if len(dbObj.ReleaseConfig) != 0 {
		err := json.Unmarshal([]byte(dbObj.ReleaseConfig), &releaseConfig)
		if err != nil {
			return nil, err
		}
	}

	return &bean.DeploymentConfig{
		Id:                   dbObj.Id,
		AppId:                dbObj.AppId,
		EnvironmentId:        dbObj.EnvironmentId,
		DeploymentAppType:    dbObj.DeploymentAppType,
		ConfigType:           dbObj.ConfigType,
		Active:               dbObj.Active,
		ReleaseMode:          dbObj.ReleaseMode,
		RepoURL:              dbObj.RepoUrl,
		ReleaseConfiguration: &releaseConfig,
	}, nil
}

func NewAppLevelReleaseConfigFromChart(gitRepoURL, chartLocation string) *bean.ReleaseConfiguration {
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

func NewFluxSpecReleaseConfig(clusterId int, namespace, gitRepositoryName, helmReleaseName, gitOpsSecretName, ChartLocation, ChartVersion,
	RevisionTarget, RepoUrl string, DevtronValueFileName string, HelmReleaseValuesFiles []string) *bean.ReleaseConfiguration {
	return &bean.ReleaseConfiguration{
		Version: bean.Version,
		FluxCDSpec: bean.FluxCDSpec{
			ClusterId:              clusterId,
			Namespace:              namespace,
			GitRepositoryName:      gitRepositoryName,
			HelmReleaseName:        helmReleaseName,
			GitOpsSecretName:       gitOpsSecretName,
			ChartLocation:          ChartLocation,
			ChartVersion:           ChartVersion,
			RevisionTarget:         RevisionTarget,
			RepoUrl:                RepoUrl,
			DevtronValueFile:       DevtronValueFileName,
			HelmReleaseValuesFiles: HelmReleaseValuesFiles,
		}}
}

func GetDeploymentConfigType(isCustomGitOpsRepo bool) string {
	if isCustomGitOpsRepo {
		return string(bean.CUSTOM)
	}
	return string(bean.SYSTEM_GENERATED)
}

func GetDevtronArgoCdAppInfo(acdAppName string, acdAppClusterId int, acdDefaultNamespace string) *bean.DevtronArgoCdAppInfo {
	return &bean.DevtronArgoCdAppInfo{
		ArgoCdAppName:    acdAppName,
		ArgoAppClusterId: acdAppClusterId,
		ArgoAppNamespace: acdDefaultNamespace,
	}
}
