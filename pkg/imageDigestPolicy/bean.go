package imageDigestPolicy

type imageDigestPolicy string

const (
	ALL_ENVIRONMENTS      imageDigestPolicy = "all_environments"
	SPECIFIC_ENVIRONMENTS imageDigestPolicy = "specific_environments"
)

type ClusterId = int
type EnvironmentId = int

type PolicyRequest struct {
	ClusterDetails []*ClusterDetail `json:"clusterDetails"`
	AllClusters    bool             `json:"allClusters"`
	UserId         int32
}

type ClusterDetail struct {
	ClusterId    int               `json:"clusterId"`
	Environments []int             `json:"environments"`
	PolicyType   imageDigestPolicy `json:"policyType" validate:"oneof=all_environments specific_environments"`
}
