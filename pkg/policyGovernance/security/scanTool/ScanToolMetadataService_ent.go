package scanTool

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/repository"
)

type ScanToolMetadataService_ent interface {
	MartToolActiveOrInActiveByNameAndVersion(toolName, version string, isActive bool) error
	RegisterScanTools(registerScanToolDto *bean.RegisterScanToolsDto, userId int32) error
	GetActiveTool() (*repository.ScanToolMetadata, error)
}

func (impl *ScanToolMetadataServiceImpl) MartToolActiveOrInActiveByNameAndVersion(toolName, version string, isActive bool) error {
	return nil
}

func (impl *ScanToolMetadataServiceImpl) RegisterScanTools(registerScanToolDto *bean.RegisterScanToolsDto, userId int32) error {
	return nil
}

func (impl *ScanToolMetadataServiceImpl) GetActiveTool() (*repository.ScanToolMetadata, error) {
	return nil, nil
}
