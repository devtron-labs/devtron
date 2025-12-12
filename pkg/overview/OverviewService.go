/*
 * Copyright (c) 2024. Devtron Inc.
 */

package overview

import (
	"context"

	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/devtron/pkg/overview/bean"
	"github.com/devtron-labs/devtron/pkg/overview/cache"
	"github.com/devtron-labs/devtron/pkg/overview/constants"
)

type OverviewService interface {
	// New Apps Overview
	GetAppsOverview(ctx context.Context) (*bean.AppsOverviewResponse, error)

	// New Workflow Overview
	GetWorkflowOverview(ctx context.Context) (*bean.WorkflowOverviewResponse, error)

	// Build and Deployment Activity
	GetBuildDeploymentActivity(ctx context.Context, request *bean.BuildDeploymentActivityRequest) (*bean.BuildDeploymentActivityResponse, error)
	GetBuildDeploymentActivityDetailed(ctx context.Context, request *bean.BuildDeploymentActivityDetailedRequest) (*bean.BuildDeploymentActivityDetailedResponse, error)

	// DORA Metrics
	GetDoraMetrics(ctx context.Context, request *bean.DoraMetricsRequest) (*bean.DoraMetricsResponse, error)

	// Insights
	GetInsights(ctx context.Context, request *bean.InsightsRequest) (*bean.InsightsResponse, error)

	// Cluster Management Overview
	GetClusterOverview(ctx context.Context) (*bean.ClusterOverviewResponse, error)
	DeleteClusterOverviewCache(ctx context.Context) error
	RefreshClusterOverviewCache(ctx context.Context) error

	// Cluster Overview Detailed Drill-down API (unified endpoint for all node view group types)
	GetClusterOverviewDetailedNodeInfo(ctx context.Context, request *bean.ClusterOverviewDetailRequest) (*bean.ClusterOverviewNodeDetailedResponse, error)

	// Security Overview APIs
	GetSecurityOverview(ctx context.Context, request *bean.SecurityOverviewRequest) (*bean.SecurityOverviewResponse, error)
	GetSeverityInsights(ctx context.Context, request *bean.SeverityInsightsRequest) (*bean.SeverityInsightsResponse, error)
	GetDeploymentSecurityStatus(ctx context.Context, request *bean.DeploymentSecurityStatusRequest) (*bean.DeploymentSecurityStatusResponse, error)
	GetVulnerabilityTrend(ctx context.Context, currentTimeRange *utils.TimeRangeRequest, envType bean.EnvType, aggregationType constants.AggregationType) (*bean.VulnerabilityTrendResponse, error)
	GetBlockedDeploymentsTrend(ctx context.Context, currentTimeRange *utils.TimeRangeRequest, aggregationType constants.AggregationType) (*bean.BlockedDeploymentsTrendResponse, error)
}

type OverviewServiceImpl struct {
	appManagementService    AppManagementService
	doraMetricsService      DoraMetricsService
	insightsService         InsightsService
	clusterOverviewService  ClusterOverviewService
	clusterCacheService     cache.ClusterCacheService
	securityOverviewService SecurityOverviewService
}

func NewOverviewServiceImpl(
	appManagementService AppManagementService,
	doraMetricsService DoraMetricsService,
	insightsService InsightsService,
	clusterOverviewService ClusterOverviewService,
	clusterCacheService cache.ClusterCacheService,
	securityOverviewService SecurityOverviewService,
) *OverviewServiceImpl {
	return &OverviewServiceImpl{
		appManagementService:    appManagementService,
		doraMetricsService:      doraMetricsService,
		insightsService:         insightsService,
		clusterOverviewService:  clusterOverviewService,
		clusterCacheService:     clusterCacheService,
		securityOverviewService: securityOverviewService,
	}
}

func (impl *OverviewServiceImpl) GetAppsOverview(ctx context.Context) (*bean.AppsOverviewResponse, error) {
	return impl.appManagementService.GetAppsOverview(ctx)
}

