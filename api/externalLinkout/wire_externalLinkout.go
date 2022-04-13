package externalLinkout

import (
	"github.com/devtron-labs/devtron/pkg/externalLinkout"
	"github.com/google/wire"
)

//depends on sql,
//TODO integrate user auth module

var ExternalLinkoutWireSet = wire.NewSet(
	externalLinkout.NewExternalLinkoutToolsRepositoryImpl,
	wire.Bind(new(externalLinkout.ExternalLinkoutToolsRepository), new(*externalLinkout.ExternalLinkoutToolsRepositoryImpl)),
	externalLinkout.NewExternalLinksClustersRepositoryImpl,
	wire.Bind(new(externalLinkout.ExternalLinksClustersRepository), new(*externalLinkout.ExternalLinksClustersRepositoryImpl)),
	externalLinkout.NewExternalLinksRepositoryImpl,
	wire.Bind(new(externalLinkout.ExternalLinksRepository), new(*externalLinkout.ExternalLinksRepositoryImpl)),


	externalLinkout.NewExternalLinkoutServiceImpl,
	wire.Bind(new(externalLinkout.ExternalLinkoutService), new(*externalLinkout.ExternalLinkoutServiceImpl)),
	NewExternalLinkoutRestHandlerImpl,
	wire.Bind(new(ExternalLinkoutRestHandler), new(*ExternalLinkoutRestHandlerImpl)),
	NewExternalLinkoutRouterImpl,
	wire.Bind(new(ExternalLinkoutRouter), new(*ExternalLinkoutRouterImpl)),
)
