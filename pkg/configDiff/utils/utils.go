package utils

import (
	"encoding/json"
	bean2 "github.com/devtron-labs/devtron/pkg/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"strings"
)

func ConvertToJsonRawMessage(request interface{}) (json.RawMessage, error) {
	var r json.RawMessage
	configMapByte, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	err = r.UnmarshalJSON(configMapByte)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func ConvertToString(req interface{}) (string, error) {
	reqByte, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return string(reqByte), nil
}

/*
GetKeyValMapForSecretConfigDataAndMaskData
1. unmarshall secret data
2. prepare secret's key val map
3. create new masked secret data and replace to original data
*/
func GetKeyValMapForSecretConfigDataAndMaskData(configDataList []*bean.ConfigData) (map[string]map[string]string, error) {
	keyValMapForSecretConfig := make(map[string]map[string]string)
	for _, secretConfigData := range configDataList {
		if secretConfigData.IsESOExternalSecretType() || secretConfigData.External {
			continue
		}
		secretRawData := secretConfigData.Data
		if secretConfigData.Global {
			secretRawData = secretConfigData.DefaultData
		}
		var secretData map[string]string
		if err := json.Unmarshal(secretRawData, &secretData); err != nil {
			return nil, err
		}
		newMaskedSecretData := make(map[string]string, len(secretData))
		for key, val := range secretData {
			if keyValMapForSecretConfig[secretConfigData.Name] == nil {
				keyValMapForSecretConfig[secretConfigData.Name] = make(map[string]string)
			}
			keyValMapForSecretConfig[secretConfigData.Name][key] = val
			newMaskedSecretData[key] = bean2.SecretMaskedValue
		}
		maskedSecretJson, err := json.Marshal(newMaskedSecretData)
		if err != nil {
			return nil, err
		}
		if secretConfigData.Global {
			secretConfigData.DefaultData = maskedSecretJson
		} else {
			secretConfigData.Data = maskedSecretJson
		}
	}
	return keyValMapForSecretConfig, nil
}

/*
CompareAndMaskSecretValuesInConfigData
1.unmarshall secrets data
2. mask secret values based on some checks
3. marshall masked secret and replace original configData
*/
func CompareAndMaskSecretValuesInConfigData(configDataList []*bean.ConfigData, keyValMapForSecretConfig1 map[string]map[string]string) error {
	for _, secretConfigData := range configDataList {
		if secretConfigData.IsESOExternalSecretType() || secretConfigData.External {
			continue
		}
		secretConfig := secretConfigData.Data
		if secretConfigData.Global {
			secretConfig = secretConfigData.DefaultData
		}
		var secretDataMap map[string]string
		if err := json.Unmarshal(secretConfig, &secretDataMap); err != nil {
			return err
		}
		if _, ok := keyValMapForSecretConfig1[secretConfigData.Name]; ok {
			newMaskedSecretData := make(map[string]string, len(secretDataMap))
			for key, val := range secretDataMap {
				if val1, ok := keyValMapForSecretConfig1[secretConfigData.Name][key]; ok {
					if strings.Compare(val, val1) == 0 {
						newMaskedSecretData[key] = bean2.SecretMaskedValue
					} else {
						//same key name exists with diff value, mask this with SecretMaskedValueLong (i.e. "************")
						newMaskedSecretData[key] = bean2.SecretMaskedValueLong
					}
				} else {
					newMaskedSecretData[key] = bean2.SecretMaskedValue
				}
			}
			maskedSecretJson, err := json.Marshal(newMaskedSecretData)
			if err != nil {
				return err
			}
			if secretConfigData.Global {
				secretConfigData.DefaultData = maskedSecretJson
			} else {
				secretConfigData.Data = maskedSecretJson
			}

		} else {
			//mask all the secret values with SecretMaskedValue(i.e. "********")
			newMaskedSecretData := make(map[string]string, len(secretDataMap))
			for key, _ := range secretDataMap {
				newMaskedSecretData[key] = bean2.SecretMaskedValue
			}
			maskedSecretJson, err := json.Marshal(newMaskedSecretData)
			if err != nil {
				return err
			}
			if secretConfigData.Global {
				secretConfigData.DefaultData = maskedSecretJson
			} else {
				secretConfigData.Data = maskedSecretJson
			}

		}
	}
	return nil
}
