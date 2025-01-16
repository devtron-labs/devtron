package pipeline

import (
	"github.com/devtron-labs/common-lib/imageScan/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/repository"
)

func (impl *CiServiceImpl) fetchImageScanExecutionMedium() (*repository.ScanToolMetadata, bean.ExecutionMedium, error) {
	return &repository.ScanToolMetadata{}, "", nil
}

func (impl *CiServiceImpl) fetchImageScanExecutionStepsForWfRequest(scanToolMetadata *repository.ScanToolMetadata) ([]*types.ImageScanningSteps, []*bean2.RefPluginObject, error) {
	return nil, nil, nil
}
