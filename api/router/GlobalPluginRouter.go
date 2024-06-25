/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type GlobalPluginRouter interface {
	initGlobalPluginRouter(globalPluginRouter *mux.Router)
}

func NewGlobalPluginRouter(logger *zap.SugaredLogger, globalPluginRestHandler restHandler.GlobalPluginRestHandler) *GlobalPluginRouterImpl {
	return &GlobalPluginRouterImpl{
		logger:                  logger,
		globalPluginRestHandler: globalPluginRestHandler,
	}
}

type GlobalPluginRouterImpl struct {
	logger                  *zap.SugaredLogger
	globalPluginRestHandler restHandler.GlobalPluginRestHandler
}

func (impl *GlobalPluginRouterImpl) initGlobalPluginRouter(globalPluginRouter *mux.Router) {
	globalPluginRouter.Path("").
		HandlerFunc(impl.globalPluginRestHandler.PatchPlugin).Methods("POST")
	globalPluginRouter.Path("/detail/all").
		HandlerFunc(impl.globalPluginRestHandler.GetAllDetailedPluginInfo).Methods("GET")
	globalPluginRouter.Path("/detail/{pluginId}").
		HandlerFunc(impl.globalPluginRestHandler.GetDetailedPluginInfoByPluginId).Methods("GET")

	globalPluginRouter.Path("/list/global-variable").
		HandlerFunc(impl.globalPluginRestHandler.GetAllGlobalVariables).Methods("GET")

	globalPluginRouter.Path("/list").
		HandlerFunc(impl.globalPluginRestHandler.ListAllPlugins).Methods("GET")

	globalPluginRouter.Path("/{pluginId}").
		HandlerFunc(impl.globalPluginRestHandler.GetPluginDetailById).Methods("GET")
}
