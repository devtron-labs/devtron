package client

import (
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/google/wire"
)

var HelmAppWireSet = wire.NewSet(
	NewHelmAppClientImpl,
	wire.Bind(new(HelmAppClient), new(*HelmAppClientImpl)),
	NewHelmAppServiceImpl,
	wire.Bind(new(HelmAppService), new(*HelmAppServiceImpl)),
	NewHelmAppRestHandlerImpl,
	wire.Bind(new(HelmAppRestHandler), new(*HelmAppRestHandlerImpl)),
	NewHelmAppRouterImpl,
	wire.Bind(new(HelmAppRouter), new(*HelmAppRouterImpl)),
	GetConfig,
	rbac.NewEnforcerUtilHelmImpl,
	wire.Bind(new(rbac.EnforcerUtilHelm), new(*rbac.EnforcerUtilHelmImpl)),

)
