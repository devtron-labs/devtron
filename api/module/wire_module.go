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
	"github.com/devtron-labs/devtron/pkg/module"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	moduleDataStore "github.com/devtron-labs/devtron/pkg/module/store"
	"github.com/google/wire"
)

var ModuleWireSet = wire.NewSet(
	module.NewModuleActionAuditLogRepositoryImpl,
	wire.Bind(new(module.ModuleActionAuditLogRepository), new(*module.ModuleActionAuditLogRepositoryImpl)),
	moduleRepo.NewModuleRepositoryImpl,
	wire.Bind(new(moduleRepo.ModuleRepository), new(*moduleRepo.ModuleRepositoryImpl)),
	moduleRepo.NewModuleResourceStatusRepositoryImpl,
	wire.Bind(new(moduleRepo.ModuleResourceStatusRepository), new(*moduleRepo.ModuleResourceStatusRepositoryImpl)),
	module.ParseModuleEnvConfig,
	moduleDataStore.InitModuleDataStore,
	module.NewModuleServiceHelperImpl,
	wire.Bind(new(module.ModuleServiceHelper), new(*module.ModuleServiceHelperImpl)),
	module.NewModuleServiceImpl,
	wire.Bind(new(module.ModuleService), new(*module.ModuleServiceImpl)),
	module.NewModuleCronServiceImpl,
	wire.Bind(new(module.ModuleCronService), new(*module.ModuleCronServiceImpl)),
	module.NewModuleCacheServiceImpl,
	wire.Bind(new(module.ModuleCacheService), new(*module.ModuleCacheServiceImpl)),
	NewModuleRestHandlerImpl,
	wire.Bind(new(ModuleRestHandler), new(*ModuleRestHandlerImpl)),
	NewModuleRouterImpl,
	wire.Bind(new(ModuleRouter), new(*ModuleRouterImpl)),
)
