package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type HistoryRouter interface {
	InitHistoryRouter(historyRouter *mux.Router)
}
type HistoryRouterImpl struct {
	historyRestHandler restHandler.HistoryRestHandler
}

func NewHistoryRouterImpl(historyRestHandler restHandler.HistoryRestHandler) *HistoryRouterImpl {
	return &HistoryRouterImpl{historyRestHandler: historyRestHandler}
}
func (impl HistoryRouterImpl) InitHistoryRouter(historyRouter *mux.Router) {

	historyRouter.Path("/cm-cs/{pipelineId}").
		HandlerFunc(impl.historyRestHandler.FetchDeployedCmCsHistory).
		Methods("GET")

	historyRouter.Path("/charts/{pipelineId}").
		HandlerFunc(impl.historyRestHandler.FetchDeployedChartsHistory).
		Methods("GET")

	historyRouter.Path("/strategy/{pipelineId}").
		HandlerFunc(impl.historyRestHandler.FetchDeployedStrategyHistory).
		Methods("GET")

	historyRouter.Path("/cd-config/{pipelineId}").
		HandlerFunc(impl.historyRestHandler.FetchDeployedStrategyHistory).
		Methods("GET")
}
