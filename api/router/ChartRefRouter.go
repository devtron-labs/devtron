/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type ChartRefRouter interface {
	initChartRefRouter(helmRouter *mux.Router)
}

type ChartRefRouterImpl struct {
	chartRefRestHandler restHandler.ChartRefRestHandler
}

func NewChartRefRouterImpl(chartRefRestHandler restHandler.ChartRefRestHandler) *ChartRefRouterImpl {
	router := &ChartRefRouterImpl{
		chartRefRestHandler: chartRefRestHandler,
	}
	return router
}

func (router ChartRefRouterImpl) initChartRefRouter(userAuthRouter *mux.Router) {

	userAuthRouter.Path("/autocomplete").
		HandlerFunc(router.chartRefRestHandler.ChartRefAutocomplete).Methods("GET")

	userAuthRouter.Path("/autocomplete/{appId}").
		HandlerFunc(router.chartRefRestHandler.ChartRefAutocompleteForApp).Methods("GET")

	userAuthRouter.Path("/autocomplete/chart/{chartRefId}").
		HandlerFunc(router.chartRefRestHandler.ChartRefAutocompleteByChartId).Methods("GET")

	userAuthRouter.Path("/autocomplete/{appId}/{environmentId}").
		HandlerFunc(router.chartRefRestHandler.ChartRefAutocompleteForEnv).Methods("GET")

}
