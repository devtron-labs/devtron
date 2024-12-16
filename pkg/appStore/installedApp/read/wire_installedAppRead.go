package read

import "github.com/google/wire"

// WireSet for InstalledAppReadService
// WireSet is registered for FULL MODE services
var WireSet = wire.NewSet(
	NewInstalledAppReadServiceImpl,
	wire.Bind(new(InstalledAppReadService), new(*InstalledAppReadServiceImpl)),
)

// EAWireSet is used for InstalledAppReadServiceEA
// EAWireSet is registered for EA MODE services
var EAWireSet = wire.NewSet(
	NewInstalledAppReadServiceEAImpl,
	wire.Bind(new(InstalledAppReadServiceEA), new(*InstalledAppReadServiceEAImpl)),
)
