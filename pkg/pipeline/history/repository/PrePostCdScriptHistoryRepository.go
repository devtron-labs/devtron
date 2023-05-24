package repository

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type CdStageType string

const (
	PRE_CD_TYPE  CdStageType = "PRE_CD"
	POST_CD_TYPE CdStageType = "POST_CD"
)

type PrePostCdScriptHistory struct {
	tableName            struct{}                   `sql:"pre_post_cd_script_history" pg:",discard_unknown_columns"`
	Id                   int                        `sql:"id,pk"`
	PipelineId           int                        `sql:"pipeline_id, notnull"`
	Script               string                     `sql:"script"`
	Stage                CdStageType                `sql:"stage"`
	TriggerType          pipelineConfig.TriggerType `sql:"trigger_type"`
	ConfigMapSecretNames string                     `sql:"configmap_secret_names"`
	ConfigMapData        string                     `sql:"configmap_data"`
	SecretData           string                     `sql:"secret_data"`
	ExecInEnv            bool                       `sql:"exec_in_env,notnull"`
	Deployed             bool                       `sql:"deployed"`
	DeployedOn           time.Time                  `sql:"deployed_on"`
	DeployedBy           int32                      `sql:"deployed_by"`
	sql.AuditLog
}

type PrePostCdScriptHistoryRepository interface {
	CreateHistoryWithTxn(history *PrePostCdScriptHistory, tx *pg.Tx) (*PrePostCdScriptHistory, error)
	CreateHistory(history *PrePostCdScriptHistory) (*PrePostCdScriptHistory, error)
	GetHistoryForDeployedPrePostScriptByStage(pipelineId int, stage CdStageType) ([]*PrePostCdScriptHistory, error)
}

type PrePostCdScriptHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPrePostCdScriptHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *PrePostCdScriptHistoryRepositoryImpl {
	return &PrePostCdScriptHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl PrePostCdScriptHistoryRepositoryImpl) CreateHistoryWithTxn(history *PrePostCdScriptHistory, tx *pg.Tx) (*PrePostCdScriptHistory, error) {
	err := tx.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating cd script history entry", "err", err, "history", history)
		return nil, err
	}
	return history, nil
}

func (impl PrePostCdScriptHistoryRepositoryImpl) CreateHistory(history *PrePostCdScriptHistory) (*PrePostCdScriptHistory, error) {
	err := impl.dbConnection.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating cd script history entry", "err", err, "history", history)
		return nil, err
	}
	return history, nil
}

func (impl PrePostCdScriptHistoryRepositoryImpl) GetHistoryForDeployedPrePostScriptByStage(pipelineId int, stage CdStageType) ([]*PrePostCdScriptHistory, error) {
	var histories []*PrePostCdScriptHistory
	err := impl.dbConnection.Model(&histories).Where("pipeline_id = ?", pipelineId).
		Where("stage = ?", stage).
		Where("deployed = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("err in getting cd script history", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	return histories, nil
}
