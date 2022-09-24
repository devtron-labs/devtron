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
	"go.uber.org/zap"
)

type AppRouter interface {
	initAppRouter(router *mux.Router)
}

type AppRouterImpl struct {
	logger  *zap.SugaredLogger
	handler restHandler.AppRestHandlerHandler
}

func NewAppRouterImpl(logger *zap.SugaredLogger, handler restHandler.AppRestHandlerHandler) *AppRouterImpl {
	router := &AppRouterImpl{
		logger:  logger,
		handler: handler,
	}
	return router
}

func (router AppRouterImpl) initAppRouter(appRouter *mux.Router) {
	appRouter.Path("/labels/list").
		HandlerFunc(router.handler.GetAllLabels).Methods("GET")
	appRouter.Path("/meta/info/{appId}").
		HandlerFunc(router.handler.GetAppMetaInfo).Methods("GET")
	appRouter.Path("/edit").
		HandlerFunc(router.handler.UpdateApp).Methods("POST")
	appRouter.Path("/edit/projects").
		HandlerFunc(router.handler.UpdateProjectForApps).Methods("POST")
}
