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
	var materials []*types.CiPipelineMaterialModel
	if err != nil {
		impl.logger.Errorw("error occurred while fetching pipeline by id", "pipelineId", pipelineId, "err", err)
		return materials, err
	}
	return adapters.ConvertToPipelineMaterials(materialEntities), nil
}

func (impl *MaterialServiceImpl) GetCheckoutPath(gitMaterialId int) (string, error) {
	return impl.repository.GetCheckoutPath(gitMaterialId)
}
