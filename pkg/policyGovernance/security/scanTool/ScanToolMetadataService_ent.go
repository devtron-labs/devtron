package scanTool

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/bean"
)

type ScanToolMetadataService_ent interface {
	RegisterScanTools(registerScanToolDto *bean.RegisterScanToolsDto, userId int32) error
}

func (impl *ScanToolMetadataServiceImpl) RegisterScanTools(registerScanToolDto *bean.RegisterScanToolsDto, userId int32) error {
	return nil
}
