package helper

import (
	bean2 "github.com/devtron-labs/devtron/pkg/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

func GetCmCsAppAndEnvLevelMap(cMCSNamesAppLevel, cMCSNamesEnvLevel []bean.ConfigNameAndType) (map[string]*bean2.ConfigProperty, map[string]*bean2.ConfigProperty) {
	cMCSNamesAppLevelMap, cMCSNamesEnvLevelMap := make(map[string]*bean2.ConfigProperty, len(cMCSNamesAppLevel)), make(map[string]*bean2.ConfigProperty, len(cMCSNamesEnvLevel))

	for _, cmcs := range cMCSNamesAppLevel {
		property := GetConfigProperty(cmcs.Name, cmcs.Type, bean2.PublishedConfigState)
		cMCSNamesAppLevelMap[property.GetKey()] = property
	}
	for _, cmcs := range cMCSNamesEnvLevel {
		property := GetConfigProperty(cmcs.Name, cmcs.Type, bean2.PublishedConfigState)
		cMCSNamesEnvLevelMap[property.GetKey()] = property
	}
	return cMCSNamesAppLevelMap, cMCSNamesEnvLevelMap
}

func GetConfigProperty(name string, configType bean.ResourceType, State bean2.ConfigState) *bean2.ConfigProperty {
	return &bean2.ConfigProperty{
		Name:        name,
		Type:        configType,
		ConfigState: State,
	}
}

//func GetUniqueConfigPropertyList(cmcsKeyVsPropertyMap map[string]*bean2.ConfigProperty, combinedProperties []*bean2.ConfigProperty) []*bean2.ConfigProperty {
//	properties := make([]*bean2.ConfigProperty, 0)
//	if len(cmcsKeyVsPropertyMap) == 0 {
//		return properties
//	}
//
//	combinedNames := make(map[string]bool)
//	for _, config := range combinedProperties {
//		combinedNames[config.GetKey()] = true
//	}
//
//	for key, property := range cmcsKeyVsPropertyMap {
//		if !combinedNames[key] {
//			properties = append(properties, property)
//		} else {
//			cmcsKeyVsPropertyMap[key] =
//		}
//	}
//	return properties
//}

func GetCombinedPropertiesMap(cmcsKeyPropertyAppLevelMap, cmcsKeyPropertyEnvLevelMap map[string]*bean2.ConfigProperty) []*bean2.ConfigProperty {
	combinedPropertiesMap := make(map[string]*bean2.ConfigProperty, len(cmcsKeyPropertyAppLevelMap)+len(cmcsKeyPropertyEnvLevelMap))
	for key, property := range cmcsKeyPropertyAppLevelMap {
		if property.IsConfigPropertyGlobal() {
			// only append global =true in combined for app level cmcs, all overridden cmcs would be handled in cmcsKeyPropertyEnvLevelMap
			combinedPropertiesMap[key] = property
		}
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
