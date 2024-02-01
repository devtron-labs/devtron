package config

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"go.uber.org/zap"
)

type DeploymentConfigurationService interface {
	ConfigAutoComplete(appId int, envId int) (*ConfigDataResponse, error)
}

type DeploymentConfigurationServiceImpl struct {
	logger              *zap.SugaredLogger
	configMapRepository chartConfig.ConfigMapRepository
}

func NewDeploymentConfigurationServiceImpl(logger *zap.SugaredLogger, configMapRepository chartConfig.ConfigMapRepository) (*DeploymentConfigurationServiceImpl, error) {
	deploymentConfigurationService := &DeploymentConfigurationServiceImpl{
		logger:              logger,
		configMapRepository: configMapRepository,
	}

	return deploymentConfigurationService, nil
}

func (impl DeploymentConfigurationServiceImpl) ConfigAutoComplete(appId int, envId int) (*ConfigDataResponse, error) {
	var configDataResponse *ConfigDataResponse
	var cMCSNamesEnvLevel []chartConfig.CMCSNames
	cMCSNamesAppLevel, err := impl.configMapRepository.GetConfigNamesAppLevel(appId)
	if err != nil {
		return nil, err
	}
	if envId > 0 {
		cMCSNamesEnvLevel, err = impl.configMapRepository.GetConfigNamesEnvLevel(appId, envId)
	}
	configDataResponse = setConfigDataResponse(cMCSNamesAppLevel)
	configDataResponse = setConfigDataResponse(cMCSNamesEnvLevel)

	return configDataResponse, nil
}

func setConfigDataResponse(cMCSNames []chartConfig.CMCSNames) *ConfigDataResponse {
	var configDataResponse *ConfigDataResponse
	for i, name := range cMCSNames {
		configDataResponse.ResourceConfig[i].Name = name.CMName
		configDataResponse.ResourceConfig[i].Type = CM
		configDataResponse.ResourceConfig[i].ConfigState = PublishedConfigState
	}
	for i, name := range cMCSNames {
		configDataResponse.ResourceConfig[i].Name = name.CSName
		configDataResponse.ResourceConfig[i].Type = CS
		configDataResponse.ResourceConfig[i].ConfigState = PublishedConfigState
	}
	return configDataResponse
}
