package apiToken

import (
	apiTokenAuth "github.com/devtron-labs/authenticator/apiToken"
	"github.com/devtron-labs/devtron/pkg/apiToken"
	"github.com/google/wire"
)

var ApiTokenSecretWireSet = wire.NewSet(
	apiTokenAuth.InitApiTokenSecretStore,
	apiToken.NewApiTokenSecretServiceImpl,
	wire.Bind(new(apiToken.ApiTokenSecretService), new(*apiToken.ApiTokenSecretServiceImpl)),
)
