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
	"github.com/devtron-labs/devtron/api/devtron-integration-manager/api/rest-handler"
	"github.com/gorilla/mux"
)

type IntegrationManagerRouter interface {
	initAttributesRouter(helmRouter *mux.Router)
}

type IntegrationManagerRouterImpl struct {
	integrationManagerRestHandler restHandler.IntegrationManagerRestHandler
}

func NewIntegrationManagerRouterImpl(integrationManagerRestHandler restHandler.IntegrationManagerRestHandler) *IntegrationManagerRouterImpl {
	router := &IntegrationManagerRouterImpl{
		integrationManagerRestHandler: integrationManagerRestHandler,
	}
	return router
}

func (router IntegrationManagerRouterImpl) initIntegrationManagerRouter(integrationRouter *mux.Router) {
	integrationRouter.Path("/install").
		HandlerFunc(router.integrationManagerRestHandler.InstallModule).Methods("POST")
	integrationRouter.Path("/status").
		HandlerFunc(router.integrationManagerRestHandler.GetModulesStatus).Methods("POST")
	integrationRouter.Path("").
		HandlerFunc(router.integrationManagerRestHandler.GetAllModules).Methods("GET")
}
