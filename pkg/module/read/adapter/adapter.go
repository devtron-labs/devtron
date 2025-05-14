/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package adapter

import (
	moduleBean "github.com/devtron-labs/devtron/pkg/module/bean"
	"github.com/devtron-labs/devtron/pkg/module/read/bean"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
)

func GetModuleInfoMin(module *moduleRepo.Module) *bean.ModuleInfoMin {
	return &bean.ModuleInfoMin{
		Name:       module.Name,
		Status:     module.Status,
		Enabled:    module.Enabled,
		ModuleType: module.ModuleType,
	}
}

func GetDefaultModuleInfo(moduleName string) *bean.ModuleInfoMin {
	return &bean.ModuleInfoMin{
		Name:   moduleName,
		Status: moduleBean.ModuleStatusNotInstalled,
	}
}
