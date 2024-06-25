package devtronResource

import "github.com/gorilla/mux"

type DevtronResourceRouter interface {
	InitDevtronResourceRouter(devtronResourceRouter *mux.Router)
}

type DevtronResourceRouterImpl struct {
	historyRouter HistoryRouter
}

func NewDevtronResourceRouterImpl(historyRouter HistoryRouter) *DevtronResourceRouterImpl {
	return &DevtronResourceRouterImpl{
		historyRouter: historyRouter,
	}
}

func (router *DevtronResourceRouterImpl) InitDevtronResourceRouter(devtronResourceRouter *mux.Router) {
	historyRouter := devtronResourceRouter.PathPrefix("/history").Subrouter()
	router.historyRouter.InitDtResourceHistoryRouter(historyRouter)
}
