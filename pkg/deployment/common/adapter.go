package common

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
)

func ConvertDeploymentConfigDTOToDbObj(config *bean.DeploymentConfig) (*deploymentConfig.DeploymentConfig, error) {
	return &deploymentConfig.DeploymentConfig{
		Id:                config.Id,
		AppId:             config.AppId,
		EnvironmentId:     config.EnvironmentId,
		DeploymentAppType: config.DeploymentAppType,
		ConfigType:        config.ConfigType,
		Active:            config.Active,
		ReleaseMode:       config.ReleaseMode,
		ReleaseConfig:     string(config.ReleaseConfiguration.JSON()),
	}, nil
}

func ConvertDeploymentConfigDbObjToDTO(dbObj *deploymentConfig.DeploymentConfig) (*bean.DeploymentConfig, error) {

	var releaseConfig bean.ReleaseConfiguration
	err := json.Unmarshal([]byte(dbObj.ReleaseConfig), &releaseConfig)
	if err != nil {
		return nil, err
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
