package adaptor

import (
	bean2 "github.com/devtron-labs/devtron/pkg/plugin/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/bean"
	"time"
)

func GetPluginMetadataAndStepsDetail(scanToolPluginMetadataDto *bean.ScanToolPluginMetadataDto, scanToolUrl string, version string) *bean2.PluginParentMetadataDto {
	pluginParentObj := &bean2.PluginParentMetadataDto{
		Name:             scanToolPluginMetadataDto.Name,
		PluginIdentifier: scanToolPluginMetadataDto.PluginIdentifier,
		Description:      scanToolPluginMetadataDto.Description,
		Type:             bean2.SHARED.ToString(),
		Icon:             scanToolUrl,
	}
	pluginMetadataDto := &bean2.PluginMetadataDto{
		Tags:        []string{"Security"},
		PluginStage: "SCANNER",
		PluginSteps: scanToolPluginMetadataDto.PluginSteps,
	}
	pluginVersionDetail := &bean2.PluginsVersionDetail{
		PluginMetadataDto: pluginMetadataDto,
		Version:           version,
		IsLatest:          true,
		CreatedOn:         time.Now(),
	}
	pluginParentObj.Versions = &bean2.PluginVersions{DetailedPluginVersionData: []*bean2.PluginsVersionDetail{pluginVersionDetail}}
	return pluginParentObj
}
