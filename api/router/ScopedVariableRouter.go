package router

import (
	"github.com/devtron-labs/devtron/api/restHandler/scopedVariable"
	"github.com/gorilla/mux"
)

type ScopedVariableRouter interface {
	InitScopedVariableRouter(router *mux.Router)
}

type ScopedVariableRouterImpl struct {
	scopedVariableRestHandler scopedVariable.ScopedVariableRestHandler
}

func NewScopedVariableRouterImpl(scopedVariableRestHandler scopedVariable.ScopedVariableRestHandler) *ScopedVariableRouterImpl {
	router := &ScopedVariableRouterImpl{
		scopedVariableRestHandler: scopedVariableRestHandler,
	}
	return router
}

func (impl ScopedVariableRouterImpl) InitScopedVariableRouter(router *mux.Router) {
	router.Path("/variables").
		HandlerFunc(impl.scopedVariableRestHandler.CreateVariables).
		Methods("POST")
	router.Path("/variables").
		HandlerFunc(impl.scopedVariableRestHandler.GetScopedVariables).
		Methods("GET")

}
