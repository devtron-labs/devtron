package appStatus

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppStatusContainer struct {
	TableName      struct{}  `sql:"app_status" pg:",discard_unknown_columns"`
	AppId          int       `sql:"app_id"`
	AppName        string    `sql:"app_name"`         //unknown
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
	DeleteWithAppId(tx *pg.Tx, appId int) error
	DeleteWithEnvId(tx *pg.Tx, envId int) error
	Get(appId, envId int) (AppStatusContainer, error)
	GetConnection() *pg.DB
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
func (repo *AppStatusRepositoryImpl) Create(tx *pg.Tx, container AppStatusContainer) error {
	container.UpdatedOn = time.Now()
	err := tx.Insert(container)
	return err
}

func (repo *AppStatusRepositoryImpl) Update(tx *pg.Tx, container AppStatusContainer) error {
	container.UpdatedOn = time.Now()
	err := tx.Update(container)
	return err
}
func (repo *AppStatusRepositoryImpl) GetAllDevtronAppStatuses(appIds []int) ([]AppStatusContainer, error) {
	appStatusContainers := make([]AppStatusContainer, 0)
	query := "SELECT ps.*,app.app_name,env.environment_identifier as env_identifier ( SELECT * " +
		"FROM app_status WHERE app_id IN ? ) ps " +
		"INNER JOIN app ON app.id = ps.app_id AND app.active=true " +
		"INNER JOIN environment env ON environment.id = ps.env_id AND env.active=true;"
	_, err := repo.dbConnection.Query(&appStatusContainers, query, pg.In(appIds))
	return appStatusContainers, err
}

func (repo *AppStatusRepositoryImpl) GetAllInstalledAppStatuses(installedAppIds []int) ([]AppStatusContainer, error) {
	appStatusContainers := make([]AppStatusContainer, 0)
	query := "SELECT ps.*,ia.id AS installed_app_id,app.app_name,env.environment_name " +
		"FROM app_status ps " +
		"INNER JOIN ( SELECT id,app_id FROM installed_apps WHERE id IN ? AND active = true ) ia " +
		"ON ps.app_id = ia.app_id " +
		"INNER JOIN app ON app.id = ps.app_id AND app.active=true " +
		"INNER JOIN environment env ON environment.id = aas.env_id AND env.active=true;"
	_, err := repo.dbConnection.Query(&appStatusContainers, query, pg.In(installedAppIds))
	return appStatusContainers, err
}
func (repo *AppStatusRepositoryImpl) Delete(tx *pg.Tx, container AppStatusContainer) error {
	err := tx.Delete(container)
	return err
}

func (repo *AppStatusRepositoryImpl) DeleteWithAppId(tx *pg.Tx, appId int) error {
	container := AppStatusContainer{}
	query := "DELETE FROM app_status WHERE app_id = ?;"
	_, err := tx.Query(&container, query, appId)
	return err
}

func (repo *AppStatusRepositoryImpl) DeleteWithEnvId(tx *pg.Tx, envId int) error {
	container := AppStatusContainer{}
	query := "DELETE FROM app_status WHERE env_id = ?;"
	_, err := tx.Query(&container, query, envId)
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
