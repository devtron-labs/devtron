/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package sso

import (
	"github.com/gorilla/mux"
)

type SsoLoginRouter interface {
	InitSsoLoginRouter(router *mux.Router)
}

type SsoLoginRouterImpl struct {
	handler SsoLoginRestHandler
}

func NewSsoLoginRouterImpl(handler SsoLoginRestHandler) *SsoLoginRouterImpl {
	router := &SsoLoginRouterImpl{
		handler: handler,
	}
	return router
}

func (router SsoLoginRouterImpl) InitSsoLoginRouter(userAuthRouter *mux.Router) {
	userAuthRouter.Path("/create").
		HandlerFunc(router.handler.CreateSSOLoginConfig).Methods("POST")
	userAuthRouter.Path("/update").
		HandlerFunc(router.handler.UpdateSSOLoginConfig).Methods("PUT")
	userAuthRouter.Path("/list").
		HandlerFunc(router.handler.GetAllSSOLoginConfig).Methods("GET")
	userAuthRouter.Path("/{id}").
		HandlerFunc(router.handler.GetSSOLoginConfig).Methods("GET")
	userAuthRouter.Path("").Methods("GET").
		Queries("name", "{name}").HandlerFunc(router.handler.GetSSOLoginConfigByName)
}
