package imageDigestPolicy

import (
	"fmt"
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

func (impl ImageDigestPolicyServiceImpl) CreatePolicyForPipelineIfNotExist(tx *pg.Tx, pipelineId int, UserId int32) (int, error) {

	var qualifierMappingId int

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()

	isDigestPolicyConfiguredForPipeline, err := impl.IsPolicyConfiguredForPipeline(pipelineId)
	if err != nil {
		impl.logger.Errorw("Error in checking if isDigestPolicyConfiguredForPipeline", "err", err)
		return 0, err
	}

	if !isDigestPolicyConfiguredForPipeline {
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
		_, err = impl.qualifierMappingService.CreateQualifierMappings([]*resourceQualifiers.QualifierMapping{qualifierMapping}, tx)
		if err != nil {
			impl.logger.Errorw("error in creating image digest qualifier mapping for pipeline", "err", err)
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
	}
	if err == pg.ErrNoRows || len(qualifierMappings) == 0 {
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
	err := impl.qualifierMappingService.DeleteByResourceTypeIdentifierKeyAndValue(resourceQualifiers.ImageDigest, devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID], pipelineId, auditLog, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting image digest policy for pipeline", "err", err, "pipeline id", pipelineId)
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
