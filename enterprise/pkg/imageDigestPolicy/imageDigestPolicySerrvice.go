package imageDigestPolicy

type ImageDigestQualifierMappingService interface {
	//CreateOrUpdateQualifierMapping creates or updates image digest qualifier mapping for given cluster and environments
	CreateOrUpdateQualifierMapping(qualifierMappingRequest ClusterDetailRequest) (ClusterDetailRequest, error)

	//GetAllQualifierMappings get all cluster and environment configured for pull using image digest
	GetAllQualifierMappings() (ClusterDetailRequest, error)

	//IsPullUsingImageDigestConfigured returns true if env or cluster has image digest policy enabled
	IsPullUsingImageDigestConfigured(envId, clusterId int) (bool, error)
}
