package deployedApp

import (
	"github.com/google/wire"
)

var DeployedAppWireSet = wire.NewSet(
	NewDeployedAppServiceImpl,
	wire.Bind(new(DeployedAppService), new(*DeployedAppServiceImpl)),
)
