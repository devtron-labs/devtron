package argoAppStatus

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ArgoAppStatusContainer struct {
	TableName      struct{}  `sql:"argo_app_status" pg:",discard_unknown_columns"`
	AppId          int       `sql:"app_id"`
	InstalledAppId int       `sql:"installed_app_id"` //unknown
	EnvId          int       `sql:"env_id"`
	Status         int       `sql:"status"`
	AppStore       bool      `sql:"app_store"` //unknown
	UpdatedOn      time.Time `sql:"status"`
	Active         bool      `sql:"active"`
}

type ArgoAppStatusRepository interface {
	Create(container ArgoAppStatusContainer) error
	Update(container ArgoAppStatusContainer) error
	GetAllDevtronAppStatuses(appIds []int) ([]ArgoAppStatusContainer, error)
	GetAllInstalledAppStatuses(installedAppIds []int) ([]ArgoAppStatusContainer, error)
	Delete(container ArgoAppStatusContainer) error
}

type ArgoAppStatusRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewArgoAppStatusRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ArgoAppStatusRepositoryImpl {
	return &ArgoAppStatusRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (repo *ArgoAppStatusRepositoryImpl) Create(container ArgoAppStatusContainer) error {
	return nil
}

func (repo *ArgoAppStatusRepositoryImpl) Update(container ArgoAppStatusContainer) error {
	return nil
}
func (repo *ArgoAppStatusRepositoryImpl) GetAllDevtronAppStatuses(appIds []int) ([]ArgoAppStatusContainer, error) {
	appStatusContainers := make([]ArgoAppStatusContainer, 0)
	query := "select * " +
		"from argo_app_status where app_id in ?;"
	_, err := repo.dbConnection.Query(&appStatusContainers, query, pg.In(appIds))
	return appStatusContainers, err
}

func (repo *ArgoAppStatusRepositoryImpl) GetAllInstalledAppStatuses(installedAppIds []int) ([]ArgoAppStatusContainer, error) {
	appStatusContainers := make([]ArgoAppStatusContainer, 0)
	query := "SELECT aas.*,ia.id AS installed_app_id " +
		"FROM argo_app_status aas INNER JOIN (SELECT id,app_id WHERE id IN ?) ia ON aas.app_id = ia.app_id;"
	_, err := repo.dbConnection.Query(&appStatusContainers, query, pg.In(installedAppIds))
	return appStatusContainers, err
}
func (repo *ArgoAppStatusRepositoryImpl) Delete(container ArgoAppStatusContainer) error {
	return nil
}
