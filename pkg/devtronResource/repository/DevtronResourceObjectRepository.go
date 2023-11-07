package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DevtronResourceObjectRepository interface {
	Save(model *DevtronResourceObject) (*DevtronResourceObject, error)
	Update(model *DevtronResourceObject) (*DevtronResourceObject, error)
	FindByOldObjectId(oldObjectId, devtronResourceId, devtronResourceSchemaId int) (*DevtronResourceObject, error)
	FindByObjectName(name string, devtronResourceId, devtronResourceSchemaId int) (*DevtronResourceObject, error)
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
	Deleted                 bool     `sql:"deleted"`
	sql.AuditLog
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

func (repo *DevtronResourceObjectRepositoryImpl) FindByOldObjectId(oldObjectId, devtronResourceId, devtronResourceSchemaId int) (*DevtronResourceObject, error) {
	var devtronResourceObject DevtronResourceObject
	err := repo.dbConnection.Model(&devtronResourceObject).
		Where("old_object_id =?", oldObjectId).Where("devtron_resource_id = ?", devtronResourceId).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).Select()
	if err != nil {
		repo.logger.Errorw("error in getting devtronResourceSchema by oldObjectId", "err", err,
			"oldObjectId", oldObjectId, "devtronResourceId", devtronResourceId, "devtronResourceSchemaId", devtronResourceSchemaId)
		return nil, err
	}
	return &devtronResourceObject, nil
}

func (repo *DevtronResourceObjectRepositoryImpl) FindByObjectName(name string, devtronResourceId, devtronResourceSchemaId int) (*DevtronResourceObject, error) {
	var devtronResourceObject DevtronResourceObject
	err := repo.dbConnection.Model(&devtronResourceObject).
		Where("name =?", name).Where("devtron_resource_id = ?", devtronResourceId).
		Where("devtron_resource_schema_id = ?", devtronResourceSchemaId).Select()
	if err != nil {
		repo.logger.Errorw("error in getting devtronResourceSchema by name", "err", err,
			"name", name, "devtronResourceId", devtronResourceId, "devtronResourceSchemaId", devtronResourceSchemaId)
		return nil, err
	}
	return &devtronResourceObject, nil
}
