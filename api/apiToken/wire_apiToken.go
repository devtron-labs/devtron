package apiToken

import (
	"github.com/devtron-labs/devtron/pkg/apiToken"
	"github.com/google/wire"
)

var ApiTokenWireSet = wire.NewSet(
	apiToken.NewApiTokenRepositoryImpl,
	wire.Bind(new(apiToken.ApiTokenRepository), new(*apiToken.ApiTokenRepositoryImpl)),
	apiToken.NewApiTokenServiceImpl,
	wire.Bind(new(apiToken.ApiTokenService), new(*apiToken.ApiTokenServiceImpl)),
	NewApiTokenRestHandlerImpl,
	wire.Bind(new(ApiTokenRestHandler), new(*ApiTokenRestHandlerImpl)),
	NewApiTokenRouterImpl,
	wire.Bind(new(ApiTokenRouter), new(*ApiTokenRouterImpl)),
)
