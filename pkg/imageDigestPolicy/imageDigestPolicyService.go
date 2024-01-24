package imageDigestPolicy

import (
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ImageDigestPolicyService interface {

	//CreatePolicyForPipeline creates image digest policy for pipeline even if it exists.
	CreatePolicyForPipeline(tx *pg.Tx, pipelineId int, pipelineName string, UserId int32) (int, error)

	//CreatePolicyForPipelineIfNotExist creates image digest policy for pipeline if not already created
	CreatePolicyForPipelineIfNotExist(tx *pg.Tx, pipelineId int, pipelineName string, UserId int32) (int, error)

	//GetDigestPolicyConfigurations returns true if pipeline or env or cluster has image digest policy enabled
	GetDigestPolicyConfigurations(digestConfigurationRequest DigestPolicyConfigurationRequest) (digestPolicyConfiguration DigestPolicyConfigurationResponse, err error)

	//DeletePolicyForPipeline deletes image digest policy for a pipeline
	DeletePolicyForPipeline(tx *pg.Tx, pipelineId int, userId int32) (int, error)

	//CreateOrUpdatePolicyForCluster creates or updates image digest qualifier mapping for given cluster and environments
	CreateOrUpdatePolicyForCluster(policyRequest *PolicyBean) (*PolicyBean, error)

	//GetAllPoliciesConfiguredForClusterOrEnv get all cluster and environment configured for pull using image digest
	GetAllPoliciesConfiguredForClusterOrEnv() (*PolicyBean, error)
}

type ImageDigestPolicyServiceImpl struct {
	logger                       *zap.SugaredLogger
	qualifierMappingService      resourceQualifiers.QualifierMappingService
	devtronResourceSearchableKey devtronResource.DevtronResourceSearchableKeyService
	environmentRepository        repository.EnvironmentRepository
	clusterRepository            repository.ClusterRepository
	dbConnection                 *pg.DB
}

func NewImageDigestPolicyServiceImpl(
	logger *zap.SugaredLogger,
	qualifierMappingService resourceQualifiers.QualifierMappingService,
	devtronResourceSearchableKey devtronResource.DevtronResourceSearchableKeyService,
	environmentRepository repository.EnvironmentRepository,
	clusterRepository repository.ClusterRepository,
	dbConnection *pg.DB,
) *ImageDigestPolicyServiceImpl {
	return &ImageDigestPolicyServiceImpl{
		logger:                       logger,
		qualifierMappingService:      qualifierMappingService,
		devtronResourceSearchableKey: devtronResourceSearchableKey,
		environmentRepository:        environmentRepository,
		clusterRepository:            clusterRepository,
		dbConnection:                 dbConnection,
	}
}

func (impl ImageDigestPolicyServiceImpl) CreatePolicyForPipeline(tx *pg.Tx, pipelineId int, pipelineName string, UserId int32) (int, error) {

	var qualifierMappingId int

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()

	identifierKey := devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID]
	identifierValueId := pipelineId
	identifierValueString := pipelineName
	qualifierMapping := digestPolicyQualifierMappingDao(int(resourceQualifiers.PIPELINE_QUALIFIER), identifierKey, identifierValueId, identifierValueString, UserId)
	_, err := impl.qualifierMappingService.CreateQualifierMappings([]*resourceQualifiers.QualifierMapping{qualifierMapping}, tx)
	if err != nil {
		impl.logger.Errorw("error in creating image digest policy for pipeline", "err", err, "pipelineId", pipelineId)
		return qualifierMapping.Id, err
	}
	qualifierMappingId = qualifierMapping.Id

	return qualifierMappingId, nil
}

func (impl ImageDigestPolicyServiceImpl) CreatePolicyForPipelineIfNotExist(tx *pg.Tx, pipelineId int, pipelineName string, UserId int32) (int, error) {

	var qualifierMappingId int

	policyConfigurationRequest := DigestPolicyConfigurationRequest{PipelineId: pipelineId}
	digestPolicyConfigurations, err := impl.GetDigestPolicyConfigurations(policyConfigurationRequest)
	if err != nil {
		impl.logger.Errorw("Error in checking if isDigestPolicyConfiguredForPipeline", "err", err, "pipelineId", pipelineId)
		return 0, err
	}

	if !digestPolicyConfigurations.DigestConfiguredForPipeline {
		qualifierMappingId, err = impl.CreatePolicyForPipeline(tx, pipelineId, pipelineName, UserId)
		if err != nil {
			impl.logger.Errorw("error in creating policy for pipeline", "err", "pipelineId", pipelineId)
			return qualifierMappingId, nil
		}
	}
	return qualifierMappingId, nil
}

