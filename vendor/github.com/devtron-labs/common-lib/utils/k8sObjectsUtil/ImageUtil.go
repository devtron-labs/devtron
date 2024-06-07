package k8sObjectsUtil

import (
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	yamlUtil "github.com/devtron-labs/common-lib/utils/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func getPath(item string, path []string) []string {
	return append(path, item)
}

func ExtractImages(obj unstructured.Unstructured) []string {
	images := make([]string, 0)

	subPath := commonBean.GetContainerSubPathForKind(obj.GetKind())
	allContainers := make([]interface{}, 0)
	containers, _, _ := unstructured.NestedSlice(obj.Object, getPath(commonBean.Containers, subPath)...)
	if len(containers) > 0 {
		allContainers = append(allContainers, containers...)
	}
	iContainers, _, _ := unstructured.NestedSlice(obj.Object, getPath(commonBean.InitContainers, subPath)...)
	if len(iContainers) > 0 {
		allContainers = append(allContainers, iContainers...)
	}
	ephContainers, _, _ := unstructured.NestedSlice(obj.Object, getPath(commonBean.EphemeralContainers, subPath)...)
	if len(ephContainers) > 0 {
		allContainers = append(allContainers, ephContainers...)
	}
	for _, container := range allContainers {
		containerMap := container.(map[string]interface{})
		if image, ok := containerMap[commonBean.Image].(string); ok {
			images = append(images, image)
		}
	}
	return images
}

func ExtractImageFromManifestYaml(manifestYaml string) []string {
	var dockerImagesFinal []string
	parsedManifests, err := yamlUtil.SplitYAMLs([]byte(manifestYaml))
	if err != nil {

		return dockerImagesFinal
	}
	for _, manifest := range parsedManifests {
		dockerImages := ExtractImages(manifest)
		dockerImagesFinal = append(dockerImagesFinal, dockerImages...)
	}
	return dockerImagesFinal
}
