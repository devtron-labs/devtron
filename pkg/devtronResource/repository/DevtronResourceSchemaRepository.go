package repository

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DevtronResourceSchemaRepository interface {
	Save(model *DevtronResourceSchema) error
	Update(model *DevtronResourceSchema) error
	FindByResourceId(id int) (*DevtronResourceSchema, error)
	FindSchemaByKindSubKindAndVersion(kind string, subKind string, version string) (*DevtronResourceSchema, error)
	GetAll() ([]*DevtronResourceSchema, error)
}

type DevtronResourceSchemaRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewDevtronResourceSchemaRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *DevtronResourceSchemaRepositoryImpl {
	return &DevtronResourceSchemaRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type DevtronResourceSchema struct {
	tableName         struct{} `sql:"devtron_resource_schema" pg:",discard_unknown_columns"`
	Id                int      `sql:"id,pk"`
	DevtronResourceId int      `sql:"devtron_resource_id"`
	Version           string   `sql:"version"`
	Schema            string   `sql:"schema"`
	Latest            bool     `sql:"latest,notnull"`
	sql.AuditLog
	DevtronResource DevtronResource
}

func (impl DevtronResourceSchemaRepositoryImpl) Save(model *DevtronResourceSchema) error {
	return impl.dbConnection.Insert(model)
}

func (impl DevtronResourceSchemaRepositoryImpl) Update(model *DevtronResourceSchema) error {
	return impl.dbConnection.Update(model)
}

func (impl DevtronResourceSchemaRepositoryImpl) FindByResourceId(id int) (*DevtronResourceSchema, error) {
	devtronResourceSchema := &DevtronResourceSchema{}
	err := impl.dbConnection.
		Model(devtronResourceSchema).
		Where("devtron_resource_id =?", id).
		Limit(1).
		Select()
	return devtronResourceSchema, err
}

func (impl DevtronResourceSchemaRepositoryImpl) FindSchemaByKindSubKindAndVersion(kind string, subKind string, version string) (*DevtronResourceSchema, error) {
	devtronResourceSchema := &DevtronResourceSchema{}
	query := `select devtron_resource_schema.* from devtron_resource dr1 
    			left join devtron_resource dr2 on dr1.parent_kind_id = dr2.id 
		 		left join devtron_resource_schema on dr1.id = devtron_resource_schema.devtron_resource_id 
		 		where devtron_resource_schema.version = ? and devtron_resource_schema.latest = ? and `
	if len(subKind) > 0 {
		query += fmt.Sprintf(" dr1.kind = '%s' and", subKind)
		query += fmt.Sprintf(" dr2.kind = '%s';", kind)
	} else {
		query += fmt.Sprintf(" dr1.kind = '%s';", kind)
	}
	_, err := impl.dbConnection.Query(devtronResourceSchema, query, version, true)
	return devtronResourceSchema, err
}

func (repo *DevtronResourceSchemaRepositoryImpl) GetAll() ([]*DevtronResourceSchema, error) {
	var models []*DevtronResourceSchema
	err := repo.dbConnection.Model(&models).
		Column("devtron_resource_schema.*", "DevtronResource").
		Where("latest = ?", true).Select()
	if err != nil {
		repo.logger.Errorw("error in getting all devtron resources schema", "err", err)
		return nil, err
	}
	return models, nil
}
