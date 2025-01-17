package bean

import (
	bean2 "github.com/devtron-labs/devtron/pkg/plugin/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/repository"
)

type RegisterScanToolsDto struct {
	ScanToolMetadata       *ScanToolsMetadataDto      `json:"scanToolMetadata" validate:"dive,required"`
	ScanToolPluginMetadata *ScanToolPluginMetadataDto `json:"scanToolPluginMetadata,omitempty" validate:"dive,required"`
}

type ScanToolsMetadataDto struct {
	Name                     string                    `json:"name" validate:"required"`
	Version                  string                    `json:"version" validate:"required"`
	ServerBaseUrl            string                    `json:"serverBaseUrl,omitempty"`
	ResultDescriptorTemplate string                    `json:"resultDescriptorTemplate,omitempty"`
	ScanTarget               repository.ScanTargetType `json:"scanTarget"`
	ToolMetaData             string                    `json:"toolMetadata,omitempty"`
	ScanToolUrl              string                    `json:"scanToolUrl"`
}

type ScanToolPluginMetadataDto struct {
	Name             string                  `json:"name" validate:"required,min=3,max=100,global-entity-name"`
	PluginIdentifier string                  `json:"pluginIdentifier" validate:"required,min=3,max=100,global-entity-name"`
	Description      string                  `json:"description" validate:"max=300"`
	PluginSteps      []*bean2.PluginStepsDto `json:"pluginSteps,omitempty"`
}

const (
	DevtronImageScanningIntegratorPluginIdentifier = "devtron-image-scanning-integrator"
)
