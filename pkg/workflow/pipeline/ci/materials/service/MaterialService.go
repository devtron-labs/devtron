package service

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/workflow/pipeline/ci/materials/types"
	"github.com/devtron-labs/devtron/pkg/workflow/pipeline/ci/materials/types/adapters"
	"go.uber.org/zap"
)

type MaterialService interface {
	GetByPipelineId(id int) ([]*types.CiPipelineMaterialModel, error)
	GetCheckoutPath(gitMaterialId int) (string, error)
}

type MaterialServiceImpl struct {
	logger     *zap.SugaredLogger
	repository pipelineConfig.CiPipelineMaterialRepository
}

func NewMaterialServiceImpl(logger *zap.SugaredLogger, repository pipelineConfig.CiPipelineMaterialRepository) *MaterialServiceImpl {
	return &MaterialServiceImpl{logger: logger, repository: repository}
}

func (impl *MaterialServiceImpl) GetByPipelineId(pipelineId int) ([]*types.CiPipelineMaterialModel, error) {
	materialEntities, err := impl.repository.GetByPipelineId(pipelineId)
	if err != nil {
		// TODO KB: log here
	}
	return adapters.ConvertToPipelineMaterials(materialEntities), nil
}

func (impl *MaterialServiceImpl) GetCheckoutPath(gitMaterialId int) (string, error) {
	return impl.repository.GetCheckoutPath(gitMaterialId)
}
