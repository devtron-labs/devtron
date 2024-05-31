/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type CoreAppRouter interface {
	initCoreAppRouter(configRouter *mux.Router)
}

type CoreAppRouterImpl struct {
	restHandler restHandler.CoreAppRestHandler
}

func NewCoreAppRouterImpl(restHandler restHandler.CoreAppRestHandler) *CoreAppRouterImpl {
	return &CoreAppRouterImpl{restHandler: restHandler}

}

func (router CoreAppRouterImpl) initCoreAppRouter(configRouter *mux.Router) {
	configRouter.Path("/v1beta1/application").HandlerFunc(router.restHandler.CreateApp).Methods("POST")
	configRouter.Path("/v1beta1/application/{appId}").HandlerFunc(router.restHandler.GetAppAllDetail).Methods("GET")
	configRouter.Path("/v1beta1/application/workflow").HandlerFunc(router.restHandler.CreateAppWorkflow).Methods("POST")
	configRouter.Path("/v1beta1/application/workflow/{appId}").HandlerFunc(router.restHandler.GetAppWorkflow).Methods("GET")
	configRouter.Path("/v1beta1/application/workflow/{appId}/sample").HandlerFunc(router.restHandler.GetAppWorkflowAndOverridesSample).Methods("GET")

}
