package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/bean/common"
)

type CiHandlerEnt interface {
}

func (impl *CiHandlerImpl) updateRuntimeParamsForAutoCI(ciPipelineId int, runtimeParameters *common.RuntimeParameters) (*common.RuntimeParameters, error) {
	return runtimeParameters, nil
}

func (impl *CiHandlerImpl) updateResourceStatusInCache(ciWorkflowId int, podName string, namespace string, status string) {
	//do nothing
}

func (impl *CiHandlerImpl) getRuntimeParamsForBuildingManualTriggerHashes(ciTriggerRequest bean.CiTriggerRequest) *common.RuntimeParameters {
	return common.NewRuntimeParameters()
}

func (impl *CiHandlerImpl) getPipelineIdForTriggerView(pipeline *pipelineConfig.CiPipeline) (pipelineId int) {
	if pipeline.ParentCiPipeline == 0 {
		pipelineId = pipeline.Id
	} else {
		pipelineId = pipeline.ParentCiPipeline
	}
	return pipelineId
}
