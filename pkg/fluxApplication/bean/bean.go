package bean

type FluxApplicationListDto struct {
	ClusterId  int `json:"clusterId"`
	FluxAppDto []*FluxApplicationDto
}
type FluxApplicationDto struct {
	Name               string             `json:"appName"`
	HealthStatus       string             `json:"appStatus"`
	SyncStatus         string             `json:"syncStatus"`
	EnvironmentDetails *EnvironmentDetail `json:"environmentDetail"`
	IsKustomizeApp     bool               `json:"isKustomizeApp"`
}
type EnvironmentDetail struct {
	ClusterId   int    `json:"clusterId"`
	ClusterName string `json:"clusterName"`
	Namespace   string `json:"namespace"`
}
