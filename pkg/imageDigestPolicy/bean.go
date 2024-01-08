package imageDigestPolicy

type imageDigestPolicy string

const (
	ALL_ENVIRONMENTS      imageDigestPolicy = "all_environments"
	SELECTED_ENVIRONMENTS imageDigestPolicy = "selected_environments"
)

type PolicyRequest struct {
	ClusterDetails []ClusterDetail `json:"clusterDetails"`
	AllClusters    bool            `json:"allClusters"`
}

type ClusterDetail struct {
	ClusterName  string            `json:"clusterName"`
	Environments []int             `json:"environments"`
	PolicyType   imageDigestPolicy `json:"policyType" validate:"oneof=all_environments selected_environments"`
}
