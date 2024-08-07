package adaptor

import (
	bean2 "github.com/devtron-labs/devtron/pkg/plugin/bean"
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
)

func GetPluginParentMetadataDbObject(pluginDto *bean2.PluginParentMetadataDto, userId int32) *repository.PluginParentMetadata {
	return repository.NewPluginParentMetadata().CreateAuditLog(userId).
		WithBasicMetadata(pluginDto.Name, pluginDto.PluginIdentifier, pluginDto.Description, pluginDto.Icon, pluginDto.Type)
}

func GetPluginVersionMetadataDbObject(pluginDto *bean2.PluginParentMetadataDto, userId int32) *repository.PluginMetadata {
	versionDto := pluginDto.Versions.DetailedPluginVersionData[0]
	return repository.NewPluginVersionMetadata().CreateAuditLog(userId).WithBasicMetadata(pluginDto.Name, versionDto.Description, versionDto.Version, versionDto.DocLink)
}
