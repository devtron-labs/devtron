/*
 * Copyright (c) 2024. Devtron Inc.
 */

package overview

import (
	"github.com/devtron-labs/devtron/pkg/overview/cache"
	"github.com/devtron-labs/devtron/pkg/overview/config"
	"github.com/google/wire"
)

// OverviewWireSet provides wire set for overview module
var OverviewWireSet = wire.NewSet(
	config.GetClusterOverviewConfig,

	// Service layer
	NewAppManagementServiceImpl,
	wire.Bind(new(AppManagementService), new(*AppManagementServiceImpl)),

	NewDoraMetricsServiceImpl,
	wire.Bind(new(DoraMetricsService), new(*DoraMetricsServiceImpl)),

	NewInsightsServiceImpl,
	wire.Bind(new(InsightsService), new(*InsightsServiceImpl)),

	// Cluster cache service
	cache.NewClusterCacheServiceImpl,
	wire.Bind(new(cache.ClusterCacheService), new(*cache.ClusterCacheServiceImpl)),

	// Cluster overview service (uses background refresh worker)
	NewClusterOverviewServiceImpl,
	wire.Bind(new(ClusterOverviewService), new(*ClusterOverviewServiceImpl)),

	// Security overview service (uses existing image scanning repositories)
	NewSecurityOverviewServiceImpl,
	wire.Bind(new(SecurityOverviewService), new(*SecurityOverviewServiceImpl)),

	// Main overview service
	NewOverviewServiceImpl,
	wire.Bind(new(OverviewService), new(*OverviewServiceImpl)),
)
