package module

import (
	"fmt"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	"github.com/devtron-labs/devtron/util"
)

type ModuleServiceHelper interface {
	GetModuleMetadata(moduleName string) ([]byte, error)
}

type ModuleServiceHelperImpl struct {
	serverEnvConfig *serverEnvConfig.ServerEnvConfig
}

func NewModuleServiceHelperImpl(serverEnvConfig *serverEnvConfig.ServerEnvConfig) *ModuleServiceHelperImpl {
	return &ModuleServiceHelperImpl{
		serverEnvConfig: serverEnvConfig,
	}
}

func (impl ModuleServiceHelperImpl) GetModuleMetadata(moduleName string) ([]byte, error) {
	moduleMetaData, err := util.ReadFromUrlWithRetry(impl.buildModuleMetaDataUrl(moduleName))
	return moduleMetaData, err
}

func (impl ModuleServiceHelperImpl) buildModuleMetaDataUrl(moduleName string) string {
	return fmt.Sprintf(impl.serverEnvConfig.ModuleMetaDataApiUrl, moduleName)
}