func (impl ImageDigestPolicyServiceImpl) GetDigestPolicyConfigurations(digestConfigurationRequest DigestPolicyConfigurationRequest) (digestPolicyConfiguration DigestPolicyConfigurationResponse, err error) {

	resourceIds := []int{resourceQualifiers.ImageDigestResourceId}

	scope := digestConfigurationRequest.getQualifierMappingScope()

	policyMappings, err := impl.qualifierMappingService.GetQualifierMappings(resourceQualifiers.ImageDigest, scope, resourceIds)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting saved policy mappings", "err", err)
		return digestPolicyConfiguration, err
	}
	if err == pg.ErrNoRows || len(policyMappings) == 0 {
		return digestPolicyConfiguration, nil
	}

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()
	clusterIdentifierKey := devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID]
	envIdentifierKey := devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID]
	pipelineIdentifierKey := devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID]

	for _, policy := range policyMappings {
		switch policy.IdentifierKey {
		case clusterIdentifierKey, envIdentifierKey:
			digestPolicyConfiguration.DigestConfiguredForEnvOrCluster = true
		case pipelineIdentifierKey:
			digestPolicyConfiguration.DigestConfiguredForPipeline = true
		}
	}

	return digestPolicyConfiguration, nil
}

func (impl ImageDigestPolicyServiceImpl) DeletePolicyForPipeline(tx *pg.Tx, pipelineId int, userId int32) (int, error) {
	auditLog := sql.AuditLog{
		CreatedOn: time.Now(),
		CreatedBy: userId,
		UpdatedOn: time.Now(),
		UpdatedBy: userId,
	}
	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()
	err := impl.qualifierMappingService.DeleteByIdentifierKeyAndValue(resourceQualifiers.ImageDigest, devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID], pipelineId, auditLog, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting image digest policy for pipeline", "err", err, "pipelineId", pipelineId)
		return pipelineId, err
	}
	return pipelineId, nil
}

func (impl ImageDigestPolicyServiceImpl) CreateOrUpdatePolicyForCluster(policyRequest *PolicyBean) (*PolicyBean, error) {

	tx, err := impl.dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in transaction begin", "err", err)
		return nil, err
	}
	defer func() {
		err = tx.Rollback()
		if err != nil {
			impl.logger.Errorw("error in transaction rollback", "err", err)
			return
		}
	}()

	if policyRequest.EnableDigestForAllClusters == true {
		err := impl.saveImageDigestPolicyForAllClusters(tx, policyRequest.UserId)
		if err != nil {
			impl.logger.Errorw("Error in saving image digest policy for all clusters", "err", err)
			return policyRequest, err
		}
		err = tx.Commit()
		if err != nil {
			impl.logger.Errorw("error in commiting transaction", "err", err)
			return policyRequest, err
		}
		return policyRequest, nil
	}

	// fetching already configured policies
	existingDigestMappings, err := impl.qualifierMappingService.GetQualifierMappingsByResourceType(resourceQualifiers.ImageDigest)
	if err != nil {
		impl.logger.Errorw("error in getting configured image digest policies", "resourceType: ", resourceQualifiers.ImageDigest)
		return nil, err
	}

	// saving image digest policy for new clusters and environments
	savePolicyRequest := newPolicySaveRequest{
		requestPolicies:            policyRequest,
		existingConfiguredPolicies: existingDigestMappings,
	}
	newConfiguredClusters, newConfiguredEnvs, err := impl.saveNewPolicies(tx, savePolicyRequest)
	if err != nil {
		impl.logger.Errorw("error in creating image digest policies", "err", err)
		return nil, err
	}

	// removing policies present in db but not present in request
	policyRemoveRequest := oldPolicyRemoveRequest{
		existingConfiguredPolicies: existingDigestMappings,
		newConfiguredClusters:      newConfiguredClusters,
		newConfiguredEnvs:          newConfiguredEnvs,
		userId:                     policyRequest.UserId,
	}
	err = impl.removePoliciesNotPresentInRequest(tx, policyRemoveRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting policies not present in request but present in DB", "err", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return policyRequest, err
	}

	return policyRequest, nil
}

