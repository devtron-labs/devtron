package resourceQualifiers

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type QualifiersMappingRepository interface {
	//transaction util funcs
	sql.TransactionWrapper
	CreateQualifierMappings(qualifierMappings []*QualifierMapping, tx *pg.Tx) ([]*QualifierMapping, error)
	GetQualifierMappings(resourceType ResourceType, scope *Scope, searchableIdMap map[bean.DevtronResourceSearchableKeyName]int, resourceIds []int) ([]*QualifierMapping, error)
	DeleteAllQualifierMappings(ResourceType, sql.AuditLog, *pg.Tx) error
	DeleteByResourceTypeIdentifierKeyAndValue(resourceType ResourceType, identifierKey int, identifierValue int, auditLog sql.AuditLog, tx *pg.Tx) error
	GetDbConnection() *pg.DB
}

type QualifiersMappingRepositoryImpl struct {
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
	logger *zap.SugaredLogger
}

func NewQualifiersMappingRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) (*QualifiersMappingRepositoryImpl, error) {
	return &QualifiersMappingRepositoryImpl{
		dbConnection:        dbConnection,
		logger:              logger,
		TransactionUtilImpl: sql.NewTransactionUtilImpl(dbConnection),
	}, nil
}

func (repo *QualifiersMappingRepositoryImpl) CreateQualifierMappings(qualifierMappings []*QualifierMapping, tx *pg.Tx) ([]*QualifierMapping, error) {
	err := tx.Insert(&qualifierMappings)
	if err != nil {
		return nil, err
	}
	return qualifierMappings, nil
}

func (repo *QualifiersMappingRepositoryImpl) addScopeWhereClause(query *orm.Query, scope *Scope, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) *orm.Query {
	return query.Where(
		"( (identifier_key = ? AND identifier_value_int = ?)  AND qualifier_id = ?) "+
			"OR (qualifier_id = ? ) ",
		searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID], scope.PipelineId, PIPELINE_QUALIFIER,
		GLOBAL_QUALIFIER)
}

func (repo *QualifiersMappingRepositoryImpl) GetQualifierMappings(resourceType ResourceType, scope *Scope, searchableIdMap map[bean.DevtronResourceSearchableKeyName]int, resourceIds []int) ([]*QualifierMapping, error) {
	var qualifierMappings []*QualifierMapping
	query := repo.dbConnection.Model(&qualifierMappings).
		Where("active = ?", true).
		Where("resource_type = ?", resourceType)

	if len(resourceIds) > 0 {
		query = query.Where("resource_id IN (?)", pg.In(resourceIds))
	}

	if scope != nil {
		query = repo.addScopeWhereClause(query, scope, searchableIdMap)
	}

	err := query.Select()
	if err != nil {
		return nil, err
	}
	return qualifierMappings, nil
}

func (repo *QualifiersMappingRepositoryImpl) DeleteAllQualifierMappings(resourceType ResourceType, auditLog sql.AuditLog, tx *pg.Tx) error {
	_, err := repo.getQualifierMappingDeleteQuery(resourceType, tx, auditLog).
		Update()
	return err
}

func (repo *QualifiersMappingRepositoryImpl) DeleteByResourceTypeIdentifierKeyAndValue(resourceType ResourceType, identifierKey int, identifierValue int, auditLog sql.AuditLog, tx *pg.Tx) error {
	_, err := repo.getQualifierMappingDeleteQuery(resourceType, tx, auditLog).
		Where("identifier_key = ?", identifierKey).
		Where("identifier_value_int = ?", identifierValue).
		Update()
	return err
}

func (repo *QualifiersMappingRepositoryImpl) getQualifierMappingDeleteQuery(resourceType ResourceType, tx *pg.Tx, auditLog sql.AuditLog) *orm.Query {
	return tx.Model(&QualifierMapping{}).
		Set("updated_by = ?", auditLog.UpdatedBy).
		Set("updated_on = ?", auditLog.UpdatedOn).
		Set("active = ?", false).
		Where("active = ?", true).
		Where("resource_type = ? ", resourceType)
}

func (repo *QualifiersMappingRepositoryImpl) GetDbConnection() *pg.DB {
	return repo.dbConnection
}
