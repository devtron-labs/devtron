package argoAppStatus

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppStatusContainer struct {
	TableName      struct{}  `sql:"argo_app_status" pg:",discard_unknown_columns"`
	AppId          int       `sql:"app_id"`
	AppName        string    `sql:"app_name"`
	EnvName        string    `sql:"env_name"`
	InstalledAppId int       `sql:"installed_app_id"` //unknown
	EnvId          int       `sql:"env_id"`
	Status         int       `sql:"status"`
	AppStore       bool      `sql:"app_store"` //unknown
	UpdatedOn      time.Time `sql:"status"`
	Active         bool      `sql:"active"`
}

type AppStatusRepository interface {
	Create(container AppStatusContainer) error
	Update(container AppStatusContainer) error
	GetAllDevtronAppStatuses(appIds []int) ([]AppStatusContainer, error)
	GetAllInstalledAppStatuses(installedAppIds []int) ([]AppStatusContainer, error)
	Delete(container AppStatusContainer) error
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

func (repo *AppStatusRepositoryImpl) Create(container AppStatusContainer) error {
	return nil
}

func (repo *AppStatusRepositoryImpl) Update(container AppStatusContainer) error {
	return nil
}
func (repo *AppStatusRepositoryImpl) GetAllDevtronAppStatuses(appIds []int) ([]AppStatusContainer, error) {
	appStatusContainers := make([]AppStatusContainer, 0)
	query := "SELECT aas.*,app.app_name,env.environment_name as env_name ( SELECT * " +
		"FROM argo_app_status WHERE app_id IN ? AND argo_app_status.active = true) aas " +
		"INNER JOIN app ON app.id = aas.app_id AND app.active=true " +
		"INNER JOIN environment env ON environment.id = aas.env_id AND env.active=true;"
	_, err := repo.dbConnection.Query(&appStatusContainers, query, pg.In(appIds))
	return appStatusContainers, err
}

func (repo *AppStatusRepositoryImpl) GetAllInstalledAppStatuses(installedAppIds []int) ([]AppStatusContainer, error) {
	appStatusContainers := make([]AppStatusContainer, 0)
	query := "SELECT aas.*,ia.id AS installed_app_id,app.app_name,env.environment_name " +
		"FROM argo_app_status aas INNER JOIN ( SELECT id,app_id FROM installed_apps WHERE id IN ? AND active = true ) ia " +
		"ON aas.app_id = ia.app_id AND aas.active=true " +
		"INNER JOIN app ON app.id = aas.app_id AND app.active=true " +
		"INNER JOIN environment env ON environment.id = aas.env_id AND env.active=true;"
	_, err := repo.dbConnection.Query(&appStatusContainers, query, pg.In(installedAppIds))
	return appStatusContainers, err
}
func (repo *AppStatusRepositoryImpl) Delete(container AppStatusContainer) error {
	return nil
}
