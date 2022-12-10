package module

import (
	"github.com/gorilla/mux"
)

type ModuleRouter interface {
	Init(configRouter *mux.Router)
}

type ModuleRouterImpl struct {
	moduleRestHandler ModuleRestHandler
}

func NewModuleRouterImpl(moduleRestHandler ModuleRestHandler) *ModuleRouterImpl {
	return &ModuleRouterImpl{moduleRestHandler: moduleRestHandler}
}

func (router *ModuleRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(router.moduleRestHandler.GetModuleInfo).Queries("name", "{name}").Methods("GET")
	configRouter.Path("/config").HandlerFunc(router.moduleRestHandler.GetModuleConfig).Queries("name", "{name}").Methods("GET")
	configRouter.Path("").HandlerFunc(router.moduleRestHandler.HandleModuleAction).Queries("name", "{name}").Methods("POST")
}
