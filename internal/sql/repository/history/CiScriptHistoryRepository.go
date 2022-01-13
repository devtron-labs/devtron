package history

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type CiScriptHistory struct {
	tableName           struct{}  `sql:"ci_script_history" pg:",discard_unknown_columns"`
	Id                  int       `sql:"id,pk"`
	CiPipelineScriptsId int       `sql:"ci_pipeline_scripts_id, notnull"`
	Script              string    `sql:"script"`
	Stage               string    `sql:"stage"`
	Built               bool      `sql:"built"`
	BuiltOn             time.Time `sql:"built_on"`
	BuiltBy             int32     `sql:"built_by"`
	sql.AuditLog
}

type CiScriptHistoryRepository interface {
	CreateHistoryWithTxn(history *CiScriptHistory, tx *pg.Tx) (*CiScriptHistory, error)
	CreateHistory(history *CiScriptHistory) (*CiScriptHistory, error)
	UpdateHistory(history *CiScriptHistory) (*CiScriptHistory, error)
}

type CiScriptHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCiScriptHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *CiScriptHistoryRepositoryImpl {
	return &CiScriptHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl CiScriptHistoryRepositoryImpl) CreateHistoryWithTxn(history *CiScriptHistory, tx *pg.Tx) (*CiScriptHistory, error) {
	err := tx.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating ci script history entry", "err", err)
		return nil, err
	}
	return history, nil
}
func (impl CiScriptHistoryRepositoryImpl) CreateHistory(history *CiScriptHistory) (*CiScriptHistory, error) {
	err := impl.dbConnection.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating ci script history entry", "err", err)
		return nil, err
	}
	return history, nil
}
func (impl CiScriptHistoryRepositoryImpl) UpdateHistory(history *CiScriptHistory) (*CiScriptHistory, error) {
	err := impl.dbConnection.Update(history)
	if err != nil {
		impl.logger.Errorw("err in updating ci script history entry", "err", err)
		return nil, err
	}
	return history, nil
}
