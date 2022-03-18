package deployment

import "github.com/google/wire"

var ClusterWireSet = wire.NewSet(
	NewDeploymentRouterImpl,
	wire.Bind(new(DeploymentRouter), new(*DeploymentRouterImpl)),
	NewDeploymentRestHandlerImpl,
	wire.Bind(new(DeploymentRestHandler), new(*DeploymentRestHandlerImpl)),
)