func (impl ImageDigestPolicyServiceImpl) saveImageDigestPolicyForAllClusters(tx *pg.Tx, userId int32) error {

	// step1: Delete all existing individual cluster and env level image digest policy mappings
	// step2: create image digest policy for all clusters by setting qualifierId = int(resourceQualifiers.GLOBAL_QUALIFIER)

	// deleting all cluster and env level policies
	err := impl.qualifierMappingService.DeleteAllByResourceTypeAndQualifierIds(
		resourceQualifiers.ImageDigest,
		resourceQualifiers.ImageDigestResourceId,
		[]int{int(resourceQualifiers.CLUSTER_QUALIFIER), int(resourceQualifiers.ENV_QUALIFIER)},
		userId,
		tx)
	if err != nil {
		impl.logger.Errorw("error in deleting image digest policies for all envs and clusters", "err", err)
		return err
	}

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()
	// creating image digest policy at global level
	qualifierId := int(resourceQualifiers.CLUSTER_QUALIFIER)
	identifierKey := devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID]
	identifierValueInt := resourceQualifiers.AllExistingAndFutureEnvsInt
	identifierValueString := resourceQualifiers.AllExistingAndFutureEnvsString

	globalQualifierMapping := digestPolicyQualifierMappingDao(qualifierId, identifierKey, identifierValueInt, identifierValueString, userId)
	_, err = impl.qualifierMappingService.CreateQualifierMappings([]*resourceQualifiers.QualifierMapping{globalQualifierMapping}, tx)
	if err != nil {
		impl.logger.Errorw("error in creating global image digest policy", "err", err)
		return err
	}

	return nil
}

func getConfiguredClustersAndEnvironments(existingDigestMappings []*resourceQualifiers.QualifierMapping,
	devtronResourceSearchableKeyMap map[bean.DevtronResourceSearchableKeyName]int) (map[ClusterId]bool, map[EnvironmentId]bool) {

	existingConfiguredClusters := make(map[ClusterId]bool)
	existingConfiguredEnvironments := make(map[EnvironmentId]bool)

	for _, existingMapping := range existingDigestMappings {

		switch existingMapping.IdentifierKey {

		case devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID]:
			existingConfiguredClusters[existingMapping.IdentifierValueInt] = true

		case devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID]:
			existingConfiguredEnvironments[existingMapping.IdentifierValueInt] = true

		}
	}
	return existingConfiguredClusters, existingConfiguredEnvironments
}

func (impl ImageDigestPolicyServiceImpl) saveNewPolicies(tx *pg.Tx, savePolicyRequest newPolicySaveRequest) (map[ClusterId]bool, map[EnvironmentId]bool, error) {

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()

	requestPolicies := savePolicyRequest.requestPolicies

	// exiting cluster and environments already having imageDigest configured
	existingConfiguredClusters, existingConfiguredEnvs := getConfiguredClustersAndEnvironments(
		savePolicyRequest.existingConfiguredPolicies, devtronResourceSearchableKeyMap)

	newPolicies := make([]*resourceQualifiers.QualifierMapping, 0)
	newClustersConfigured := make(map[ClusterId]bool)
	newEnvsConfigured := make(map[EnvironmentId]bool)

	clusterIdNameMap, envIdNameMap, err := impl.getIdToNameMappings(requestPolicies)
	if err != nil {
		impl.logger.Errorw("error in saving policies", "err", err)
		return newClustersConfigured, newEnvsConfigured, nil
	}

	for _, policy := range requestPolicies.ClusterDetails {

		switch policy.PolicyType {

		case ALL_EXISTING_AND_FUTURE_ENVIRONMENTS:

			if _, ok := existingConfiguredClusters[policy.ClusterId]; !ok {

				qualifierId := int(resourceQualifiers.CLUSTER_QUALIFIER)
				identifierKey := devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID]
				identifierValueInt := policy.ClusterId
				identifierValueString := clusterIdNameMap[policy.ClusterId]

				newQualifierMapping := digestPolicyQualifierMappingDao(
					qualifierId, identifierKey, identifierValueInt, identifierValueString, requestPolicies.UserId)
				newPolicies = append(newPolicies, newQualifierMapping)

			}
			newClustersConfigured[policy.ClusterId] = true

		case SPECIFIC_ENVIRONMENTS:

			for _, envId := range policy.EnvironmentIds {
				if _, ok := existingConfiguredEnvs[envId]; !ok {

					qualifierId := int(resourceQualifiers.ENV_QUALIFIER)
					identifierKey := devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID]
					identifierValueInt := envId
					identifierValueString := envIdNameMap[envId]

					newQualifierMapping := digestPolicyQualifierMappingDao(
						qualifierId, identifierKey, identifierValueInt, identifierValueString, requestPolicies.UserId)
					newPolicies = append(newPolicies, newQualifierMapping)
				}
				newEnvsConfigured[envId] = true
			}
		}
	}
	if len(newPolicies) > 0 {
		_, err = impl.qualifierMappingService.CreateQualifierMappings(newPolicies, tx)
		if err != nil {
			impl.logger.Errorw("error in creating image digest policy", "err", err)
			return newClustersConfigured, newEnvsConfigured, err
		}
	}
	return newClustersConfigured, newEnvsConfigured, nil
}

