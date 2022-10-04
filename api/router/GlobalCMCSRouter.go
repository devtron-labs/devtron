package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type GlobalCMCSRouter interface {
	initGlobalCMCSRouter(configRouter *mux.Router)
}
type GlobalCMCSRouterImpl struct {
	restHandler restHandler.GlobalCMCSRestHandler
}

func NewGlobalCMCSRouterImpl(restHandler restHandler.GlobalCMCSRestHandler) *GlobalCMCSRouterImpl {
	return &GlobalCMCSRouterImpl{restHandler: restHandler}

}

func (router GlobalCMCSRouterImpl) initGlobalCMCSRouter(configRouter *mux.Router) {
	configRouter.Path("").
		HandlerFunc(router.restHandler.CreateGlobalCMCSConfig).Methods("POST")
}
