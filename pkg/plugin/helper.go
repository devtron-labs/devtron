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

func paginatePluginParentMetadataFromDb(allPluginParentMetadata []*repository.PluginParentMetadata, size, offset int) []*repository.PluginParentMetadata {
	if size > 0 {
		if offset+size <= len(allPluginParentMetadata) {
			allPluginParentMetadata = allPluginParentMetadata[offset : offset+size]
		} else {
			allPluginParentMetadata = allPluginParentMetadata[offset:]
		}
	}
	return allPluginParentMetadata
}
