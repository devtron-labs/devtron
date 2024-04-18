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
	Save(tx *pg.Tx, model *DevtronResourceObject) error
	Update(tx *pg.Tx, model *DevtronResourceObject) error
	UpdateIdentifier(id, devtronResourceSchemaId int, identifier string) error

	// FindByIdAndSchemaId will fetch the DevtronResourceObject by DevtronResourceObject.Id and DevtronResourceObject.DevtronResourceSchemaId
	//
	// DevtronResourceObject.Id is unique and incremental for all kinds of resource (Applications/Job/Release/ReleaseTrack)
	FindByIdAndSchemaId(id, devtronResourceSchemaId int) (*DevtronResourceObject, error)
	// GetAllWithSchemaId will list out all the objects specific to a resource schema
	GetAllWithSchemaId(devtronResourceSchemaId int) ([]*DevtronResourceObject, error)
	// FindByOldObjectId will fetch the DevtronResourceObject by DevtronResourceObject.OldObjectId
	//
	// DevtronResourceObject.OldObjectId refers the id column of the resource's own table
	//
	// For example: In DevtronResourceObject OldObjectId for Application resource -> app.id
	FindByOldObjectId(oldObjectId, devtronResourceSchemaId int) (*DevtronResourceObject, error)
	GetAllObjectByIdsOrOldObjectIds(objectIds, oldObjectIds []int, devtronResourceSchemaId int) ([]*DevtronResourceObject, error)
	FindAllObjects() ([]*DevtronResourceObject, error)
	FindByObjectName(name string, devtronResourceSchemaId int) (*DevtronResourceObject, error)
	FindByObjectIdentifier(name string, devtronResourceSchemaId int) (*DevtronResourceObject, error)

	CheckIfExistById(id, devtronResourceSchemaId int) (bool, error)
	CheckIfExistByName(name string, devtronResourceSchemaId int) (bool, error)
	CheckIfExistByIdentifier(identifier string, devtronResourceSchemaId int) (bool, error)

	SoftDeleteById(id, devtronResourceSchemaId int) (*DevtronResourceObject, error)
	SoftDeleteByName(name string, devtronResourceSchemaId int) (*DevtronResourceObject, error)
	SoftDeleteByIdentifier(name string, devtronResourceSchemaId int) (*DevtronResourceObject, error)

	GetChildObjectsByParentArgAndSchemaId(argumentValue interface{}, argumentType string,
		devtronResourceSchemaId int) ([]*DevtronResourceObject, error)
	GetDownstreamObjectsByParentArgAndSchemaIds(argumentValues []interface{}, argumentTypes []string,
		devtronResourceSchemaIds []int) ([]*DevtronResourceObject, error)
	GetObjectsByArgAndSchemaIds(allArgumentValues []interface{},
		allArgumentTypes []string, devtronSchemaIdsForArgsForAllArguments []int) ([]*DevtronResourceObject, error)

	DeleteObject(tx *pg.Tx, oldObjectId, devtronResourceId int, updatedBy int32) error
	DeleteDependencyInObjectData(tx *pg.Tx, oldObjectId, devtronResourceId int, updatedBy int32) error
	DeleteKeysFromObjectData(tx *pg.Tx, pathsToRemove []string, resourceId int, userId int) error
	sql.TransactionWrapper
}

type DevtronResourceObjectRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}

func NewDevtronResourceObjectRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *DevtronResourceObjectRepositoryImpl {
	return &DevtronResourceObjectRepositoryImpl{
		logger:              logger,
		dbConnection:        dbConnection,
		TransactionUtilImpl: sql.NewTransactionUtilImpl(dbConnection),
	}
}

