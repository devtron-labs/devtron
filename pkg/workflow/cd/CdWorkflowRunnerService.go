package cd

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/adapter"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/bean"
	"go.uber.org/zap"
)

type CdWorkflowRunnerService interface {
	FindWorkflowRunnerById(wfrId int) (*bean.CdWorkflowRunnerDto, error)
}

type CdWorkflowRunnerServiceImpl struct {
	logger               *zap.SugaredLogger
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository
}

func NewCdWorkflowRunnerServiceImpl(logger *zap.SugaredLogger,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository) *CdWorkflowRunnerServiceImpl {
	return &CdWorkflowRunnerServiceImpl{
		logger:               logger,
		cdWorkflowRepository: cdWorkflowRepository,
	}
}

func (impl *CdWorkflowRunnerServiceImpl) FindWorkflowRunnerById(wfrId int) (*bean.CdWorkflowRunnerDto, error) {
	cdWfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(wfrId)
	if err != nil {
		impl.logger.Errorw("error in getting cd workflow runner by id", "err", err, "id", wfrId)
		return nil, err
	}
	return adapter.ConvertCdWorkflowRunnerDbObjToDto(cdWfr), nil

}
