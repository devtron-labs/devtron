/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	user "github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type UserAttributesRouter interface {
	InitUserAttributesRouter(helmRouter *mux.Router)
}

type UserAttributesRouterImpl struct {
	userAttributesRestHandler user.UserAttributesRestHandler
}

func NewUserAttributesRouterImpl(userAttributesRestHandler user.UserAttributesRestHandler) *UserAttributesRouterImpl {
	router := &UserAttributesRouterImpl{
		userAttributesRestHandler: userAttributesRestHandler,
	}
	return router
}

func (router UserAttributesRouterImpl) InitUserAttributesRouter(attributesRouter *mux.Router) {
	attributesRouter.Path("/update").
		HandlerFunc(router.userAttributesRestHandler.UpdateUserAttributes).Methods("POST")
	attributesRouter.Path("/get").
		HandlerFunc(router.userAttributesRestHandler.GetUserAttribute).Queries("key", "{key}").Methods("GET")
}
