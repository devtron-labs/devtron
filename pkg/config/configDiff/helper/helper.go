package helper

import (
	"encoding/json"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/config/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/config/configDiff/utils"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
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

func GetKeysToDelete(cmcsData map[string]*bean3.ConfigData, resourceName string) []string {
	keysToDelete := make([]string, 0, len(cmcsData))
	for key, _ := range cmcsData {
		if key != resourceName {
			keysToDelete = append(keysToDelete, key)
		}
	}
	return keysToDelete
}

func FilterOutMergedCmCsForResourceName(cmcsMerged *bean2.CmCsMetadataDto, resourceName string, resourceType bean.ResourceType) {
	for _, key := range GetKeysToDelete(cmcsMerged.SecretMap, resourceName) {
		delete(cmcsMerged.SecretMap, key)
	}
	for _, key := range GetKeysToDelete(cmcsMerged.CmMap, resourceName) {
		delete(cmcsMerged.CmMap, key)
	}

	// handle the case when a cm and a cs can have a same name, in that case, check from resource type if correct key is filtered out or not
	if resourceType == bean.CS {
		if len(cmcsMerged.CmMap) > 0 {
			// delete all elements from cmMap as requested resource is of secret type
			for key, _ := range cmcsMerged.CmMap {
				delete(cmcsMerged.CmMap, key)
			}
		}
	} else if resourceType == bean.CM {
		if len(cmcsMerged.SecretMap) > 0 {
			// delete all elements from secretMap as requested resource is of secret type
			for key, _ := range cmcsMerged.SecretMap {
				delete(cmcsMerged.SecretMap, key)
			}
		}
	}
}

func GetConfigDataRequestJsonRawMessage(configDataList []*bean.ConfigData) (json.RawMessage, error) {
	configDataReq := &bean.ConfigDataRequest{ConfigData: configDataList}
	configDataJson, err := utils.ConvertToJsonRawMessage(configDataReq)
	if err != nil {
		return nil, err
	}
	return configDataJson, nil
}

func GetJsonRawMessageFromConfigDataRequest(configDataList []*bean.ConfigData) (json.RawMessage, error) {
	configDataReq := &bean.ConfigDataRequest{ConfigData: configDataList}
	// doing this because FE wants string version of resolved value
	strReq, err := utils.ConvertToString(configDataReq)
	if err != nil {
		return nil, err
	}
	configDataJson, err := utils.ConvertToJsonRawMessage(strReq)
	if err != nil {
		return nil, err
	}
	return configDataJson, nil
}
