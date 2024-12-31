/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	DeleteAllVariablesForEntities(tx *pg.Tx, entities []Entity, userId int32) error
	DeleteVariablesForEntity(tx *pg.Tx, variableIDs []string, entity Entity, userId int32) error
}

func NewVariableEntityMappingRepository(logger *zap.SugaredLogger, dbConnection *pg.DB, TransactionUtilImpl *sql.TransactionUtilImpl) *VariableEntityMappingRepositoryImpl {
	return &VariableEntityMappingRepositoryImpl{
		logger:              logger,
		dbConnection:        dbConnection,
		TransactionUtilImpl: TransactionUtilImpl,
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

func (impl *VariableEntityMappingRepositoryImpl) DeleteAllVariablesForEntities(tx *pg.Tx, entities []Entity, userId int32) error {

	var connection orm.DB
	connection = tx
	if tx == nil {
		connection = impl.dbConnection
	}

	_, err := connection.Model((*VariableEntityMapping)(nil)).
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
