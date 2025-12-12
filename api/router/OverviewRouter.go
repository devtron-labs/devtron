/*
 * Copyright (c) 2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type OverviewRouter interface {
	InitOverviewRouter(overviewRouter *mux.Router)
}

type OverviewRouterImpl struct {
	overviewRestHandler restHandler.OverviewRestHandler
}

func NewOverviewRouterImpl(overviewRestHandler restHandler.OverviewRestHandler) *OverviewRouterImpl {
	return &OverviewRouterImpl{
		overviewRestHandler: overviewRestHandler,
	}
}

func (router OverviewRouterImpl) InitOverviewRouter(overviewRouter *mux.Router) {
	// New Apps Overview API
	overviewRouter.Path("/apps-overview").
		HandlerFunc(router.overviewRestHandler.GetAppsOverview).
		Methods("GET")

	// New Workflow Overview API
	overviewRouter.Path("/workflow-overview").
		HandlerFunc(router.overviewRestHandler.GetWorkflowOverview).
		Methods("GET")

	// Build and Deployment Activity
	overviewRouter.Path("/build-deployment-activity").
		HandlerFunc(router.overviewRestHandler.GetBuildDeploymentActivity).
		Methods("GET")

	// Build and Deployment Activity Detailed
	overviewRouter.Path("/build-deployment-activity/detailed").
		HandlerFunc(router.overviewRestHandler.GetBuildDeploymentActivityDetailed).
		Methods("GET")

	// DORA Metrics
	overviewRouter.Path("/dora-metrics").
		HandlerFunc(router.overviewRestHandler.GetDoraMetrics).
		Methods("GET")

	// Pipeline Insights
	overviewRouter.Path("/pipeline-insights").
		HandlerFunc(router.overviewRestHandler.GetInsights).
		Methods("GET")

	// Infra Overview Subrouter
	infraOverviewRouter := overviewRouter.PathPrefix("/infra").Subrouter()

	// Cluster Management Overview
	infraOverviewRouter.Path("").
		HandlerFunc(router.overviewRestHandler.GetClusterOverview).
		Methods("GET")

	// Delete Cluster Overview Cache
	infraOverviewRouter.Path("/cache").
		HandlerFunc(router.overviewRestHandler.DeleteClusterOverviewCache).
		Methods("DELETE")

	// Refresh Cluster Overview Cache
	infraOverviewRouter.Path("/refresh").
		HandlerFunc(router.overviewRestHandler.RefreshClusterOverviewCache).
		Methods("GET")

	// Cluster Overview Detailed Node Info
	infraOverviewRouter.Path("/node-list").
		HandlerFunc(router.overviewRestHandler.GetClusterOverviewDetailedNodeInfo).
		Methods("GET")

	// Security Overview Subrouter
	securityOverviewRouter := overviewRouter.PathPrefix("/security").Subrouter()

	// Security Overview - "At a Glance" metrics (organization-wide)
	securityOverviewRouter.Path("/security-glance").
		HandlerFunc(router.overviewRestHandler.GetSecurityOverview).
		Methods("GET")

	// Severity Insights - With prod/non-prod filtering
	securityOverviewRouter.Path("/severity-insights").
		HandlerFunc(router.overviewRestHandler.GetSeverityInsights).
		Methods("GET")

	// Deployment Security Status
	securityOverviewRouter.Path("/deployment-security-status").
		HandlerFunc(router.overviewRestHandler.GetDeploymentSecurityStatus).
		Methods("GET")

	// Vulnerability Trend - Time-series with prod/non-prod filtering
	securityOverviewRouter.Path("/vulnerability-trend").
		HandlerFunc(router.overviewRestHandler.GetVulnerabilityTrend).
		Methods("GET")

	// Blocked Deployments Trend - Organization-wide
	securityOverviewRouter.Path("/blocked-deployments-trend").
		HandlerFunc(router.overviewRestHandler.GetBlockedDeploymentsTrend).
		Methods("GET")

}
