package common

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
)

func ConvertDeploymentConfigDTOToDbObj(config *bean.DeploymentConfig) *deploymentConfig.DeploymentConfig {
	return &deploymentConfig.DeploymentConfig{
		Id:                config.Id,
		AppId:             config.AppId,
		EnvironmentId:     config.EnvironmentId,
		DeploymentAppType: config.DeploymentAppType,
		ConfigType:        config.ConfigType,
		RepoUrl:           config.RepoURL,
		Active:            config.Active,
	}
}

func ConvertDeploymentConfigDbObjToDTO(dbObj *deploymentConfig.DeploymentConfig) *bean.DeploymentConfig {
	return &bean.DeploymentConfig{
		Id:                dbObj.Id,
		AppId:             dbObj.AppId,
		EnvironmentId:     dbObj.EnvironmentId,
		DeploymentAppType: dbObj.DeploymentAppType,
		ConfigType:        dbObj.ConfigType,
		RepoURL:           dbObj.RepoUrl,
		Active:            dbObj.Active,
	}
}
