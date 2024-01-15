package imageDigestPolicy

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	http2 "net/http"
	"time"
)

type ImageDigestPolicyService interface {
	//CreateOrDeletePolicyForPipeline created policy for enforcing pull using digest at pipeline level
	CreateOrDeletePolicyForPipeline(pipelineId int, isImageDigestEnforcedForPipeline bool, UserId int32, tx *pg.Tx) error

	//IsPolicyConfiguredForPipeline returns true if pipeline or env or cluster has image digest policy enabled
	IsPolicyConfiguredForPipeline(pipelineId int) (bool, error)

	//CreateOrUpdatePolicyForCluster creates or updates image digest qualifier mapping for given cluster and environments
	CreateOrUpdatePolicyForCluster(policyRequest *PolicyRequest) (*PolicyRequest, error)

	//GetAllConfiguredGlobalPolicies get all cluster and environment configured for pull using image digest
	GetAllConfiguredGlobalPolicies() (*PolicyRequest, error)

	//IsPolicyConfiguredAtGlobalLevel for env or cluster or for all clusters
	IsPolicyConfiguredAtGlobalLevel(envId int, clusterId int) (bool, error)

	//IsPolicyConfiguredAtGlobalOrPipeline for env or cluster or for all clusters or for pipeline
	IsPolicyConfiguredAtGlobalOrPipeline(envId, clusterId, pipelineId int) (bool, error)
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

func (impl ImageDigestPolicyServiceImpl) CreateOrDeletePolicyForPipeline(pipelineId int, isImageDigestEnforcedForPipeline bool, UserId int32, tx *pg.Tx) error {

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()

	isDigestPolicyConfiguredForPipeline, err := impl.IsPolicyConfiguredForPipeline(pipelineId)
	if err != nil {
		impl.logger.Errorw("Error in checking if isDigestPolicyConfiguredForPipeline", "err", err)
		return err
	}

	if !isDigestPolicyConfiguredForPipeline && isImageDigestEnforcedForPipeline {

		qualifierMapping := &resourceQualifiers.QualifierMapping{
			ResourceId:            resourceQualifiers.ImageDigestResourceId,
			ResourceType:          resourceQualifiers.ImageDigest,
			QualifierId:           int(resourceQualifiers.PIPELINE_QUALIFIER),
			IdentifierKey:         devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID],
			IdentifierValueInt:    pipelineId,
			Active:                true,
			IdentifierValueString: fmt.Sprintf("%d", pipelineId),
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: UserId,
				UpdatedOn: time.Now(),
				UpdatedBy: UserId,
			},
		}
		_, err := impl.qualifierMappingService.CreateQualifierMappings([]*resourceQualifiers.QualifierMapping{qualifierMapping}, tx)
		if err != nil {
			impl.logger.Errorw("error in creating image digest qualifier mapping for pipeline", "err", err)
			return err
		}

	} else if isDigestPolicyConfiguredForPipeline && !isImageDigestEnforcedForPipeline {
		auditLog := sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: UserId,
			UpdatedOn: time.Now(),
			UpdatedBy: UserId,
		}
		err := impl.qualifierMappingService.DeleteAllQualifierMappingsByIdentifierKeyAndValue(devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID], pipelineId, auditLog, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting image digest policy for pipeline", "err", err, "pipeline id", pipelineId)
			return err
		}
	}
	return nil
}

func (impl ImageDigestPolicyServiceImpl) IsPolicyConfiguredForPipeline(pipelineId int) (bool, error) {
	scope := &resourceQualifiers.Scope{PipelineId: pipelineId}
	resourceIds := []int{resourceQualifiers.ImageDigestResourceId}
	qualifierMappings, err := impl.qualifierMappingService.GetQualifierMappings(resourceQualifiers.ImageDigest, scope, resourceIds)
	if err != nil && err != pg.ErrNoRows {
		return false, err
	}
	if err == pg.ErrNoRows || len(qualifierMappings) == 0 {
		return false, nil
	}
	return true, nil
}

