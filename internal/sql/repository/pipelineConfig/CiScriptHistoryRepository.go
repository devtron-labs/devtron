package pipelineConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type CiScriptHistory struct {
	tableName           struct{}  `sql:"ci_script_history" pg:",discard_unknown_columns"`
	Id                  int       `sql:"id,pk"`
	CiPipelineScriptsId int       `sql:"ci_pipeline_scripts_id"`
	Script              string    `sql:"script"`
	Stage               string    `sql:"stage"`
	Latest              bool      `sql:"latest"`
	Built               bool      `sql:"built"`
	BuiltOn             time.Time `sql:"built_on"`
	BuiltBy             int32     `sql:"built_by"`
	sql.AuditLog
}

type CiScriptHistoryRepository interface {
	CreateHistory(history *CiScriptHistory) (*CiScriptHistory, error)
	UpdateHistory(history *CiScriptHistory) (*CiScriptHistory, error)
	GetLatestByCiPipelineScriptsId(ciPipelineScriptsId int) (*CiScriptHistory, error)
}

type CiScriptHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCiScriptHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *CiScriptHistoryRepositoryImpl {
	return &CiScriptHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
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
	err := impl.dbConnection.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in updating ci script history entry", "err", err)
		return nil, err
	}
	return history, nil
}

func (impl CiScriptHistoryRepositoryImpl) GetLatestByCiPipelineScriptsId(ciPipelineScriptsId int) (*CiScriptHistory, error) {
	var scriptHistory *CiScriptHistory
	err := impl.dbConnection.Model(&scriptHistory).Where("ci_pipeline_scripts_id = ?", ciPipelineScriptsId).
		Where("latest = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("err in getting latest entry for ci script history", "err", err, "ci_pipeline_scripts_id", ciPipelineScriptsId)
		return nil, err
	}
	return scriptHistory, nil
}
