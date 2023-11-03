package bean

import (
	"github.com/devtron-labs/common-lib/utils/k8s"
)

const (
	DEFAULT_NAMESPACE = "default"
	EVENT_K8S_KIND    = "Event"
	LIST_VERB         = "list"
	Delete            = "delete"
)

const (
	// App Type Identifiers
	DevtronAppType = 0 // Identifier for Devtron Apps
	HelmAppType    = 1 // Identifier for Helm Apps
	ArgoAppType    = 2
	// Deployment Type Identifiers
	HelmInstalledType = 0 // Identifier for Helm deployment
	ArgoInstalledType = 1 // Identifier for ArgoCD deployment
)

type ResourceInfo struct {
	PodName string `json:"podName"`
}

type DevtronAppIdentifier struct {
	ClusterId int `json:"clusterId"`
	AppId     int `json:"appId"`
	EnvId     int `json:"envId"`
}

type Response struct {
	Kind     string   `json:"kind"`
	Name     string   `json:"name"`
	PointsTo string   `json:"pointsTo"`
	Urls     []string `json:"urls"`
}

type RotatePodResourceResponse struct {
	k8s.ResourceIdentifier
	ErrorResponse string `json:"errorResponse"`
}
