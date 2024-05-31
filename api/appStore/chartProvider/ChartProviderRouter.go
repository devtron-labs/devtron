/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package chartProvider

import (
	"github.com/gorilla/mux"
)

type ChartProviderRouter interface {
	Init(configRouter *mux.Router)
}

type ChartProviderRouterImpl struct {
	chartProviderRestHandler ChartProviderRestHandler
}

func NewChartProviderRouterImpl(chartProviderRestHandler ChartProviderRestHandler) *ChartProviderRouterImpl {
	return &ChartProviderRouterImpl{
		chartProviderRestHandler: chartProviderRestHandler,
	}
}

func (router ChartProviderRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("/list").
		HandlerFunc(router.chartProviderRestHandler.GetChartProviderList).Methods("GET")
	configRouter.Path("/update").
		HandlerFunc(router.chartProviderRestHandler.ToggleChartProvider).Methods("POST")
	configRouter.Path("/sync-chart").
		HandlerFunc(router.chartProviderRestHandler.SyncChartProvider).Methods("POST")
}
