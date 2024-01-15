package lockConfiguation

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration"
	"github.com/google/wire"
)

var LockConfigWireSet = wire.NewSet(
	lockConfiguration.NewRepositoryImpl,
	wire.Bind(new(lockConfiguration.LockConfigurationRepository), new(*lockConfiguration.RepositoryImpl)),
	lockConfiguration.NewLockConfigurationServiceImpl,
	wire.Bind(new(lockConfiguration.LockConfigurationService), new(*lockConfiguration.LockConfigurationServiceImpl)),
	NewLockConfigRestHandlerImpl,
	wire.Bind(new(LockConfigRestHandler), new(*LockConfigRestHandlerImpl)),
	NewLockConfigurationRouterImpl,
	wire.Bind(new(LockConfigurationRouter), new(*LockConfigurationRouterImpl)),
)
