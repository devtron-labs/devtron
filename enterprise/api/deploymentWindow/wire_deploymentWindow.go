/*
 * Copyright (c) 2024. Devtron Inc.
 */

package deploymentWindow

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/deploymentWindow"
	"github.com/google/wire"
)

var DeploymentWindowWireSet = wire.NewSet(
	NewDeploymentWindowRouterImpl,
	wire.Bind(new(DeploymentWindowRouter), new(*DeploymentWindowRouterImpl)),
	NewDeploymentWindowRestHandlerImpl,
	wire.Bind(new(DeploymentWindowRestHandler), new(*DeploymentWindowRestHandlerImpl)),
	deploymentWindow.NewDeploymentWindowServiceImpl,
	wire.Bind(new(deploymentWindow.DeploymentWindowService), new(*deploymentWindow.DeploymentWindowServiceImpl)),
)
