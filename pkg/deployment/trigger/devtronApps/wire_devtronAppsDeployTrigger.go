package devtronApps

import (
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest"
	"github.com/google/wire"
)

var DevtronAppsDeployTriggerWireSet = wire.NewSet(
	userDeploymentRequest.WireSet,
	NewTriggerServiceImpl,
	wire.Bind(new(TriggerService), new(*TriggerServiceImpl)),
)
