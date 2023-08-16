package repository

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type VariableEntityMappingRepository interface {
	GetConnection() (dbConnection *pg.DB)
	GetVariablesForEntities(entities []Entity) ([]*VariableEntityMapping, error)
	SaveVariableEntityMappings(tx *pg.Tx, mappings []*VariableEntityMapping) error
	DeleteAllVariablesForEntities(entities []Entity) error
	DeleteVariablesForEntity(tx *pg.Tx, variableIDs []int, entity Entity) error
}

func NewVariableEntityMappingRepository(logger *zap.SugaredLogger, dbConnection *pg.DB) *VariableEntityMappingRepositoryImpl {
	return &VariableEntityMappingRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type VariableEntityMappingRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func (impl *VariableEntityMappingRepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
}

func (impl *VariableEntityMappingRepositoryImpl) SaveVariableEntityMappings(tx *pg.Tx, mappings []*VariableEntityMapping) error {
	err := tx.Insert(mappings)
	if err != nil {
		impl.logger.Errorw("err in saving variable entity mappings", "err", err)
		return err
	}
	return nil
}

func (impl *VariableEntityMappingRepositoryImpl) GetVariablesForEntities(entities []Entity) ([]*VariableEntityMapping, error) {
	var mappings []*VariableEntityMapping

	err := impl.dbConnection.Model(&mappings).
		Where("is_deleted = ?", false).
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			for _, entity := range entities {
				q = q.WhereOr("entity_id = ? AND entity_type = ?", entity.EntityId, entity.EntityType)
			}
			return q, nil
		}).Select()
	if err != nil {
		impl.logger.Errorw("err in getting variables for entities", "err", err)
		return nil, err
	}
	return mappings, nil
}

func (impl *VariableEntityMappingRepositoryImpl) DeleteVariablesForEntity(tx *pg.Tx, variableIDs []int, entity Entity) error {

	_, err := tx.Model((*VariableEntityMapping)(nil)).
		Set("is_deleted = ?", true).
		Where("variable_id IN (?)", pg.In(variableIDs)).
		Where("entity_id = ? AND entity_type = ?", entity.EntityId, entity.EntityType).
		//WhereGroup(func(q *orm.Query) (*orm.Query, error) {
		//	for _, entity := range entities {
		//		q = q.WhereOr("entity_id = ? AND entity_type = ?", entity.EntityId, entity.EntityType)
		//	}
		//	return q, nil
		//}).
		Update()
	if err != nil {
		impl.logger.Errorw("err in deleting variable entity mappings", "err", err)
		return err
	}
	return nil
}

func (impl *VariableEntityMappingRepositoryImpl) DeleteAllVariablesForEntities(entities []Entity) error {
	_, err := impl.dbConnection.Model((*VariableEntityMapping)(nil)).
		Set("is_deleted = ?", true).
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
