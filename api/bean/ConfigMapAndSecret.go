/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package bean

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/util"
)

type ConfigMapRootJson struct {
	ConfigMapJson ConfigMapJson `json:"ConfigMaps"`
}
type ConfigMapJson struct {
	Enabled bool              `json:"enabled"`
	Maps    []ConfigSecretMap `json:"maps"`
}

type ConfigSecretRootJson struct {
	ConfigSecretJson ConfigSecretJson `json:"ConfigSecrets"`
}
type ConfigSecretJson struct {
	Enabled bool               `json:"enabled"`
	Secrets []*ConfigSecretMap `json:"secrets"`
}

type ConfigMapAndSecretJson struct {
	ConfigMapJson    ConfigMapJson    `json:"configMapJson"`
	ConfigSecretJson ConfigSecretJson `json:"configSecretJson"`
}

type ConfigSecretMap struct {
	Name           string          `json:"name"`
	Type           string          `json:"type"`
	External       bool            `json:"external"`
	MountPath      string          `json:"mountPath"`
	Data           json.RawMessage `json:"data,omitempty"`
	ESOSecretData  json.RawMessage `json:"esoSecretData,omitempty"`
	ExternalType   string          `json:"externalType"`
	RoleARN        string          `json:"roleARN"`
	SecretData     json.RawMessage `json:"secretData,omitempty"`
	SubPath        bool            `json:"subPath"`
	FilePermission string          `json:"filePermission"`
}

func (configSecret ConfigSecretMap) GetDataMap() (map[string]string, error) {
	var datamap map[string]string
	err := json.Unmarshal(configSecret.Data, &datamap)
	return datamap, err
}
func (configSecretJson ConfigSecretJson) GetDereferencedSecrets() []ConfigSecretMap {
	return util.GetDeReferencedArray(configSecretJson.Secrets)
}

func (configSecretJson *ConfigSecretJson) SetReferencedSecrets(secrets []ConfigSecretMap) {
	configSecretJson.Secrets = util.GetReferencedArray(secrets)
}

func (ConfigSecretRootJson) GetTransformedDataForSecretData(data string, mode util.SecretTransformMode) (string, error) {
	secretsJson := ConfigSecretRootJson{}
	err := json.Unmarshal([]byte(data), &secretsJson)
	if err != nil {
		return "", err
	}

	for _, configData := range secretsJson.ConfigSecretJson.Secrets {
		configData.Data, err = util.GetDecodedAndEncodedData(configData.Data, mode)
		if err != nil {
			return "", err
		}
	}

	marshal, err := json.Marshal(secretsJson)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}
