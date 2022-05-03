package module

import (
	"github.com/devtron-labs/devtron/pkg/module"
	"github.com/google/wire"
)

var ModuleWireSet = wire.NewSet(
	module.NewModuleActionAuditLogRepositoryImpl,
	wire.Bind(new(module.ModuleActionAuditLogRepository), new(*module.ModuleActionAuditLogRepositoryImpl)),
	module.NewModuleRepositoryImpl,
	wire.Bind(new(module.ModuleRepository), new(*module.ModuleRepositoryImpl)),
	module.ParseModuleEnvConfig,
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
