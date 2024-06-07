/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package appStoreDiscover

import (
	"github.com/gorilla/mux"
)

type AppStoreDiscoverRouter interface {
	Init(configRouter *mux.Router)
}

type AppStoreDiscoverRouterImpl struct {
	appStoreRestHandler AppStoreRestHandler
}

func NewAppStoreDiscoverRouterImpl(appStoreRestHandler AppStoreRestHandler) *AppStoreDiscoverRouterImpl {
	return &AppStoreDiscoverRouterImpl{
		appStoreRestHandler: appStoreRestHandler,
	}
}

func (router AppStoreDiscoverRouterImpl) Init(configRouter *mux.Router) {

	configRouter.Path("/").
		HandlerFunc(router.appStoreRestHandler.FindAllApps).Methods("GET")

	configRouter.Path("/application/{id}").
		HandlerFunc(router.appStoreRestHandler.GetChartDetailsForVersion).Methods("GET")

	configRouter.Path("/application/{appStoreId}/version/autocomplete").
		HandlerFunc(router.appStoreRestHandler.GetChartVersions).Methods("GET")

	configRouter.Path("/application/chartInfo/{appStoreApplicationVersionId}").
		HandlerFunc(router.appStoreRestHandler.GetChartInfo).Methods("GET")

	configRouter.Path("/search").
		HandlerFunc(router.appStoreRestHandler.SearchAppStoreChartByName).Queries("chartName", "{chartName}").
		Methods("GET")

}
