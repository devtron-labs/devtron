package appStatus

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppStatusContainer struct {
	AppId          int       `json:"app_id"`
	AppName        string    `json:"app_name"`
	EnvIdentifier  string    `json:"env_identifier"`
	InstalledAppId int       `json:"installed_app_id"`
	EnvId          int       `json:"env_id"`
	Status         string    `json:"status"`
	AppType        int       `json:"app_type"`
	UpdatedOn      time.Time `json:"updated_on"`
}

type AppStatusDto struct {
	TableName struct{}  `sql:"app_status" pg:",discard_unknown_columns"`
	AppId     int       `sql:"app_id,pk"`
	EnvId     int       `sql:"env_id,pk"`
	Status    string    `sql:"status"`
	UpdatedOn time.Time `sql:"updated_on"`
}

type AppStatusRepository interface {
	Create(container AppStatusContainer) error
	Update(container AppStatusContainer) error
	Delete(tx *pg.Tx, appId, envId int) error
	DeleteWithEnvId(tx *pg.Tx, envId int) error
	Get(appId, envId int) (AppStatusContainer, error)
	GetConnection() *pg.DB
	GetByEnvId(envId int) ([]*AppStatusDto, error)

	//GetAllDevtronAppStatuses(appIds []int) ([]AppStatusContainer, error)
	//GetAllInstalledAppStatuses(installedAppIds []int) ([]AppStatusContainer, error)
}

type AppStatusRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewAppStatusRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *AppStatusRepositoryImpl {
	return &AppStatusRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (repo *AppStatusRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}
func (repo *AppStatusRepositoryImpl) Create(container AppStatusContainer) error {
	model := AppStatusDto{
		AppId:     container.AppId,
		EnvId:     container.EnvId,
		Status:    container.Status,
		UpdatedOn: time.Now(),
	}
	err := repo.dbConnection.Insert(&model)
	return err
}

func (repo *AppStatusRepositoryImpl) Update(container AppStatusContainer) error {
	model := AppStatusDto{
		AppId:     container.AppId,
		EnvId:     container.EnvId,
		Status:    container.Status,
		UpdatedOn: time.Now(),
	}
	err := repo.dbConnection.Update(&model)
	return err
}

func (repo *AppStatusRepositoryImpl) Delete(tx *pg.Tx, appId, envId int) error {
	model := AppStatusDto{
		AppId: appId,
		EnvId: envId,
	}
	err := tx.Delete(&model)
	return err
}

func (repo *AppStatusRepositoryImpl) DeleteWithEnvId(tx *pg.Tx, envId int) error {
	model := AppStatusDto{
		EnvId: envId,
	}
	//TODO : Change to ORM query
	query := "DELETE FROM app_status WHERE env_id = ?;"
	_, err := tx.Query(&model, query, envId)
	return err
}

func (repo *AppStatusRepositoryImpl) Get(appId, envId int) (AppStatusContainer, error) {
	model := AppStatusDto{}
	err := repo.dbConnection.Model(&model).
		Where("app_id = ?", appId).
		Where("env_id = ?", envId).
		Select()
	container := AppStatusContainer{
		AppId:     model.AppId,
		EnvId:     model.EnvId,
		Status:    model.Status,
		UpdatedOn: model.UpdatedOn,
	}
	return container, err
}

func (repo *AppStatusRepositoryImpl) GetByEnvId(envId int) ([]*AppStatusDto, error) {
	var models []*AppStatusDto
	err := repo.dbConnection.Model(&models).
		Where("env_id = ?", envId).
		Select()
	return models, err
}
