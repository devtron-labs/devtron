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

func (impl ModuleRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(impl.moduleRestHandler.GetModuleInfo).Queries("name", "{name}").Methods("GET")
	configRouter.Path("").HandlerFunc(impl.moduleRestHandler.HandleModuleAction).Queries("name", "{name}").Methods("POST")
}
