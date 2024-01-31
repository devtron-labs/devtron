package devtronApps

import "github.com/google/wire"

var DevtronAppsDeployTriggerWireSet = wire.NewSet(
	NewTriggerServiceImpl,
	wire.Bind(new(TriggerService), new(*TriggerServiceImpl)),
)
