package bean

import "github.com/devtron-labs/devtron/api/helm-app/gRPC"

type FluxApplicationListDto struct {
	ClusterId  int `json:"clusterId"`
	FluxAppDto []*FluxApplication
}
type FluxApplication struct {
	Name                  string `json:"appName"`
	HealthStatus          string `json:"appStatus"`
	SyncStatus            string `json:"syncStatus"`
	ClusterId             int    `json:"clusterId"`
	ClusterName           string `json:"clusterName"`
	Namespace             string `json:"namespace"`
	FluxAppDeploymentType string `json:"fluxAppDeploymentType"`
}

type FluxAppList struct {
	ClusterId *[]int32           `json:"clusterIds,omitempty"`
	FluxApps  *[]FluxApplication `json:"fluxApplication,omitempty"`
	Errored   *bool              `json:"errored,omitempty"`
	ErrorMsg  *string            `json:"errorMsg,omitempty"`
}

type FluxAppIdentifier struct {
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	ClusterId      int    `json:"clusterId"`
	IsKustomizeApp bool   `json:"isKustomizeApp"`
}

type FluxApplicationDetailDto struct {
	*FluxApplication
	FluxAppStatusDetail  *FluxAppStatusDetail
	ResourceTreeResponse *gRPC.ResourceTreeResponse `json:"resourceTree"`
}

type FluxAppStatusDetail struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
}
