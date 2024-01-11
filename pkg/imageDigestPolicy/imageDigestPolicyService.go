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
	//CreateOrDeletePolicyForPipeline created policy for enforcing pull using digest at pipeline level
	CreateOrDeletePolicyForPipeline(pipelineId int, isImageDigestEnforcedForPipeline bool, UserId int32, tx *pg.Tx) error

	//IsPolicyConfiguredForPipeline returns true if pipeline or env or cluster has image digest policy enabled
	IsPolicyConfiguredForPipeline(pipelineId int) (bool, error)
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

func (impl ImageDigestPolicyServiceImpl) CreateOrDeletePolicyForPipeline(pipelineId int, isImageDigestEnforcedForPipeline bool, UserId int32, tx *pg.Tx) error {

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()

	qualifierMappings, err := impl.getQualifierMappingForPipeline(pipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching qualifier mappings for resourceType: imageDigest by pipelineId", "pipelineId", pipelineId)
		return err
	}

	if len(qualifierMappings) == 0 && isImageDigestEnforcedForPipeline {

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

	} else if !isImageDigestEnforcedForPipeline && len(qualifierMappings) > 0 {
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
	qualifierMappings, err := impl.getQualifierMappingForPipeline(pipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching qualifier mappings for resourceType: imageDigest by pipelineId", "pipelineId", pipelineId)
		return false, err
	}
	if err == pg.ErrNoRows || len(qualifierMappings) == 0 {
		return false, nil
	}
	return true, nil
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
