package repository

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type ScopedVariableRepository interface {
	//transaction util funcs
	sql.TransactionWrapper
	// Create
	CreateVariableDefinition(variableDefinition []*VariableDefinition, tx *pg.Tx) ([]*VariableDefinition, error)
	CreateVariableScope(variableDefinition []*VariableScope, tx *pg.Tx) ([]*VariableScope, error)
	CreateVariableData(variableDefinition []*VariableData, tx *pg.Tx) error

	// Get
	GetAllVariables() ([]*VariableDefinition, error)
	GetAllVariableMetadata() ([]*VariableDefinition, error)
	GetVariablesForVarIds(ids []int) ([]*VariableDefinition, error)
	GetVariablesByNames(vars []string) ([]*VariableDefinition, error)
	GetAllVariableScopeAndDefinition() ([]*VariableDefinition, error)
	GetScopedVariableData(scope models.Scope, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int, varIds []int) ([]*VariableScope, error)
	GetDataForScopeIds(scopeIds []int) ([]*VariableData, error)

	// Delete
	DeleteVariables(auditLog sql.AuditLog, tx *pg.Tx) error
}

type ScopedVariableRepositoryImpl struct {
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
	logger *zap.SugaredLogger
}

func NewScopedVariableRepository(dbConnection *pg.DB, logger *zap.SugaredLogger) *ScopedVariableRepositoryImpl {
	return &ScopedVariableRepositoryImpl{
		dbConnection:        dbConnection,
		logger:              logger,
		TransactionUtilImpl: sql.NewTransactionUtilImpl(dbConnection),
	}
}

func (impl *ScopedVariableRepositoryImpl) CreateVariableDefinition(variableDefinition []*VariableDefinition, tx *pg.Tx) ([]*VariableDefinition, error) {
	err := tx.Insert(&variableDefinition)
	if err != nil {
		return nil, err
	}
	return variableDefinition, nil
}

func (impl *ScopedVariableRepositoryImpl) CreateVariableScope(variableScope []*VariableScope, tx *pg.Tx) ([]*VariableScope, error) {
	err := tx.Insert(&variableScope)
	if err != nil {
		return nil, err
	}
	return variableScope, nil

}

func (impl *ScopedVariableRepositoryImpl) CreateVariableData(variableDefinition []*VariableData, tx *pg.Tx) error {
	return tx.Insert(&variableDefinition)
}

func (impl *ScopedVariableRepositoryImpl) GetAllVariables() ([]*VariableDefinition, error) {
	variableDefinition := make([]*VariableDefinition, 0)
	err := impl.
		dbConnection.Model(&variableDefinition).
		Where("variable_definition.active = ?", true).
		Select()
	return variableDefinition, err
}

func (impl *ScopedVariableRepositoryImpl) GetAllVariableMetadata() ([]*VariableDefinition, error) {
	variableDefinition := make([]*VariableDefinition, 0)
	err := impl.
		dbConnection.Model(&variableDefinition).
		Column("id", "name", "data_type", "var_type,short_description").
		Where("active = ?", true).
		Select()
	if err == pg.ErrNoRows {
		err = nil
	}
	return variableDefinition, err
}

func (impl *ScopedVariableRepositoryImpl) GetVariablesForVarIds(ids []int) ([]*VariableDefinition, error) {
	var variableDefinition []*VariableDefinition
	err := impl.
		dbConnection.Model(&variableDefinition).
		Where("id in (?)", pg.In(ids)).
		Where("variable_definition.active = ?", true).
		Select()
	return variableDefinition, err
}

func (impl *ScopedVariableRepositoryImpl) GetVariablesByNames(vars []string) ([]*VariableDefinition, error) {
	var variableDefinition []*VariableDefinition
	err := impl.dbConnection.Model(&variableDefinition).Where("active = ?", true).
		Where("name in (?)", pg.In(vars)).Select()
	impl.logger.Info("variableDefinition: ", variableDefinition)
	return variableDefinition, err
}

func (impl *ScopedVariableRepositoryImpl) GetAllVariableScopeAndDefinition() ([]*VariableDefinition, error) {
	var variableDefinition []*VariableDefinition
	err := impl.dbConnection.
		Model(&variableDefinition).
		Column("variable_definition.*", "VariableScope", "VariableScope.VariableData").
		Relation("VariableScope", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("variable_scope.active = ?", true), nil
		}).
		Where("variable_definition.active = ?", true).
		Select()
	if err != nil {
		return nil, err
	}
	return variableDefinition, err

}

func (impl *ScopedVariableRepositoryImpl) addScopeWhereClause(query *orm.Query, scope models.Scope, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) *orm.Query {
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
func (impl *ScopedVariableRepositoryImpl) GetScopedVariableData(scope models.Scope, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int, varIds []int) ([]*VariableScope, error) {
	var variableScopes []*VariableScope
	query := impl.dbConnection.Model(&variableScopes).
		Where("active = ?", true)

	//Enterprise only
	query = impl.addScopeWhereClause(query, scope, searchableKeyNameIdMap)

	if len(varIds) > 0 {
		query = query.Where("variable_definition_id IN (?)", pg.In(varIds))
	}

	err := query.Select()
	if err != nil {
		return nil, err
	}
	return variableScopes, nil
}

func (impl *ScopedVariableRepositoryImpl) GetDataForScopeIds(scopeIds []int) ([]*VariableData, error) {
	var variableData []*VariableData
	err := impl.dbConnection.
		Model(&variableData).
		Where("variable_scope_id in(?)", pg.In(scopeIds)).
		Select()
	return variableData, err

}

func (impl *ScopedVariableRepositoryImpl) DeleteVariables(auditLog sql.AuditLog, tx *pg.Tx) error {
	_, err := tx.Model(&VariableScope{}).
		Set("updated_by = ?", auditLog.UpdatedBy).
		Set("updated_on = ?", auditLog.UpdatedOn).
		Set("active = ?", false).
		Where("active = ?", true).
		Update()
	if err != nil {
		return err
	}
	_, err = tx.Model(&VariableDefinition{}).
		Set("updated_by = ?", auditLog.UpdatedBy).
		Set("updated_on = ?", auditLog.UpdatedOn).
		Set("active = ?", false).
		Where("active = ?", true).
		Update()
	return err
}
