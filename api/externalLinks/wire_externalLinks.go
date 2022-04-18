package externalLinks

import (
	"github.com/devtron-labs/devtron/pkg/externalLinks"
	"github.com/google/wire"
)

//depends on sql,
//TODO integrate user auth module

var ExternalLinksWireSet = wire.NewSet(
	externalLinks.NewExternalLinksToolsRepositoryImpl,
	wire.Bind(new(externalLinks.ExternalLinksToolsRepository), new(*externalLinks.ExternalLinksToolsRepositoryImpl)),
	externalLinks.NewExternalLinksClustersRepositoryImpl,
	wire.Bind(new(externalLinks.ExternalLinksClustersRepository), new(*externalLinks.ExternalLinksClustersRepositoryImpl)),
	externalLinks.NewExternalLinksRepositoryImpl,
	wire.Bind(new(externalLinks.ExternalLinksRepository), new(*externalLinks.ExternalLinksRepositoryImpl)),


	externalLinks.NewExternalLinksServiceImpl,
	wire.Bind(new(externalLinks.ExternalLinksService), new(*externalLinks.ExternalLinksServiceImpl)),
	NewExternalLinksRestHandlerImpl,
	wire.Bind(new(ExternalLinksRestHandler), new(*ExternalLinksRestHandlerImpl)),
	NewExternalLinksRouterImpl,
	wire.Bind(new(ExternalLinksRouter), new(*ExternalLinksRouterImpl)),
)
