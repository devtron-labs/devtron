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

	//CreatePolicyForPipelineIfNotExist creates image digest policy for pipeline if not already created
	CreatePolicyForPipelineIfNotExist(tx *pg.Tx, pipelineId int, UserId int32) (int, error)

	//IsPolicyConfiguredForPipeline returns true if pipeline or env or cluster has image digest policy enabled
	IsPolicyConfiguredForPipeline(pipelineId int) (bool, error)

	//DeletePolicyForPipeline deletes image digest policy for a pipeline
	DeletePolicyForPipeline(tx *pg.Tx, pipelineId int, userId int32) (int, error)

	//CreateOrUpdatePolicyForCluster creates or updates image digest qualifier mapping for given cluster and environments
	CreateOrUpdatePolicyForCluster(policyRequest *PolicyBean) (*PolicyBean, error)

	//GetAllPoliciesConfiguredForClusterOrEnv get all cluster and environment configured for pull using image digest
	GetAllPoliciesConfiguredForClusterOrEnv() (*PolicyBean, error)

	//IsPolicyConfiguredForEnvOrCluster for env or cluster or for all clusters
	IsPolicyConfiguredForEnvOrCluster(envId int, clusterId int) (bool, error)

	//IsPolicyConfiguredForClusterOrEnvOrPipeline for env or cluster or for all clusters or for pipeline
	IsPolicyConfiguredForClusterOrEnvOrPipeline(envId, clusterId, pipelineId int) (bool, error)
}

type ImageDigestPolicyServiceImpl struct {
	logger                       *zap.SugaredLogger
	qualifierMappingService      resourceQualifiers.QualifierMappingService
	devtronResourceSearchableKey devtronResource.DevtronResourceSearchableKeyService
	environmentRepository        repository.EnvironmentRepository
}

func NewImageDigestPolicyServiceImpl(
	logger *zap.SugaredLogger,
	qualifierMappingService resourceQualifiers.QualifierMappingService,
	devtronResourceSearchableKey devtronResource.DevtronResourceSearchableKeyService,
	environmentRepository repository.EnvironmentRepository,
) *ImageDigestPolicyServiceImpl {
	return &ImageDigestPolicyServiceImpl{
		logger:                       logger,
		qualifierMappingService:      qualifierMappingService,
		devtronResourceSearchableKey: devtronResourceSearchableKey,
		environmentRepository:        environmentRepository,
	}
}

func (impl ImageDigestPolicyServiceImpl) CreatePolicyForPipelineIfNotExist(tx *pg.Tx, pipelineId int, UserId int32) (int, error) {

	var qualifierMappingId int

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()

	isDigestPolicyConfiguredForPipeline, err := impl.IsPolicyConfiguredForPipeline(pipelineId)
	if err != nil {
		impl.logger.Errorw("Error in checking if isDigestPolicyConfiguredForPipeline", "err", err, "pipelineId", pipelineId)
		return 0, err
	}

	if !isDigestPolicyConfiguredForPipeline {
		identifierKey := devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID]
		identifierValue := pipelineId
		qualifierMapping := QualifierMappingDao(int(resourceQualifiers.PIPELINE_QUALIFIER),
			identifierKey,
			identifierValue,
			UserId,
		)
		_, err = impl.qualifierMappingService.CreateQualifierMappings([]*resourceQualifiers.QualifierMapping{qualifierMapping}, tx)
		if err != nil {
			impl.logger.Errorw("error in creating image digest policy for pipeline", "err", err, "identifierKey", "pipelineId", pipelineId)
			return qualifierMapping.Id, err
		}
		qualifierMappingId = qualifierMapping.Id
	}
	return qualifierMappingId, nil
}

func (impl ImageDigestPolicyServiceImpl) IsPolicyConfiguredForPipeline(pipelineId int) (bool, error) {
	scope := &resourceQualifiers.Scope{PipelineId: pipelineId}
	resourceIds := []int{resourceQualifiers.ImageDigestResourceId}
	qualifierMappings, err := impl.qualifierMappingService.GetQualifierMappings(resourceQualifiers.ImageDigest, scope, resourceIds)
	if err != nil && err != pg.ErrNoRows {
		return false, err
	} else if err == pg.ErrNoRows || len(qualifierMappings) == 0 {
		return false, nil
	}
	return true, nil
}

