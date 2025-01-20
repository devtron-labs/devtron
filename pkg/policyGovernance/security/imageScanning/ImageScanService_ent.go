package imageScanning

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/adapter"
	bean3 "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/helper/parser"
)

func (impl ImageScanServiceImpl) GetScanResults(resourceScanQueryParams *bean3.ResourceScanQueryParams) (resp parser.ResourceScanResponseDto, err error) {
	request := &bean3.ImageScanRequest{
		ArtifactId: resourceScanQueryParams.ArtifactId,
		AppId:      resourceScanQueryParams.AppId,
		EnvId:      resourceScanQueryParams.EnvId,
	}
	respFromExecutionDetail, err := impl.FetchExecutionDetailResult(request)
	if err != nil {
		impl.Logger.Errorw("error encountered in GetScanResults", "req", request, "err", err)
		return resp, err
	}
	// build an adapter to convert the respFromExecutionDetail to the required ResourceScanResponseDto format
	return adapter.ExecutionDetailsToResourceScanResponseDto(respFromExecutionDetail), nil

}
