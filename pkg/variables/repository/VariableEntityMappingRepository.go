package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
	"time"
)

type VariableEntityMappingRepository interface {
	sql.TransactionWrapper
	GetVariablesForEntities(entities []Entity) ([]*VariableEntityMapping, error)
	SaveVariableEntityMappings(tx *pg.Tx, mappings []*VariableEntityMapping) error
	DeleteAllVariablesForEntities(entities []Entity, userId int32) error
	DeleteVariablesForEntity(tx *pg.Tx, variableIDs []string, entity Entity, userId int32) error
}

func NewVariableEntityMappingRepository(logger *zap.SugaredLogger, dbConnection *pg.DB) *VariableEntityMappingRepositoryImpl {
	return &VariableEntityMappingRepositoryImpl{
		logger:              logger,
		dbConnection:        dbConnection,
		TransactionUtilImpl: sql.NewTransactionUtilImpl(dbConnection),
	}
}

type VariableEntityMappingRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}

func (impl *VariableEntityMappingRepositoryImpl) SaveVariableEntityMappings(tx *pg.Tx, mappings []*VariableEntityMapping) error {
	err := tx.Insert(&mappings)
	if err != nil {
		impl.logger.Errorw("err in saving variable entity mappings", "err", err)
		return err
	}
	return nil
}

func (impl *VariableEntityMappingRepositoryImpl) GetVariablesForEntities(entities []Entity) ([]*VariableEntityMapping, error) {
	mappings := make([]*VariableEntityMapping, 0)

	err := impl.dbConnection.Model(&mappings).
		Where("is_deleted = ?", false).
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			for _, entity := range entities {
				q = q.WhereOr("entity_id = ? AND entity_type = ?", entity.EntityId, entity.EntityType)
			}
			return q, nil
		}).Select()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting variables for entities", "err", err)
		return nil, err
	}
	return mappings, nil
}

func (impl *VariableEntityMappingRepositoryImpl) DeleteVariablesForEntity(tx *pg.Tx, variableNames []string, entity Entity, userId int32) error {

	_, err := tx.Model((*VariableEntityMapping)(nil)).
		Set("is_deleted = ?", true).
		Set("updated_by = ?", userId).
		Set("updated_on = ?", time.Now()).
		Where("variable_name IN (?)", pg.In(variableNames)).
		Where("is_deleted = ?", false).
		Where("entity_id = ? AND entity_type = ?", entity.EntityId, entity.EntityType).
		Update()
	if err != nil {
		impl.logger.Errorw("err in deleting variable entity mappings", "err", err)
		return err
	}
	return nil
}

func (impl *VariableEntityMappingRepositoryImpl) DeleteAllVariablesForEntities(entities []Entity, userId int32) error {
	_, err := impl.dbConnection.Model((*VariableEntityMapping)(nil)).
		Set("is_deleted = ?", true).
		Set("updated_by = ?", userId).
		Set("updated_on = ?", time.Now()).
		Where("is_deleted = ?", false).
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			for _, entity := range entities {
				q = q.WhereOr("entity_id = ? AND entity_type = ?", entity.EntityId, entity.EntityType)
			}
			return q, nil
		}).
		Update()
	if err != nil {
		impl.logger.Errorw("err in deleting variable entity mappings", "err", err)
		return err
	}
	return nil
}
