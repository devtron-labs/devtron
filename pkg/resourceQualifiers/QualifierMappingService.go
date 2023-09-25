package resourceQualifiers

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type QualifierMappingService interface {
	CreateQualifierMappings(qualifierMappings []*QualifierMapping, tx *pg.Tx) ([]*QualifierMapping, error)
	GetQualifierMappings(resourceType ResourceType, resourceIds []int) ([]*QualifierMapping, error)
	DeleteAllQualifierMappings(sql.AuditLog, *pg.Tx) error
}

type QualifierMappingServiceImpl struct {
	logger                     *zap.SugaredLogger
	qualifierMappingRepository QualifiersMappingRepository
}

func NewQualifierMappingServiceImpl(logger *zap.SugaredLogger, qualifierMappingRepository QualifiersMappingRepository) (*QualifierMappingServiceImpl, error) {
	return &QualifierMappingServiceImpl{
		logger:                     logger,
		qualifierMappingRepository: qualifierMappingRepository,
	}, nil
}

func (impl QualifierMappingServiceImpl) CreateQualifierMappings(qualifierMappings []*QualifierMapping, tx *pg.Tx) ([]*QualifierMapping, error) {
	return impl.qualifierMappingRepository.CreateQualifierMappings(qualifierMappings, tx)
}

func (impl QualifierMappingServiceImpl) GetQualifierMappings(resourceType ResourceType, resourceIds []int) ([]*QualifierMapping, error) {
	return impl.qualifierMappingRepository.GetQualifierMappings(resourceType, resourceIds)
}

func (impl QualifierMappingServiceImpl) DeleteAllQualifierMappings(auditLog sql.AuditLog, tx *pg.Tx) error {
	return impl.qualifierMappingRepository.DeleteAllQualifierMappings(auditLog, tx)
}
