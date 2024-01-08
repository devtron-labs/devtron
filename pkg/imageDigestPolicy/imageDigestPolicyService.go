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

type ImageDigestQualifierMappingService interface {
	//CreateOrDeletePolicyForPipeline created policy for enforcing pull using digest at pipeline level
	CreateOrDeletePolicyForPipeline(pipelineId int, isImageDigestEnforcedForPipeline bool, UserId int32) error

	//IsPolicyConfiguredForPipeline returns true if pipeline or env or cluster has image digest policy enabled
	IsPolicyConfiguredForPipeline(pipelineId int) (bool, error)
}

type ImageDigestQualifierMappingServiceImpl struct {
	logger                       *zap.SugaredLogger
	qualifierMappingService      resourceQualifiers.QualifierMappingService
	devtronResourceSearchableKey devtronResource.DevtronResourceSearchableKeyService
}

func NewImageDigestQualifierMappingServiceImpl(
	logger *zap.SugaredLogger,
	qualifierMappingService resourceQualifiers.QualifierMappingService,
	devtronResourceSearchableKey devtronResource.DevtronResourceSearchableKeyService,
) *ImageDigestQualifierMappingServiceImpl {
	return &ImageDigestQualifierMappingServiceImpl{
		logger:                       logger,
		qualifierMappingService:      qualifierMappingService,
		devtronResourceSearchableKey: devtronResourceSearchableKey,
	}
}

func (impl ImageDigestQualifierMappingServiceImpl) CreateOrDeletePolicyForPipeline(pipelineId int, isImageDigestEnforcedInRequest bool, UserId int32) error {

	devtronResourceSearchableKeyMap := impl.devtronResourceSearchableKey.GetAllSearchableKeyNameIdMap()

	qualifierMappings, err := impl.getQualifierMappingForPipeline(pipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching qualifier mappings for resourceType: imageDigest by pipelineId", "pipelineId", pipelineId)
		return err
	}

	if err == pg.ErrNoRows && isImageDigestEnforcedInRequest {

		qualifierMapping := &resourceQualifiers.QualifierMapping{
			ResourceId:            resourceQualifiers.ImageDigestResourceId,
			ResourceType:          resourceQualifiers.ImageDigest,
			QualifierId:           int(resourceQualifiers.APP_AND_ENV_QUALIFIER),
			IdentifierKey:         devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID],
			IdentifierValueInt:    pipelineId,
			Active:                false,
			IdentifierValueString: fmt.Sprintf("%d", pipelineId),
			AuditLog:              sql.AuditLog{},
		}

		dbConnection := impl.qualifierMappingService.GetDbConnection()
		tx, _ := dbConnection.Begin()
		_, err := impl.qualifierMappingService.CreateQualifierMappings([]*resourceQualifiers.QualifierMapping{qualifierMapping}, tx)
		if err != nil {
			impl.logger.Errorw("error in creating image digest qualifier mapping for pipeline", "err", err)
			return err
		}
		_ = tx.Commit()

	} else if !isImageDigestEnforcedInRequest && len(qualifierMappings) > 0 {

		dbConnection := impl.qualifierMappingService.GetDbConnection()
		tx, _ := dbConnection.Begin()
		auditLog := sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: UserId,
			UpdatedOn: time.Now(),
			UpdatedBy: UserId,
		}
		err := impl.qualifierMappingService.DeleteAllQualifierMappingsByIdentifierKeyAndValue(devtronResourceSearchableKeyMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID], pipelineId, auditLog, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting image digest policy for pipeline", "err", err, "pipeline id", pipelineId)
		}
		_ = tx.Commit()

	}
	return nil
}

func (impl ImageDigestQualifierMappingServiceImpl) IsPolicyConfiguredForPipeline(pipelineId int) (bool, error) {
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

func (impl ImageDigestQualifierMappingServiceImpl) getQualifierMappingForPipeline(pipelineId int) ([]*resourceQualifiers.QualifierMapping, error) {
	scope := &resourceQualifiers.Scope{PipelineId: pipelineId}
	resourceIds := []int{resourceQualifiers.ImageDigestResourceId}
	qualifierMappings, err := impl.qualifierMappingService.GetQualifierMappings(resourceQualifiers.ImageDigest, scope, resourceIds)
	if err != nil && err != pg.ErrNoRows {
		return qualifierMappings, err
	}
	return qualifierMappings, nil
}
