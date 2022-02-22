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

type PluginRouter interface {
	initPluginRouter(helmRouter *mux.Router)
}

type PluginRouterImpl struct {
	pluginRestHandler restHandler.PluginRestHandler
}

func NewPluginRouterImpl(pluginRestHandler restHandler.PluginRestHandler) *PluginRouterImpl {
	router := &PluginRouterImpl{
		pluginRestHandler: pluginRestHandler,
	}
	return router
}

func (router PluginRouterImpl) initPluginRouter(pluginRouter *mux.Router) {

	pluginRouter.Path("/plugin").
		HandlerFunc(router.pluginRestHandler.SavePlugin).
		Methods("POST")

	pluginRouter.Path("/plugin").
		HandlerFunc(router.pluginRestHandler.UpdatePlugin).
		Methods("PUT")

	pluginRouter.Path("/plugin/{Id}").
		HandlerFunc(router.pluginRestHandler.DeletePlugin).
		Methods("DELETE")

	pluginRouter.Path("/plugin/{Id}").
		HandlerFunc(router.pluginRestHandler.FindByPlugin).
		Methods("GET")
}
