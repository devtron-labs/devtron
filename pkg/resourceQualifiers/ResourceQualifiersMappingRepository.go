package resourceQualifiers

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type QualifiersMappingRepository interface {
	// transaction util funcs
	sql.TransactionWrapper
	CreateQualifierMappings(qualifierMappings []*QualifierMapping, tx *pg.Tx) ([]*QualifierMapping, error)
	GetQualifierMappings(resourceType ResourceType, scope *Scope, searchableIdMap map[bean.DevtronResourceSearchableKeyName]int, resourceIds []int) ([]*QualifierMapping, error)
	GetQualifierMappingsForFilter(scope Scope, searchableIdMap map[bean.DevtronResourceSearchableKeyName]int) ([]*QualifierMapping, error)
	GetQualifierMappingsForFilterById(resourceId int) ([]*QualifierMapping, error)
	DeleteAllQualifierMappings(ResourceType, sql.AuditLog, *pg.Tx) error
	DeleteAllQualifierMappingsByResourceTypeAndId(resourceType ResourceType, resourceId int, auditLog sql.AuditLog, tx *pg.Tx) error
	DeleteGivenQualifierMappingsByResourceType(resourceType ResourceType, identifierKey int, identifierValueInts []int, auditLog sql.AuditLog, tx *pg.Tx) error
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

func (impl *QualifiersMappingRepositoryImpl) addScopeWhereClauseForFilter(query *orm.Query, scope Scope, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) *orm.Query {
	return query.Where(
		"((identifier_key = ? AND identifier_value_int = ?) "+
			"OR (identifier_key = ? AND identifier_value_int IN (?)) "+
			"OR (identifier_key = ? AND identifier_value_int = ?) "+
			"OR (identifier_key = ? AND identifier_value_int IN (?)))",
		searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID], scope.AppId,
		searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_ID], pg.In([]int{scope.ProjectId, AllProjectsInt}),
		searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID], scope.EnvId,
		searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID], pg.In([]int{scope.ClusterId, GetEnvIdentifierValue(scope)}),
	)
}

func (repo *QualifiersMappingRepositoryImpl) GetQualifierMappingsForFilter(scope Scope, searchableIdMap map[bean.DevtronResourceSearchableKeyName]int) ([]*QualifierMapping, error) {
	var qualifierMappings []*QualifierMapping
	query := repo.dbConnection.Model(&qualifierMappings).
		Where("active = ?", true).
		Where("resource_type = ?", Filter)

	query = repo.addScopeWhereClauseForFilter(query, scope, searchableIdMap)
	err := query.Select()
	if err == pg.ErrNoRows {
		repo.logger.Errorw("no qualifier mappings found", "scope", scope)
		err = nil
	}
	return qualifierMappings, err
}
func (repo *QualifiersMappingRepositoryImpl) GetQualifierMappingsForFilterById(resourceId int) ([]*QualifierMapping, error) {
	var qualifierMappings []*QualifierMapping
	err := repo.dbConnection.Model(&qualifierMappings).
		Where("active = ?", true).
		Where("resource_type = ?", Filter).
		Where("resource_id = ?", resourceId).
		Select()
	if err == pg.ErrNoRows {
		return qualifierMappings, errors.New(fmt.Sprintf("no qualifier mappings found for given filter id %v", resourceId))
	}
	return qualifierMappings, nil
}

func (repo *QualifiersMappingRepositoryImpl) addScopeWhereClause(query *orm.Query, scope *Scope, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) *orm.Query {
	return query.Where(
		"(((identifier_key = ? AND identifier_value_int = ?) OR (identifier_key = ? AND identifier_value_int = ?)) AND qualifier_id = ?) "+
			"OR (qualifier_id = ? AND identifier_key = ? AND identifier_value_int = ?) "+
			"OR (qualifier_id = ? AND identifier_key = ? AND identifier_value_int = ?) "+
			"OR (qualifier_id = ? AND identifier_key = ? AND identifier_value_int = ?) "+
			"OR (qualifier_id = ?)",
		searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID], scope.AppId, searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID], scope.EnvId, APP_AND_ENV_QUALIFIER,
		APP_QUALIFIER, searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID], scope.AppId,
		ENV_QUALIFIER, searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID], scope.EnvId,
		CLUSTER_QUALIFIER, searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID], scope.ClusterId,
		GLOBAL_QUALIFIER,
	)
}

func (repo *QualifiersMappingRepositoryImpl) GetQualifierMappings(resourceType ResourceType, scope *Scope, searchableIdMap map[bean.DevtronResourceSearchableKeyName]int, resourceIds []int) ([]*QualifierMapping, error) {
	var qualifierMappings []*QualifierMapping
	query := repo.dbConnection.Model(&qualifierMappings).
		Where("active = ?", true).
		Where("resource_type = ?", resourceType)

	if len(resourceIds) > 0 {
		query = query.Where("resource_id IN (?)", pg.In(resourceIds))
	}

	// Enterprise Only
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
	_, err := tx.Model(&QualifierMapping{}).
		Set("updated_by = ?", auditLog.UpdatedBy).
		Set("updated_on = ?", auditLog.UpdatedOn).
		Set("active = ?", false).
		Where("active = ?", true).
		Where("resource_type = ?", resourceType).
		Update()
	return err
}
func (repo *QualifiersMappingRepositoryImpl) DeleteAllQualifierMappingsByResourceTypeAndId(resourceType ResourceType, resourceId int, auditLog sql.AuditLog, tx *pg.Tx) error {
	_, err := tx.Model(&QualifierMapping{}).
		Set("updated_by = ?", auditLog.UpdatedBy).
		Set("updated_on = ?", auditLog.UpdatedOn).
		Set("active = ?", false).
		Where("active = ?", true).
		Where("resource_type = ?", resourceType).
		Where("resource_id = ?", resourceId).
		Update()
	return err
}

func (repo *QualifiersMappingRepositoryImpl) DeleteGivenQualifierMappingsByResourceType(resourceType ResourceType, identifierKey int, identifierValueInts []int, auditLog sql.AuditLog, tx *pg.Tx) error {
	_, err := tx.Model(&QualifierMapping{}).
		Set("updated_by=?", auditLog.UpdatedBy).
		Set("updated_on=?", auditLog.UpdatedOn).
		Set("active=?", false).
		Where("active=?", true).
		Where("resource_type=?", resourceType).
		Where("identifier_value_int IN (?)", pg.In(identifierValueInts)).
		Where("identifier_key=?", identifierKey).
		Update()
	return err
}
