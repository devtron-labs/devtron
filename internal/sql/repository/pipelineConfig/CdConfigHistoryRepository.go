package pipelineConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
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
	PipelineId           int         `sql:"pipeline_id"`
	Config               string      `sql:"config"`
	Stage                CdStageType `sql:"stage"`
	ConfigMapSecretNames string      `sql:"configmap_secret_names"`
	Latest               bool        `sql:"latest"`
	Deployed             bool        `sql:"deployed"`
	DeployedOn           time.Time   `sql:"deployed_on"`
	DeployedBy           int32       `sql:"deployed_by"`
	models.AuditLog
}

type CdConfigHistoryRepository interface {
	CreateHistory(history *CdConfigHistory) (*CdConfigHistory, error)
	UpdateHistory(history *CdConfigHistory) (*CdConfigHistory, error)
	GetLatestByStageTypeAndPipelineId(stage CdStageType, pipelineId int) (*CdConfigHistory, error)
}

type CdConfigHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCdConfigHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *CdConfigHistoryRepositoryImpl {
	return &CdConfigHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl CdConfigHistoryRepositoryImpl) CreateHistory(history *CdConfigHistory) (*CdConfigHistory, error) {
	err := impl.dbConnection.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating cd config history entry", "err", err)
		return nil, err
	}
	return history, nil
}

func (impl CdConfigHistoryRepositoryImpl) UpdateHistory(history *CdConfigHistory) (*CdConfigHistory, error) {
	err := impl.dbConnection.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in updating cd config history entry", "err", err)
		return nil, err
	}
	return history, nil
}

func (impl CdConfigHistoryRepositoryImpl) GetLatestByStageTypeAndPipelineId(stage CdStageType, pipelineId int) (*CdConfigHistory, error) {
	var scriptHistory *CdConfigHistory
	err := impl.dbConnection.Model(&scriptHistory).Where("pipeline_id = ?", pipelineId).
		Where("stage = ?", stage).Where("latest = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("err in getting latest entry for cd config history", "err", err, "pipeline_id", pipelineId, "stage", stage)
		return nil, err
	}
	return scriptHistory, nil
}
