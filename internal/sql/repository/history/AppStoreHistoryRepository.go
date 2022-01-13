package history

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type InstalledAppHistory struct {
	tableName             struct{}  `sql:"installed_app_history" pg:",discard_unknown_columns"`
	Id                    int       `sql:"id,pk"`
	InstalledAppVersionId int       `sql:"installed_app_version_id, notnull"`
	Values                string    `sql:"values_yaml"`
	DeployedOn            time.Time `sql:"deployed_on"`
	DeployedBy            int32     `sql:"deployed_by"`
	sql.AuditLog
}

type InstalledAppHistoryRepository interface {
	CreateHistory(chart *InstalledAppHistory, tx *pg.Tx) (*InstalledAppHistory, error)
	UpdateHistory(chart *InstalledAppHistory) (*InstalledAppHistory, error)
}

type InstalledAppHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewInstalledAppHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *InstalledAppHistoryRepositoryImpl {
	return &InstalledAppHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl InstalledAppHistoryRepositoryImpl) CreateHistory(history *InstalledAppHistory, tx *pg.Tx) (*InstalledAppHistory, error) {
	err := tx.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating installed app history entry", "err", err, "history", history)
		return history, err
	}
	return history, nil
}

func (impl InstalledAppHistoryRepositoryImpl) UpdateHistory(history *InstalledAppHistory) (*InstalledAppHistory, error) {
	err := impl.dbConnection.Update(history)
	if err != nil {
		impl.logger.Errorw("err in updating installed app history entry", "err", err, "history", history)
		return history, err
	}
	return history, nil
}
