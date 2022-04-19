package externalLinks

import (
	"github.com/devtron-labs/devtron/pkg/externalLinks"
	"github.com/google/wire"
)

//depends on sql,
//TODO integrate user auth module

var ExternalLinksWireSet = wire.NewSet(
	externalLinks.NewExternalLinkMonitoringToolRepositoryImpl,
	wire.Bind(new(externalLinks.ExternalLinkMonitoringToolRepository), new(*externalLinks.ExternalLinkMonitoringToolRepositoryImpl)),
	externalLinks.NewExternalLinkClusterMappingRepositoryImpl,
	wire.Bind(new(externalLinks.ExternalLinkClusterMappingRepository), new(*externalLinks.ExternalLinkClusterMappingRepositoryImpl)),
	externalLinks.NewExternalLinkRepositoryImpl,
	wire.Bind(new(externalLinks.ExternalLinkRepository), new(*externalLinks.ExternalLinkRepositoryImpl)),


	externalLinks.NewExternalLinkServiceImpl,
	wire.Bind(new(externalLinks.ExternalLinkService), new(*externalLinks.ExternalLinkServiceImpl)),
	NewExternalLinkRestHandlerImpl,
	wire.Bind(new(ExternalLinkRestHandler), new(*ExternalLinkRestHandlerImpl)),
	NewExternalLinkRouterImpl,
	wire.Bind(new(ExternalLinkRouter), new(*ExternalLinkRouterImpl)),
)
