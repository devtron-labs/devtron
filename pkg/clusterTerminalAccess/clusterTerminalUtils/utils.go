package clusterTerminalUtils

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	ServiceAccountNameKey    = "serviceAccountName"
	NodeSelectorKey          = "nodeSelector"
	ContainersKey            = "containers"
	ImageKey                 = "image"
	NamespaceKey             = "namespace"
	MetadataKey              = "metadata"
	NameKey                  = "name"
	SubjectsKey              = "subjects"
	NodeNameKey              = "nodeName"
	Key                      = "key"
	TolerationsKey           = "tolerations"
	TolerationEffectKey      = "effect"
	TolerationOperator       = "operator"
	TolerationOperatorExists = "exists"
	ServiceAccountKind       = "ServiceAccount"
	ClusterRoleBindingKind   = "ClusterRoleBinding"
	PodKind                  = "Pod"
	TerminalNodeDebugPodName = "terminal-node-debug-pod"
)

func IsNodeDebugPod(pod *v1.Pod) bool {
	if pod == nil || len(pod.Spec.Containers) == 0 || len(pod.Spec.Containers[0].VolumeMounts) == 0 || len(pod.Spec.Volumes) == 0 {
		return false
	}
	container := pod.Spec.Containers[0]
	volumeMatch := false
	hostVolumeMountVolumeName := ""
	for _, volumeMount := range container.VolumeMounts {
		if volumeMount.MountPath == "/host" {
			hostVolumeMountVolumeName = volumeMount.Name
		}
	}
	if hostVolumeMountVolumeName == "" {
		return false
	}
	for _, volume := range pod.Spec.Volumes {
		if volume.Name == hostVolumeMountVolumeName && volume.HostPath.Path == "/" {
			volumeMatch = true
		}
	}

	return volumeMatch && container.TTY && container.Stdin &&
		(container.SecurityContext == nil) && (pod.Spec.NodeName == "") &&
		pod.Spec.HostIPC && pod.Spec.HostPID && pod.Spec.HostNetwork

}

func updatePodTemplate(templateDataMap map[string]interface{}, podNameVar string, nodeName string, baseImage string, isAutoSelect bool, taints []models.NodeTaints) (string, error) {
	//adding pod name in metadata
	if val, ok := templateDataMap[MetadataKey]; ok && len(podNameVar) > 0 {
		metadataMap := val.(map[string]interface{})
		if _, ok1 := metadataMap[NameKey]; ok1 {
			metadataMap[NameKey] = interface{}(podNameVar)
		}
	}
	//adding service account and nodeName in pod spec
	if val, ok := templateDataMap["spec"]; ok {
		specMap := val.(map[string]interface{})
		if _, ok1 := specMap[ServiceAccountNameKey]; ok1 && len(podNameVar) > 0 {
			name := fmt.Sprintf("%s-sa", podNameVar)
			specMap[ServiceAccountNameKey] = interface{}(name)
		}
		//TODO: remove the below line after changing pod manifest data in DB
		delete(specMap, NodeSelectorKey)
		if !isAutoSelect {
			specMap[NodeNameKey] = interface{}(nodeName)
		}

		//adding container data in pod spec
		if containers, ok1 := specMap[ContainersKey]; ok1 {
			containersData := containers.([]interface{})
			for _, containerData := range containersData {
				containerDataMap := containerData.(map[string]interface{})
				if _, ok2 := containerDataMap[ImageKey]; ok2 {
					containerDataMap[ImageKey] = interface{}(baseImage)
				}
			}
		}

		//adding pod toleration's for the given node if autoSelect = false
		tolerationData := make([]interface{}, 0)
		if !isAutoSelect {
			for _, taint := range taints {
				toleration := make(map[string]interface{})
				toleration[Key] = interface{}(taint.Key)
				toleration[TolerationOperator] = interface{}(TolerationOperatorExists)
				toleration[TolerationEffectKey] = interface{}(taint.Effect)
				tolerationData = append(tolerationData, interface{}(toleration))
			}
		}
		specMap[TolerationsKey] = interface{}(tolerationData)
	}
	bytes, err := json.Marshal(&templateDataMap)
	return string(bytes), err
}
func updateClusterRoleBindingTemplate(templateDataMap map[string]interface{}, podNameVar string, namespace string) (string, error) {
	if val, ok := templateDataMap[MetadataKey]; ok {
		metadataMap := val.(map[string]interface{})
		if _, ok1 := metadataMap[NameKey]; ok1 {
			name := fmt.Sprintf("%s-crb", podNameVar)
			metadataMap[NameKey] = name
		}
	}

	if subjects, ok := templateDataMap[SubjectsKey]; ok {
		for _, subject := range subjects.([]interface{}) {
			subjectMap := subject.(map[string]interface{})
			if _, ok1 := subjectMap[NameKey]; ok1 {
				name := fmt.Sprintf("%s-sa", podNameVar)
				subjectMap[NameKey] = interface{}(name)
			}

			if _, ok2 := subjectMap[NamespaceKey]; ok2 {
				subjectMap[NamespaceKey] = interface{}(namespace)
			}
		}
	}

	bytes, err := json.Marshal(&templateDataMap)
	return string(bytes), err
}
func updateServiceAccountTemplate(templateDataMap map[string]interface{}, podNameVar string, namespace string) (string, error) {
	if val, ok := templateDataMap[MetadataKey]; ok {
		metadataMap := val.(map[string]interface{})
		if _, ok1 := metadataMap[NameKey]; ok1 {
			name := fmt.Sprintf("%s-sa", podNameVar)
			metadataMap[NameKey] = interface{}(name)
		}

		if _, ok2 := metadataMap[NamespaceKey]; ok2 {
			metadataMap[NamespaceKey] = interface{}(namespace)
		}

	}
	bytes, err := json.Marshal(&templateDataMap)
	return string(bytes), err
}

func ReplaceTemplateData(templateData string, podNameVar string, namespace string, nodeName string, baseImage string, isAutoSelect bool, taints []models.NodeTaints) (string, error) {
	templateDataMap := map[string]interface{}{}
	template := templateData

	err := yaml.Unmarshal([]byte(template), &templateDataMap)
	if err != nil {
		return templateData, err
	}
	if _, ok := templateDataMap["kind"]; ok {
		kind := templateDataMap["kind"]
		if kind == ServiceAccountKind {
			return updateServiceAccountTemplate(templateDataMap, podNameVar, namespace)
		} else if kind == ClusterRoleBindingKind {
			return updateClusterRoleBindingTemplate(templateDataMap, podNameVar, namespace)
		} else if kind == PodKind {
			return updatePodTemplate(templateDataMap, podNameVar, nodeName, baseImage, isAutoSelect, taints)
		}
	}
	return templateData, nil
}
