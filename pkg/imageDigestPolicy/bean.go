package imageDigestPolicy

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"gopkg.in/go-playground/validator.v9"
	"k8s.io/utils/strings/slices"
)

type imageDigestPolicy string

const (

	//ALL_EXISTING_AND_FUTURE_ENVIRONMENTS - For this policy type resource qualifier mapping will have one entry with identifierKey=<ClusterKey>  and identifierValue = <clusterId>
	ALL_EXISTING_AND_FUTURE_ENVIRONMENTS imageDigestPolicy = "all_existing_and_future_environments"

	//SPECIFIC_ENVIRONMENTS - For this policy we will have entry for each environment of that cluster in resource qualifier mapping. identifierKey=<environmentKey> and identifierValue=<EnvironmentId>
	SPECIFIC_ENVIRONMENTS imageDigestPolicy = "specific_environments"
)

type ClusterId = int
type EnvironmentId = int

var EmptyClusterDetailsErr = errors.New("enableDigestForAllClusters is false but cluster details not provided")

type PolicyBean struct {

	//if EnableDigestForAllClusters is false, ClusterDetails will have details of cluster level policy
	ClusterDetails []*ClusterDetail `json:"clusterDetails" validate:"omitempty,dive"`

	//EnableDigestForAllClusters if true ClusterDetails field will be ignored and image digest policy will be
	//configured for all existing and future CLUSTERS. In this resource qualifier mapping will have qualifier_id = <GLOBAL_QUALIFIER>
	EnableDigestForAllClusters bool `json:"enableDigestForAllClusters,notnull"`

	UserId int32 `json:"-"`
}

type ClusterDetail struct {
	ClusterId      int               `json:"clusterId"`
	EnvironmentIds []int             `json:"environmentIds,notnull"`
	PolicyType     imageDigestPolicy `json:"policyType" validate:"validate-image-digest-policy-type"`
}

func ValidateImageDigestPolicyType(fl validator.FieldLevel) bool {
	validPolicyType := []string{string(ALL_EXISTING_AND_FUTURE_ENVIRONMENTS), string(SPECIFIC_ENVIRONMENTS)}
	if slices.Contains(validPolicyType, fl.Field().String()) {
		return true
	}
	return false
}

type newPolicySaveRequest struct {
	requestPolicies            *PolicyBean
	existingConfiguredPolicies []*resourceQualifiers.QualifierMapping
}

type oldPolicyRemoveRequest struct {
	existingConfiguredPolicies []*resourceQualifiers.QualifierMapping
	newConfiguredClusters      map[ClusterId]bool
	newConfiguredEnvs          map[EnvironmentId]bool
	userId                     int32
}
