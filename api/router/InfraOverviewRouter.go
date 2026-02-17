/*
 * Copyright (c) 2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type InfraOverviewRouter interface {
	InitInfraOverviewRouter(infraOverviewRouter *mux.Router)
}

type InfraOverviewRouterImpl struct {
	infraOverviewRestHandler restHandler.InfraOverviewRestHandler
}

func NewInfraOverviewRouterImpl(infraOverviewRestHandler restHandler.InfraOverviewRestHandler) *InfraOverviewRouterImpl {
	return &InfraOverviewRouterImpl{
		infraOverviewRestHandler: infraOverviewRestHandler,
	}
}

func (router InfraOverviewRouterImpl) InitInfraOverviewRouter(infraOverviewRouter *mux.Router) {
	// Cluster Management Overview
	infraOverviewRouter.Path("").
		HandlerFunc(router.infraOverviewRestHandler.GetClusterOverview).
		Methods("GET")

	// Delete Cluster Overview Cache
	infraOverviewRouter.Path("/cache").
		HandlerFunc(router.infraOverviewRestHandler.DeleteClusterOverviewCache).
		Methods("DELETE")

	// Refresh Cluster Overview Cache
	infraOverviewRouter.Path("/refresh").
		HandlerFunc(router.infraOverviewRestHandler.RefreshClusterOverviewCache).
		Methods("GET")

	// Cluster Overview Detailed Node Info
	infraOverviewRouter.Path("/node-list").
		HandlerFunc(router.infraOverviewRestHandler.GetClusterOverviewDetailedNodeInfo).
		Methods("GET")
}
