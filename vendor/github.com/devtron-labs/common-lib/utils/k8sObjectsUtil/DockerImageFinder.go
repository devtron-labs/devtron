/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package k8sObjectsUtil

import (
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	appsV1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func ExtractAllDockerImages(manifests []unstructured.Unstructured) ([]string, error) {
	var dockerImages []string
	for _, manifest := range manifests {
		switch manifest.GroupVersionKind() {
		case schema.GroupVersionKind{Group: "", Version: "v1", Kind: k8sCommonBean.PodKind}:
			var pod coreV1.Pod
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(manifest.UnstructuredContent(), &pod)
			if err != nil {
				return nil, err
			}
			dockerImages = append(dockerImages, extractImagesFromPodTemplate(pod.Spec)...)
		case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: k8sCommonBean.DeploymentKind}:
			var deployment appsV1.Deployment
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(manifest.UnstructuredContent(), &deployment)
			if err != nil {
				return nil, err
			}
			dockerImages = append(dockerImages, extractImagesFromPodTemplate(deployment.Spec.Template.Spec)...)
		case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: k8sCommonBean.ReplicaSetKind}:
			var replicaSet appsV1.ReplicaSet
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(manifest.UnstructuredContent(), &replicaSet)
			if err != nil {
				return nil, err
			}
			dockerImages = append(dockerImages, extractImagesFromPodTemplate(replicaSet.Spec.Template.Spec)...)
		case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: k8sCommonBean.StatefulSetKind}:
			var statefulSet appsV1.StatefulSet
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(manifest.UnstructuredContent(), &statefulSet)
			if err != nil {
				return nil, err
			}
			dockerImages = append(dockerImages, extractImagesFromPodTemplate(statefulSet.Spec.Template.Spec)...)
		case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: k8sCommonBean.DaemonSetKind}:
			var daemonSet appsV1.DaemonSet
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(manifest.UnstructuredContent(), &daemonSet)
			if err != nil {
				return nil, err
			}
			dockerImages = append(dockerImages, extractImagesFromPodTemplate(daemonSet.Spec.Template.Spec)...)
		case schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: k8sCommonBean.JobKind}:
			var job batchV1.Job
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(manifest.UnstructuredContent(), &job)
			if err != nil {
				return nil, err
			}
			dockerImages = append(dockerImages, extractImagesFromPodTemplate(job.Spec.Template.Spec)...)
		case schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "CronJob"}:
			var cronJob batchV1.CronJob
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(manifest.UnstructuredContent(), &cronJob)
			if err != nil {
				return nil, err
			}
			dockerImages = append(dockerImages, extractImagesFromPodTemplate(cronJob.Spec.JobTemplate.Spec.Template.Spec)...)
		case schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ReplicationController"}:
			var replicationController coreV1.ReplicationController
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(manifest.UnstructuredContent(), &replicationController)
			if err != nil {
				return nil, err
			}
			dockerImages = append(dockerImages, extractImagesFromPodTemplate(replicationController.Spec.Template.Spec)...)
		case schema.GroupVersionKind{Group: "argoproj.io", Version: "v1alpha1", Kind: "Rollout"}:
			var rolloutSpec map[string]interface{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(manifest.UnstructuredContent(), &rolloutSpec)
			if err != nil {
				return nil, err
			}
			dockerImages = append(dockerImages, extractImagesFromRolloutTemplate(rolloutSpec)...)
		}
	}

	return dockerImages, nil
}

func extractImagesFromPodTemplate(podSpec coreV1.PodSpec) []string {
	var dockerImages []string
	for _, container := range podSpec.Containers {
		dockerImages = append(dockerImages, container.Image)
	}
	for _, initContainer := range podSpec.InitContainers {
		dockerImages = append(dockerImages, initContainer.Image)
	}
	for _, ephContainer := range podSpec.EphemeralContainers {
		dockerImages = append(dockerImages, ephContainer.Image)
	}

	return dockerImages
}

func extractImagesFromRolloutTemplate(rolloutSpec map[string]interface{}) []string {
	var dockerImages []string
	if rolloutSpec != nil && rolloutSpec["spec"] != nil {
		spec := rolloutSpec["spec"].(map[string]interface{})
		if spec != nil && spec["template"] != nil {
			template := spec["template"].(map[string]interface{})
			if template != nil && template["spec"] != nil {
				templateSpec := template["spec"].(map[string]interface{})
				if templateSpec != nil && templateSpec["containers"] != nil {
					containers := templateSpec["containers"].([]interface{})
					for _, item := range containers {
						container := item.(map[string]interface{})
						images := container["image"].(interface{})
						dockerImages = append(dockerImages, images.(string))
					}
				}
			}
		}
	}
	return dockerImages
}