func (impl ImageDigestPolicyServiceImpl) getIdToNameMappings(clusterDetails *PolicyBean) (clusterIdNameMap map[ClusterId]ClusterName, envIdNameMap map[EnvironmentId]EnvName, err error) {

	clusterIdNameMap = make(map[ClusterId]ClusterName)
	envIdNameMap = make(map[EnvironmentId]EnvName)

	clusterIds := make([]int, 0)
	envIds := make([]*int, 0)
	for _, clusterDetail := range clusterDetails.ClusterDetails {
		clusterIds = append(clusterIds, clusterDetail.ClusterId)
		for _, envId := range clusterDetail.EnvironmentIds {
			envIds = append(envIds, &envId)
		}
	}

	environments, err := impl.environmentRepository.FindByIds(envIds)
	if err != nil {
		impl.logger.Errorw("error in fetching envs by ids", "envIds", envIds, "err", err)
		return clusterIdNameMap, envIdNameMap, err
	}

	clusters, err := impl.clusterRepository.FindByIds(clusterIds)
	if err != nil {
		impl.logger.Errorw("error in fetching envs by ids", "clusterIds", clusterIds, "err", err)
		return clusterIdNameMap, envIdNameMap, err
	}

	for _, env := range environments {
		envIdNameMap[env.Id] = env.Name
	}
	for _, cluster := range clusters {
		clusterIdNameMap[cluster.Id] = cluster.ClusterName
	}
	return clusterIdNameMap, envIdNameMap, nil
}

func (impl ImageDigestPolicyServiceImpl) removePoliciesNotPresentInRequest(tx *pg.Tx, policyRemoveRequest oldPolicyRemoveRequest) error {

	existingPolicies := policyRemoveRequest.existingConfiguredPolicies
	newClustersConfigured := policyRemoveRequest.newConfiguredClusters
	newEnvsConfigured := policyRemoveRequest.newConfiguredEnvs
	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()

	policiesToBeRemovedIDs := make([]int, 0)

	for _, policy := range existingPolicies {
		removePolicy := false
		if policy.IdentifierKey == devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID] {
			if _, ok := newClustersConfigured[policy.IdentifierValueInt]; !ok {
				removePolicy = true
			}
		} else if policy.IdentifierKey == devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID] {
			if _, ok := newEnvsConfigured[policy.IdentifierValueInt]; !ok {
				removePolicy = true
			}
		} else if policy.QualifierId == int(resourceQualifiers.GLOBAL_QUALIFIER) {
			// removing global policy because if we are here EnableDigestForAllClusters=false in request
			removePolicy = true
		}
		if removePolicy {
			policiesToBeRemovedIDs = append(policiesToBeRemovedIDs, policy.Id)
		}
	}

	if len(policiesToBeRemovedIDs) > 0 {
		err := impl.qualifierMappingService.DeleteAllByIds(policiesToBeRemovedIDs, policyRemoveRequest.userId, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting old policies", "err", err)
			return err
		}
	}

	return nil
}

