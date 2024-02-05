package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type DevtronResourceRepository interface {
	Save(model *DevtronResource) error
	Update(model *DevtronResource) error
	UpdateNameAndDescription(tx *pg.Tx, model *DevtronResource, userId int) error
	FindByKind(kind string) (*DevtronResource, error)
	GetAll() ([]*DevtronResource, error)
	GetById(id int) (*DevtronResource, error)
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
	DisplayName  string   `sql:"display_name"`
	Description  string   `sql:"description"`
	IsExposed    bool     `sql:"is_exposed,notnull"`
	Icon         string   `sql:"icon"`
	ParentKindId int      `sql:"parent_kind_id"`
	Deleted      bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

func (repo *DevtronResourceRepositoryImpl) Save(model *DevtronResource) error {
	return repo.dbConnection.Insert(model)
}

func (repo *DevtronResourceRepositoryImpl) Update(model *DevtronResource) error {
	return repo.dbConnection.Update(model)
}

func (repo *DevtronResourceRepositoryImpl) UpdateNameAndDescription(tx *pg.Tx, model *DevtronResource, userId int) error {
	_, err := tx.Model(model).
		Set("display_name = ?", model.DisplayName).
		Set("description = ?", model.Description).
		Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", userId).
		Where("id = ?", model.Id).
		Update()
	return err
}

func (repo *DevtronResourceRepositoryImpl) FindByKind(kind string) (*DevtronResource, error) {
	devtronResource := &DevtronResource{}
	err := repo.dbConnection.
		Model(devtronResource).
		Where("kind =?", kind).
		Select()
	return devtronResource, err
}

func (repo *DevtronResourceRepositoryImpl) GetAll() ([]*DevtronResource, error) {
	var models []*DevtronResource
	err := repo.dbConnection.Model(&models).Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting all devtron resources", "err", err)
		return nil, err
	}
	return models, nil
}

func (repo *DevtronResourceRepositoryImpl) GetById(id int) (*DevtronResource, error) {
	model := &DevtronResource{}
	err := repo.dbConnection.Model(model).
		Where("id = ?", id).
		Where("deleted = ?", false).
		Select()
	if err != nil {
		repo.logger.Errorw("error in getting devtron resource", "err", err, "id", id)
		return nil, err
	}
	return model, nil
}
