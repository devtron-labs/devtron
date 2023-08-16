package module

import (
	"github.com/devtron-labs/devtron/api/logger"
	"github.com/gorilla/mux"
)

type ModuleRouter interface {
	Init(configRouter *mux.Router)
}

type ModuleRouterImpl struct {
	moduleRestHandler ModuleRestHandler
	userAuth          logger.UserAuth
}

func NewModuleRouterImpl(moduleRestHandler ModuleRestHandler, userAuth logger.UserAuth) *ModuleRouterImpl {
	return &ModuleRouterImpl{moduleRestHandler: moduleRestHandler, userAuth: userAuth}
}

func (impl ModuleRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Use(impl.userAuth.LoggingMiddleware)
	configRouter.Path("").HandlerFunc(impl.moduleRestHandler.GetModuleInfo).Queries("name", "{name}").Methods("GET")
	configRouter.Path("").HandlerFunc(impl.moduleRestHandler.GetModuleInfo).Methods("GET")
	configRouter.Path("/config").HandlerFunc(impl.moduleRestHandler.GetModuleConfig).Queries("name", "{name}").Methods("GET")
	configRouter.Path("").HandlerFunc(impl.moduleRestHandler.HandleModuleAction).Queries("name", "{name}").Methods("POST")
	configRouter.Path("/enable").HandlerFunc(impl.moduleRestHandler.EnableModule).Queries("name", "{name}").Methods("POST")
}
