package config

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
)

type DeploymentConfigurationService interface {
	ConfigAutoComplete(appId int, envId int) (*ConfigDataResponse, error)
}

type DeploymentConfigurationServiceImpl struct {
	logger           *zap.SugaredLogger
	configMapService pipeline.ConfigMapService
}

func NewDeploymentConfigurationServiceImpl(logger *zap.SugaredLogger,
	configMapService pipeline.ConfigMapService,
) (*DeploymentConfigurationServiceImpl, error) {
	deploymentConfigurationService := &DeploymentConfigurationServiceImpl{
		logger:           logger,
		configMapService: configMapService,
	}

	return deploymentConfigurationService, nil
}

func (impl *DeploymentConfigurationServiceImpl) ConfigAutoComplete(appId int, envId int) (*ConfigDataResponse, error) {
	var configDataResponse *ConfigDataResponse
	cMCSNamesAppLevel, cMCSNamesEnvLevel, err := impl.configMapService.FetchCmCsNamesAppAndEnvLevel(appId, envId)
	if err != nil {
		return nil, err
	}
	configDataResponse = setConfigDataResponse(cMCSNamesAppLevel, configDataResponse)
	configDataResponse = setConfigDataResponse(cMCSNamesEnvLevel, configDataResponse)

	return configDataResponse, nil
}

func setConfigDataResponse(cMCSNames []chartConfig.CMCSNames, configDataResponse *ConfigDataResponse) *ConfigDataResponse {
	if cMCSNames == nil {
		return configDataResponse
	}
	configDataResponse = &ConfigDataResponse{}
	for _, name := range cMCSNames {
		if name.CMName != "" {
			// Fill in CM data if the CMName is not empty
			cmConfig := setConfigProperty(name.CMName, CM, PublishedConfigState)
			configDataResponse.ResourceConfig = append(configDataResponse.ResourceConfig, cmConfig)
		}
		if name.CSName != "" {
			// Fill in CS data if the CSName is not empty
			csConfig := setConfigProperty(name.CSName, CS, PublishedConfigState)
			configDataResponse.ResourceConfig = append(configDataResponse.ResourceConfig, csConfig)
		}

	}
	return configDataResponse
}
func setConfigProperty(name string, configType ResourceType, State ConfigState) ConfigProperty {
	return ConfigProperty{
		Name:        name,
		Type:        configType,
		ConfigState: State,
	}
}
