/*
 * Copyright (c) 2024. Devtron Inc.
 */

package trigger

import (
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	"github.com/google/wire"
)

var DeploymentTriggerWireSet = wire.NewSet(
	devtronApps.DevtronAppsDeployTriggerWireSet,
)
