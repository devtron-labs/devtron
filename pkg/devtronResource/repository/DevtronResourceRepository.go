package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DevtronResourceRepository interface {
	Save(model *DevtronResource) error
	Update(model *DevtronResource) error
	FindByKind(kind string) (*DevtronResource, error)
}

type DevtronResourceRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewDevtronResourceRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *DevtronResourceRepositoryImpl {
	return &DevtronResourceRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type DevtronResource struct {
	tableName    struct{} `sql:"devtron_resource" pg:",discard_unknown_columns"`
	Id           int      `sql:"id,pk"`
	Kind         string   `sql:"kind"`
	DisplayName  string   `sql:"displayName"`
	Icon         string   `sql:"icon"`
	ParentKindId int      `sql:"parent_kind_id"`
	Deleted      bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

func (impl DevtronResourceRepositoryImpl) Save(model *DevtronResource) error {
	return impl.dbConnection.Insert(model)
}

func (impl DevtronResourceRepositoryImpl) Update(model *DevtronResource) error {
	return impl.dbConnection.Update(model)
}

func (impl DevtronResourceRepositoryImpl) FindByKind(kind string) (*DevtronResource, error) {
	devtronResource := &DevtronResource{}
	err := impl.dbConnection.
		Model(devtronResource).
		Where("kind =?", kind).
		Limit(1).
		Select()
	return devtronResource, err
}
