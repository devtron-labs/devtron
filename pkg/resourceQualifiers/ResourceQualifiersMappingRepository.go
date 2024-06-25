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
	DeleteAllByIds(qualifierMappingIds []int, auditLog sql.AuditLog, tx *pg.Tx) error
	GetDbConnection() *pg.DB
	GetMappingsByResourceTypeAndIdsAndQualifierId(resourceType ResourceType, resourceIds []int, qualifier int) ([]*QualifierMapping, error)
	GetQualifierMappingsForListOfQualifierValues(resourceType ResourceType, valuesMap map[Qualifier][][]int, searchableIdMap map[bean.DevtronResourceSearchableKeyName]int, resourceIds []int) ([]*QualifierMapping, error)
}

const appEnvCondition = "(((identifier_key = ? AND identifier_value_int in (?)) OR (identifier_key = ? AND identifier_value_int in (?))) AND qualifier_id = ?)"
const condition = "(qualifier_id = ? AND identifier_key = ? AND identifier_value_int in (?))"

type QualifiersMappingRepositoryImpl struct {
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
	logger *zap.SugaredLogger
}

func NewQualifiersMappingRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger, TransactionUtilImpl *sql.TransactionUtilImpl) (*QualifiersMappingRepositoryImpl, error) {
	return &QualifiersMappingRepositoryImpl{
		dbConnection:        dbConnection,
		logger:              logger,
		TransactionUtilImpl: TransactionUtilImpl,
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

func (repo *QualifiersMappingRepositoryImpl) DeleteAllByIds(qualifierMappingIds []int, auditLog sql.AuditLog, tx *pg.Tx) error {
	_, err := tx.Model(&QualifierMapping{}).
		Set("updated_by = ?", auditLog.UpdatedBy).
		Set("updated_on = ?", auditLog.UpdatedOn).
		Set("active = ?", false).
		Where("active = ?", true).
		Where("id in (?)", pg.In(qualifierMappingIds)).
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

func (repo *QualifiersMappingRepositoryImpl) GetMappingsByResourceTypeAndIdsAndQualifierId(resourceType ResourceType, resourceIds []int, qualifier int) ([]*QualifierMapping, error) {
	mappings := make([]*QualifierMapping, 0)
	if len(resourceIds) == 0 {
		return mappings, nil
	}
	err := repo.dbConnection.Model(&mappings).
		Where("active = ?", true).
		Where("resource_type = ?", resourceType).
		Where("resource_id IN (?)", pg.In(resourceIds)).
		Where("qualifier_id = ?", qualifier).
		Select()
	if err == pg.ErrNoRows {
		err = nil
	}
	return mappings, err
}

func (repo *QualifiersMappingRepositoryImpl) GetQualifierMappingsForListOfQualifierValues(resourceType ResourceType, valuesMap map[Qualifier][][]int, searchableIdMap map[bean.DevtronResourceSearchableKeyName]int, resourceIds []int) ([]*QualifierMapping, error) {
	var qualifierMappings []*QualifierMapping
	query := repo.dbConnection.Model(&qualifierMappings).
		Where("active = ?", true).
		Where("resource_type = ?", resourceType)

	if len(resourceIds) > 0 {
		query = query.Where("resource_id IN (?)", pg.In(resourceIds))
	}

	// Enterprise Only
	if valuesMap != nil {
		query = repo.addScopeWhereClauseBatch(query, valuesMap, searchableIdMap)
	}

	err := query.Select()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	return qualifierMappings, nil
}

func addCond(query *orm.Query, qualifier Qualifier, valuesMap map[Qualifier][][]int, identifierKey int) *orm.Query {
	if values, ok := valuesMap[qualifier]; ok && len(values[0]) > 0 {
		query = query.WhereOr(condition,
			qualifier, identifierKey, pg.In(valuesMap[qualifier][0]),
		)
	}
	return query
}

func (repo *QualifiersMappingRepositoryImpl) addScopeWhereClauseBatch(q *orm.Query, valuesMap map[Qualifier][][]int, drs map[bean.DevtronResourceSearchableKeyName]int) *orm.Query {

	q = q.WhereGroup(func(query *orm.Query) (*orm.Query, error) {
		if len(valuesMap[APP_AND_ENV_QUALIFIER][0]) > 0 && len(valuesMap[APP_AND_ENV_QUALIFIER][1]) > 0 {
			query = query.WhereOr(appEnvCondition,
				drs[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID], pg.In(valuesMap[APP_AND_ENV_QUALIFIER][0]),
				drs[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID], pg.In(valuesMap[APP_AND_ENV_QUALIFIER][1]),
				APP_AND_ENV_QUALIFIER,
			)
		}
		query = addCond(query, APP_QUALIFIER, valuesMap, drs[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID])
		query = addCond(query, ENV_QUALIFIER, valuesMap, drs[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID])
		query = addCond(query, CLUSTER_QUALIFIER, valuesMap, drs[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID])
		query = addCond(query, PIPELINE_QUALIFIER, valuesMap, drs[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID])
		query = query.WhereOr("(qualifier_id = ?)", GLOBAL_QUALIFIER)
		return query, nil
	})
	return q
}
