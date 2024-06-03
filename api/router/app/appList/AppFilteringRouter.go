/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package appList

import (
	"github.com/devtron-labs/devtron/api/restHandler/app/appList"
	"github.com/gorilla/mux"
)

type AppFilteringRouter interface {
	InitAppFilteringRouter(helmRouter *mux.Router)
}

type AppFilteringRouterImpl struct {
	appFilteringRestHandler appList.AppFilteringRestHandler
}

func NewAppFilteringRouterImpl(appFilteringRestHandler appList.AppFilteringRestHandler) *AppFilteringRouterImpl {
	router := &AppFilteringRouterImpl{
		appFilteringRestHandler: appFilteringRestHandler,
	}
	return router
}

func (router AppFilteringRouterImpl) InitAppFilteringRouter(AppFilteringRouter *mux.Router) {
	AppFilteringRouter.Path("/autocomplete").
		HandlerFunc(router.appFilteringRestHandler.GetClusterTeamAndEnvListForAutocomplete).Methods("GET")
}
