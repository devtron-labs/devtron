package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PrePostCiScriptHistory struct {
	tableName           struct{}  `sql:"pre_post_ci_script_history" pg:",discard_unknown_columns"`
	Id                  int       `sql:"id,pk"`
	CiPipelineScriptsId int       `sql:"ci_pipeline_scripts_id, notnull"`
	Script              string    `sql:"script"`
	Stage               string    `sql:"stage"`
	Name                string    `sql:"name"`
	OutputLocation      string    `sql:"output_location"`
	Built               bool      `sql:"built"`
	BuiltOn             time.Time `sql:"built_on"`
	BuiltBy             int32     `sql:"built_by"`
	sql.AuditLog
}

type PrePostCiScriptHistoryRepository interface {
	CreateHistoryWithTxn(history *PrePostCiScriptHistory, tx *pg.Tx) (*PrePostCiScriptHistory, error)
	CreateHistory(history *PrePostCiScriptHistory) (*PrePostCiScriptHistory, error)
}

type PrePostCiScriptHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPrePostCiScriptHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *PrePostCiScriptHistoryRepositoryImpl {
	return &PrePostCiScriptHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl PrePostCiScriptHistoryRepositoryImpl) CreateHistoryWithTxn(history *PrePostCiScriptHistory, tx *pg.Tx) (*PrePostCiScriptHistory, error) {
	err := tx.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating ci script history entry", "err", err)
		return nil, err
	}
	return history, nil
}
func (impl PrePostCiScriptHistoryRepositoryImpl) CreateHistory(history *PrePostCiScriptHistory) (*PrePostCiScriptHistory, error) {
	err := impl.dbConnection.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating ci script history entry", "err", err)
		return nil, err
	}
	return history, nil
}
