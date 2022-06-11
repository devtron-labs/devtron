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

func (impl ApiTokenRouterImpl) InitApiTokenRouter(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(impl.apiTokenRestHandler.GetAllApiTokens).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.apiTokenRestHandler.CreateApiToken).Methods("POST")
	configRouter.Path("/{id}").HandlerFunc(impl.apiTokenRestHandler.UpdateApiToken).Methods("PUT")
	configRouter.Path("/{id}").HandlerFunc(impl.apiTokenRestHandler.DeleteApiToken).Methods("DELETE")
}
