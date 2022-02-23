package history

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStrategyHistoryService interface {
	CreatePipelineStrategyHistory(pipelineStrategy *chartConfig.PipelineStrategy, tx *pg.Tx) (historyModel *repository.PipelineStrategyHistory, err error)
	CreateStrategyHistoryForDeploymentTrigger(strategy *chartConfig.PipelineStrategy, deployedOn time.Time, deployedBy int32) error
	GetHistoryForDeployedStrategy(pipelineId int) ([]*PipelineStrategyHistoryDto, error)
}

type PipelineStrategyHistoryServiceImpl struct {
	logger                            *zap.SugaredLogger
	pipelineStrategyHistoryRepository repository.PipelineStrategyHistoryRepository
}

func NewPipelineStrategyHistoryServiceImpl(logger *zap.SugaredLogger, pipelineStrategyHistoryRepository repository.PipelineStrategyHistoryRepository) *PipelineStrategyHistoryServiceImpl {
	return &PipelineStrategyHistoryServiceImpl{
		logger:                            logger,
		pipelineStrategyHistoryRepository: pipelineStrategyHistoryRepository,
	}
}

func (impl PipelineStrategyHistoryServiceImpl) CreatePipelineStrategyHistory(pipelineStrategy *chartConfig.PipelineStrategy, tx *pg.Tx) (historyModel *repository.PipelineStrategyHistory, err error) {
	//creating new entry
	historyModel = &repository.PipelineStrategyHistory{
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
		impl.logger.Errorw("err in creating history entry for pipeline strategy", "err", err)
		return nil, err
	}
	return historyModel, err
}

func (impl PipelineStrategyHistoryServiceImpl) CreateStrategyHistoryForDeploymentTrigger(pipelineStrategy *chartConfig.PipelineStrategy, deployedOn time.Time, deployedBy int32) error {
	//creating new entry
	historyModel := &repository.PipelineStrategyHistory{
		PipelineId: pipelineStrategy.PipelineId,
		Strategy:   pipelineStrategy.Strategy,
		Config:     pipelineStrategy.Config,
		Default:    pipelineStrategy.Default,
		Deployed:   true,
		DeployedBy: deployedBy,
		DeployedOn: deployedOn,
		AuditLog: sql.AuditLog{
			CreatedOn: deployedOn,
			CreatedBy: deployedBy,
			UpdatedOn: deployedOn,
			UpdatedBy: deployedBy,
		},
	}
	_, err := impl.pipelineStrategyHistoryRepository.CreateHistory(historyModel)
	if err != nil {
		impl.logger.Errorw("err in creating history entry for pipeline strategy", "err", err)
		return err
	}
	return err
}

func (impl PipelineStrategyHistoryServiceImpl) GetHistoryForDeployedStrategy(pipelineId int) ([]*PipelineStrategyHistoryDto, error) {
	histories, err := impl.pipelineStrategyHistoryRepository.GetHistoryForDeployedStrategy(pipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting history for strategy", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	var historiesDto []*PipelineStrategyHistoryDto
	for _, history := range histories {
		historyDto := &PipelineStrategyHistoryDto{
			Id:         history.Id,
			PipelineId: history.PipelineId,
			Strategy:   string(history.Strategy),
			Config:     history.Config,
			Default:    history.Default,
			Deployed:   history.Deployed,
			DeployedOn: history.DeployedOn,
			DeployedBy: history.DeployedBy,
			AuditLog: sql.AuditLog{
				CreatedBy: history.CreatedBy,
				CreatedOn: history.CreatedOn,
				UpdatedBy: history.UpdatedBy,
				UpdatedOn: history.UpdatedOn,
			},
		}
		historiesDto = append(historiesDto, historyDto)
	}
	return historiesDto, nil
}
