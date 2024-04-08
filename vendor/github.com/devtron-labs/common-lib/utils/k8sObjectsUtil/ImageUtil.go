package k8sObjectsUtil

import (
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func getPath(item string, path []string) []string {
	return append(path, item)
}

func ExtractImages(obj unstructured.Unstructured) []string {
	images := make([]string, 0)

	kind := obj.GetKind()
	subPath, ok := commonBean.KindToPath[kind]
	if !ok {
		return images
	}
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
