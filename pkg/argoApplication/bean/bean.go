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

package bean

import (
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ArgoGroup                    = "argoproj.io"
	ArgoApplicationKind          = "Application"
	VersionV1Alpha1              = "v1alpha1"
	AllNamespaces                = ""
	DevtronCDNamespae            = "devtroncd"
	ArgoLabelForManagedResources = "app.kubernetes.io/instance"
)

const (
	Server      = "server"
	Destination = "destination"
	Config      = "config"
)

var GvkForArgoApplication = schema.GroupVersionKind{
	Group:   ArgoGroup,
	Kind:    ArgoApplicationKind,
	Version: VersionV1Alpha1,
}

var GvkForSecret = schema.GroupVersionKind{
	Kind:    k8sCommonBean.SecretKind,
	Version: k8sCommonBean.V1VERSION,
}

type ArgoApplicationListDto struct {
	Name         string `json:"appName"`
	ClusterId    int    `json:"clusterId"`
	ClusterName  string `json:"clusterName"`
	Namespace    string `json:"namespace"`
	HealthStatus string `json:"appStatus"`
	SyncStatus   string `json:"syncStatus"`
}

type ArgoApplicationDetailDto struct {
	*ArgoApplicationListDto
	ResourceTree *gRPC.ResourceTreeResponse `json:"resourceTree"`
	Manifest     map[string]interface{}     `json:"manifest"`
}

type ArgoManagedResource struct {
	Group     string
	Kind      string
	Version   string
	Name      string
	Namespace string
}

type ArgoClusterConfigObj struct {
	BearerToken     string `json:"bearerToken"`
	TlsClientConfig struct {
		Insecure bool   `json:"insecure"`
		KeyData  string `json:"keyData,omitempty"`
		CertData string `json:"certData,omitempty"`
		CaData   string `json:"caData,omitempty"`
	} `json:"tlsClientConfig"`
}