type DevtronResourceObject struct {
	tableName               struct{} `sql:"devtron_resource_object" pg:",discard_unknown_columns"`
	Id                      int      `sql:"id,pk"`
	OldObjectId             int      `sql:"old_object_id"` //id of object present across different tables, idea is to migrate this new object id
	name                    string   `sql:"name"`          // making this private as this will not be used in future as well, it is not unique, everything will work on id
	Identifier              string   `sql:"identifier"`    // unique identifier for identification for release - its release-track-name and version, for release-track will be name
	DevtronResourceId       int      `sql:"devtron_resource_id"`
	DevtronResourceSchemaId int      `sql:"devtron_resource_schema_id"`
	ObjectData              string   `sql:"object_data"` //json string
	Deleted                 bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

func (repo *DevtronResourceObjectRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}

func (repo *DevtronResourceObjectRepositoryImpl) Save(tx *pg.Tx, model *DevtronResourceObject) error {
	var err error
	if tx != nil {
		err = tx.Insert(model)
	} else {
		err = repo.dbConnection.Insert(model)
	}
	return err
}

func (repo *DevtronResourceObjectRepositoryImpl) Update(tx *pg.Tx, model *DevtronResourceObject) error {
	var err error
	if tx != nil {
		err = tx.Update(model)
	} else {
		err = repo.dbConnection.Update(model)
	}
	return err
}

func (repo *DevtronResourceObjectRepositoryImpl) UpdateIdentifier(id, devtronResourceSchemaId int, identifier string) error {
	var devtronResourceObject DevtronResourceObject
	_, err := repo.dbConnection.Model(&devtronResourceObject).Where("id =?", id).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).
		Set("identifier = ?", identifier).Update()
	return err
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

func (repo *DevtronResourceObjectRepositoryImpl) FindByIdAndSchemaId(id, devtronResourceSchemaId int) (*DevtronResourceObject, error) {
	var devtronResourceObject DevtronResourceObject
	err := repo.dbConnection.Model(&devtronResourceObject).Where("id =?", id).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting devtronResourceObject by id and devtronResourceSchemaId", "err", err,
			"id", id, "devtronResourceSchemaId", devtronResourceSchemaId)
		return nil, err
	}
	return &devtronResourceObject, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) GetAllWithSchemaId(devtronResourceSchemaId int) ([]*DevtronResourceObject, error) {
	var models []*DevtronResourceObject
	err := repo.dbConnection.Model(&models).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).
		Where("deleted = ?", false).
		Select()
	if err != nil {
		repo.logger.Errorw("error in getting devtronResourceObject by id", "err", err,
			"devtronResourceSchemaId", devtronResourceSchemaId)
		return nil, err
	}
	return models, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) GetAllObjectByIdsOrOldObjectIds(objectIds, oldObjectIds []int,
	devtronResourceSchemaId int) ([]*DevtronResourceObject, error) {
	var models []*DevtronResourceObject
	query := repo.dbConnection.Model(&models).Where("deleted = ?", false).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId)
	query.WhereGroup(func(query *orm.Query) (*orm.Query, error) {
		if len(objectIds) > 0 {
			query.WhereOr("id in (?)", pg.In(objectIds))
		}
		if len(oldObjectIds) > 0 {
			query.WhereOr("old_object_id in (?)", pg.In(oldObjectIds))
		}
		return query, nil
	})
	err := query.Select()
	if err != nil {
		repo.logger.Errorw("error, GetAllChildWithObjectIdOrOldObjectId", "err", err,
			"objectIds", objectIds, "oldObjectIds", oldObjectIds, "schemaId", devtronResourceSchemaId)
		return nil, err
	}
	return models, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) CheckIfExistById(id, devtronResourceSchemaId int) (bool, error) {
	var devtronResourceObject DevtronResourceObject
	exists, err := repo.dbConnection.Model(&devtronResourceObject).
		Where("id =?", id).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).
		Where("deleted = ?", false).Exists()
	if err != nil {
		repo.logger.Errorw("error in getting devtronResourceSchema by oldObjectId", "err", err,
			"id", id, "devtronResourceSchemaId", devtronResourceSchemaId)
		return false, err
	}
	return exists, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) FindAllObjects() ([]*DevtronResourceObject, error) {
	var models []*DevtronResourceObject
	err := repo.dbConnection.Model(&models).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting all resource objects", "err", err)
		return nil, err
	}
	return models, nil
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

