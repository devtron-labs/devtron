package apiToken

import (
	"github.com/devtron-labs/devtron/api/logger"
	"github.com/gorilla/mux"
)

type ApiTokenRouter interface {
	InitApiTokenRouter(configRouter *mux.Router)
}

type ApiTokenRouterImpl struct {
	apiTokenRestHandler ApiTokenRestHandler
	userAuth            logger.UserAuth
}

func NewApiTokenRouterImpl(apiTokenRestHandler ApiTokenRestHandler, userAuth logger.UserAuth) *ApiTokenRouterImpl {
	return &ApiTokenRouterImpl{apiTokenRestHandler: apiTokenRestHandler, userAuth: userAuth}
}

func (impl ApiTokenRouterImpl) InitApiTokenRouter(configRouter *mux.Router) {
	configRouter.Use(impl.userAuth.LoggingMiddleware)
	configRouter.Path("").HandlerFunc(impl.apiTokenRestHandler.GetAllApiTokens).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.apiTokenRestHandler.CreateApiToken).Methods("POST")
	configRouter.Path("/{id}").HandlerFunc(impl.apiTokenRestHandler.UpdateApiToken).Methods("PUT")
	configRouter.Path("/{id}").HandlerFunc(impl.apiTokenRestHandler.DeleteApiToken).Methods("DELETE")
	configRouter.Path("/webhook").HandlerFunc(impl.apiTokenRestHandler.GetAllApiTokensForWebhook).Methods("GET")
}
