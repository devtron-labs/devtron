package adapter

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/build/pipeline/read/bean"
)

func NewCiPipelineMin(ciPipeline *pipelineConfig.CiPipeline) (*bean.CiPipelineMin, error) {
	if ciPipeline == nil {
		return nil, errors.New("ci pipeline not found")
	}
	dto := &bean.CiPipelineMin{
		Id:               ciPipeline.Id,
		Name:             ciPipeline.Name,
		AppId:            ciPipeline.AppId,
		ParentCiPipeline: ciPipeline.ParentCiPipeline,
		CiPipelineType:   ciPipeline.PipelineType,
	}
	if ciPipeline.App != nil {
		dto.TeamId = ciPipeline.App.TeamId
	}
	return dto, nil
}
