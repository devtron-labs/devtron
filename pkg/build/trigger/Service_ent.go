package trigger

import (
	"github.com/devtron-labs/common-lib/imageScan/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/bean/common"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/repository"
)

func (impl *ServiceImpl) updateRuntimeParamsForAutoCI(ciPipelineId int, runtimeParameters *common.RuntimeParameters) (*common.RuntimeParameters, error) {
	return runtimeParameters, nil
}

func (impl *ServiceImpl) getRuntimeParamsForBuildingManualTriggerHashes(ciTriggerRequest bean3.CiTriggerRequest) *common.RuntimeParameters {
	return common.NewRuntimeParameters()
}

func (impl *ServiceImpl) fetchImageScanExecutionMedium() (*repository.ScanToolMetadata, bean.ScanExecutionMedium, error) {
	return &repository.ScanToolMetadata{}, "", nil
}

func (impl *ServiceImpl) fetchImageScanExecutionStepsForWfRequest(scanToolMetadata *repository.ScanToolMetadata) ([]*types.ImageScanningSteps, []*bean2.RefPluginObject, error) {
	return nil, nil, nil
}
