package server

import (
	"github.com/gorilla/mux"
)

type ServerRouter interface {
	Init(configRouter *mux.Router)
}

type ServerRouterImpl struct {
	serverRestHandler ServerRestHandler
}

func NewServerRouterImpl(serverRestHandler ServerRestHandler) *ServerRouterImpl {
	return &ServerRouterImpl{serverRestHandler: serverRestHandler}
}

func (router *ServerRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(router.serverRestHandler.GetServerInfo).Methods("GET")
	configRouter.Path("").HandlerFunc(router.serverRestHandler.HandleServerAction).Methods("POST")
}