func (impl ImageDigestPolicyServiceImpl) DeletePolicyForPipeline(tx *pg.Tx, pipelineId int, userId int32) (int, error) {
	auditLog := sql.AuditLog{
		CreatedOn: time.Now(),
		CreatedBy: userId,
		UpdatedOn: time.Now(),
		UpdatedBy: userId,
	}
	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()
	err := impl.qualifierMappingService.DeleteByResourceTypeIdentifierKeyAndValue(
		resourceQualifiers.ImageDigest,
		devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID],
		pipelineId,
		auditLog,
		tx)
	if err != nil {
		impl.logger.Errorw("error in deleting image digest policy for pipeline", "err", err, "pipeline id", pipelineId)
		return pipelineId, err
	}
	return pipelineId, nil
}

func (impl ImageDigestPolicyServiceImpl) CreateOrUpdatePolicyForCluster(policyRequest *PolicyBean) (*PolicyBean, error) {

	dbConnection := impl.qualifierMappingService.GetDbConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in transaction begin", "err", err)
		return nil, err
	}
	defer func() {
		err = tx.Commit()
		if err != nil {
			impl.logger.Errorw("error in commiting transaction", "err", err)
			return
		}
	}()

	if policyRequest.EnableDigestForAllClusters == true {
		err := impl.saveImageDigestPolicyForAllClusters(policyRequest.UserId, tx)
		if err != nil {
			impl.logger.Errorw("Error in saving image digest policy for all clusters", "err", err)
			return nil, err
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
	newConfiguredClusters, newConfiguredEnvs, err := impl.saveNewPolicies(savePolicyRequest, tx)
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
	err = impl.removePoliciesNotPresentInRequest(policyRemoveRequest, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting policies not present in request but present in DB", "err", err)
		return nil, err
	}

	return policyRequest, nil
}

func (impl ImageDigestPolicyServiceImpl) saveImageDigestPolicyForAllClusters(userId int32, tx *pg.Tx) error {

	// step1: create image digest policy for all clusters by setting qualifierId = int(resourceQualifiers.GLOBAL_QUALIFIER)
	// step2: Delete individual cluster and env level image digest policy mappings

	globalQualifierMapping := QualifierMappingDao(int(resourceQualifiers.GLOBAL_QUALIFIER), 0, 0, userId)

	// creating image digest policy at global level
	_, err := impl.qualifierMappingService.CreateQualifierMappings([]*resourceQualifiers.QualifierMapping{globalQualifierMapping}, tx)
	if err != nil {
		impl.logger.Errorw("error in creating global image digest policy", "err", err)
		return err
	}

	// deleting all cluster and env level policies
	err = impl.qualifierMappingService.DeleteAllByResourceTypeAndQualifierIds(
		resourceQualifiers.ImageDigest,
		resourceQualifiers.ImageDigestResourceId,
		[]int{int(resourceQualifiers.CLUSTER_QUALIFIER), int(resourceQualifiers.ENV_QUALIFIER)},
		userId,
		tx)
	if err != nil {
		impl.logger.Errorw("error in deleting image digest policies for all envs and clusters", "err", err)
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

func (impl ImageDigestPolicyServiceImpl) saveNewPolicies(savePolicyRequest newPolicySaveRequest, tx *pg.Tx) (map[ClusterId]bool, map[EnvironmentId]bool, error) {

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()

	requestPolicies := savePolicyRequest.requestPolicies

	// exiting cluster and environments already having imageDigest configured
	existingConfiguredClusters, existingConfiguredEnvs := getConfiguredClustersAndEnvironments(
		savePolicyRequest.existingConfiguredPolicies, devtronResourceSearchableKeyMap)

	newPolicies := make([]*resourceQualifiers.QualifierMapping, 0)
	newClustersConfigured := make(map[ClusterId]bool)
	newEnvsConfigured := make(map[EnvironmentId]bool)

	for _, policy := range requestPolicies.ClusterDetails {

		switch policy.PolicyType {

		case ALL_EXISTING_AND_FUTURE_ENVIRONMENTS:

			if _, ok := existingConfiguredClusters[policy.ClusterId]; !ok {

				newQualifierMapping := QualifierMappingDao(int(resourceQualifiers.CLUSTER_QUALIFIER),
					devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID],
					policy.ClusterId,
					requestPolicies.UserId,
				)
				newPolicies = append(newPolicies, newQualifierMapping)

			}
			newClustersConfigured[policy.ClusterId] = true

		case SPECIFIC_ENVIRONMENTS:

			for _, envId := range policy.EnvironmentIds {
				if _, ok := existingConfiguredEnvs[envId]; !ok {
					newQualifierMapping := QualifierMappingDao(int(resourceQualifiers.ENV_QUALIFIER),
						devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID],
						envId,
						requestPolicies.UserId,
					)
					newPolicies = append(newPolicies, newQualifierMapping)
				}
				newEnvsConfigured[envId] = true
			}
		}
	}
	if len(newPolicies) > 0 {
		_, err := impl.qualifierMappingService.CreateQualifierMappings(newPolicies, tx)
		if err != nil {
			impl.logger.Errorw("error in creating image digest policy", "err", err)
			return newClustersConfigured, newEnvsConfigured, err
		}
	}
	return newClustersConfigured, newEnvsConfigured, nil
}

func (impl ImageDigestPolicyServiceImpl) removePoliciesNotPresentInRequest(policyRemoveRequest oldPolicyRemoveRequest, tx *pg.Tx) error {

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

	for _, qualifierMapping := range imageDigestQualifierMappings {
		if qualifierMapping.QualifierId == int(resourceQualifiers.GLOBAL_QUALIFIER) {
			imageDigestPolicies.EnableDigestForAllClusters = true
			break
		}
	}

	if imageDigestPolicies.EnableDigestForAllClusters {
		return imageDigestPolicies, nil
	}

	ClusterIdToEnvIdsMapping, err := impl.getClusterIdToEnvIdsMapping(imageDigestQualifierMappings)
	if err != nil {
		impl.logger.Errorw("error in getting cluster id to envIds map", "err", err)
		return nil, err
	}

	for clusterId, envIds := range ClusterIdToEnvIdsMapping {
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

	ClusterIdToEnvIdsMapping := make(map[ClusterId][]EnvironmentId)
	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()

	// adding cluster configured for all existing and future environments
	for _, qualifierMapping := range imageDigestQualifierMappings {
		if qualifierMapping.IdentifierKey == devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID] {
			ClusterIdToEnvIdsMapping[qualifierMapping.IdentifierValueInt] = make([]EnvironmentId, 0)
		}
	}

	// adding clusters added for specific environments
	EnvToClusterMapping, err := impl.getEnvToClusterMapping(imageDigestQualifierMappings) // map[envId]clusterId
	if err != nil {
		impl.logger.Errorw("error in fetching environments to cluster maping", "err", err)
		return nil, err
	}
	for envId, clusterId := range EnvToClusterMapping {
		ClusterIdToEnvIdsMapping[clusterId] = append(ClusterIdToEnvIdsMapping[clusterId], envId)
	}

	return ClusterIdToEnvIdsMapping, nil
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

func (impl ImageDigestPolicyServiceImpl) IsPolicyConfiguredForEnvOrCluster(envId int, clusterId int) (bool, error) {
	scope := &resourceQualifiers.Scope{EnvId: envId, ClusterId: clusterId}
	resourceIds := []int{resourceQualifiers.ImageDigestResourceId}
	qualifierMappings, err := impl.qualifierMappingService.GetQualifierMappings(resourceQualifiers.ImageDigest, scope, resourceIds)
	if err != nil && err != pg.ErrNoRows {
		return false, err
	}
	if len(qualifierMappings) == 0 {
		return false, nil
	}
	return true, nil
}

func (impl ImageDigestPolicyServiceImpl) IsPolicyConfiguredForClusterOrEnvOrPipeline(envId, clusterId, pipelineId int) (bool, error) {
	isImageDigestPolicyConfiguredAtGlobalLevel, err :=
		impl.IsPolicyConfiguredForEnvOrCluster(envId, clusterId)
	if err != nil {
		impl.logger.Errorw("error in checking if image digest policy is configured or not", "err", err)
		return false, err
	}

	var isDigestPolicyConfiguredForPipeline bool
	if !isImageDigestPolicyConfiguredAtGlobalLevel {
		isDigestPolicyConfiguredForPipeline, err = impl.IsPolicyConfiguredForPipeline(pipelineId)
		if err != nil {
			impl.logger.Errorw("Error in checking if isDigestPolicyConfiguredForPipeline", "err", err)
			return false, err
		}
	}
	isDigestConfigured := isImageDigestPolicyConfiguredAtGlobalLevel || isDigestPolicyConfiguredForPipeline
	return isDigestConfigured, nil
}
