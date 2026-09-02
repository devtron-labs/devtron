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
	GetLiveUsagesForVariableNames(variableNames []string) ([]*VariableUsageRow, error)
	SaveVariableEntityMappings(tx *pg.Tx, mappings []*VariableEntityMapping) error
	DeleteAllVariablesForEntities(tx *pg.Tx, entities []Entity, userId int32) error
	DeleteVariablesForEntity(tx *pg.Tx, variableIDs []string, entity Entity, userId int32) error
}

// VariableUsageRow is one live reference of a variable, resolved to the app/env/pipeline it belongs to.
// A mapping row is considered live only when the entity that holds the reference is still the
// currently-served config (latest chart / latest env override / non-deleted pipeline or config map)
// and its app/environment is active.
type VariableUsageRow struct {
	VariableName string     `sql:"variable_name"`
	EntityType   EntityType `sql:"entity_type"`
	EntityId     int        `sql:"entity_id"`
	AppId        int        `sql:"app_id"`
	AppName      string     `sql:"app_name"`
	EnvId        int        `sql:"env_id"`
	EnvName      string     `sql:"env_name"`
	PipelineName string     `sql:"pipeline_name"`
	StageType    string     `sql:"stage_type"`
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

const liveVariableUsageQuery = `
SELECT vem.variable_name, vem.entity_type, vem.entity_id,
       a.id AS app_id, COALESCE(NULLIF(a.display_name, ''), a.app_name) AS app_name,
       NULL::int AS env_id, NULL::varchar AS env_name,
       NULL::varchar AS pipeline_name, NULL::varchar AS stage_type
FROM variable_entity_mapping vem
INNER JOIN charts c ON c.id = vem.entity_id AND c.active = true AND c.latest = true
INNER JOIN app a ON a.id = c.app_id AND a.active = true
WHERE vem.entity_type = 1 AND vem.is_deleted = false AND vem.variable_name IN (?)
UNION ALL
SELECT vem.variable_name, vem.entity_type, vem.entity_id,
       a.id, COALESCE(NULLIF(a.display_name, ''), a.app_name),
       e.id, e.environment_name, NULL, NULL
FROM variable_entity_mapping vem
INNER JOIN chart_env_config_override ceco ON ceco.id = vem.entity_id AND ceco.active = true AND ceco.latest = true
INNER JOIN charts c ON c.id = ceco.chart_id
INNER JOIN app a ON a.id = c.app_id AND a.active = true
INNER JOIN environment e ON e.id = ceco.target_environment AND e.active = true
WHERE vem.entity_type = 2 AND vem.is_deleted = false AND vem.variable_name IN (?)
UNION ALL
SELECT vem.variable_name, vem.entity_type, vem.entity_id,
       a.id, COALESCE(NULLIF(a.display_name, ''), a.app_name),
       e.id, e.environment_name,
       COALESCE(cp.name, p.pipeline_name), ps.type::varchar
FROM variable_entity_mapping vem
INNER JOIN pipeline_stage ps ON ps.id = vem.entity_id AND ps.deleted = false
LEFT JOIN ci_pipeline cp ON cp.id = ps.ci_pipeline_id AND cp.deleted = false
LEFT JOIN pipeline p ON p.id = ps.cd_pipeline_id AND p.deleted = false
LEFT JOIN environment e ON e.id = p.environment_id
INNER JOIN app a ON a.id = COALESCE(cp.app_id, p.app_id) AND a.active = true
WHERE vem.entity_type = 3 AND vem.is_deleted = false AND vem.variable_name IN (?)
  AND (cp.id IS NOT NULL OR p.id IS NOT NULL)
UNION ALL
SELECT vem.variable_name, vem.entity_type, vem.entity_id,
       a.id, COALESCE(NULLIF(a.display_name, ''), a.app_name),
       NULL, NULL, NULL, NULL
FROM variable_entity_mapping vem
INNER JOIN config_map_app_level cmal ON cmal.id = vem.entity_id
INNER JOIN app a ON a.id = cmal.app_id AND a.active = true
WHERE vem.entity_type IN (4, 6) AND vem.is_deleted = false AND vem.variable_name IN (?)
UNION ALL
SELECT vem.variable_name, vem.entity_type, vem.entity_id,
       a.id, COALESCE(NULLIF(a.display_name, ''), a.app_name),
       e.id, e.environment_name, NULL, NULL
FROM variable_entity_mapping vem
INNER JOIN config_map_env_level cmel ON cmel.id = vem.entity_id AND cmel.deleted = false
INNER JOIN app a ON a.id = cmel.app_id AND a.active = true
INNER JOIN environment e ON e.id = cmel.environment_id AND e.active = true
WHERE vem.entity_type IN (5, 7) AND vem.is_deleted = false AND vem.variable_name IN (?)
`

func (impl *VariableEntityMappingRepositoryImpl) GetLiveUsagesForVariableNames(variableNames []string) ([]*VariableUsageRow, error) {
	usages := make([]*VariableUsageRow, 0)
	if len(variableNames) == 0 {
		return usages, nil
	}
	names := pg.In(variableNames)
	_, err := impl.dbConnection.Query(&usages, liveVariableUsageQuery, names, names, names, names, names)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting live usages for variable names", "variableNames", variableNames, "err", err)
		return nil, err
	}
	return usages, nil
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
