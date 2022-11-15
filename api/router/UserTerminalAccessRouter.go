package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type UserTerminalAccessRouter interface {
	InitTerminalAccessRouter(userTerminalAccessRouter *mux.Router)
}

type UserTerminalAccessRouterImpl struct {
	userTerminalAccessRestHandler restHandler.UserTerminalAccessRestHandler
}

func NewUserTerminalAccessRouterImpl(userTerminalAccessRestHandler restHandler.UserTerminalAccessRestHandler) *UserTerminalAccessRouterImpl {
	return &UserTerminalAccessRouterImpl{
		userTerminalAccessRestHandler: userTerminalAccessRestHandler,
	}
}

func (router UserTerminalAccessRouterImpl) InitTerminalAccessRouter(userTerminalAccessRouter *mux.Router) {
	userTerminalAccessRouter.Path("/update").
		HandlerFunc(router.userTerminalAccessRestHandler.UpdateTerminalSession).Methods("POST")
	userTerminalAccessRouter.Path("/update/shell").
		HandlerFunc(router.userTerminalAccessRestHandler.UpdateTerminalShellSession).Methods("POST")
	userTerminalAccessRouter.Path("/start").
		HandlerFunc(router.userTerminalAccessRestHandler.StartTerminalSession).Methods("PUT")
	userTerminalAccessRouter.Path("/get").
		HandlerFunc(router.userTerminalAccessRestHandler.FetchTerminalStatus).Queries("terminalAccessId", "{terminalAccessId}").Methods("GET")
	userTerminalAccessRouter.Path("/disconnect").
		HandlerFunc(router.userTerminalAccessRestHandler.DisconnectTerminalSession).Queries("terminalAccessId", "{terminalAccessId}").Methods("POST")
	userTerminalAccessRouter.Path("/stop").
		HandlerFunc(router.userTerminalAccessRestHandler.StopTerminalSession).Queries("terminalAccessId", "{terminalAccessId}").Methods("POST")
	userTerminalAccessRouter.Path("/disconnectAndRetry").
		HandlerFunc(router.userTerminalAccessRestHandler.DisconnectAllTerminalSessionAndRetry).Methods("POST")
}
