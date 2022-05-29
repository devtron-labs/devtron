package apiToken

import (
	"github.com/devtron-labs/devtron/pkg/apiToken"
	"github.com/google/wire"
)

var ApiTokenWireSet = wire.NewSet(
	apiToken.NewApiTokenSecretRepositoryImpl,
	wire.Bind(new(apiToken.ApiTokenSecretRepository), new(*apiToken.ApiTokenSecretRepositoryImpl)),
	apiToken.NewApiTokenRepositoryImpl,
	wire.Bind(new(apiToken.ApiTokenRepository), new(*apiToken.ApiTokenRepositoryImpl)),
	apiToken.InitApiTokenSecretStore,
	apiToken.NewApiTokenSecretServiceImpl,
	wire.Bind(new(apiToken.ApiTokenSecretService), new(*apiToken.ApiTokenSecretServiceImpl)),
	apiToken.NewApiTokenServiceImpl,
	wire.Bind(new(apiToken.ApiTokenService), new(*apiToken.ApiTokenServiceImpl)),
	NewApiTokenRestHandlerImpl,
	wire.Bind(new(ApiTokenRestHandler), new(*ApiTokenRestHandlerImpl)),
	NewApiTokenRouterImpl,
	wire.Bind(new(ApiTokenRouter), new(*ApiTokenRouterImpl)),
)
