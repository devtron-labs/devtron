/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package bean

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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

func GetDecodedDataForSecret(data string) (string, error) {
	secretsJson := ConfigSecretRootJson{}
	err := json.Unmarshal([]byte(data), &secretsJson)
	if err != nil {
		return "", err
	}

	for _, configData := range secretsJson.ConfigSecretJson.Secrets {
		configData.SecretData = configData.GetDecodedData()
	}

	marshal, err := json.Marshal(secretsJson)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

func (configSecretMap ConfigSecretMap) GetDecodedData() []byte {
	dataMap := make(map[string]string)
	err := json.Unmarshal(configSecretMap.SecretData, &dataMap)
	if err != nil {
		return nil
	}
	var decodedData []byte
	for k, s := range dataMap {
		decodedData, err = base64.StdEncoding.DecodeString(s)
		if err != nil {
			fmt.Println("Error decoding base64:", err)
		}
		dataMap[k] = string(decodedData)
	}
	marshal, err := json.Marshal(dataMap)
	if err != nil {
		return nil
	}
	return marshal
}

func GetEncodedDataForSecret(data string) (string, error) {
	secretsJson := ConfigSecretRootJson{}
	err := json.Unmarshal([]byte(data), &secretsJson)
	if err != nil {
		return "", err
	}

	for _, configData := range secretsJson.ConfigSecretJson.Secrets {
		configData.SecretData = configData.GetEncodedData()
	}

	marshal, err := json.Marshal(secretsJson)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

func (configSecretMap ConfigSecretMap) GetEncodedData() []byte {
	dataMap := make(map[string]string)
	err := json.Unmarshal(configSecretMap.SecretData, &dataMap)
	if err != nil {
		return nil
	}
	var decodedData []byte
	for k, s := range dataMap {
		decodedData = []byte(base64.StdEncoding.EncodeToString([]byte(s)))
		if err != nil {
			fmt.Println("Error decoding base64:", err)
		}
		dataMap[k] = string(decodedData)
	}
	marshal, err := json.Marshal(dataMap)
	if err != nil {
		return nil
	}
	return marshal
}