func (impl ImageDigestPolicyServiceImpl) CreateOrUpdatePolicyForCluster(policyRequest *PolicyRequest) (*PolicyRequest, error) {

	dbConnection := impl.qualifierMappingService.GetDbConnection()
	tx, _ := dbConnection.Begin()

	if policyRequest.EnableDigestForAllClusters == true {
		err := impl.handleImageDigestPolicyForAllClusters(policyRequest.UserId, tx)
		if err != nil {
			impl.logger.Errorw("Error in saving image digest policy for all clusters", "err", err)
			return nil, err
		}
		_ = tx.Commit()
		return policyRequest, nil
	} else if len(policyRequest.ClusterDetails) == 0 {
		return policyRequest, &util.ApiError{
			HttpStatusCode:  http2.StatusBadRequest,
			InternalMessage: "all clusters = false and cluster details is also empty",
			UserMessage:     "Please provide cluster details",
		}
	}

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()

	// fetching already configured policies
	imageDigestQualifierMappings, err := impl.qualifierMappingService.GetQualifierMappingsByResourceType(resourceQualifiers.ImageDigest)
	if err != nil {
		impl.logger.Errorw("error in fetching qualifier mappings by resourceType", "resourceType: ", resourceQualifiers.ImageDigest)
		return nil, err
	}

	// exiting cluster and environments already having imageDigest configured
	ExistingClustersWithImageDigestPolicyConfigured, ExistingEnvironmentsWithImageDigestPolicyConfigured := getExistingClustersAndEnvsWithImagePullPolicyConfigured(imageDigestQualifierMappings, devtronResourceSearchableKeyMap)

	// saving image digest policy for new clusters and environments
	newClustersWithImageDigestPolicyConfigured, newEnvironmentsWithImageDigestPolicyConfigured, err :=
		impl.saveNewPolicies(
			policyRequest,
			ExistingClustersWithImageDigestPolicyConfigured, ExistingEnvironmentsWithImageDigestPolicyConfigured, devtronResourceSearchableKeyMap, tx)
	if err != nil {
		impl.logger.Errorw("error in creating image digest policies", "err", err)
		return nil, err
	}

	// removing policies present in db but not present in request
	err = impl.removePoliciesNotPresentInRequest(imageDigestQualifierMappings,
		newClustersWithImageDigestPolicyConfigured,
		newEnvironmentsWithImageDigestPolicyConfigured,
		devtronResourceSearchableKeyMap,
		policyRequest.UserId,
		tx)
	if err != nil {
		impl.logger.Errorw("error in deleting policies not present in request but present in DB", "err", err)
		return nil, err
	}

	_ = tx.Commit()

	return policyRequest, nil
}

func (impl ImageDigestPolicyServiceImpl) handleImageDigestPolicyForAllClusters(userId int32, tx *pg.Tx) error {

	// step1: create image digest policy for all clusters by setting qualifierId = int(resourceQualifiers.GLOBAL_QUALIFIER)
	// step2: Delete individual cluster and env level image digest policy mappings

	globalQualifierMapping := &resourceQualifiers.QualifierMapping{
		ResourceId:   resourceQualifiers.ImageDigestResourceId,
		ResourceType: resourceQualifiers.ImageDigest,
		QualifierId:  int(resourceQualifiers.GLOBAL_QUALIFIER),
		Active:       true,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Time{},
			CreatedBy: userId,
			UpdatedOn: time.Time{},
			UpdatedBy: userId,
		},
	}

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
		impl.logger.Errorw("error in deleting resource by resource type, id and qualifier id", "err", err)
		return err
	}
	return nil
}

func getExistingClustersAndEnvsWithImagePullPolicyConfigured(imageDigestQualifierMappings []*resourceQualifiers.QualifierMapping, devtronResourceSearchableKeyMap map[bean.DevtronResourceSearchableKeyName]int) (map[ClusterId]bool, map[EnvironmentId]bool) {
	ExistingClustersWithImageDigestPolicyConfigured := make(map[ClusterId]bool)
	ExistingEnvironmentsWithImageDigestPolicyConfigured := make(map[EnvironmentId]bool)
	for _, existingMapping := range imageDigestQualifierMappings {
		if existingMapping.IdentifierKey == devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID] {
			ExistingClustersWithImageDigestPolicyConfigured[existingMapping.IdentifierValueInt] = true
		} else if existingMapping.IdentifierKey == devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID] {
			ExistingEnvironmentsWithImageDigestPolicyConfigured[existingMapping.IdentifierValueInt] = true
		}
	}
	return ExistingClustersWithImageDigestPolicyConfigured, ExistingEnvironmentsWithImageDigestPolicyConfigured
}