func (impl ImageDigestPolicyServiceImpl) GetAllPoliciesConfiguredForClusterOrEnv() (*PolicyBean, error) {

	imageDigestQualifierMappings, err := impl.qualifierMappingService.GetQualifierMappingsByResourceType(resourceQualifiers.ImageDigest)
	if err != nil {
		impl.logger.Errorw("error in fetching qualifier mappings by resourceType", "resourceType: ", resourceQualifiers.ImageDigest)
		return nil, err
	}

	imageDigestPolicies := &PolicyBean{
		ClusterDetails:             make([]*ClusterDetail, 0),
		EnableDigestForAllClusters: false,
	}

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()
	clusterIdentifierKey := devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID]

	for _, qualifierMapping := range imageDigestQualifierMappings {
		if qualifierMapping.IdentifierKey == clusterIdentifierKey && qualifierMapping.IdentifierValueInt == resourceQualifiers.AllExistingAndFutureEnvsInt {
			imageDigestPolicies.EnableDigestForAllClusters = true
			return imageDigestPolicies, nil
		}
	}

	configuredClusterToEnvMapping, err := impl.getClusterIdToEnvIdsMapping(imageDigestQualifierMappings)
	if err != nil {
		impl.logger.Errorw("error in getting cluster id to envIds map", "err", err)
		return nil, err
	}

	for clusterId, envIds := range configuredClusterToEnvMapping {
		clusterDetail := &ClusterDetail{
			ClusterId: clusterId,
		}
		if len(envIds) == 0 {
			clusterDetail.PolicyType = ALL_EXISTING_AND_FUTURE_ENVIRONMENTS
		} else {
			clusterDetail.EnvironmentIds = envIds
			clusterDetail.PolicyType = SPECIFIC_ENVIRONMENTS
		}
		imageDigestPolicies.ClusterDetails = append(imageDigestPolicies.ClusterDetails, clusterDetail)
	}

	return imageDigestPolicies, nil
}

func (impl ImageDigestPolicyServiceImpl) getClusterIdToEnvIdsMapping(imageDigestQualifierMappings []*resourceQualifiers.QualifierMapping) (map[ClusterId][]EnvironmentId, error) {

	clusterIdToEnvIdsMapping := make(map[ClusterId][]EnvironmentId)
	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()

	// adding cluster configured for all existing and future environments
	for _, qualifierMapping := range imageDigestQualifierMappings {
		if qualifierMapping.IdentifierKey == devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID] {
			clusterIdToEnvIdsMapping[qualifierMapping.IdentifierValueInt] = make([]EnvironmentId, 0)
		}
	}

	// adding clusters added for specific environments
	envToClusterMapping, err := impl.getEnvToClusterMapping(imageDigestQualifierMappings) // map[envId]clusterId
	if err != nil {
		impl.logger.Errorw("error in fetching environments to cluster maping", "err", err)
		return nil, err
	}
	for envId, clusterId := range envToClusterMapping {
		clusterIdToEnvIdsMapping[clusterId] = append(clusterIdToEnvIdsMapping[clusterId], envId)
	}

	return clusterIdToEnvIdsMapping, nil
}

func (impl ImageDigestPolicyServiceImpl) getEnvToClusterMapping(imageDigestQualifierMappings []*resourceQualifiers.QualifierMapping) (map[EnvironmentId]ClusterId, error) {
	EnvToClusterMapping := make(map[int]int)
	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()
	environmentIds := make([]*int, 0)
	for _, qualifierMapping := range imageDigestQualifierMappings {
		if qualifierMapping.IdentifierKey == devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID] {
			environmentIds = append(environmentIds, &qualifierMapping.IdentifierValueInt)
		}
	}
	if len(environmentIds) > 0 {
		environments, err := impl.environmentRepository.FindByIds(environmentIds)
		if err != nil {
			impl.logger.Errorw("error in fetching environments by environmentIds", "err", err)
			return nil, err
		}
		for _, env := range environments {
			EnvToClusterMapping[env.Id] = env.ClusterId
		}
	}
	return EnvToClusterMapping, nil
}
