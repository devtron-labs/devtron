package bean

type FluxApplicationListDto struct {
	ClusterId  int `json:"clusterId"`
	FluxAppDto []*FluxApplication
}
type FluxApplication struct {
	Name         string `json:"appName"`
	HealthStatus string `json:"appStatus"`
	SyncStatus   string `json:"syncStatus"`
	ClusterId    int    `json:"clusterId"`
	ClusterName  string `json:"clusterName"`
	Namespace    string `json:"namespace"`
	AppType      string `json:"appType"`
}

type FluxAppList struct {
	ClusterId *int32             `json:"clusterIds,omitempty"`
	FluxApps  *[]FluxApplication `json:"fluxApplication,omitempty"`
}
