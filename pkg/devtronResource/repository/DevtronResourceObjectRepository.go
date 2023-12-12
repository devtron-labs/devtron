package repository

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
	"time"
)

type DevtronResourceObjectRepository interface {
	GetConnection() *pg.DB
	Save(model *DevtronResourceObject) (*DevtronResourceObject, error)
	Update(model *DevtronResourceObject) (*DevtronResourceObject, error)
	FindByOldObjectId(oldObjectId, devtronResourceSchemaId int) (*DevtronResourceObject, error)
	FindByObjectName(name string, devtronResourceSchemaId int) (*DevtronResourceObject, error)
	GetChildObjectsByParentArgAndSchemaId(argumentValue interface{}, argumentType string,
		devtronResourceSchemaId int) ([]*DevtronResourceObject, error)
	GetDownstreamObjectsByParentArgAndSchemaIds(argumentValues []interface{}, argumentTypes []string,
		devtronResourceSchemaIds []int) ([]*DevtronResourceObject, error)
	GetObjectsByArgAndSchemaIds(allArgumentValues []interface{},
		allArgumentTypes []string, devtronSchemaIdsForArgsForAllArguments []int) ([]*DevtronResourceObject, error)
	GetBySchemaIdAndOldObjectIdsMap(mapOfResourceSchemaIdAndDependencyIds map[int][]int) ([]*DevtronResourceObject, error)
	DeleteObject(tx *pg.Tx, oldObjectId, devtronResourceId int, updatedBy int32) error
	DeleteDependencyInObjectData(tx *pg.Tx, oldObjectId, devtronResourceId int, updatedBy int32) error
}

type DevtronResourceObjectRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewDevtronResourceObjectRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *DevtronResourceObjectRepositoryImpl {
	return &DevtronResourceObjectRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type DevtronResourceObject struct {
	tableName               struct{} `sql:"devtron_resource_object" pg:",discard_unknown_columns"`
	Id                      int      `sql:"id,pk"`
	OldObjectId             int      `sql:"old_object_id"` //id of object present across different tables, idea is to migrate this new object id
	Name                    string   `sql:"name"`
	DevtronResourceId       int      `sql:"devtron_resource_id"`
	DevtronResourceSchemaId int      `sql:"devtron_resource_schema_id"`
	ObjectData              string   `sql:"object_data"` //json string
	Deleted                 bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

func (repo *DevtronResourceObjectRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}

func (repo *DevtronResourceObjectRepositoryImpl) Save(model *DevtronResourceObject) (*DevtronResourceObject, error) {
	err := repo.dbConnection.Insert(model)
	if err != nil {
		repo.logger.Errorw("error in saving devtronResourceObject", "err", err, "model", model)
		return nil, err
	}
	return model, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) Update(model *DevtronResourceObject) (*DevtronResourceObject, error) {
	err := repo.dbConnection.Update(model)
	if err != nil {
		repo.logger.Errorw("error in updating devtronResourceObject", "err", err, "model", model)
		return nil, err
	}
	return model, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) FindByOldObjectId(oldObjectId, devtronResourceSchemaId int) (*DevtronResourceObject, error) {
	var devtronResourceObject DevtronResourceObject
	err := repo.dbConnection.Model(&devtronResourceObject).Where("old_object_id =?", oldObjectId).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting devtronResourceSchema by oldObjectId", "err", err,
			"oldObjectId", oldObjectId, "devtronResourceSchemaId", devtronResourceSchemaId)
		return nil, err
	}
	return &devtronResourceObject, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) FindByObjectName(name string, devtronResourceSchemaId int) (*DevtronResourceObject, error) {
	var devtronResourceObject DevtronResourceObject
	err := repo.dbConnection.Model(&devtronResourceObject).Where("name =?", name).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting devtronResourceSchema by name", "err", err,
			"name", name, "devtronResourceSchemaId", devtronResourceSchemaId)
		return nil, err
	}
	return &devtronResourceObject, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) GetChildObjectsByParentArgAndSchemaId(argumentValue interface{}, argumentType string,
	devtronResourceSchemaId int) ([]*DevtronResourceObject, error) {
	var models []*DevtronResourceObject
	query := repo.dbConnection.Model(&models).Where("deleted = ?", false)
	query.WhereGroup(func(query *orm.Query) (*orm.Query, error) {
		query.WhereOr(getChildWhereClauseByArgValueTypeAndSchemaId(argumentValue, argumentType, devtronResourceSchemaId))
		return query, nil
	})
	err := query.Select()
	if err != nil {
		repo.logger.Errorw("error, GetChildObjectsByParentArgAndSchemaId", "err", err,
			"argumentValue", argumentValue, "argumentType", argumentType, "devtronResourceSchemaId", devtronResourceSchemaId)

		return nil, err
	}
	return models, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) GetDownstreamObjectsByParentArgAndSchemaIds(argumentValues []interface{}, argumentTypes []string,
	devtronResourceSchemaIds []int) ([]*DevtronResourceObject, error) {
	var models []*DevtronResourceObject
	query := repo.dbConnection.Model(&models).Where("deleted = ?", false)
	query.WhereGroup(func(query *orm.Query) (*orm.Query, error) {
		for i := range argumentValues {
			query.WhereOr(getDownstreamWhereClauseByArgValueTypeAndSchemaId(argumentValues[i], argumentTypes[i], devtronResourceSchemaIds[i]))
		}
		return query, nil
	})
	err := query.Select()
	if err != nil {
		repo.logger.Errorw("error, GetDownstreamObjectsByParentArgAndSchemaIds", "err", err,
			"argumentValues", argumentValues, "argumentTypes", argumentTypes, "devtronResourceSchemaIds", devtronResourceSchemaIds)
		return nil, err
	}
	return models, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) GetObjectsByArgAndSchemaIds(allArgumentValues []interface{},
	allArgumentTypes []string, devtronSchemaIdsForArgsForAllArguments []int) ([]*DevtronResourceObject, error) {
	var models []*DevtronResourceObject
	query := repo.dbConnection.Model(&models).Where("deleted = ?", false)
	query.WhereGroup(func(query *orm.Query) (*orm.Query, error) {
		for i := range allArgumentValues {
			query.WhereOr(getObjectWhereClauseByArgValueTypeAndSchemaId(allArgumentValues[i], allArgumentTypes[i],
				devtronSchemaIdsForArgsForAllArguments[i]))
		}
		return query, nil
	})
	err := query.Select()
	if err != nil {
		repo.logger.Errorw("error, GetObjectsByArgAndSchemaIds", "err", err, "allArgumentValues", allArgumentValues,
			"allArgumentTypes", allArgumentTypes, "devtronResourceSchemaIds", devtronSchemaIdsForArgsForAllArguments)
		return nil, err
	}
	return models, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) GetBySchemaIdAndOldObjectIdsMap(mapOfResourceSchemaIdAndDependencyIds map[int][]int) ([]*DevtronResourceObject, error) {
	var models []*DevtronResourceObject
	query := repo.dbConnection.Model(&models).Where("deleted = ?", false)
	query.WhereGroup(func(query *orm.Query) (*orm.Query, error) {
		for devtronResourceSchemaId, dependencyIds := range mapOfResourceSchemaIdAndDependencyIds {
			query.WhereGroup(func(query *orm.Query) (*orm.Query, error) {
				query = query.Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).
					Where("old_object_id in (?)", pg.In(dependencyIds))
				return query, nil
			})
		}
		return query, nil
	})
	err := query.Select()
	if err != nil {
		repo.logger.Errorw("error, GetBySchemaIdAndOldObjectIdsMap", "err", err, "mapOfResourceSchemaIdAndDependencyIds", mapOfResourceSchemaIdAndDependencyIds)
		return nil, err
	}
	return models, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) DeleteObject(tx *pg.Tx, oldObjectId, devtronResourceId int, updatedBy int32) error {
	_, err := tx.Model((*DevtronResourceObject)(nil)).Set("deleted = ?", true).
		Set("updated_on = ?", time.Now()).Set("updated_by = ?", updatedBy).
		Where("old_object_id = ?", oldObjectId).Where("devtron_resource_id = ?", devtronResourceId).
		Where("deleted = ?", false).Update()
	if err != nil {
		repo.logger.Errorw("error, DeleteObject", "err", err, "oldObjectId", oldObjectId, "devtronResourceId", devtronResourceId)
		return err
	}
	return nil
}

