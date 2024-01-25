package k8s

import (
	k8s2 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/pkg/k8s/application/bean"
)

type ResourceRequestBean struct {
	AppId                string                     `json:"appId"`
	AppType              int                        `json:"appType,omitempty"`        // 0: DevtronApp, 1: HelmApp
	DeploymentType       int                        `json:"deploymentType,omitempty"` // 0: DevtronApp, 1: HelmApp
	AppIdentifier        *client.AppIdentifier      `json:"-"`
	K8sRequest           *k8s2.K8sRequestBean       `json:"k8sRequest"`
	DevtronAppIdentifier *bean.DevtronAppIdentifier `json:"-"`         // For Devtron App Resources
	ClusterId            int                        `json:"clusterId"` // clusterId is used when request is for direct cluster (not for helm release)
}

type BatchResourceResponse struct {
	ManifestResponse *k8s2.ManifestResponse
	Err              error
}

type RotatePodResponse struct {
	Responses     []*bean.RotatePodResourceResponse `json:"responses"`
	ContainsError bool                              `json:"containsError"`
}

type RotatePodRequest struct {
	ClusterId int                       `json:"clusterId"`
	Resources []k8s2.ResourceIdentifier `json:"resources"`
}
type PodContainerList struct {
	Containers          []string
	InitContainers      []string
	EphemeralContainers []string
}

type ResourceGetResponse struct {
	ManifestResponse *k8s2.ManifestResponse `json:"manifestResponse"`
	SecretViewAccess bool                   `json:"secretViewAccess"` // imp: only for resource browser, this is being used to check whether a user can see obscured secret values or not.
}
