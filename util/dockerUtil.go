package util

import "strings"

const (
	DEFAULT_TAG_VALUE = "latest"
)

func ExtractImageRepoAndTag(imagePath string) (repo string, tag string) {
	if len(imagePath) == 0 {
		return "", ""
	}
	var containerRepository, containerImageTag string
	lastColonIndex := strings.LastIndex(imagePath, ":")
	if lastColonIndex == -1 {
		containerRepository = imagePath
		containerImageTag = DEFAULT_TAG_VALUE
	} else {
		containerRepository = imagePath[:lastColonIndex]
		containerImageTag = imagePath[lastColonIndex+1:]
	}
	return containerRepository, containerImageTag
}
