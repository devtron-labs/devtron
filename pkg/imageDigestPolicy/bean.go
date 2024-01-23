package imageDigestPolicy

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"gopkg.in/go-playground/validator.v9"
	"k8s.io/utils/strings/slices"
	"time"
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
type ClusterName = string
type EnvName = string

var EmptyClusterDetailsErr = errors.New("cluster details cannot be empty if enableDigestForAllClusters=false")
var EmptyEnvDetailsErr = errors.New(fmt.Sprintf("env details cannot be empty for policyType=%s", SPECIFIC_ENVIRONMENTS))

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

func QualifierMappingDao(qualifierId, identifierKey, IdentifierValueInt int, IdentifierValueString string, userId int32) *resourceQualifiers.QualifierMapping {
	return &resourceQualifiers.QualifierMapping{
		ResourceId:            resourceQualifiers.ImageDigestResourceId,
		ResourceType:          resourceQualifiers.ImageDigest,
		QualifierId:           qualifierId,
		IdentifierKey:         identifierKey,
		IdentifierValueInt:    IdentifierValueInt,
		IdentifierValueString: IdentifierValueString,
		Active:                true,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
}

func ValidateImageDigestPolicyType(fl validator.FieldLevel) bool {
	validPolicyType := []string{string(ALL_EXISTING_AND_FUTURE_ENVIRONMENTS), string(SPECIFIC_ENVIRONMENTS)}
	if slices.Contains(validPolicyType, fl.Field().String()) {
		return true
	}
	return false
}

type DigestPolicyConfigurationRequest struct {
	PipelineId    int
	ClusterId     int
	EnvironmentId int
}

func (request DigestPolicyConfigurationRequest) getQualifierMappingScope() *resourceQualifiers.Scope {
	return &resourceQualifiers.Scope{
		EnvId:      request.EnvironmentId,
		ClusterId:  request.ClusterId,
		PipelineId: request.PipelineId,
	}
}

type DigestPolicyConfigurationResponse struct {
	DigestConfiguredForPipeline     bool
	DigestConfiguredForEnvOrCluster bool
}

func (config DigestPolicyConfigurationResponse) UseDigestForTrigger() bool {
	return config.DigestConfiguredForEnvOrCluster || config.DigestConfiguredForPipeline
}
