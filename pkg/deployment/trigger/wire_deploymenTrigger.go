package trigger

import (
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	"github.com/google/wire"
)

var DeploymentTriggerWireSet = wire.NewSet(
	devtronApps.DevtronAppsDeployTriggerWireSet,
)
