package imageDigestPolicy

type imageDigestPolicy string

const (

	//ALL_EXISTING_AND_FUTURE_ENVIRONMENTS - For this policy type resource qualifier mapping will have one entry with identifierKey=<ClusterKey>  and identifierValue = <clusterId>
	ALL_EXISTING_AND_FUTURE_ENVIRONMENTS imageDigestPolicy = "all_existing_and_future_environments"

	//SPECIFIC_ENVIRONMENTS - For this policy we will have entry for each environment of that cluster in resource qualifier mapping. identifierKey=<environmentKey> and identifierValue=<EnvironmentId>
	SPECIFIC_ENVIRONMENTS imageDigestPolicy = "specific_environments"
)

type ClusterId = int
type EnvironmentId = int

type PolicyRequest struct {
	ClusterDetails             []*ClusterDetail `json:"clusterDetails" validate:"omitempty,dive"`
	EnableDigestForAllClusters bool             `json:"enableDigestForAllClusters,notnull"`
	UserId                     int32            `json:"-"`
}

type ClusterDetail struct {
	ClusterId    int               `json:"clusterId"`
	Environments []int             `json:"environments"`
	PolicyType   imageDigestPolicy `json:"policyType" validate:"oneof=all_existing_and_future_environments specific_environments"`
}
