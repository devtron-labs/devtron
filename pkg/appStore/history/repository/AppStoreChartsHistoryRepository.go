package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppStoreChartsHistory struct {
	tableName       struct{}  `sql:"app_store_charts_history" pg:",discard_unknown_columns"`
	Id              int       `sql:"id,pk"`
	InstalledAppsId int       `sql:"installed_apps_id, notnull"`
	Values          string    `sql:"values_yaml"`
	DeployedOn      time.Time `sql:"deployed_on"`
	DeployedBy      int32     `sql:"deployed_by"`
	sql.AuditLog
}

type AppStoreChartsHistoryRepository interface {
	CreateHistoryWithTxn(chart *AppStoreChartsHistory, tx *pg.Tx) (*AppStoreChartsHistory, error)
	CreateHistory(chart *AppStoreChartsHistory) (*AppStoreChartsHistory, error)
}

type AppStoreChartsHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewAppStoreChartsHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *AppStoreChartsHistoryRepositoryImpl {
	return &AppStoreChartsHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl AppStoreChartsHistoryRepositoryImpl) CreateHistoryWithTxn(history *AppStoreChartsHistory, tx *pg.Tx) (*AppStoreChartsHistory, error) {
	err := tx.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating app store charts history entry", "err", err, "history", history)
		return history, err
	}
	return history, nil
}

func (impl AppStoreChartsHistoryRepositoryImpl) CreateHistory(history *AppStoreChartsHistory) (*AppStoreChartsHistory, error) {
	err := impl.dbConnection.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating app store charts history entry", "err", err, "history", history)
		return history, err
	}
	return history, nil
}
