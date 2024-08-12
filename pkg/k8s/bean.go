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
	helmBean "github.com/devtron-labs/devtron/api/helm-app/service/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/fluxApplication/bean"
	"github.com/devtron-labs/devtron/pkg/k8s/application/bean"
)

type ResourceRequestBean struct {
	AppId                       string                     `json:"appId"`
	AppType                     int                        `json:"appType,omitempty"`        // 0: DevtronApp, 1: HelmApp, 2:ArgoApp, 3 fluxApp
	DeploymentType              int                        `json:"deploymentType,omitempty"` // 0: DevtronApp, 1: HelmApp
	AppIdentifier               *helmBean.AppIdentifier    `json:"-"`
	K8sRequest                  *k8s.K8sRequestBean        `json:"k8sRequest"`
	DevtronAppIdentifier        *bean.DevtronAppIdentifier `json:"-"`         // For Devtron App Resources
	ClusterId                   int                        `json:"clusterId"` // clusterId is used when request is for direct cluster (not for helm release)
	ExternalArgoApplicationName string                     `json:"externalArgoApplicationName,omitempty"`
	ExternalFluxAppIdentifier   *bean2.FluxAppIdentifier   `json: "-"`
}

func (r *ResourceRequestBean) IsValidAppType() bool {
	return r.AppType == bean.DevtronAppType || r.AppType == bean.HelmAppType || r.AppType == bean.ArgoAppType || r.AppType == bean.FluxAppType
}

func (r *ResourceRequestBean) IsValidDeploymentType() bool {
	return r.DeploymentType == bean.HelmInstalledType || r.DeploymentType == bean.ArgoInstalledType || r.DeploymentType == bean.FluxInstalledType
}

type CmCsRequestBean struct {
	clusterId      int
	namespace      string
	externalCmList []string
	externalCsList []string
}

func (req *CmCsRequestBean) SetClusterId(clusterId int) *CmCsRequestBean {
	req.clusterId = clusterId
	return req
}

func (req *CmCsRequestBean) SetNamespace(namespace string) *CmCsRequestBean {
	req.namespace = namespace
	return req
}

func (req *CmCsRequestBean) SetExternalCmList(externalCmList ...string) *CmCsRequestBean {
	if len(externalCmList) == 0 {
		return req
	}
	req.externalCmList = append(req.externalCmList, externalCmList...)
	return req
}

func (req *CmCsRequestBean) SetExternalCsList(externalCsList ...string) *CmCsRequestBean {
	if len(externalCsList) == 0 {
		return req
	}
	req.externalCsList = append(req.externalCsList, externalCsList...)
	return req
}

func (req *CmCsRequestBean) GetClusterId() int {
	return req.clusterId
}

func (req *CmCsRequestBean) GetNamespace() string {
	return req.namespace
}

func (req *CmCsRequestBean) GetExternalCmList() []string {
	return req.externalCmList
}

func (req *CmCsRequestBean) GetExternalCsList() []string {
	return req.externalCsList
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
