package argoAppStatus

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppType int

const (
	DEVTRON_APP           AppType = 0
	DEVTRON_INSTALLED_APP AppType = 1
)

type ArgoAppStatusContainer struct {
	TableName struct{}  `sql:"argo_app_status" pg:",discard_unknown_columns"`
	AppId     int       `sql:"app_id"`
	EnvId     int       `sql:"env_id"`
	Status    int       `sql:"status"`
	AppType   AppType   `sql:"app_type"`
	UpdatedOn time.Time `sql:"status"`
}

type ArgoAppStatusRepository interface {
	Create(container ArgoAppStatusContainer) error
	Update(container ArgoAppStatusContainer) error
	GetAllWithAppIds(appIds, installedAppIds []int) ([]ArgoAppStatusContainer, error)
	Delete(container ArgoAppStatusContainer) error
}

type ArgoAppStatusRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func (repo *ArgoAppStatusRepositoryImpl) Create(container ArgoAppStatusContainer) error {
	return nil
}

func (repo *ArgoAppStatusRepositoryImpl) Update(container ArgoAppStatusContainer) error {
	return nil
}
func (repo *ArgoAppStatusRepositoryImpl) GetAllWithAppIds(appIds, installedAppIds []int) ([]ArgoAppStatusContainer, error) {
	appStatusContainers := make([]ArgoAppStatusContainer, 0)
	query := "select * " +
		"from argo_app_status where (app_id in ? and type = 0) or (installed_app_id in ? and type = 1)"
	_, err := repo.dbConnection.Query(&appStatusContainers, query, pg.In(appIds), pg.In(installedAppIds))
	return appStatusContainers, err
}
func (repo *ArgoAppStatusRepositoryImpl) Delete(container ArgoAppStatusContainer) error {
	return nil
}
