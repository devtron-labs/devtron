package pipeline

import (
	"github.com/devtron-labs/common-lib/imageScan/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/repository"
)

func (impl *CiServiceImpl) fetchScanVia() (*repository.ScanToolMetadata, bean.ExecutionMedium, error) {
	return &repository.ScanToolMetadata{}, "", nil
}

func (impl *CiServiceImpl) fetchImageScanExecutionSteps(scanToolMetadata *repository.ScanToolMetadata) ([]*types.ImageScanningSteps, error) {
	return nil, nil
}
