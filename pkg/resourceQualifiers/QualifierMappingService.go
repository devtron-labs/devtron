package resourceQualifiers

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type QualifierMappingService interface {
	CreateQualifierMappings(qualifierMappings []*QualifierMapping, tx *pg.Tx) ([]*QualifierMapping, error)
	GetQualifierMappingsForFilter(scope Scope) ([]*QualifierMapping, error)
	GetQualifierMappingsForFilterById(resourceId int) ([]*QualifierMapping, error)
	GetQualifierMappings(resourceType ResourceType, scope *Scope, resourceIds []int) ([]*QualifierMapping, error)
	GetQualifierMappingsByResourceType(resourceType ResourceType) ([]*QualifierMapping, error)
	DeleteAllQualifierMappings(resourceType ResourceType, auditLog sql.AuditLog, tx *pg.Tx) error
	DeleteAllQualifierMappingsByResourceTypeAndId(resourceType ResourceType, resourceId int, auditLog sql.AuditLog, tx *pg.Tx) error
	DeleteByResourceTypeIdentifierKeyAndValue(resourceType ResourceType, identifierKey int, identifierValue int, auditLog sql.AuditLog, tx *pg.Tx) error
	DeleteAllByResourceTypeAndQualifierIds(resourceType ResourceType, resourceId int, qualifierIds []int, userId int32, tx *pg.Tx) error
	DeleteAllByIds(qualifierMappingIds []int, userId int32, tx *pg.Tx) error
	GetDbConnection() *pg.DB
}

type QualifierMappingServiceImpl struct {
	logger                              *zap.SugaredLogger
	qualifierMappingRepository          QualifiersMappingRepository
	devtronResourceSearchableKeyService devtronResource.DevtronResourceSearchableKeyService
}

func NewQualifierMappingServiceImpl(logger *zap.SugaredLogger, qualifierMappingRepository QualifiersMappingRepository, devtronResourceSearchableKeyService devtronResource.DevtronResourceSearchableKeyService) (*QualifierMappingServiceImpl, error) {
	return &QualifierMappingServiceImpl{
		logger:                              logger,
		qualifierMappingRepository:          qualifierMappingRepository,
		devtronResourceSearchableKeyService: devtronResourceSearchableKeyService,
	}, nil
}

func (impl QualifierMappingServiceImpl) CreateQualifierMappings(qualifierMappings []*QualifierMapping, tx *pg.Tx) ([]*QualifierMapping, error) {
	return impl.qualifierMappingRepository.CreateQualifierMappings(qualifierMappings, tx)
}

func (impl QualifierMappingServiceImpl) GetQualifierMappings(resourceType ResourceType, scope *Scope, resourceIds []int) ([]*QualifierMapping, error) {
	searchableIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()
	return impl.qualifierMappingRepository.GetQualifierMappings(resourceType, scope, searchableIdMap, resourceIds)
}

func (impl QualifierMappingServiceImpl) GetQualifierMappingsForFilter(scope Scope) ([]*QualifierMapping, error) {
	searchableKeyNameIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()
	return impl.qualifierMappingRepository.GetQualifierMappingsForFilter(scope, searchableKeyNameIdMap)
}

func (impl QualifierMappingServiceImpl) GetQualifierMappingsForFilterById(resourceId int) ([]*QualifierMapping, error) {
	return impl.qualifierMappingRepository.GetQualifierMappingsForFilterById(resourceId)
}

func (impl QualifierMappingServiceImpl) GetQualifierMappingsByResourceType(resourceType ResourceType) ([]*QualifierMapping, error) {
	return impl.qualifierMappingRepository.GetQualifierMappingsByResourceType(resourceType)
}

func (impl QualifierMappingServiceImpl) DeleteAllQualifierMappings(resourceType ResourceType, auditLog sql.AuditLog, tx *pg.Tx) error {
	return impl.qualifierMappingRepository.DeleteAllQualifierMappings(resourceType, auditLog, tx)
}
func (impl QualifierMappingServiceImpl) DeleteAllQualifierMappingsByResourceTypeAndId(resourceType ResourceType, resourceId int, auditLog sql.AuditLog, tx *pg.Tx) error {
	return impl.qualifierMappingRepository.DeleteAllQualifierMappingsByResourceTypeAndId(resourceType, resourceId, auditLog, tx)
}

func (impl QualifierMappingServiceImpl) DeleteByResourceTypeIdentifierKeyAndValue(resourceType ResourceType, identifierKey int, identifierValue int, auditLog sql.AuditLog, tx *pg.Tx) error {
	return impl.qualifierMappingRepository.DeleteByResourceTypeIdentifierKeyAndValue(resourceType, identifierKey, identifierValue, auditLog, tx)
}

func (impl QualifierMappingServiceImpl) DeleteAllByResourceTypeAndQualifierIds(resourceType ResourceType, resourceId int, qualifierIds []int, userId int32, tx *pg.Tx) error {
	auditLog := sql.AuditLog{
		CreatedOn: time.Now(),
		CreatedBy: userId,
		UpdatedOn: time.Now(),
		UpdatedBy: userId,
	}
	return impl.qualifierMappingRepository.DeleteAllByResourceTypeAndQualifierId(resourceType, resourceId, qualifierIds, auditLog, tx)
}

func (impl QualifierMappingServiceImpl) DeleteAllByIds(qualifierMappingIds []int, userId int32, tx *pg.Tx) error {
	auditLog := sql.AuditLog{
		CreatedOn: time.Now(),
		CreatedBy: userId,
		UpdatedOn: time.Now(),
		UpdatedBy: userId,
	}
	return impl.qualifierMappingRepository.DeleteAllByIds(qualifierMappingIds, auditLog, tx)
}

func (impl QualifierMappingServiceImpl) GetDbConnection() *pg.DB {
	return impl.qualifierMappingRepository.GetDbConnection()
}
