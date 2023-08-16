package router

import (
	"github.com/devtron-labs/devtron/api/logger"
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type GlobalCMCSRouter interface {
	initGlobalCMCSRouter(configRouter *mux.Router)
}
type GlobalCMCSRouterImpl struct {
	restHandler restHandler.GlobalCMCSRestHandler
	userAuth    logger.UserAuth
}

func NewGlobalCMCSRouterImpl(restHandler restHandler.GlobalCMCSRestHandler, userAuth logger.UserAuth) *GlobalCMCSRouterImpl {
	return &GlobalCMCSRouterImpl{restHandler: restHandler, userAuth: userAuth}

}

func (router GlobalCMCSRouterImpl) initGlobalCMCSRouter(configRouter *mux.Router) {
	configRouter.Use(router.userAuth.LoggingMiddleware)
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
