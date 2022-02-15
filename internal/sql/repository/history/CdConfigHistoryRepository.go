package history

import (
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

type CdConfigHistory struct {
	tableName            struct{}    `sql:"cd_config_history" pg:",discard_unknown_columns"`
	Id                   int         `sql:"id,pk"`
	PipelineId           int         `sql:"pipeline_id, notnull"`
	Config               string      `sql:"config"`
	Stage                CdStageType `sql:"stage"`
	ConfigMapSecretNames string      `sql:"configmap_secret_names"`
	ExecInEnv			 bool 		 `sql:"exec_in_env"`
	Deployed             bool        `sql:"deployed"`
	DeployedOn           time.Time   `sql:"deployed_on"`
	DeployedBy           int32       `sql:"deployed_by"`
	sql.AuditLog
}

type CdConfigHistoryRepository interface {
	CreateHistoryWithTxn(history *CdConfigHistory, tx *pg.Tx) (*CdConfigHistory, error)
	CreateHistory(history *CdConfigHistory) (*CdConfigHistory, error)
	UpdateHistory(history *CdConfigHistory) (*CdConfigHistory, error)
	GetHistoryForDeployedCdConfigByStage(pipelineId int, stage CdStageType) ([]*CdConfigHistory, error)
}

type CdConfigHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCdConfigHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *CdConfigHistoryRepositoryImpl {
	return &CdConfigHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl CdConfigHistoryRepositoryImpl) CreateHistoryWithTxn(history *CdConfigHistory, tx *pg.Tx) (*CdConfigHistory, error) {
	err := tx.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating cd config history entry", "err", err, "history", history)
		return nil, err
	}
	return history, nil
}

func (impl CdConfigHistoryRepositoryImpl) CreateHistory(history *CdConfigHistory) (*CdConfigHistory, error) {
	err := impl.dbConnection.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating cd config history entry", "err", err, "history", history)
		return nil, err
	}
	return history, nil
}

func (impl CdConfigHistoryRepositoryImpl) UpdateHistory(history *CdConfigHistory) (*CdConfigHistory, error) {
	err := impl.dbConnection.Update(history)
	if err != nil {
		impl.logger.Errorw("err in updating cd config history entry", "err", err, "history", history)
		return nil, err
	}
	return history, nil
}

func (impl CdConfigHistoryRepositoryImpl) GetHistoryForDeployedCdConfigByStage(pipelineId int, stage CdStageType) ([]*CdConfigHistory, error) {
	var histories []*CdConfigHistory
	err := impl.dbConnection.Model(&histories).Where("pipeline_id = ?", pipelineId).
		Where("stage = ?", stage).
		Where("deployed = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("err in getting cd config history", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	return histories, nil
}
