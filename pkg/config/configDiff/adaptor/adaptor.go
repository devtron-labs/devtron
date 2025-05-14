/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package adaptor

import (
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/config/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/adapter"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

func GetConfigProperty(id int, name string, configType bean.ResourceType, State bean2.ConfigState) *bean2.ConfigProperty {
	return &bean2.ConfigProperty{
		Id:          id,
		Name:        name,
		Type:        configType,
		ConfigState: State,
	}
}

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

func ConfigListConvertor(r bean3.ConfigList) bean.ConfigsList {
	pipelineConfigData := make([]*bean.ConfigData, 0, len(r.ConfigData))
	for _, item := range r.ConfigData {
		pipelineConfigData = append(pipelineConfigData, adapter.ConvertConfigDataToPipelineConfigData(item))
	}
	return bean.ConfigsList{ConfigData: pipelineConfigData}
}

func SecretListConvertor(r bean3.SecretList) bean.SecretsList {
	pipelineConfigData := make([]*bean.ConfigData, 0, len(r.ConfigData))
	for _, item := range r.ConfigData {
		pipelineConfigData = append(pipelineConfigData, adapter.ConvertConfigDataToPipelineConfigData(item))
	}
	return bean.SecretsList{ConfigData: pipelineConfigData}
}

func ReverseConfigListConvertor(r bean.ConfigsList) bean3.ConfigList {
	configData := make([]*bean3.ConfigData, 0, len(r.ConfigData))
	for _, item := range r.ConfigData {
		configData = append(configData, adapter.ConvertPipelineConfigDataToConfigData(item))
	}
	return bean3.ConfigList{ConfigData: configData}
}

func ReverseSecretListConvertor(r bean.SecretsList) bean3.SecretList {
	configData := make([]*bean3.ConfigData, 0, len(r.ConfigData))
	for _, item := range r.ConfigData {
		configData = append(configData, adapter.ConvertPipelineConfigDataToConfigData(item))
	}
	return bean3.SecretList{ConfigData: configData}
}
