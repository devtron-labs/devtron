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
	"net/http"
)

type CDRouter interface {
	initCDRouter(helmRouter *mux.Router)
}

type CDRouterImpl struct {
	logger        *zap.SugaredLogger
	cdRestHandler restHandler.CDRestHandler
}

func NewCDRouterImpl(logger *zap.SugaredLogger, cdRestHandler restHandler.CDRestHandler) *CDRouterImpl {
	router := &CDRouterImpl{
		logger:        logger,
		cdRestHandler: cdRestHandler,
	}
	return router
}

func (router CDRouterImpl) initCDRouter(cdRouter *mux.Router) {
	cdRouter.Path("/").
		HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			router.writeSuccess("Welcome @Devtron", writer)
		}).Methods("GET")

	cdRouter.Path("/tree/{app-name}").
		HandlerFunc(router.cdRestHandler.FetchResourceTree).
		Methods("GET")

	cdRouter.Path("/logs/{app-name}/pods/{pod-name}").
		HandlerFunc(router.cdRestHandler.FetchPodContainerLogs).
		Methods("GET")
}

func (router CDRouterImpl) writeSuccess(message string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(message))
	if err != nil {
		router.logger.Error(err)
	}
}
