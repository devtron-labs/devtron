package module

import (
	"github.com/devtron-labs/devtron/pkg/module"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	"github.com/google/wire"
)

var ModuleWireSet = wire.NewSet(
	module.NewModuleActionAuditLogRepositoryImpl,
	wire.Bind(new(module.ModuleActionAuditLogRepository), new(*module.ModuleActionAuditLogRepositoryImpl)),
	moduleRepo.NewModuleRepositoryImpl,
	wire.Bind(new(moduleRepo.ModuleRepository), new(*moduleRepo.ModuleRepositoryImpl)),
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
