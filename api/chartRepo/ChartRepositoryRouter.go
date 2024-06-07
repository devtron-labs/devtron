/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package chartRepo

import (
	"github.com/gorilla/mux"
)

type ChartRepositoryRouter interface {
	Init(configRouter *mux.Router)
}

type ChartRepositoryRouterImpl struct {
	chartRepositoryRestHandler ChartRepositoryRestHandler
}

func NewChartRepositoryRouterImpl(chartRepositoryRestHandler ChartRepositoryRestHandler) *ChartRepositoryRouterImpl {
	return &ChartRepositoryRouterImpl{chartRepositoryRestHandler: chartRepositoryRestHandler}
}

func (router ChartRepositoryRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("/sync-charts").
		HandlerFunc(router.chartRepositoryRestHandler.TriggerChartSyncManual).Methods("POST")
	configRouter.Path("/list").
		HandlerFunc(router.chartRepositoryRestHandler.GetChartRepoList).Methods("GET")
	configRouter.Path("/list/min").
		HandlerFunc(router.chartRepositoryRestHandler.GetChartRepoListMin).Methods("GET")
	configRouter.Path("/{id}").
		HandlerFunc(router.chartRepositoryRestHandler.GetChartRepoById).Methods("GET")
	configRouter.Path("/create").
		HandlerFunc(router.chartRepositoryRestHandler.CreateChartRepo).Methods("POST")
	configRouter.Path("/update").
		HandlerFunc(router.chartRepositoryRestHandler.UpdateChartRepo).Methods("POST")
	configRouter.Path("/validate").
		HandlerFunc(router.chartRepositoryRestHandler.ValidateChartRepo).Methods("POST")
	configRouter.Path("").
		HandlerFunc(router.chartRepositoryRestHandler.DeleteChartRepo).Methods("DELETE")
}
