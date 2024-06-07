package devtronResource

import "github.com/gorilla/mux"

type HistoryRouter interface {
	InitDtResourceHistoryRouter(devtronResourceRouter *mux.Router)
}

type HistoryRouterImpl struct {
	dtResourceHistoryRestHandler HistoryRestHandler
}

func NewHistoryRouterImpl(dtResourceHistoryRestHandler HistoryRestHandler) *HistoryRouterImpl {
	return &HistoryRouterImpl{dtResourceHistoryRestHandler: dtResourceHistoryRestHandler}
}

func (router *HistoryRouterImpl) InitDtResourceHistoryRouter(historyRouter *mux.Router) {
	historyRouter.Path("/deployment/config/{kind:[a-zA-Z0-9/-]+}/{version:[a-zA-Z0-9]+}").
		HandlerFunc(router.dtResourceHistoryRestHandler.GetDeploymentHistoryConfigList).Methods("GET")

	historyRouter.Path("/deployment/{kind:[a-zA-Z0-9/-]+}/{version:[a-zA-Z0-9]+}").
		HandlerFunc(router.dtResourceHistoryRestHandler.GetDeploymentHistory).Methods("GET")
}