func (impl *OverviewServiceImpl) GetWorkflowOverview(ctx context.Context) (*bean.WorkflowOverviewResponse, error) {
	return impl.appManagementService.GetWorkflowOverview(ctx)
}

func (impl *OverviewServiceImpl) GetBuildDeploymentActivity(ctx context.Context, request *bean.BuildDeploymentActivityRequest) (*bean.BuildDeploymentActivityResponse, error) {
	return impl.appManagementService.GetBuildDeploymentActivity(ctx, request)
}

func (impl *OverviewServiceImpl) GetBuildDeploymentActivityDetailed(ctx context.Context, request *bean.BuildDeploymentActivityDetailedRequest) (*bean.BuildDeploymentActivityDetailedResponse, error) {
	return impl.appManagementService.GetBuildDeploymentActivityDetailed(ctx, request)
}

func (impl *OverviewServiceImpl) GetDoraMetrics(ctx context.Context, request *bean.DoraMetricsRequest) (*bean.DoraMetricsResponse, error) {
	return impl.doraMetricsService.GetDoraMetrics(ctx, request)
}

func (impl *OverviewServiceImpl) GetInsights(ctx context.Context, request *bean.InsightsRequest) (*bean.InsightsResponse, error) {
	return impl.insightsService.GetInsights(ctx, request)
}

func (impl *OverviewServiceImpl) GetClusterOverview(ctx context.Context) (*bean.ClusterOverviewResponse, error) {
	return impl.clusterOverviewService.GetClusterOverview(ctx)
}

func (impl *OverviewServiceImpl) DeleteClusterOverviewCache(ctx context.Context) error {
	impl.clusterCacheService.InvalidateClusterOverview()
	return nil
}

func (impl *OverviewServiceImpl) RefreshClusterOverviewCache(ctx context.Context) error {
	return impl.clusterOverviewService.RefreshClusterOverviewCache(ctx)
}

func (impl *OverviewServiceImpl) GetClusterOverviewDetailedNodeInfo(ctx context.Context, request *bean.ClusterOverviewDetailRequest) (*bean.ClusterOverviewNodeDetailedResponse, error) {
	return impl.clusterOverviewService.GetClusterOverviewDetailedNodeInfo(ctx, request)
}

// ============================================================================
// Security Overview APIs
// ============================================================================

func (impl *OverviewServiceImpl) GetSecurityOverview(ctx context.Context, request *bean.SecurityOverviewRequest) (*bean.SecurityOverviewResponse, error) {
	return impl.securityOverviewService.GetSecurityOverview(ctx, request)
}

func (impl *OverviewServiceImpl) GetSeverityInsights(ctx context.Context, request *bean.SeverityInsightsRequest) (*bean.SeverityInsightsResponse, error) {
	return impl.securityOverviewService.GetSeverityInsights(ctx, request)
}

func (impl *OverviewServiceImpl) GetDeploymentSecurityStatus(ctx context.Context, request *bean.DeploymentSecurityStatusRequest) (*bean.DeploymentSecurityStatusResponse, error) {
	return impl.securityOverviewService.GetDeploymentSecurityStatus(ctx, request)
}

func (impl *OverviewServiceImpl) GetVulnerabilityTrend(ctx context.Context, currentTimeRange *utils.TimeRangeRequest, envType bean.EnvType, aggregationType constants.AggregationType) (*bean.VulnerabilityTrendResponse, error) {
	return impl.securityOverviewService.GetVulnerabilityTrend(ctx, currentTimeRange, envType, aggregationType)
}

func (impl *OverviewServiceImpl) GetBlockedDeploymentsTrend(ctx context.Context, currentTimeRange *utils.TimeRangeRequest, aggregationType constants.AggregationType) (*bean.BlockedDeploymentsTrendResponse, error) {
	return impl.securityOverviewService.GetBlockedDeploymentsTrend(ctx, currentTimeRange, aggregationType)
}
