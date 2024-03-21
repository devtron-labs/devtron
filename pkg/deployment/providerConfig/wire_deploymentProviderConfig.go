package providerConfig

import (
	"github.com/google/wire"
)

var DeploymentProviderConfigWireSet = wire.NewSet(
	NewDeploymentTypeOverrideServiceImpl,
	wire.Bind(new(DeploymentTypeOverrideService), new(*DeploymentTypeOverrideServiceImpl)),
)
