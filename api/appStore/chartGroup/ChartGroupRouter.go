/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package chartGroup

import (
	"github.com/gorilla/mux"
)

type ChartGroupRouterImpl struct {
	ChartGroupRestHandler ChartGroupRestHandler
}
type ChartGroupRouter interface {
	InitChartGroupRouter(helmRouter *mux.Router)
}

func NewChartGroupRouterImpl(ChartGroupRestHandler ChartGroupRestHandler) *ChartGroupRouterImpl {
	return &ChartGroupRouterImpl{ChartGroupRestHandler: ChartGroupRestHandler}

}

func (impl *ChartGroupRouterImpl) InitChartGroupRouter(chartGroupRouter *mux.Router) {
	chartGroupRouter.Path("/").
		HandlerFunc(impl.ChartGroupRestHandler.CreateChartGroup).Methods("POST")
	chartGroupRouter.Path("/").
		HandlerFunc(impl.ChartGroupRestHandler.UpdateChartGroup).Methods("PUT")
	chartGroupRouter.Path("/entries").
		HandlerFunc(impl.ChartGroupRestHandler.SaveChartGroupEntries).Methods("PUT")
	chartGroupRouter.Path("/list").
		HandlerFunc(impl.ChartGroupRestHandler.GetChartGroupList).Methods("GET")
	chartGroupRouter.Path("/{chartGroupId}").
		HandlerFunc(impl.ChartGroupRestHandler.GetChartGroupWithChartMetaData).Methods("GET")
	chartGroupRouter.Path("/installation-detail/{chartGroupId}").
		HandlerFunc(impl.ChartGroupRestHandler.GetChartGroupInstallationDetail).Methods("GET")

	chartGroupRouter.Path("/list/min").
		HandlerFunc(impl.ChartGroupRestHandler.GetChartGroupListMin).Methods("GET")
	chartGroupRouter.Path("").
		HandlerFunc(impl.ChartGroupRestHandler.DeleteChartGroup).Methods("DELETE")
}
