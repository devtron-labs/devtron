package helper

import (
	bean2 "github.com/devtron-labs/devtron/pkg/configDiff/bean"
)

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
