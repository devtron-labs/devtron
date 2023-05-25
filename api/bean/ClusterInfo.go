package bean

type ClusterInfo struct {
	ClusterId   int    `json:"clusterId"`
	ClusterName string `json:"clusterName"`
	BearerToken string `json:"bearerToken"`
	ServerUrl   string `json:"serverUrl"`
}
