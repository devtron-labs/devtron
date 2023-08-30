package repository

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

type VariableDefinition struct {
	tableName     struct{} `sql:"variable_definition" pg:",discard_unknown_columns"`
	Id            int      `sql:"id,pk"`
	Name          string   `sql:"name"`
	DataType      string   `sql:"data_type"`
	VarType       string   `sql:"var_type"`
	Active        bool     `sql:"active"`
	Description   string   `sql:"description"`
	VariableScope []*VariableScope
	sql.AuditLog
}

type VariableScope struct {
	tableName             struct{} `sql:"variable_scope" pg:",discard_unknown_columns"`
	Id                    int      `sql:"id,pk"`
	VariableDefinitionId  int      `sql:"variable_definition_id"`
	QualifierId           int      `sql:"qualifier_id"`
	IdentifierKey         int      `sql:"identifier_key"`
	IdentifierValueInt    int      `sql:"identifier_value_int"`
	Active                bool     `sql:"active"`
	IdentifierValueString string   `sql:"identifier_value_string"`
	ParentIdentifier      int      `sql:"parent_identifier"`
	CompositeKey          string   `sql:"-"`
	VariableData          *VariableData
	sql.AuditLog
}

type VariableData struct {
	tableName       struct{} `sql:"variable_data" pg:",discard_unknown_columns"`
	Id              int      `sql:"id,pk"`
	VariableScopeId int      `sql:"variable_scope_id"`
	Data            string   `sql:"data"`
	sql.AuditLog
}

type Qualifier int

const (
	APP_AND_ENV_QUALIFIER Qualifier = 1
	APP_QUALIFIER         Qualifier = 2
	ENV_QUALIFIER         Qualifier = 3
	CLUSTER_QUALIFIER     Qualifier = 4
	GLOBAL_QUALIFIER      Qualifier = 5
)

var CompoundQualifiers = []Qualifier{APP_AND_ENV_QUALIFIER}

type ScopedVariableRepository interface {
	//transaction util funcs
	sql.TransactionWrapper
	CreateVariableDefinition(variableDefinition []*VariableDefinition, tx *pg.Tx) ([]*VariableDefinition, error)
	CreateVariableScope(variableDefinition []*VariableScope, tx *pg.Tx) ([]*VariableScope, error)
	CreateVariableData(variableDefinition []*VariableData, tx *pg.Tx) error
	GetAllVariables() ([]*VariableDefinition, error)
	GetVariablesForVarIds(ids []int) ([]*VariableDefinition, error)
	GetVariablesByNames(vars []string) ([]*VariableDefinition, error)
	GetAllVariableScopeAndDefinition() ([]*VariableDefinition, error)
	GetScopedVariableData(scope models.Scope, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int, varIds []int) ([]*VariableScope, error)
	GetDataForScopeIds(scopeIds []int) ([]*VariableData, error)
	DeleteVariables() error
}

type ScopedVariableRepositoryImpl struct {
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}

func NewScopedVariableRepository(dbConnection *pg.DB) *ScopedVariableRepositoryImpl {
	return &ScopedVariableRepositoryImpl{
		dbConnection:        dbConnection,
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
	var variableDefinition []*VariableDefinition
	err := impl.
		dbConnection.Model(&variableDefinition).
		Where("variable_definition.active = ?", true).
		Select()
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
		Where("name in (?)", helper.GetCommaSepratedString(vars)).Select()
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
func (impl *ScopedVariableRepositoryImpl) GetScopedVariableData(scope models.Scope, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int, varIds []int) ([]*VariableScope, error) {
	var variableScopes []*VariableScope
	query := impl.dbConnection.Model(&variableScopes).
		Where("active = ?", true).
		Where(
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

func (impl *ScopedVariableRepositoryImpl) DeleteVariables() error {
	tx, err := impl.dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer func(tx *pg.Tx) {
		err = tx.Rollback()
		if err != nil {
			return
		}
	}(tx)
	_, err = tx.Model(&VariableScope{}).
		Set("active = ?", false).
		Where("active = ?", true).
		Update()

	_, err = tx.Model(&VariableDefinition{}).
		Set("active = ?", false).
		Where("active = ?", true).
		Update()
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
