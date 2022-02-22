package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppStoreChartsHistory struct {
	tableName             struct{}  `sql:"app_store_charts_history" pg:",discard_unknown_columns"`
	Id                    int       `sql:"id,pk"`
	InstalledAppVersionId int       `sql:"installed_app_version_id, notnull"`
	Values                string    `sql:"values_yaml"`
	DeployedOn            time.Time `sql:"deployed_on"`
	DeployedBy            int32     `sql:"deployed_by"`
	sql.AuditLog
}

type AppStoreChartsHistoryRepository interface {
	CreateHistory(chart *AppStoreChartsHistory, tx *pg.Tx) (*AppStoreChartsHistory, error)
}

type AppStoreChartsHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewChartsHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *AppStoreChartsHistoryRepositoryImpl {
	return &AppStoreChartsHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl AppStoreChartsHistoryRepositoryImpl) CreateHistory(history *AppStoreChartsHistory, tx *pg.Tx) (*AppStoreChartsHistory, error) {
	err := tx.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating charts history entry", "err", err, "history", history)
		return history, err
	}
	return history, nil
}
