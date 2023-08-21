package protect

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/protect"
	"github.com/google/wire"
)

var ProtectWireSet = wire.NewSet(
	protect.NewResourceProtectionRepositoryImpl,
	wire.Bind(new(protect.ResourceProtectionRepository), new(*protect.ResourceProtectionRepositoryImpl)),
	protect.NewResourceProtectionServiceImpl,
	wire.Bind(new(protect.ResourceProtectionService), new(*protect.ResourceProtectionServiceImpl)),
	NewResourceProtectionRestHandlerImpl,
	wire.Bind(new(ResourceProtectionRestHandler), new(*ResourceProtectionRestHandlerImpl)),
	NewResourceProtectionRouterImpl,
	wire.Bind(new(ResourceProtectionRouter), new(*ResourceProtectionRouterImpl)),
)
