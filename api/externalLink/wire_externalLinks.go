package externalLink

import (
	"github.com/devtron-labs/devtron/pkg/externalLink"
	"github.com/google/wire"
)

//depends on sql,
//TODO integrate user auth module

var ExternalLinkWireSet = wire.NewSet(
	externalLink.NewExternalLinkMonitoringToolRepositoryImpl,
	wire.Bind(new(externalLink.ExternalLinkMonitoringToolRepository), new(*externalLink.ExternalLinkMonitoringToolRepositoryImpl)),
	externalLink.NewExternalLinkClusterMappingRepositoryImpl,
	wire.Bind(new(externalLink.ExternalLinkClusterMappingRepository), new(*externalLink.ExternalLinkClusterMappingRepositoryImpl)),
	externalLink.NewExternalLinkRepositoryImpl,
	wire.Bind(new(externalLink.ExternalLinkRepository), new(*externalLink.ExternalLinkRepositoryImpl)),


	externalLink.NewExternalLinkServiceImpl,
	wire.Bind(new(externalLink.ExternalLinkService), new(*externalLink.ExternalLinkServiceImpl)),
	NewExternalLinkRestHandlerImpl,
	wire.Bind(new(ExternalLinkRestHandler), new(*ExternalLinkRestHandlerImpl)),
	NewExternalLinkRouterImpl,
	wire.Bind(new(ExternalLinkRouter), new(*ExternalLinkRouterImpl)),
)
