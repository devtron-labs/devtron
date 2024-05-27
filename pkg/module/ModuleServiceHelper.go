/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
