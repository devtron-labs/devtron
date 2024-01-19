package imageDigestPolicy

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ImageDigestPolicyService interface {

	//CreatePolicyForPipeline creates image digest policy for pipeline
	CreatePolicyForPipeline(pipelineId int, pipelineName string, tx *pg.Tx, UserId int32) (int, error)

	//CreatePolicyForPipelineIfNotExist creates image digest policy for pipeline if not already created
	CreatePolicyForPipelineIfNotExist(tx *pg.Tx, pipelineId int, pipelineName string, UserId int32) (int, error)

	//GetDigestPolicyConfigurations returns true if pipeline or env or cluster has image digest policy enabled
	GetDigestPolicyConfigurations(digestConfigurationRequest DigestPolicyConfigurationRequest) (digestPolicyConfiguration DigestPolicyConfiguration, err error)

	//DeletePolicyForPipeline deletes image digest policy for a pipeline
	DeletePolicyForPipeline(tx *pg.Tx, pipelineId int, userId int32) (int, error)
}

type ImageDigestPolicyServiceImpl struct {
	logger                       *zap.SugaredLogger
	qualifierMappingService      resourceQualifiers.QualifierMappingService
	devtronResourceSearchableKey devtronResource.DevtronResourceSearchableKeyService
}

func NewImageDigestPolicyServiceImpl(
	logger *zap.SugaredLogger,
	qualifierMappingService resourceQualifiers.QualifierMappingService,
	devtronResourceSearchableKey devtronResource.DevtronResourceSearchableKeyService,
) *ImageDigestPolicyServiceImpl {
	return &ImageDigestPolicyServiceImpl{
		logger:                       logger,
		qualifierMappingService:      qualifierMappingService,
		devtronResourceSearchableKey: devtronResourceSearchableKey,
	}
}

func (impl ImageDigestPolicyServiceImpl) CreatePolicyForPipeline(pipelineId int, pipelineName string, tx *pg.Tx, UserId int32) (int, error) {

	var qualifierMappingId int

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()

	identifierKey := devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID]
	identifierValue := pipelineId
	qualifierMapping := QualifierMappingDao(int(resourceQualifiers.PIPELINE_QUALIFIER),
		identifierKey,
		identifierValue,
		UserId,
	)
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
		qualifierMappingId, err = impl.CreatePolicyForPipeline(pipelineId, "", tx, UserId)
		if err != nil {
			impl.logger.Errorw("error in creating policy for pipeline", "err", "pipelineId", pipelineId)
			return qualifierMappingId, nil
		}
	}
	return qualifierMappingId, nil
}

func (impl ImageDigestPolicyServiceImpl) GetDigestPolicyConfigurations(digestConfigurationRequest DigestPolicyConfigurationRequest) (digestPolicyConfiguration DigestPolicyConfiguration, err error) {

	resourceIds := []int{resourceQualifiers.ImageDigestResourceId}

	scope := &resourceQualifiers.Scope{
		PipelineId: digestConfigurationRequest.PipelineId,
	}
	policyMappings, err := impl.qualifierMappingService.GetQualifierMappings(resourceQualifiers.ImageDigest, scope, resourceIds)
	if err != nil && err != pg.ErrNoRows {
		return digestPolicyConfiguration, err
	}
	if err == pg.ErrNoRows || len(policyMappings) == 0 {
		return digestPolicyConfiguration, nil
	}

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()
	clusterIdentifierKey := devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID]
	envIdentifierKey := devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID]
	pipelineIdentifierKey := devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID]
	globalQualifierId := int(resourceQualifiers.GLOBAL_QUALIFIER)

	for _, policy := range policyMappings {
		if policy.QualifierId == globalQualifierId || policy.IdentifierKey == clusterIdentifierKey || pipelineIdentifierKey == envIdentifierKey {
			digestPolicyConfiguration.DigestConfiguredForEnvOrCluster = true
		} else if policy.IdentifierKey == pipelineIdentifierKey {
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
	err := impl.qualifierMappingService.DeleteByIdentifierKeyValue(resourceQualifiers.ImageDigest, devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID], pipelineId, auditLog, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting image digest policy for pipeline", "err", err, "pipelineId", pipelineId)
		return pipelineId, err
	}
	return pipelineId, nil
}

func (impl ImageDigestPolicyServiceImpl) getQualifierMappingForPipeline(pipelineId int) ([]*resourceQualifiers.QualifierMapping, error) {
	scope := &resourceQualifiers.Scope{PipelineId: pipelineId}
	resourceIds := []int{resourceQualifiers.ImageDigestResourceId}
	qualifierMappings, err := impl.qualifierMappingService.GetQualifierMappings(resourceQualifiers.ImageDigest, scope, resourceIds)
	if err != nil && err != pg.ErrNoRows {
		return qualifierMappings, err
	}
	return qualifierMappings, err
}
