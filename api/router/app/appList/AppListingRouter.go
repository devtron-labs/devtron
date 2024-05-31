/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package appList

import (
	"github.com/devtron-labs/devtron/api/restHandler/app/appList"
	"github.com/gorilla/mux"
)

type AppListingRouter interface {
	InitAppListingRouter(helmRouter *mux.Router)
}

type AppListingRouterImpl struct {
	appListingRestHandler appList.AppListingRestHandler
}

func NewAppListingRouterImpl(appListingRestHandler appList.AppListingRestHandler) *AppListingRouterImpl {
	router := &AppListingRouterImpl{
		appListingRestHandler: appListingRestHandler,
	}
	return router
}

func (router AppListingRouterImpl) InitAppListingRouter(appListingRouter *mux.Router) {
	appListingRouter.Path("").
		HandlerFunc(router.appListingRestHandler.FetchAppsByEnvironmentV2).
		Methods("POST")

	appListingRouter.Path("/v2").
		HandlerFunc(router.appListingRestHandler.FetchAppsByEnvironmentV2).
		Methods("POST")

	appListingRouter.Path("/group/{env-id}").
		HandlerFunc(router.appListingRestHandler.FetchOverviewAppsByEnvironment).
		Methods("GET")
}
