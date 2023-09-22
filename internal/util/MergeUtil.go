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

package util

import (
	"encoding/json"
	"slices"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/util"
	jsonpatch "github.com/evanphx/json-patch"
	"go.uber.org/zap"
)

type MergeUtil struct {
	Logger *zap.SugaredLogger
}

/*
//returns json representation of merged values
func (m MergeUtil) MergeOverride(helmValues string, override []byte) ([]byte, error) {
	cf, err := conflate.FromData([]byte(helmValues), override)
	if err != nil {
		m.Logger.Errorw("error in merging config",
			"original", helmValues,
			"override", override,
			"error", err)
		return nil, err
	}
	jsonBytes, err := cf.MarshalJSON()
	if err != nil {
		m.Logger.Errorw("error in marshaling yaml ",
			"cf", cf,
			"error", err)
		return nil, err
	}
	dst := new(bytes.Buffer)
	err = json.Compact(dst, jsonBytes)
	if err != nil {
		return nil, err
	}
	jsonBytes = dst.Bytes()
	m.Logger.Infow("merged config ",
		"original", helmValues,
		"override", override,
		"yaml", jsonBytes,
	)
	return jsonBytes, nil
}

func (m MergeUtil) MergeOverrideVal(data ...[]byte) ([]byte, error) {
	cf, err := conflate.FromData(data...)
	if err != nil {
		m.Logger.Errorw("error in merging config",
			"val", data,
			"error", err)
		return nil, err
	}
	jsonBytes, err := cf.MarshalJSON()
	if err != nil {
		m.Logger.Errorw("error in marshaling yaml ",
			"cf", cf,
			"error", err)
		return nil, err
	}
	dst := new(bytes.Buffer)
	err = json.Compact(dst, jsonBytes)
	if err != nil {
		return nil, err
	}
	jsonBytes = dst.Bytes()
	return jsonBytes, nil
}
*/
//merges two json objects
func (m MergeUtil) JsonPatch(target, patch []byte) (data []byte, err error) {
	data, err = jsonpatch.MergePatch(target, patch)
	if err != nil {
		m.Logger.Debugw("error in merging json ", "target", target, "patch", patch, "err", err)
	}
	return data, err
}

func (m MergeUtil) ConfigMapMerge(appLevelConfigMapJson string, envLevelConfigMapJson string) (data string, err error) {
	appLevelConfigMap := bean.ConfigMapJson{}
	envLevelConfigMap := bean.ConfigMapJson{}
	configResponse := bean.ConfigMapJson{}
	if appLevelConfigMapJson != "" {
		err = json.Unmarshal([]byte(appLevelConfigMapJson), &appLevelConfigMap)
		if err != nil {
			m.Logger.Debugw("error in Unmarshal ", "appLevelConfigMapJson", appLevelConfigMapJson, "envLevelConfigMapJson", envLevelConfigMapJson, "err", err)
		}
	}
	if envLevelConfigMapJson != "" {
		err = json.Unmarshal([]byte(envLevelConfigMapJson), &envLevelConfigMap)
		if err != nil {
			m.Logger.Debugw("error in Unmarshal ", "appLevelConfigMapJson", appLevelConfigMapJson, "envLevelConfigMapJson", envLevelConfigMapJson, "err", err)
		}
	}
	if len(appLevelConfigMap.Maps) > 0 || len(envLevelConfigMap.Maps) > 0 {
		configResponse.Enabled = true
	}

	configResponse.Maps = mergeConfigMapsAndSecrets(envLevelConfigMap.Maps, appLevelConfigMap.Maps)
	byteData, err := json.Marshal(configResponse)
	if err != nil {
		m.Logger.Debugw("error in marshal ", "err", err)
	}
	return string(byteData), err
}

func (m MergeUtil) ConfigSecretMerge(appLevelSecretJson string, envLevelSecretJson string, chartMajorVersion int, chartMinorVersion int, isJob bool) (data string, err error) {
	appLevelSecret := bean.ConfigSecretJson{}
	envLevelSecret := bean.ConfigSecretJson{}
	secretResponse := bean.ConfigSecretJson{}
	var finalMaps []bean.ConfigSecretMap
	if appLevelSecretJson != "" {
		err = json.Unmarshal([]byte(appLevelSecretJson), &appLevelSecret)
		if err != nil {
			m.Logger.Debugw("error in Unmarshal ", "appLevelSecretJson", appLevelSecretJson, "envLevelSecretJson", envLevelSecretJson, "err", err)
		}
	}
	if envLevelSecretJson != "" {
		err = json.Unmarshal([]byte(envLevelSecretJson), &envLevelSecret)
		if err != nil {
			m.Logger.Debugw("error in Unmarshal ", "appLevelSecretJson", appLevelSecretJson, "envLevelSecretJson", envLevelSecretJson, "err", err)
		}
	}
	if len(appLevelSecret.Secrets) > 0 || len(envLevelSecret.Secrets) > 0 {
		secretResponse.Enabled = true
	}

	finalMaps = mergeConfigMapsAndSecrets(envLevelSecret.GetDereferenceSecrets(), appLevelSecret.GetDereferenceSecrets())
	for _, finalMap := range finalMaps {
		finalMap = m.processExternalSecrets(finalMap, chartMajorVersion, chartMinorVersion, isJob)
	}
	secretResponse.SetReferenceSecrets(finalMaps)
	byteData, err := json.Marshal(secretResponse)
	if err != nil {
		m.Logger.Debugw("error in marshal ", "err", err)
	}
	return string(byteData), err
}

func mergeConfigMapsAndSecrets(envLevelCMCS []bean.ConfigSecretMap, appLevelSecretCMCS []bean.ConfigSecretMap) []bean.ConfigSecretMap {
	envSecretNames := make([]string, 0)
	var finalMaps []bean.ConfigSecretMap
	for _, item := range envLevelCMCS {
		envSecretNames = append(envSecretNames, item.Name)
	}
	for i, _ := range appLevelSecretCMCS {
		//else ignoring this value as override from configB
		if !slices.Contains(envSecretNames, appLevelSecretCMCS[i].Name) {
			finalMaps = append(finalMaps, appLevelSecretCMCS[i])
		}
	}
	for i, _ := range envLevelCMCS {
		finalMaps = append(finalMaps, envLevelCMCS[i])
	}
	return finalMaps
}

func (m MergeUtil) processExternalSecrets(secret bean.ConfigSecretMap, chartMajorVersion int, chartMinorVersion int, isJob bool) bean.ConfigSecretMap {
	if secret.ExternalType == util.AWSSecretsManager || secret.ExternalType == util.AWSSystemManager || secret.ExternalType == util.HashiCorpVault {
		if secret.SecretData != nil && ((chartMajorVersion <= 3 && chartMinorVersion < 8) || isJob) {
			var es []map[string]interface{}
			esNew := make(map[string]interface{})
			err := json.Unmarshal(secret.SecretData, &es)
			if err != nil {
				m.Logger.Debugw("error in Unmarshal ", "SecretData", secret.SecretData, "external secret", es, "err", err)
			}
			for _, item := range es {
				keyProp := item["name"].(string)
				valueProp := item["key"]
				esNew[keyProp] = valueProp
			}
			byteData, err := json.Marshal(esNew)
			if err != nil {
				m.Logger.Debugw("error in marshal ", "err", err)
			}
			secret.Data = byteData
			secret.SecretData = nil
		}
	}
	return secret
}
