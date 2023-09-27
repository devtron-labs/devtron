package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ScopedVariableRepository interface {
	//transaction util funcs
	sql.TransactionWrapper
	// Create
	CreateVariableDefinition(variableDefinition []*VariableDefinition, tx *pg.Tx) ([]*VariableDefinition, error)
	CreateVariableData(variableDefinition []*VariableData, tx *pg.Tx) error

	// Get
	GetAllVariables() ([]*VariableDefinition, error)
	GetAllVariableMetadata() ([]*VariableDefinition, error)
	GetVariablesForVarIds(ids []int) ([]*VariableDefinition, error)
	GetVariablesByNames(vars []string) ([]*VariableDefinition, error)
	GetAllVariableDefinition() ([]*VariableDefinition, error)
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
	impl.logger.Debug("variableDefinition: ", variableDefinition)
	return variableDefinition, err
}

func (impl *ScopedVariableRepositoryImpl) GetAllVariableDefinition() ([]*VariableDefinition, error) {
	var variableDefinition []*VariableDefinition
	err := impl.dbConnection.
		Model(&variableDefinition).
		Column("variable_definition.*").
		Where("variable_definition.active = ?", true).
		Select()
	if err != nil {
		return nil, err
	}
	return variableDefinition, err

}

func (impl *ScopedVariableRepositoryImpl) GetDataForScopeIds(scopeIds []int) ([]*VariableData, error) {
	var variableData []*VariableData
	if len(scopeIds) == 0 {
		return variableData, nil
	}
	err := impl.dbConnection.
		Model(&variableData).
		Where("variable_scope_id in(?)", pg.In(scopeIds)).
		Select()
	return variableData, err

}

func (impl *ScopedVariableRepositoryImpl) DeleteVariables(auditLog sql.AuditLog, tx *pg.Tx) error {
	_, err := tx.Model(&VariableDefinition{}).
		Set("updated_by = ?", auditLog.UpdatedBy).
		Set("updated_on = ?", auditLog.UpdatedOn).
		Set("active = ?", false).
		Where("active = ?", true).
		Update()
	return err
}