func (repo *DevtronResourceObjectRepositoryImpl) DeleteDependencyInObjectData(tx *pg.Tx, oldObjectId, devtronResourceId int, updatedBy int32) error {
	query := `UPDATE devtron_resource_object
				SET object_data = jsonb_set(
   				 object_data,
    			'{dependencies}',
    			(
					SELECT jsonb_agg(elem)
					FROM jsonb_array_elements(object_data->'dependencies') elem
					WHERE (elem->>'id')::int <> '?' OR (elem->>'devtronResourceId')::int <> '?'
				)::jsonb
			), updated_on = ?, updated_by = ?
			WHERE deleted = ?;`
	_, err := tx.Query((*DevtronResourceObject)(nil), query, oldObjectId, devtronResourceId, time.Now(), updatedBy, false)
	if err != nil {
		repo.logger.Errorw("error, DeleteDependencyInObjectData", "err", err, "oldObjectId", oldObjectId, "devtronResourceId", devtronResourceId)
		return err
	}
	return nil
}

func getChildWhereClauseByArgValueTypeAndSchemaId(arg interface{}, argType string, schemaId int) string {
	return fmt.Sprintf(`object_data -> 'dependencies' @> '[{"%s": %v, "devtronResourceSchemaId": %d, "typeOfDependency": "parent"}]'`, argType, arg, schemaId)
}

func getDownstreamWhereClauseByArgValueTypeAndSchemaId(arg interface{}, argType string, schemaId int) string {
	return fmt.Sprintf(`object_data -> 'dependencies' @> '[{"%s": %v, "devtronResourceSchemaId": %d, "typeOfDependency": "upstream"}]'`, argType, arg, schemaId)
}

func getObjectWhereClauseByArgValueTypeAndSchemaId(arg interface{}, argType string, schemaId int) string {
	return fmt.Sprintf(`%s = %v and devtron_resource_schema_id = %d`, argType, arg, schemaId)
}
