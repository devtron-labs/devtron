package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/history"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type PipelineStrategyHistoryService interface {
	CreatePipelineStrategyHistory(pipelineStrategy *chartConfig.PipelineStrategy, tx *pg.Tx) (historyModel *history.PipelineStrategyHistory, err error)
}

type PipelineStrategyHistoryServiceImpl struct {
	logger                            *zap.SugaredLogger
	pipelineStrategyHistoryRepository history.PipelineStrategyHistoryRepository
}

func NewPipelineStrategyHistoryServiceImpl(logger *zap.SugaredLogger, pipelineStrategyHistoryRepository history.PipelineStrategyHistoryRepository) *PipelineStrategyHistoryServiceImpl {
	return &PipelineStrategyHistoryServiceImpl{
		logger:                            logger,
		pipelineStrategyHistoryRepository: pipelineStrategyHistoryRepository,
	}
}

func (impl PipelineStrategyHistoryServiceImpl) CreatePipelineStrategyHistory(pipelineStrategy *chartConfig.PipelineStrategy, tx *pg.Tx) (historyModel *history.PipelineStrategyHistory, err error) {
	//creating new entry
	historyModel = &history.PipelineStrategyHistory{
		PipelineId: pipelineStrategy.PipelineId,
		Strategy:   pipelineStrategy.Strategy,
		Config:     pipelineStrategy.Config,
		Default:    pipelineStrategy.Default,
		Deployed:   false,
		AuditLog: sql.AuditLog{
			CreatedOn: pipelineStrategy.CreatedOn,
			CreatedBy: pipelineStrategy.CreatedBy,
			UpdatedOn: pipelineStrategy.UpdatedOn,
			UpdatedBy: pipelineStrategy.UpdatedBy,
		},
	}
	if tx != nil {
		_, err = impl.pipelineStrategyHistoryRepository.CreateHistoryWithTxn(historyModel, tx)
	} else {
		_, err = impl.pipelineStrategyHistoryRepository.CreateHistory(historyModel)
	}
	if err != nil {
		impl.logger.Errorw("err in creating history entry for ci script", "err", err)
		return nil, err
	}
	return historyModel, err
}