func (impl ImageDigestPolicyServiceImpl) saveNewPolicies(
	policyRequest *PolicyRequest, ExistingClustersWithImageDigestPolicyConfigured map[ClusterId]bool,
	ExistingEnvironmentsWithImageDigestPolicyConfigured map[EnvironmentId]bool,
	devtronResourceSearchableKeyMap map[bean.DevtronResourceSearchableKeyName]int, tx *pg.Tx) (map[ClusterId]bool, map[EnvironmentId]bool, error) {

	newPolicies := make([]*resourceQualifiers.QualifierMapping, 0)
	newClustersWithImageDigestPolicyConfigured := make(map[ClusterId]bool)
	newEnvironmentsWithImageDigestPolicyConfigured := make(map[EnvironmentId]bool)

	for _, policy := range policyRequest.ClusterDetails {
		if policy.PolicyType == ALL_EXISTING_AND_FUTURE_ENVIRONMENTS {
			if _, ok := ExistingClustersWithImageDigestPolicyConfigured[policy.ClusterId]; !ok {
				qualifierMapping := &resourceQualifiers.QualifierMapping{
					ResourceId:         resourceQualifiers.ImageDigestResourceId,
					ResourceType:       resourceQualifiers.ImageDigest,
					QualifierId:        int(resourceQualifiers.CLUSTER_QUALIFIER),
					IdentifierKey:      devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID],
					IdentifierValueInt: policy.ClusterId,
					Active:             true,
					AuditLog: sql.AuditLog{
						CreatedOn: time.Now(),
						CreatedBy: policyRequest.UserId,
						UpdatedOn: time.Now(),
						UpdatedBy: policyRequest.UserId,
					},
				}
				newPolicies = append(newPolicies, qualifierMapping)
			}
			newClustersWithImageDigestPolicyConfigured[policy.ClusterId] = true
		} else if policy.PolicyType == SPECIFIC_ENVIRONMENTS {
			for _, envId := range policy.EnvironmentIds {
				if _, ok := ExistingEnvironmentsWithImageDigestPolicyConfigured[policy.ClusterId]; !ok {
					qualifierMapping := &resourceQualifiers.QualifierMapping{
						ResourceId:         resourceQualifiers.ImageDigestResourceId,
						ResourceType:       resourceQualifiers.ImageDigest,
						QualifierId:        int(resourceQualifiers.ENV_QUALIFIER),
						IdentifierKey:      devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID],
						IdentifierValueInt: envId,
						Active:             true,
						AuditLog: sql.AuditLog{
							CreatedOn: time.Now(),
							CreatedBy: policyRequest.UserId,
							UpdatedOn: time.Now(),
							UpdatedBy: policyRequest.UserId,
						},
					}
					newPolicies = append(newPolicies, qualifierMapping)
				}
				newEnvironmentsWithImageDigestPolicyConfigured[envId] = true
			}
		}
	}
	if len(newPolicies) > 0 {
		_, err := impl.qualifierMappingService.CreateQualifierMappings(newPolicies, tx)
		if err != nil {
			impl.logger.Errorw("error in creating qualifier mappings for image digest policy", "err", err)
			return newClustersWithImageDigestPolicyConfigured, newEnvironmentsWithImageDigestPolicyConfigured, err
		}
	}
	return newClustersWithImageDigestPolicyConfigured, newEnvironmentsWithImageDigestPolicyConfigured, nil
}

func (impl ImageDigestPolicyServiceImpl) removePoliciesNotPresentInRequest(imageDigestQualifierMappings []*resourceQualifiers.QualifierMapping,
	newClustersWithImageDigestPolicyConfigured map[ClusterId]bool,
	newEnvironmentsWithImageDigestPolicyConfigured map[EnvironmentId]bool,
	devtronResourceSearchableKeyMap map[bean.DevtronResourceSearchableKeyName]int,
	UserId int32,
	tx *pg.Tx) error {

	policiesToBeRemovedIDs := make([]int, 0)

	for _, existingMapping := range imageDigestQualifierMappings {
		removePolicy := false
		if existingMapping.IdentifierKey == devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID] {
			if _, ok := newClustersWithImageDigestPolicyConfigured[existingMapping.IdentifierValueInt]; !ok {
				removePolicy = true
			}
		} else if existingMapping.IdentifierKey == devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID] {
			if _, ok := newEnvironmentsWithImageDigestPolicyConfigured[existingMapping.IdentifierValueInt]; !ok {
				removePolicy = true
			}
		} else if existingMapping.QualifierId == int(resourceQualifiers.GLOBAL_QUALIFIER) {
			// removing global policy because if we are here EnableDigestForAllClusters=false in request
			removePolicy = true
		}
		if removePolicy {
			policiesToBeRemovedIDs = append(policiesToBeRemovedIDs, existingMapping.Id)
		}
	}

	if len(policiesToBeRemovedIDs) > 0 {
		err := impl.qualifierMappingService.DeleteAllByIds(policiesToBeRemovedIDs, UserId, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting old policies", "err", err)
			return err
		}
	}

	return nil
}

func (impl ImageDigestPolicyServiceImpl) GetAllConfiguredGlobalPolicies() (*PolicyRequest, error) {

	imageDigestQualifierMappings, err := impl.qualifierMappingService.GetQualifierMappingsByResourceType(resourceQualifiers.ImageDigest)
	if err != nil {
		impl.logger.Errorw("error in fetching qualifier mappings by resourceType", "resourceType: ", resourceQualifiers.ImageDigest)
		return nil, err
	}

	imageDigestPolicies := &PolicyRequest{
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

func (impl ImageDigestPolicyServiceImpl) IsPolicyConfiguredAtGlobalLevel(envId int, clusterId int) (bool, error) {
	if clusterId == 0 {
		env, err := impl.environmentRepository.FindById(envId)
		if err != nil {
			impl.logger.Errorw("error in fetching environment by envId", "err", err, "envId", envId)
		}
		clusterId = env.ClusterId
	}
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

func (impl ImageDigestPolicyServiceImpl) IsPolicyConfiguredAtGlobalOrPipeline(envId, clusterId, pipelineId int) (bool, error) {
	isImageDigestPolicyConfiguredAtGlobalLevel, err :=
		impl.IsPolicyConfiguredAtGlobalLevel(envId, clusterId)
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
