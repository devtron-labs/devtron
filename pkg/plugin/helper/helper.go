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

package helper

import (
	"github.com/devtron-labs/devtron/pkg/plugin/bean"
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/plugin/utils"
)

func GetAllUniqueTags(tags []*repository.PluginTag) []string {
	uniqueTagsMap := make(map[string]bool, len(tags))
	for _, tag := range tags {
		if _, ok := uniqueTagsMap[tag.Name]; ok {
			continue
		} else {
			uniqueTagsMap[tag.Name] = true
		}
	}

	uniqueTags := make([]string, 0, len(uniqueTagsMap))
	for tagName, _ := range uniqueTagsMap {
		uniqueTags = append(uniqueTags, tagName)
	}
	return uniqueTags
}

func PaginatePluginParentMetadata(allPluginParentMetadata []*repository.PluginParentMetadata, size, offset int) []*repository.PluginParentMetadata {
	if size > 0 {
		if offset+size <= len(allPluginParentMetadata) {
			allPluginParentMetadata = allPluginParentMetadata[offset : offset+size]
		} else {
			allPluginParentMetadata = allPluginParentMetadata[offset:]
		}
	}
	return allPluginParentMetadata
}

func GetParentPluginDtoMappings(pluginsParentMetadata []*repository.PluginParentMetadata) (map[int]*bean.PluginParentMetadataDto, map[int]bool) {
	pluginParentIdVsPluginParentDtoMap := make(map[int]*bean.PluginParentMetadataDto, len(pluginsParentMetadata))
	pluginIdMap := make(map[int]bool, len(pluginsParentMetadata))

	for _, metadata := range pluginsParentMetadata {
		if _, ok := pluginIdMap[metadata.Id]; !ok {
			pluginIdMap[metadata.Id] = true
		}
		if _, ok := pluginParentIdVsPluginParentDtoMap[metadata.Id]; !ok {
			pluginParentDto := bean.NewPluginParentMetadataDto()
			pluginParentDto.WithNameAndId(metadata.Name, metadata.Id).
				WithIcon(metadata.Icon).
				WithDescription(metadata.Description).
				WithPluginIdentifier(metadata.Identifier).
				WithType(string(metadata.Type))
			pluginParentIdVsPluginParentDtoMap[metadata.Id] = pluginParentDto
		}
	}
	return pluginParentIdVsPluginParentDtoMap, pluginIdMap
}

func GetPluginVersionAndDetailsMapping(pluginVersionsMetadata []*repository.PluginMetadata, userIdVsEmailMap map[int32]string) map[int]map[int]*bean.PluginsVersionDetail {
	pluginVersionsVsPluginsVersionDetailMap := make(map[int]map[int]*bean.PluginsVersionDetail)
	for _, versionMetadata := range pluginVersionsMetadata {
		pluginVersionDetails := bean.NewPluginsVersionDetail()
		pluginVersionDetails.SetMinimalPluginsVersionDetail(versionMetadata)
		pluginVersionDetails.WithLastUpdatedEmail(userIdVsEmailMap[versionMetadata.UpdatedBy])
		pluginVersionDetails.WithCreatedOn(versionMetadata.CreatedOn)

		if _, ok := pluginVersionsVsPluginsVersionDetailMap[versionMetadata.PluginParentMetadataId]; !ok {
			pluginVersionsVsPluginsVersionDetailMap[versionMetadata.PluginParentMetadataId] = make(map[int]*bean.PluginsVersionDetail)
		}
		pluginVersionsVsPluginsVersionDetailMap[versionMetadata.PluginParentMetadataId][versionMetadata.Id] = pluginVersionDetails
	}
	return pluginVersionsVsPluginsVersionDetailMap
}

func AppendMinimalVersionDetailsInParentMetadataDtos(pluginParentIdVsPluginParentDtoMap map[int]*bean.PluginParentMetadataDto,
	pluginVersionsVsPluginsVersionDetailMap map[int]map[int]*bean.PluginsVersionDetail) {

	for parentPluginId, versionMap := range pluginVersionsVsPluginsVersionDetailMap {
		minimalPluginVersionsMetadataDtos := make([]*bean.PluginsVersionDetail, 0, len(versionMap))
		pluginVersion := bean.NewPluginVersions()
		for _, versionDetail := range versionMap {
			minimalPluginVersionsMetadataDtos = append(minimalPluginVersionsMetadataDtos, versionDetail)
		}
		utils.SortPluginsVersionDetailSliceByCreatedOn(minimalPluginVersionsMetadataDtos)

		pluginVersion.WithMinimalPluginVersionData(minimalPluginVersionsMetadataDtos)
		pluginParentIdVsPluginParentDtoMap[parentPluginId].WithVersions(pluginVersion)
	}
}

