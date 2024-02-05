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
	cMCSNamesAppLevel, cMCSNamesEnvLevel, err := impl.configMapService.FetchCmCsNamesAppAndEnvLevel(appId, envId)
	if err != nil {
		return nil, err
	}
	combinedProperties := make([]ConfigProperty, 0)
	//App level cm/cs
	combinedProperties = append(combinedProperties, getConfigPropertyList(cMCSNamesAppLevel)...)
	//env level cm/cs
	combinedProperties = append(combinedProperties, getConfigPropertyList(cMCSNamesEnvLevel)...)
	//DT
	combinedProperties = append(combinedProperties, getConfigProperty("", DeploymentTemplate, PublishedConfigState))
	combinedProperties = append(combinedProperties)

	return &ConfigDataResponse{ResourceConfig: combinedProperties}, nil
}
func getConfigPropertyList(cMCSNames []chartConfig.CMCSNames) []ConfigProperty {
	properties := make([]ConfigProperty, 0)
	if len(cMCSNames) == 0 {
		return properties
	}
	for _, name := range cMCSNames {
		// Fill in CM property if the CMName is not empty
		if name.CMName != "" {
			properties = append(properties, getConfigProperty(name.CMName, CM, PublishedConfigState))
		}
		// Fill in CS property if the CSName is not empty
		if name.CSName != "" {
			properties = append(properties, getConfigProperty(name.CSName, CS, PublishedConfigState))

		}
	}
	return properties
}
func getConfigProperty(name string, configType ResourceType, State ConfigState) ConfigProperty {
	return ConfigProperty{
		Name:        name,
		Type:        configType,
		ConfigState: State,
	}
}
