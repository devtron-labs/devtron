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
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type ConfigMapRouter interface {
	initConfigMapRouter(configRouter *mux.Router)
}
type ConfigMapRouterImpl struct {
	restHandler restHandler.ConfigMapRestHandler
}

func NewConfigMapRouterImpl(restHandler restHandler.ConfigMapRestHandler) *ConfigMapRouterImpl {
	return &ConfigMapRouterImpl{restHandler: restHandler}

}

func (router ConfigMapRouterImpl) initConfigMapRouter(configRouter *mux.Router) {
	configRouter.Path("/global/cm").
		HandlerFunc(router.restHandler.CMGlobalAddUpdate).Methods("POST")
	configRouter.Path("/environment/cm").
		HandlerFunc(router.restHandler.CMEnvironmentAddUpdate).Methods("POST")

	configRouter.Path("/global/cm/{appId}").
		HandlerFunc(router.restHandler.CMGlobalFetch).Methods("GET")
	configRouter.Path("/environment/cm/{appId}/{envId}").
		HandlerFunc(router.restHandler.CMEnvironmentFetch).Methods("GET")

	configRouter.Path("/global/cs").
		HandlerFunc(router.restHandler.CSGlobalAddUpdate).Methods("POST")
	configRouter.Path("/environment/cs").
		HandlerFunc(router.restHandler.CSEnvironmentAddUpdate).Methods("POST")

	configRouter.Path("/global/cs/{appId}").
		HandlerFunc(router.restHandler.CSGlobalFetch).Methods("GET")
	configRouter.Path("/environment/cs/{appId}/{envId}").
		HandlerFunc(router.restHandler.CSEnvironmentFetch).Methods("GET")

	configRouter.Path("/global/cm/{appId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CMGlobalDelete).Methods("DELETE")
	configRouter.Path("/environment/cm/{appId}/{envId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CMEnvironmentDelete).Methods("DELETE")
	configRouter.Path("/global/cs/{appId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CSGlobalDelete).Methods("DELETE")
	configRouter.Path("/environment/cs/{appId}/{envId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CSEnvironmentDelete).Methods("DELETE")

	configRouter.Path("/global/cs/edit/{appId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CSGlobalFetchForEdit).Methods("GET")
	configRouter.Path("/environment/cs/edit/{appId}/{envId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CSEnvironmentFetchForEdit).Methods("GET")

	configRouter.Path("/bulk/patch").
		HandlerFunc(router.restHandler.ConfigSecretBulkPatch).Methods("POST")
}
