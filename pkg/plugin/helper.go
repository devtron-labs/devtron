package plugin

import "github.com/devtron-labs/devtron/pkg/plugin/repository"

func getAllUniqueTags(tags []*repository.PluginTag) []string {
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

func paginatePluginParentMetadata(allPluginParentMetadata []*repository.PluginParentMetadata, size, offset int) []*repository.PluginParentMetadata {
	if size > 0 {
		if offset+size <= len(allPluginParentMetadata) {
			allPluginParentMetadata = allPluginParentMetadata[offset : offset+size]
		} else {
			allPluginParentMetadata = allPluginParentMetadata[offset:]
		}
	}
	return allPluginParentMetadata
}

func filterOnlyRequiredPluginVersions(versionIdVsPluginsVersionDetailMap map[int]map[int]*PluginsVersionDetail, pluginVersionsIdMap map[int]bool) {
	for pluginParentId, versionMap := range versionIdVsPluginsVersionDetailMap {
		for pluginVersionId, _ := range versionMap {
			if _, ok := pluginVersionsIdMap[pluginVersionId]; !ok {
				delete(versionIdVsPluginsVersionDetailMap[pluginParentId], pluginVersionId)
			}
		}
	}
}

func getParentPluginDtoMappings(pluginsParentMetadata []*repository.PluginParentMetadata) (map[int]*PluginParentMetadataDto, []int) {
	pluginParentIdVsPluginParentDtoMap := make(map[int]*PluginParentMetadataDto, len(pluginsParentMetadata))
	pluginParentIds := make([]int, 0, len(pluginsParentMetadata))
	for _, metadata := range pluginsParentMetadata {
		pluginParentIds = append(pluginParentIds, metadata.Id)
		if _, ok := pluginParentIdVsPluginParentDtoMap[metadata.Id]; !ok {
			pluginParentDto := NewPluginParentMetadataDto()
			pluginParentDto.WithNameAndId(metadata.Name, metadata.Id).
				WithIcon(metadata.Icon).
				WithDescription(metadata.Description).
				WithPluginIdentifier(metadata.Identifier).
				WithType(string(metadata.Type))
			pluginParentIdVsPluginParentDtoMap[metadata.Id] = pluginParentDto
		}
	}
	return pluginParentIdVsPluginParentDtoMap, pluginParentIds
}

func getPluginVersionAndDetailsMapping(pluginVersionsMetadata []*repository.PluginMetadata, userIdVsEmailMap map[int32]string) map[int]map[int]*PluginsVersionDetail {
	pluginVersionsVsPluginsVersionDetailMap := make(map[int]map[int]*PluginsVersionDetail)
	for _, versionMetadata := range pluginVersionsMetadata {
		pluginVersionDetails := NewPluginsVersionDetail()
		pluginVersionDetails.SetMinimalPluginsVersionDetail(versionMetadata)
		pluginVersionDetails.WithLastUpdatedEmail(userIdVsEmailMap[versionMetadata.UpdatedBy])

		if _, ok := pluginVersionsVsPluginsVersionDetailMap[versionMetadata.PluginParentMetadataId]; !ok {
			pluginVersionsVsPluginsVersionDetailMap[versionMetadata.PluginParentMetadataId] = make(map[int]*PluginsVersionDetail)
		}
		pluginVersionsVsPluginsVersionDetailMap[versionMetadata.PluginParentMetadataId][versionMetadata.Id] = pluginVersionDetails
	}
	return pluginVersionsVsPluginsVersionDetailMap
}

func appendMinimalVersionDetailsInParentMetadataDtos(pluginParentIdVsPluginParentDtoMap map[int]*PluginParentMetadataDto,
	pluginVersionsVsPluginsVersionDetailMap map[int]map[int]*PluginsVersionDetail) {

	for parentPluginId, versionMap := range pluginVersionsVsPluginsVersionDetailMap {
		minimalPluginVersionsMetadataDtos := make([]*PluginsVersionDetail, 0, len(versionMap))
		pluginVersion := NewPluginVersions()
		for _, versionDetail := range versionMap {
			minimalPluginVersionsMetadataDtos = append(minimalPluginVersionsMetadataDtos, versionDetail)
		}
		pluginVersion.WithMinimalPluginVersionData(minimalPluginVersionsMetadataDtos)
		pluginParentIdVsPluginParentDtoMap[parentPluginId].WithVersions(pluginVersion)
	}
}
