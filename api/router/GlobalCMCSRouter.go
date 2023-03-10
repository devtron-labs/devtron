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
	configRouter.Path("/{config_name}/{config_type}").
		HandlerFunc(router.restHandler.GetGlobalCMCSDataByConfigTypeAndName).Methods("GET")
	configRouter.Path("/all").
		HandlerFunc(router.restHandler.GetAllGlobalCMCSData).Methods("GET")
	configRouter.Path("").
		HandlerFunc(router.restHandler.CreateGlobalCMCSConfig).Methods("POST")
	configRouter.Path("/update").
		HandlerFunc(router.restHandler.UpdateGlobalCMCSDataById).Methods("PUT")
	configRouter.Path("/delete/{id}").
		HandlerFunc(router.restHandler.DeleteByID).Methods("DELETE")
}
