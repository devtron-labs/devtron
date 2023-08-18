package k8s

import (
	"github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/pkg/k8s/application/bean"
	"github.com/devtron-labs/devtron/util/k8s"
)

const (
	SecretKind                   = "Secret"
	ServiceKind                  = "Service"
	ServiceAccountKind           = "ServiceAccount"
	EndpointsKind                = "Endpoints"
	DeploymentKind               = "Deployment"
	ReplicaSetKind               = "ReplicaSet"
	StatefulSetKind              = "StatefulSet"
	DaemonSetKind                = "DaemonSet"
	IngressKind                  = "Ingress"
	JobKind                      = "Job"
	PersistentVolumeClaimKind    = "PersistentVolumeClaim"
	CustomResourceDefinitionKind = "CustomResourceDefinition"
	PodKind                      = "Pod"
	APIServiceKind               = "APIService"
	NamespaceKind                = "Namespace"
	HorizontalPodAutoscalerKind  = "HorizontalPodAutoscaler"
)

const (
	Group   = "group"
	Version = "version"
)

type ResourceRequestBean struct {
	AppId                string                     `json:"appId"`
	AppType              int                        `json:"appType,omitempty"`        // 0: DevtronApp, 1: HelmApp
	DeploymentType       int                        `json:"deploymentType,omitempty"` // 0: DevtronApp, 1: HelmApp
	AppIdentifier        *client.AppIdentifier      `json:"-"`
	K8sRequest           *k8s.K8sRequestBean        `json:"k8sRequest"`
	DevtronAppIdentifier *bean.DevtronAppIdentifier `json:"-"`         // For Devtron App Resources
	ClusterId            int                        `json:"clusterId"` // clusterId is used when request is for direct cluster (not for helm release)
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
