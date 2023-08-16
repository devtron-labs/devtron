/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package router

import (
	"github.com/devtron-labs/devtron/api/logger"
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type CoreAppRouter interface {
	initCoreAppRouter(configRouter *mux.Router)
}

type CoreAppRouterImpl struct {
	restHandler restHandler.CoreAppRestHandler
	userAuth    logger.UserAuth
}

func NewCoreAppRouterImpl(restHandler restHandler.CoreAppRestHandler, userAuth logger.UserAuth) *CoreAppRouterImpl {
	return &CoreAppRouterImpl{restHandler: restHandler, userAuth: userAuth}

}

func (router CoreAppRouterImpl) initCoreAppRouter(configRouter *mux.Router) {
	configRouter.Use(router.userAuth.LoggingMiddleware)
	configRouter.Path("/v1beta1/application").HandlerFunc(router.restHandler.CreateApp).Methods("POST")
	configRouter.Path("/v1beta1/application/{appId}").HandlerFunc(router.restHandler.GetAppAllDetail).Methods("GET")
	configRouter.Path("/v1beta1/application/workflow").HandlerFunc(router.restHandler.CreateAppWorkflow).Methods("POST")
	configRouter.Path("/v1beta1/application/workflow/{appId}").HandlerFunc(router.restHandler.GetAppWorkflow).Methods("GET")
	configRouter.Path("/v1beta1/application/workflow/{appId}/sample").HandlerFunc(router.restHandler.GetAppWorkflowAndOverridesSample).Methods("GET")

}
