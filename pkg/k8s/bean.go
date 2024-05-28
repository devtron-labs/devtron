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

package k8s

import (
	"github.com/devtron-labs/common-lib/utils/k8s"
	client "github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/pkg/k8s/application/bean"
)

type ResourceRequestBean struct {
	AppId                       string                     `json:"appId"`
	AppType                     int                        `json:"appType,omitempty"`        // 0: DevtronApp, 1: HelmApp, 2:ArgoApp
	DeploymentType              int                        `json:"deploymentType,omitempty"` // 0: DevtronApp, 1: HelmApp
	AppIdentifier               *client.AppIdentifier      `json:"-"`
	K8sRequest                  *k8s.K8sRequestBean        `json:"k8sRequest"`
	DevtronAppIdentifier        *bean.DevtronAppIdentifier `json:"-"`         // For Devtron App Resources
	ClusterId                   int                        `json:"clusterId"` // clusterId is used when request is for direct cluster (not for helm release)
	ExternalArgoApplicationName string                     `json:"externalArgoApplicationName,omitempty"`
	Filter                      string                     `json:"filter,omitempty"`
	LabelSelector               []string                   `json:"labelSelector,omitempty"`
	FieldSelector               []string                   `json:"fieldSelector,omitempty"`
}

type LogsDownloadBean struct {
	FileName string `json:"fileName"`
	LogsData string `json:"data"`
}

type BatchResourceResponse struct {
	ManifestResponse *k8s.ManifestResponse
	Err              error
}

type RotatePodResponse struct {
	Responses     []*bean.RotatePodResourceResponse `json:"responses"`
	ContainsError bool                              `json:"containsError"`
}

type RotatePodRequest struct {
	ClusterId int                      `json:"clusterId"`
	Resources []k8s.ResourceIdentifier `json:"resources"`
}
type PodContainerList struct {
	Containers          []string
	InitContainers      []string
	EphemeralContainers []string
}

type ResourceGetResponse struct {
	ManifestResponse *k8s.ManifestResponse `json:"manifestResponse"`
	SecretViewAccess bool                  `json:"secretViewAccess"` // imp: only for resource browser, this is being used to check whether a user can see obscured secret values or not.
}

var (
	ResourceNotFoundErr = "Unable to locate Kubernetes resource."
)
