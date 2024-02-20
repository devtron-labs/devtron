package appStoreDeployment

import "github.com/google/wire"

var TestAppStoreDeploymentWireSet = wire.NewSet(
	NewAppStoreDeploymentRestHandlerImpl,
	wire.Bind(new(AppStoreDeploymentRestHandler), new(*AppStoreDeploymentRestHandlerImpl)),
)
