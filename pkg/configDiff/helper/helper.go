package helper

import (
	bean2 "github.com/devtron-labs/devtron/pkg/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

func GetCmCsAppAndEnvLevelMap(cMCSNamesAppLevel, cMCSNamesEnvLevel []bean.ConfigNameAndType) (map[string]*bean2.ConfigProperty, map[string]*bean2.ConfigProperty) {
	cMCSNamesAppLevelMap, cMCSNamesEnvLevelMap := make(map[string]*bean2.ConfigProperty, len(cMCSNamesAppLevel)), make(map[string]*bean2.ConfigProperty, len(cMCSNamesEnvLevel))

	for _, cmcs := range cMCSNamesAppLevel {
		property := GetConfigProperty(cmcs.Id, cmcs.Name, cmcs.Type, bean2.PublishedConfigState)
		cMCSNamesAppLevelMap[property.GetKey()] = property
	}
	for _, cmcs := range cMCSNamesEnvLevel {
		property := GetConfigProperty(cmcs.Id, cmcs.Name, cmcs.Type, bean2.PublishedConfigState)
		cMCSNamesEnvLevelMap[property.GetKey()] = property
	}
	return cMCSNamesAppLevelMap, cMCSNamesEnvLevelMap
}

func GetConfigProperty(id int, name string, configType bean.ResourceType, State bean2.ConfigState) *bean2.ConfigProperty {
	return &bean2.ConfigProperty{
		Id:          id,
		Name:        name,
		Type:        configType,
		ConfigState: State,
	}
}

func GetCombinedPropertiesMap(cmcsKeyPropertyAppLevelMap, cmcsKeyPropertyEnvLevelMap map[string]*bean2.ConfigProperty) []*bean2.ConfigProperty {
	combinedPropertiesMap := make(map[string]*bean2.ConfigProperty, len(cmcsKeyPropertyAppLevelMap)+len(cmcsKeyPropertyEnvLevelMap))
	for key, property := range cmcsKeyPropertyAppLevelMap {
		combinedPropertiesMap[key] = property
	}
	for key, property := range cmcsKeyPropertyEnvLevelMap {
		combinedPropertiesMap[key] = property
	}
	combinedProperties := make([]*bean2.ConfigProperty, 0, len(cmcsKeyPropertyAppLevelMap)+len(cmcsKeyPropertyEnvLevelMap))
	for _, property := range combinedPropertiesMap {
		combinedProperties = append(combinedProperties, property)
	}
	return combinedProperties
}
