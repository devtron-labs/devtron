package appStatus

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppStatusContainer struct {
	TableName      struct{}  `sql:"pipeline_status" pg:",discard_unknown_columns"`
	AppId          int       `sql:"app_id"`
	AppName        string    `sql:"app_name"`
	EnvIdentifier  string    `sql:"env_identifier"`   //unknown
	InstalledAppId int       `sql:"installed_app_id"` //unknown
	EnvId          int       `sql:"env_id"`
	Status         string    `sql:"status"`
	AppStore       bool      `sql:"app_store"` //unknown
	UpdatedOn      time.Time `sql:"status"`
	Active         bool      `sql:"active"`
}

type AppStatusRepository interface {
	Create(tx *pg.Tx, container AppStatusContainer) error
	Update(tx *pg.Tx, container AppStatusContainer) error
	GetAllDevtronAppStatuses(appIds []int) ([]AppStatusContainer, error)
	GetAllInstalledAppStatuses(installedAppIds []int) ([]AppStatusContainer, error)
	Delete(tx *pg.Tx, container AppStatusContainer) error
	Get(appId, envId int) (AppStatusContainer, error)
	GetConnection() *pg.DB
}

type AppStatusRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewArgoAppStatusRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *AppStatusRepositoryImpl {
	return &AppStatusRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (repo *AppStatusRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}
func (repo *AppStatusRepositoryImpl) Create(tx *pg.Tx, container AppStatusContainer) error {
	err := tx.Insert(container)
	return err
}

func (repo *AppStatusRepositoryImpl) Update(tx *pg.Tx, container AppStatusContainer) error {
	err := tx.Update(container)
	return err
}
func (repo *AppStatusRepositoryImpl) GetAllDevtronAppStatuses(appIds []int) ([]AppStatusContainer, error) {
	appStatusContainers := make([]AppStatusContainer, 0)
	query := "SELECT ps.*,app.app_name,env.environment_identifier as env_identifier ( SELECT * " +
		"FROM pipeline_status WHERE app_id IN ? AND argo_app_status.active = true) ps " +
		"INNER JOIN app ON app.id = ps.app_id AND app.active=true " +
		"INNER JOIN environment env ON environment.id = ps.env_id AND env.active=true;"
	_, err := repo.dbConnection.Query(&appStatusContainers, query, pg.In(appIds))
	return appStatusContainers, err
}

func (repo *AppStatusRepositoryImpl) GetAllInstalledAppStatuses(installedAppIds []int) ([]AppStatusContainer, error) {
	appStatusContainers := make([]AppStatusContainer, 0)
	query := "SELECT ps.*,ia.id AS installed_app_id,app.app_name,env.environment_name " +
		"FROM pipeline_status ps " +
		"INNER JOIN ( SELECT id,app_id FROM installed_apps WHERE id IN ? AND active = true ) ia " +
		"ON ps.app_id = ia.app_id AND ps.active=true " +
		"INNER JOIN app ON app.id = ps.app_id AND app.active=true " +
		"INNER JOIN environment env ON environment.id = aas.env_id AND env.active=true;"
	_, err := repo.dbConnection.Query(&appStatusContainers, query, pg.In(installedAppIds))
	return appStatusContainers, err
}
func (repo *AppStatusRepositoryImpl) Delete(tx *pg.Tx, container AppStatusContainer) error {
	container.Active = false
	err := repo.Update(tx, container)
	return err
}

func (repo *AppStatusRepositoryImpl) Get(appId, envId int) (AppStatusContainer, error) {
	appStatusContainer := AppStatusContainer{}
	err := repo.dbConnection.Model(&appStatusContainer).
		Where("app_id = ?", appId).
		Where("env_id = ?", envId).
		Select()
	return appStatusContainer, err
}
