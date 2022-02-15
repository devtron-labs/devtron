package history

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/history"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStrategyHistoryService interface {
	CreatePipelineStrategyHistory(pipelineStrategy *chartConfig.PipelineStrategy, tx *pg.Tx) (historyModel *history.PipelineStrategyHistory, err error)
	CreateStrategyHistoryForDeploymentTrigger(strategy *chartConfig.PipelineStrategy, deployedOn time.Time, deployedBy int32) error
	GetHistoryForDeployedStrategy(pipelineId int) ([]*history.PipelineStrategyHistory, error)
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

func (impl PipelineStrategyHistoryServiceImpl) CreateStrategyHistoryForDeploymentTrigger(pipelineStrategy *chartConfig.PipelineStrategy, deployedOn time.Time, deployedBy int32) error {
	//creating new entry
	historyModel := &history.PipelineStrategyHistory{
		PipelineId: pipelineStrategy.PipelineId,
		Strategy:   pipelineStrategy.Strategy,
		Config:     pipelineStrategy.Config,
		Default:    pipelineStrategy.Default,
		Deployed:   true,
		DeployedBy: deployedBy,
		DeployedOn: deployedOn,
		AuditLog: sql.AuditLog{
			CreatedOn: pipelineStrategy.CreatedOn,
			CreatedBy: pipelineStrategy.CreatedBy,
			UpdatedOn: pipelineStrategy.UpdatedOn,
			UpdatedBy: pipelineStrategy.UpdatedBy,
		},
	}
	_, err := impl.pipelineStrategyHistoryRepository.CreateHistory(historyModel)
	if err != nil {
		impl.logger.Errorw("err in creating history entry for ci script", "err", err)
		return err
	}
	return err
}

func (impl PipelineStrategyHistoryServiceImpl) GetHistoryForDeployedStrategy(pipelineId int) ([]*history.PipelineStrategyHistory, error) {
	histories, err := impl.GetHistoryForDeployedStrategy(pipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting history for strategy", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	return histories, nil
}
