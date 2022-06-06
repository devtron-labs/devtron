package apiToken

import (
	apiTokenAuth "github.com/devtron-labs/authenticator/apiToken"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/apiToken"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/google/wire"
)

var ApiTokenSecretWireSet = wire.NewSet(
	attributes.NewAttributesServiceImpl,
	wire.Bind(new(attributes.AttributesService), new(*attributes.AttributesServiceImpl)),
	repository.NewAttributesRepositoryImpl,
	wire.Bind(new(repository.AttributesRepository), new(*repository.AttributesRepositoryImpl)),
	apiTokenAuth.InitApiTokenSecretStore,
	apiToken.NewApiTokenSecretServiceImpl,
	wire.Bind(new(apiToken.ApiTokenSecretService), new(*apiToken.ApiTokenSecretServiceImpl)),
)
