package apiToken

import (
	"github.com/gorilla/mux"
)

type ApiTokenRouter interface {
	InitApiTokenRouter(configRouter *mux.Router)
}

type ApiTokenRouterImpl struct {
	apiTokenRestHandler ApiTokenRestHandler
}

func NewApiTokenRouterImpl(apiTokenRestHandler ApiTokenRestHandler) *ApiTokenRouterImpl {
	return &ApiTokenRouterImpl{apiTokenRestHandler: apiTokenRestHandler}
}

func (router *ApiTokenRouterImpl) InitApiTokenRouter(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(router.apiTokenRestHandler.GetAllApiTokens).Methods("GET")
	configRouter.Path("").HandlerFunc(router.apiTokenRestHandler.CreateApiToken).Methods("POST")
	configRouter.Path("/{id}").HandlerFunc(router.apiTokenRestHandler.UpdateApiToken).Methods("PUT")
	configRouter.Path("/{id}").HandlerFunc(router.apiTokenRestHandler.DeleteApiToken).Methods("DELETE")
	configRouter.Path("/webhook").HandlerFunc(router.apiTokenRestHandler.GetAllApiTokensForWebhook).Methods("GET")
}
