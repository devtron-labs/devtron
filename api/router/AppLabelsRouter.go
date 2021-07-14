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

type AppLabelRouter interface {
	initLabelRouter(router *mux.Router)
}

type AppLabelRouterImpl struct {
	logger  *zap.SugaredLogger
	handler restHandler.AppLabelRestHandler
}

func NewAppLabelRouterImpl(logger *zap.SugaredLogger, handler restHandler.AppLabelRestHandler) *AppLabelRouterImpl {
	router := &AppLabelRouterImpl{
		logger:  logger,
		handler: handler,
	}
	return router
}

func (router AppLabelRouterImpl) initLabelRouter(appLabelsRouter *mux.Router) {
	appLabelsRouter.Path("/labels/list").
		HandlerFunc(router.handler.GetAllLabels).Methods("GET")
	appLabelsRouter.Path("/meta/info/{appId}").
		HandlerFunc(router.handler.GetAppMetaInfo).Methods("GET")
	appLabelsRouter.Path("/labels").
		HandlerFunc(router.handler.UpdateLabelsInApp).Methods("POST")
}
