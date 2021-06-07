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
	"github.com/argoproj/argo-cd/util/settings"
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/pkg/dex"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type SsoLoginRouter interface {
	initSsoLoginRouter(router *mux.Router)
}

type SsoLoginRouterImpl struct {
	logger   *zap.SugaredLogger
	handler  restHandler.SsoLoginRestHandler
	cdProxy  func(writer http.ResponseWriter, request *http.Request)
	dexProxy func(writer http.ResponseWriter, request *http.Request)
}

func NewSsoLoginRouterImpl(logger *zap.SugaredLogger, handler restHandler.SsoLoginRestHandler, cdCfg *argocdServer.Config, dexCfg *dex.Config, settings *settings.ArgoCDSettings, userService user.UserService) *SsoLoginRouterImpl {
	tlsConfig := settings.TLSConfig()
	if tlsConfig != nil {
		tlsConfig.InsecureSkipVerify = true
	}
	router := &SsoLoginRouterImpl{
		handler: handler,
	}
	return router
}

func (router SsoLoginRouterImpl) initSsoLoginRouter(userAuthRouter *mux.Router) {
	userAuthRouter.Path("/create").
		HandlerFunc(router.handler.CreateSSOLoginConfig).Methods("POST")
	userAuthRouter.Path("/update").
		HandlerFunc(router.handler.UpdateSSOLoginConfig).Methods("PUT")
	userAuthRouter.Path("/list").
		HandlerFunc(router.handler.GetAllSSOLoginConfig).Methods("GET")
	userAuthRouter.Path("/{id}").
		HandlerFunc(router.handler.GetSSOLoginConfig).Methods("GET")
	userAuthRouter.Path("").Methods("GET").
		Queries("name", "{name}").HandlerFunc(router.handler.GetSSOLoginConfigByName)
}
