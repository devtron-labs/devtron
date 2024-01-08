package imageDigestPolicy

import (
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/go-pg/pg"
)

type ImageDigestQualifierMappingService interface {

	//CreateOrUpdatePolicyForPipeline created policy for enforcing pull using digest at pipeline level
	CreateOrUpdatePolicyForPipeline(pipelineId int, isImageDigestEnforcedInRequest bool, tx *pg.Tx) error

	//IsPolicyConfiguredForPipeline returns true if pipeline has image digest policy enabled
	IsPolicyConfiguredForPipeline(pipelineId int) (bool, error)

	//GetAllConfiguredPolicies get all cluster and environment configured for pull using image digest
	GetAllConfiguredPolicies() (PolicyRequest, error)

	//CreateOrUpdatePolicyForCluster creates or updates image digest qualifier mapping for given cluster and environments
	CreateOrUpdatePolicyForCluster(qualifierMappingRequest PolicyRequest) (PolicyRequest, error)

	//IsPolicyConfiguredAtGlobalLevel for env or cluster or for all clusters
	IsPolicyConfiguredAtGlobalLevel(envId, clusterId int) (bool, error)
}

type ImageDigestQualifierMappingServiceImpl struct {
	qualifierMappingService resourceQualifiers.QualifierMappingService
	clusterRepository       clusterRepository.ClusterRepository
	environmentRepository   clusterRepository.EnvironmentRepository
}

func NewImageDigestQualifierMappingServiceImpl(
	qualifierMappingService resourceQualifiers.QualifierMappingService,
	clusterRepository clusterRepository.ClusterRepository,
	environmentRepository clusterRepository.EnvironmentRepository,
) *ImageDigestQualifierMappingServiceImpl {
	return &ImageDigestQualifierMappingServiceImpl{
		qualifierMappingService: qualifierMappingService,
		clusterRepository:       clusterRepository,
		environmentRepository:   environmentRepository,
	}
}
