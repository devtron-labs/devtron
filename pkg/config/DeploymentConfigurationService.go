package config

import (
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
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
		impl.logger.Errorw("error in fetching CM and CS names at app or env level", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	combinedProperties := make([]*ConfigProperty, 0)
	combinedProperties = append(combinedProperties,
		getUniqueConfigPropertyList(cMCSNamesAppLevel, combinedProperties)...)
	combinedProperties = append(combinedProperties,
		getUniqueConfigPropertyList(cMCSNamesEnvLevel, combinedProperties)...)
	combinedProperties = append(combinedProperties,
		getConfigProperty("", bean.DeploymentTemplate, PublishedConfigState))
	return &ConfigDataResponse{ResourceConfig: combinedProperties}, nil
}

func getUniqueConfigPropertyList(cMCSNames []bean.ConfigNameAndType, combinedProperties []*ConfigProperty) []*ConfigProperty {
	properties := make([]*ConfigProperty, 0)
	if len(cMCSNames) == 0 {
		return properties
	}
	combinedNames := make(map[string]bool)
	for _, config := range combinedProperties {
		combinedNames[config.getKey()] = true
	}
	for _, config := range cMCSNames {
		// Fill in CM and CS property
		property := getConfigProperty(config.Name, config.Type, PublishedConfigState)
		if !combinedNames[property.getKey()] {
			properties = append(properties, property)
		}
	}
	return properties
}

func getConfigProperty(name string, configType bean.ResourceType, State ConfigState) *ConfigProperty {
	return &ConfigProperty{
		Name:        name,
		Type:        configType,
		ConfigState: State,
	}
}
