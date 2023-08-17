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

func (impl ServerRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(impl.serverRestHandler.GetServerInfo).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.serverRestHandler.HandleServerAction).Methods("POST")
}
