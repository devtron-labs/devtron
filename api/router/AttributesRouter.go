/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type AttributesRouter interface {
	InitAttributesRouter(helmRouter *mux.Router)
}

type AttributesRouterImpl struct {
	attributesRestHandler restHandler.AttributesRestHandler
}

func NewAttributesRouterImpl(attributesRestHandler restHandler.AttributesRestHandler) *AttributesRouterImpl {
	router := &AttributesRouterImpl{
		attributesRestHandler: attributesRestHandler,
	}
	return router
}

func (router AttributesRouterImpl) InitAttributesRouter(attributesRouter *mux.Router) {
	attributesRouter.Path("/create").
		HandlerFunc(router.attributesRestHandler.AddAttributes).Methods("POST")
	attributesRouter.Path("/update").
		HandlerFunc(router.attributesRestHandler.UpdateAttributes).Methods("PUT")
	attributesRouter.Path("").Queries("key", "{key}").
		HandlerFunc(router.attributesRestHandler.GetAttributesByKey).Methods("GET")
	attributesRouter.Path("/{id}").
		HandlerFunc(router.attributesRestHandler.GetAttributesById).Methods("GET")
	attributesRouter.Path("/active/list").
		HandlerFunc(router.attributesRestHandler.GetAttributesActiveList).Methods("GET")
	attributesRouter.Path("/create/deploymentConfig").
		HandlerFunc(router.attributesRestHandler.AddDeploymentEnforcementConfig).Methods("POST")
}
