package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
)

type CiHandlerEnt interface {
}

func (impl *CiHandlerImpl) updateResourceStatusInCache(ciWorkflowId int, podName string, namespace string, status string) {
	//do nothing
}

func (impl *CiHandlerImpl) getPipelineIdForTriggerView(pipeline *pipelineConfig.CiPipeline) (pipelineId int) {
	if pipeline.ParentCiPipeline == 0 {
		pipelineId = pipeline.Id
	} else {
		pipelineId = pipeline.ParentCiPipeline
	}
	return pipelineId
}
