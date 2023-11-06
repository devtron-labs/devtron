package devtronResource

import "github.com/gorilla/mux"

type DevtronResourceRouter interface {
	InitDevtronResourceRouter(devtronResourceRouter *mux.Router)
}

type DevtronResourceRouterImpl struct {
	devtronResourceRestHandler DevtronResourceRestHandler
}

func NewDevtronResourceRouterImpl(devtronResourceRestHandler DevtronResourceRestHandler) *DevtronResourceRouterImpl {
	return &DevtronResourceRouterImpl{devtronResourceRestHandler: devtronResourceRestHandler}
}

func (router *DevtronResourceRouterImpl) InitDevtronResourceRouter(devtronResourceRouter *mux.Router) {
	devtronResourceRouter.Path("/list").
		HandlerFunc(router.devtronResourceRestHandler.GetResourceList).Methods("GET")

	//regex in path allows to have sub-kinds, for ex - "/applications/devtron-apps/v1" & "/cluster/v1" both will be accepted
	devtronResourceRouter.Path("/{kind:[a-zA-Z0-9/-]+}/{version:[a-zA-Z0-9]+}").
		HandlerFunc(router.devtronResourceRestHandler.GetResourceObject).Methods("GET")

	devtronResourceRouter.Path("/{kind:[a-zA-Z0-9/-]+}/{version:[a-zA-Z0-9]+}").
		HandlerFunc(router.devtronResourceRestHandler.UpdateResourceObject).Methods("PUT")
}
