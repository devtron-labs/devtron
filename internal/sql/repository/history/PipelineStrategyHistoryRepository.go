package history

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStrategyHistoryRepository interface {
	CreateHistory(model *PipelineStrategyHistory) (*PipelineStrategyHistory, error)
	CreateHistoryWithTxn(model *PipelineStrategyHistory, tx *pg.Tx) (*PipelineStrategyHistory, error)
	UpdateHistory(model *PipelineStrategyHistory) (*PipelineStrategyHistory, error)
}

type PipelineStrategyHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPipelineStrategyHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *PipelineStrategyHistoryRepositoryImpl {
	return &PipelineStrategyHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

type PipelineStrategyHistory struct {
	TableName  struct{}                          `sql:"pipeline_strategy_history" pg:",discard_unknown_columns"`
	Id         int                               `sql:"id,pk"`
	PipelineId int                               `sql:"pipeline_id, notnull"`
	Strategy   pipelineConfig.DeploymentTemplate `sql:"strategy,notnull"`
	Config     string                            `sql:"config"`
	Default    bool                              `sql:"default"`
	Deployed   bool                              `sql:"deployed"`
	DeployedOn time.Time                         `sql:"deployed_on"`
	DeployedBy int32                             `sql:"deployed_by"`
	sql.AuditLog
}

func (impl PipelineStrategyHistoryRepositoryImpl) CreateHistory(model *PipelineStrategyHistory) (*PipelineStrategyHistory, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.logger.Errorw("err in creating strategy history entry", "err", err)
		return model, err
	}
	return model, nil
}

func (impl PipelineStrategyHistoryRepositoryImpl) CreateHistoryWithTxn(model *PipelineStrategyHistory, tx *pg.Tx) (*PipelineStrategyHistory, error) {
	err := tx.Insert(model)
	if err != nil {
		impl.logger.Errorw("err in creating strategy history entry", "err", err)
		return model, err
	}
	return model, nil
}

func (impl PipelineStrategyHistoryRepositoryImpl) UpdateHistory(model *PipelineStrategyHistory) (*PipelineStrategyHistory, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.logger.Errorw("err in updating strategy history entry", "err", err)
		return model, err
	}
	return model, nil
}