func (repo *DevtronResourceObjectRepositoryImpl) FindByObjectIdentifier(identifier string, devtronResourceSchemaId int) (*DevtronResourceObject, error) {
	var devtronResourceObject DevtronResourceObject
	err := repo.dbConnection.Model(&devtronResourceObject).Where("identifier =?", identifier).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting devtronResourceSchema by name", "err", err,
			"identifier", identifier, "devtronResourceSchemaId", devtronResourceSchemaId)
		return nil, err
	}
	return &devtronResourceObject, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) CheckIfExistByName(name string, devtronResourceSchemaId int) (bool, error) {
	var devtronResourceObject DevtronResourceObject
	exists, err := repo.dbConnection.Model(&devtronResourceObject).Where("name =?", name).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).
		Where("deleted = ?", false).Exists()
	if err != nil {
		repo.logger.Errorw("error in getting CheckIfExistByName", "err", err,
			"name", name, "devtronResourceSchemaId", devtronResourceSchemaId)
		return false, err
	}
	return exists, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) CheckIfExistByIdentifier(identifier string, devtronResourceSchemaId int) (bool, error) {
	var devtronResourceObject DevtronResourceObject
	exists, err := repo.dbConnection.Model(&devtronResourceObject).Where("identifier =?", identifier).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).
		Where("deleted = ?", false).Exists()
	if err != nil {
		repo.logger.Errorw("error in getting CheckIfExistByIdentifier", "err", err,
			"identifier", identifier, "devtronResourceSchemaId", devtronResourceSchemaId)
		return false, err
	}
	return exists, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) SoftDeleteById(id, devtronResourceSchemaId int) (*DevtronResourceObject, error) {
	var devtronResourceObject DevtronResourceObject
	_, err := repo.dbConnection.Model(&devtronResourceObject).Where("id =?", id).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).
		Set("deleted = ?", true).Update()
	return &devtronResourceObject, err
}

func (repo *DevtronResourceObjectRepositoryImpl) SoftDeleteByName(name string, devtronResourceSchemaId int) (*DevtronResourceObject, error) {
	var devtronResourceObject DevtronResourceObject
	_, err := repo.dbConnection.Model(&devtronResourceObject).Where("name =?", name).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).
		Set("deleted = ?", true).Update()
	return &devtronResourceObject, err
}

func (repo *DevtronResourceObjectRepositoryImpl) SoftDeleteByIdentifier(identifier string, devtronResourceSchemaId int) (*DevtronResourceObject, error) {
	var devtronResourceObject DevtronResourceObject
	_, err := repo.dbConnection.Model(&devtronResourceObject).Where("identifier =?", identifier).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).
		Set("deleted = ?", true).Update()
	return &devtronResourceObject, err
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
    			coalesce((
					SELECT jsonb_agg(elem)
					FROM jsonb_array_elements(object_data->'dependencies') elem
					WHERE (elem->>'id')::int <> '?' OR (elem->>'devtronResourceId')::int <> '?'
				)::jsonb, '[]')
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

func (repo *DevtronResourceObjectRepositoryImpl) DeleteKeysFromObjectData(tx *pg.Tx, pathsToRemove []string, resourceId int, userId int) error {
	var models []*DevtronResourceObject
	//using fmt.Sprintf for setting o
	query := fmt.Sprintf(`UPDATE devtron_resource_object  
				SET object_data = %s, updated_by = ?, updated_on = ?
		 			WHERE devtron_resource_id = ? AND deleted = ?;`, getNewObjectDataWithRemovedPaths(pathsToRemove))
	_, err := tx.Query(models, query, userId, time.Now(), resourceId, false)
	return err
}

func getNewObjectDataWithRemovedPaths(pathsToRemove []string) string {
	jsonQuery := "object_data "
	if len(pathsToRemove) == 0 || pathsToRemove[0] == "" {
		jsonQuery = "NULL "
	} else {
		for _, path := range pathsToRemove {
			jsonQuery += fmt.Sprintf("#- '{%s}' ", path)
		}
	}
	return jsonQuery
}