func GetPluginVersionAndParentPluginIdsMap(pluginVersionIds, parentPluginIds []int) (map[int]bool, map[int]bool) {
	pluginVersionIdsMap, parentPluginIdsMap := make(map[int]bool, len(pluginVersionIds)), make(map[int]bool, len(parentPluginIds))
	for _, item := range pluginVersionIds {
		pluginVersionIdsMap[item] = true
	}
	for _, item := range parentPluginIds {
		parentPluginIdsMap[item] = true
	}
	return pluginVersionIdsMap, parentPluginIdsMap
}

func GetPluginVersionsMetadataByVersionAndParentPluginIds(pluginVersionsMetadata []*repository.PluginMetadata, pluginVersionIdsMap,
	parentPluginIdsMap map[int]bool) []*repository.PluginMetadata {

	filteredPluginVersionMetadata := make([]*repository.PluginMetadata, 0, len(pluginVersionIdsMap)+len(parentPluginIdsMap))
	for _, pluginVersion := range pluginVersionsMetadata {
		if len(pluginVersionIdsMap) > 0 {
			if _, ok := pluginVersionIdsMap[pluginVersion.Id]; ok {
				filteredPluginVersionMetadata = append(filteredPluginVersionMetadata, pluginVersion)
			}
		}
		if len(parentPluginIdsMap) > 0 {
			if _, ok := parentPluginIdsMap[pluginVersion.PluginParentMetadataId]; ok {
				filteredPluginVersionMetadata = append(filteredPluginVersionMetadata, pluginVersion)
			}
		}
	}
	return filteredPluginVersionMetadata
}

type IdGetter interface {
	// ID gets id field from struct wherever this interface is  implemented, see PluginStep sql object for example
	ID() int
}

// GetIDs is a generic function to get IDs
func GetIDs[T IdGetter](items []T) []int {
	ids := make([]int, len(items))
	for i, item := range items {
		ids[i] = item.ID()
	}
	return ids
}

func GetPluginStepIdVsPluginStepVariablesMap(pluginStepVariables []*bean.PluginVariableDto) map[int][]*bean.PluginVariableDto {
	pluginStepIdVsPluginStepVariablesMap := make(map[int][]*bean.PluginVariableDto, len(pluginStepVariables))
	for _, stepVar := range pluginStepVariables {
		pluginStepIdVsPluginStepVariablesMap[stepVar.PluginStepId] = append(pluginStepIdVsPluginStepVariablesMap[stepVar.PluginStepId], stepVar)
	}
	return pluginStepIdVsPluginStepVariablesMap
}

func GetScriptIdVsScriptArgsDetailsMap(scriptArgDetails []*bean.ScriptPathArgPortMapping) map[int][]*bean.ScriptPathArgPortMapping {
	scriptIdVsScriptArgsDetailsMap := make(map[int][]*bean.ScriptPathArgPortMapping, len(scriptArgDetails))
	for _, scriptArgDetail := range scriptArgDetails {
		scriptIdVsScriptArgsDetailsMap[scriptArgDetail.ScriptId] = append(scriptIdVsScriptArgsDetailsMap[scriptArgDetail.ScriptId], scriptArgDetail)
	}
	return scriptIdVsScriptArgsDetailsMap
}

func GetScriptIdList(stepsDto []*bean.PluginStepsDto) []int {
	scriptIds := make([]int, 0, len(stepsDto))
	for _, step := range stepsDto {
		scriptIds = append(scriptIds, step.ScriptId)
	}
	return scriptIds
}

func GetScriptIdVsPluginScript(pluginScripts []*bean.PluginPipelineScript) map[int]*bean.PluginPipelineScript {
	scriptIdVsPluginScript := make(map[int]*bean.PluginPipelineScript, len(pluginScripts))
	for _, pluginScript := range pluginScripts {
		scriptIdVsPluginScript[pluginScript.Id] = pluginScript
	}
	return scriptIdVsPluginScript
}
