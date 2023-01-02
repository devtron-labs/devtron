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
	AppStore       bool      `json:"app_store"`
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
	Create(tx *pg.Tx, container AppStatusContainer) error
	Update(tx *pg.Tx, container AppStatusContainer) error
	//GetAllDevtronAppStatuses(appIds []int) ([]AppStatusContainer, error)
	//GetAllInstalledAppStatuses(installedAppIds []int) ([]AppStatusContainer, error)
	Delete(tx *pg.Tx, appId, envId int) error
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
	model := AppStatusDto{
		AppId:     container.AppId,
		EnvId:     container.EnvId,
		Status:    container.Status,
		UpdatedOn: time.Now(),
	}
	err := tx.Insert(&model)
	return err
}

func (repo *AppStatusRepositoryImpl) Update(tx *pg.Tx, container AppStatusContainer) error {
	model := AppStatusDto{
		AppId:     container.AppId,
		EnvId:     container.EnvId,
		Status:    container.Status,
		UpdatedOn: time.Now(),
	}
	err := tx.Update(&model)
	return err
}

//func (repo *AppStatusRepositoryImpl) GetAllDevtronAppStatuses(appIds []int) ([]AppStatusContainer, error) {
//	appStatusContainers := make([]AppStatusContainer, 0)
//	query := "SELECT ps.*,app.app_name,env.environment_identifier as env_identifier ( SELECT * " +
//		"FROM app_status WHERE app_id IN ? ) ps " +
//		"INNER JOIN app ON app.id = ps.app_id AND app.active=true " +
//		"INNER JOIN environment env ON environment.id = ps.env_id AND env.active=true;"
//	_, err := repo.dbConnection.Query(&appStatusContainers, query, pg.In(appIds))
//	return appStatusContainers, err
//}
//
//func (repo *AppStatusRepositoryImpl) GetAllInstalledAppStatuses(installedAppIds []int) ([]AppStatusContainer, error) {
//	appStatusContainers := make([]AppStatusContainer, 0)
//	query := "SELECT ps.*,ia.id AS installed_app_id,app.app_name,env.environment_name " +
//		"FROM app_status ps " +
//		"INNER JOIN ( SELECT id,app_id FROM installed_apps WHERE id IN ? AND active = true ) ia " +
//		"ON ps.app_id = ia.app_id " +
//		"INNER JOIN app ON app.id = ps.app_id AND app.active=true " +
//		"INNER JOIN environment env ON environment.id = aas.env_id AND env.active=true;"
//	_, err := repo.dbConnection.Query(&appStatusContainers, query, pg.In(installedAppIds))
//	return appStatusContainers, err
//}
func (repo *AppStatusRepositoryImpl) Delete(tx *pg.Tx, appId, envId int) error {
	model := AppStatusDto{
		AppId: appId,
		EnvId: envId,
	}
	query := "DELETE FROM app_status WHERE app_id = ? and env_id = ?;"
	_, err := tx.Query(&model, query, appId, envId)
	return err
}

func (repo *AppStatusRepositoryImpl) DeleteWithAppId(tx *pg.Tx, appId int) error {
	model := AppStatusDto{
		AppId: appId,
	}
	query := "DELETE FROM app_status WHERE app_id = ?;"
	_, err := tx.Query(&model, query, appId)
	return err
}

func (repo *AppStatusRepositoryImpl) DeleteWithEnvId(tx *pg.Tx, envId int) error {
	model := AppStatusDto{
		EnvId: envId,
	}
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
