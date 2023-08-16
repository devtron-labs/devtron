package server

import (
	"github.com/devtron-labs/devtron/api/logger"
	"github.com/gorilla/mux"
)

type ServerRouter interface {
	Init(configRouter *mux.Router)
}

type ServerRouterImpl struct {
	serverRestHandler ServerRestHandler
	userAuth          logger.UserAuth
}

func NewServerRouterImpl(serverRestHandler ServerRestHandler, userAuth logger.UserAuth) *ServerRouterImpl {
	return &ServerRouterImpl{serverRestHandler: serverRestHandler, userAuth: userAuth}
}

func (impl ServerRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Use(impl.userAuth.LoggingMiddleware)
	configRouter.Path("").HandlerFunc(impl.serverRestHandler.GetServerInfo).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.serverRestHandler.HandleServerAction).Methods("POST")
}
