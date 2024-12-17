package read

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"go.uber.org/zap"
)

type CdWorkflowReadService interface {
	CheckIfLatestWf(pipelineId, cdWfId int) (latest bool, err error)
}

type CdWorkflowReadServiceImpl struct {
	logger               *zap.SugaredLogger
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository
}

func NewCdWorkflowReadServiceImpl(logger *zap.SugaredLogger,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository) *CdWorkflowReadServiceImpl {
	return &CdWorkflowReadServiceImpl{
		logger:               logger,
		cdWorkflowRepository: cdWorkflowRepository,
	}
}

func (impl *CdWorkflowReadServiceImpl) CheckIfLatestWf(pipelineId, cdWfId int) (latest bool, err error) {
	latest, err = impl.cdWorkflowRepository.IsLatestWf(pipelineId, cdWfId)
	if err != nil {
		impl.logger.Errorw("error in checking if wf is latest", "pipelineId", pipelineId, "cdWfId", cdWfId, "err", err)
		return false, err
	}
	return latest, nil
}
