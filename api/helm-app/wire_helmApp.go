package client

import (
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
)
