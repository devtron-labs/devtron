package k8s

import (
	"github.com/devtron-labs/common-lib/utils/k8s"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/pkg/k8s/application/bean"
)

type ResourceRequestBean struct {
	AppId                string                     `json:"appId"`
	AppType              int                        `json:"appType,omitempty"`        // 0: DevtronApp, 1: HelmApp, 2:ArgoApp
	DeploymentType       int                        `json:"deploymentType,omitempty"` // 0: DevtronApp, 1: HelmApp
	AppIdentifier        *client.AppIdentifier      `json:"-"`
	K8sRequest           *k8s.K8sRequestBean        `json:"k8sRequest"`
	DevtronAppIdentifier *bean.DevtronAppIdentifier `json:"-"`         // For Devtron App Resources
	ClusterId            int                        `json:"clusterId"` // clusterId is used when request is for direct cluster (not for helm release)
	IsArgoApplication    bool                       `json:"-"`
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
