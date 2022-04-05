package history

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStrategyHistoryService interface {
	CreatePipelineStrategyHistory(pipelineStrategy *chartConfig.PipelineStrategy, tx *pg.Tx) (historyModel *repository.PipelineStrategyHistory, err error)
	CreateStrategyHistoryForDeploymentTrigger(strategy *chartConfig.PipelineStrategy, deployedOn time.Time, deployedBy int32) error
	GetHistoryForDeployedStrategyById(id, pipelineId int) (*PipelineStrategyHistoryDto, error)
	GetDeploymentDetailsForDeployedStrategyHistory(pipelineId int) ([]*PipelineStrategyHistoryDto, error)
}

type PipelineStrategyHistoryServiceImpl struct {
	logger                            *zap.SugaredLogger
	pipelineStrategyHistoryRepository repository.PipelineStrategyHistoryRepository
	userService                       user.UserService
}

func NewPipelineStrategyHistoryServiceImpl(logger *zap.SugaredLogger,
	pipelineStrategyHistoryRepository repository.PipelineStrategyHistoryRepository,
	userService user.UserService) *PipelineStrategyHistoryServiceImpl {
	return &PipelineStrategyHistoryServiceImpl{
		logger:                            logger,
		pipelineStrategyHistoryRepository: pipelineStrategyHistoryRepository,
		userService:                       userService,
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

func (impl PipelineStrategyHistoryServiceImpl) GetHistoryForDeployedStrategyById(id, pipelineId int) (*PipelineStrategyHistoryDto, error) {
	history, err := impl.pipelineStrategyHistoryRepository.GetHistoryForDeployedStrategyById(id, pipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting history for strategy", "err", err, "id", id, "pipelineId", pipelineId)
		return nil, err
	}
	user, err := impl.userService.GetById(history.DeployedBy)
	if err != nil {
		impl.logger.Errorw("unable to find user by id", "err", err, "id", history.Id)
		return nil, err
	}
	historyDto := &PipelineStrategyHistoryDto{
		Id:         history.Id,
		PipelineId: history.PipelineId,
		Strategy:   string(history.Strategy),
		Config:     history.Config,
		Default:    history.Default,
		Deployed:   history.Deployed,
		DeployedOn: history.DeployedOn,
		DeployedBy: history.DeployedBy,
		EmailId:    user.EmailId,
	}
	return historyDto, nil
}

func (impl PipelineStrategyHistoryServiceImpl) GetDeploymentDetailsForDeployedStrategyHistory(pipelineId int) ([]*PipelineStrategyHistoryDto, error) {
	histories, err := impl.pipelineStrategyHistoryRepository.GetDeploymentDetailsForDeployedStrategyHistory(pipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting history for strategy", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	var historiesDto []*PipelineStrategyHistoryDto
	for _, history := range histories {
		user, err := impl.userService.GetById(history.DeployedBy)
		if err != nil {
			impl.logger.Errorw("unable to find user by id", "err", err, "id", history.Id)
			return nil, err
		}
		historyDto := &PipelineStrategyHistoryDto{
			Id:         history.Id,
			PipelineId: history.PipelineId,
			Deployed:   history.Deployed,
			DeployedOn: history.DeployedOn,
			DeployedBy: history.DeployedBy,
			EmailId:    user.EmailId,
		}
		historiesDto = append(historiesDto, historyDto)
	}
	return historiesDto, nil
}
