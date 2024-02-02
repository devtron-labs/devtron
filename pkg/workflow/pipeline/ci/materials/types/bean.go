package types

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

type CiPipelineMaterialModel struct {
	//TODO KB: check model bean would work or not ??
	CiMaterialId  int
	Type          pipelineConfig.SourceType
	Value         string
	GitTag        string
	GitMaterialId int
	GitMaterial   *bean.GitMaterialModel
	GitOptions    bean2.GitOptions
	Active        bool
}
