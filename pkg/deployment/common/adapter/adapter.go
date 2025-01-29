package adapter

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
)

func NewDeploymentConfigMin(deploymentAppType, releaseMode string) *bean.DeploymentConfigMin {
	return &bean.DeploymentConfigMin{
		DeploymentAppType: deploymentAppType,
		ReleaseMode:       releaseMode,
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
