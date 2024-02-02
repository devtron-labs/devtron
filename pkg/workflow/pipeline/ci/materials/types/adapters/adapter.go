package adapters

import (
	pc "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/workflow/pipeline/ci/materials/types"
)

func ConvertToPipelineMaterials(materialEntities []*pc.CiPipelineMaterialEntity) []*types.CiPipelineMaterialModel {
	var models []*types.CiPipelineMaterialModel
	for _, materialEntity := range materialEntities {
		models = append(models, ConvertToPipelineMaterial(materialEntity))
	}
	return models
}

func ConvertToPipelineMaterial(materialEntity *pc.CiPipelineMaterialEntity) *types.CiPipelineMaterialModel {
	//TODO KB: fix this
	model := &types.CiPipelineMaterialModel{}
	return model
}
