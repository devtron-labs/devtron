package globalConfig

import (
	auth "github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig/repository"
	"github.com/google/wire"
)

var GlobalConfigWireSet = wire.NewSet(
	NewGlobalAuthorisationConfigRestHandlerImpl,
	wire.Bind(new(AuthorisationConfigRestHandler), new(*AuthorisationConfigRestHandlerImpl)),

	NewGlobalConfigAuthorisationConfigRouterImpl,
	wire.Bind(new(AuthorisationConfigRouter), new(*AuthorisationConfigRouterImpl)),

	repository.NewGlobalAuthorisationConfigRepositoryImpl,
	wire.Bind(new(repository.GlobalAuthorisationConfigRepository), new(*repository.GlobalAuthorisationConfigRepositoryImpl)),

	auth.NewGlobalAuthorisationConfigServiceImpl,
	wire.Bind(new(auth.GlobalAuthorisationConfigService), new(*auth.GlobalAuthorisationConfigServiceImpl)),
)
