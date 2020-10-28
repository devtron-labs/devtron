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
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/util"
	"encoding/json"
	"github.com/evanphx/json-patch"
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
	commonMaps := map[string]bean.Map{}
	var finalMaps []bean.Map
	if len(appLevelConfigMapJson) != 0 {
		err = json.Unmarshal([]byte(appLevelConfigMapJson), &appLevelConfigMap)
		if err != nil {
			m.Logger.Debugw("error in Unmarshal ", "appLevelConfigMapJson", appLevelConfigMapJson, "envLevelConfigMapJson", envLevelConfigMapJson, "err", err)
		}
	}
	if len(envLevelConfigMapJson) != 0 {
		err = json.Unmarshal([]byte(envLevelConfigMapJson), &envLevelConfigMap)
		if err != nil {
			m.Logger.Debugw("error in Unmarshal ", "appLevelConfigMapJson", appLevelConfigMapJson, "envLevelConfigMapJson", envLevelConfigMapJson, "err", err)
		}
	}
	if len(appLevelConfigMap.Maps) > 0 || len(envLevelConfigMap.Maps) > 0 {
		configResponse.Enabled = true
	}

	for _, item := range envLevelConfigMap.Maps {
		commonMaps[item.Name] = item
	}
	for _, item := range appLevelConfigMap.Maps {
		if _, ok := commonMaps[item.Name]; ok {
			//ignoring this value as override from configB
		} else {
			commonMaps[item.Name] = item
		}
	}
	for _, v := range commonMaps {
		finalMaps = append(finalMaps, v)
	}
	configResponse.Maps = finalMaps
	byteData, err := json.Marshal(configResponse)
	if err != nil {
		m.Logger.Debugw("error in marshal ", "err", err)
	}
	return string(byteData), err
}

func (m MergeUtil) ConfigSecretMerge(appLevelSecretJson string, envLevelSecretJson string, chartMajorVersion int, chartMinorVersion int) (data string, err error) {
	appLevelSecret := bean.ConfigSecretJson{}
	envLevelSecret := bean.ConfigSecretJson{}
	secretResponse := bean.ConfigSecretJson{}
	commonSecrets := map[string]*bean.Map{}
	var finalMaps []*bean.Map
	if len(appLevelSecretJson) != 0 {
		err = json.Unmarshal([]byte(appLevelSecretJson), &appLevelSecret)
		if err != nil {
			m.Logger.Debugw("error in Unmarshal ", "appLevelSecretJson", appLevelSecretJson, "envLevelSecretJson", envLevelSecretJson, "err", err)
		}
	}
	if len(envLevelSecretJson) != 0 {
		err = json.Unmarshal([]byte(envLevelSecretJson), &envLevelSecret)
		if err != nil {
			m.Logger.Debugw("error in Unmarshal ", "appLevelSecretJson", appLevelSecretJson, "envLevelSecretJson", envLevelSecretJson, "err", err)
		}
	}
	if len(appLevelSecret.Secrets) > 0 || len(envLevelSecret.Secrets) > 0 {
		secretResponse.Enabled = true
	}

	for _, item := range envLevelSecret.Secrets {
		commonSecrets[item.Name] = item
	}
	for _, item := range appLevelSecret.Secrets {
		//else ignoring this value as override from configB
		if _, ok := commonSecrets[item.Name]; !ok {
			commonSecrets[item.Name] = item
		}
	}

	for _, item := range commonSecrets {
		if item.ExternalType == util.AWSSecretsManager || item.ExternalType == util.AWSSystemManager || item.ExternalType == util.HashiCorpVault {
			if item.SecretData != nil && chartMajorVersion <= 3 && chartMinorVersion < 8 {
				var es []map[string]interface{}
				esNew := make(map[string]interface{})
				err = json.Unmarshal(item.SecretData, &es)
				if err != nil {
					m.Logger.Debugw("error in Unmarshal ", "appLevelSecretJson", appLevelSecretJson, "envLevelSecretJson", envLevelSecretJson, "err", err)
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
				item.Data = byteData
				item.SecretData = nil
			}
		}
	}

	for _, v := range commonSecrets {
		finalMaps = append(finalMaps, v)
	}
	secretResponse.Secrets = finalMaps
	byteData, err := json.Marshal(secretResponse)
	if err != nil {
		m.Logger.Debugw("error in marshal ", "err", err)
	}
	return string(byteData), err
}
