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
	"fmt"
	"github.com/argoproj/argo-cd/util/settings"
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

type UserAuthRouter interface {
	initUserAuthRouter(helmRouter *mux.Router)
}

type UserAuthRouterImpl struct {
	logger          *zap.SugaredLogger
	userAuthHandler restHandler.UserAuthHandler
	cdProxy         func(writer http.ResponseWriter, request *http.Request)
}

func NewUserAuthRouterImpl(logger *zap.SugaredLogger, userAuthHandler restHandler.UserAuthHandler, cdCfg *argocdServer.Config, settings *settings.ArgoCDSettings, userService user.UserService) *UserAuthRouterImpl {
	tlsConfig := settings.TLSConfig()
	if tlsConfig != nil {
		tlsConfig.InsecureSkipVerify = true
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	cdProxy := argocdServer.NewCDHTTPReverseProxy(fmt.Sprintf("https://%s:%s", cdCfg.Host, cdCfg.Port), client.Transport, userService.GetUserByToken)
	router := &UserAuthRouterImpl{
		userAuthHandler: userAuthHandler,
		cdProxy:         cdProxy,
		logger:          logger,
	}
	return router
}

func (router UserAuthRouterImpl) initUserAuthRouter(userAuthRouter *mux.Router) {
	userAuthRouter.Path("/").
		HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			router.writeSuccess("Welcome @Devtron", writer)
		}).Methods("GET")

	userAuthRouter.Path("/login").HandlerFunc(router.cdProxy)
	userAuthRouter.Path("/auth/login").HandlerFunc(router.cdProxy)
	userAuthRouter.Path("/auth/callback").HandlerFunc(router.cdProxy)

	userAuthRouter.Path("/api/v1/session").HandlerFunc(router.userAuthHandler.LoginHandler)
	userAuthRouter.Path("/orchestrator/api/dex/auth").HandlerFunc(router.cdProxy)
	userAuthRouter.Path("/orchestrator/api/dex/auth/google").HandlerFunc(router.cdProxy)
	userAuthRouter.Path("/orchestrator/api/dex/approval").HandlerFunc(router.cdProxy)
	userAuthRouter.Path("/orchestrator/api/dex/callback").HandlerFunc(router.cdProxy)
	userAuthRouter.Path("/refresh").HandlerFunc(router.userAuthHandler.RefreshTokenHandler)

	// Policies Setup
	userAuthRouter.Path("/admin/policy").
		HandlerFunc(router.userAuthHandler.AddPolicy).Methods("POST")
	userAuthRouter.Path("/admin/policy").
		HandlerFunc(router.userAuthHandler.RemovePolicy).Methods("DELETE")
	userAuthRouter.Path("/admin/policy/subject").
		HandlerFunc(router.userAuthHandler.GetAllSubjectsFromCasbin).Methods("GET")
	userAuthRouter.Path("/admin/policy/roles").
		Queries("user", "{user}").
		HandlerFunc(router.userAuthHandler.GetRolesForUserFromCasbin).Methods("GET")
	userAuthRouter.Path("/admin/policy/users").
		Queries("role", "{role}").
		HandlerFunc(router.userAuthHandler.GetUserByRoleFromCasbin).Methods("GET")
	userAuthRouter.Path("/admin/policy/subject").
		Queries("user", "{user}").
		Queries("role", "{role}").
		HandlerFunc(router.userAuthHandler.DeleteRoleForUserFromCasbin).Methods("DELETE")

	// Policies mapping in orchestrator
	userAuthRouter.Path("/admin/role").
		HandlerFunc(router.userAuthHandler.CreateRoleFromOrchestrator).Methods("POST")
	userAuthRouter.Path("/admin/role").
		HandlerFunc(router.userAuthHandler.UpdateRoleFromOrchestrator).Methods("PUT")
	userAuthRouter.Path("/admin/role/user/{userId}").
		HandlerFunc(router.userAuthHandler.GetRolesByUserIdFromOrchestrator).Methods("GET")
	userAuthRouter.Path("/admin/role").
		HandlerFunc(router.userAuthHandler.GetAllRoleFromOrchestrator).Methods("GET")
	userAuthRouter.Path("/admin/role/filter").
		Queries("team", "{team}", "app", "{app}", "env", "{env}", "act", "{act}").
		HandlerFunc(router.userAuthHandler.GetRoleByFilterFromOrchestrator).Methods("GET")
	userAuthRouter.Path("/admin/role").
		Queries("role", "{role}").
		HandlerFunc(router.userAuthHandler.DeleteRoleFromOrchestrator).Methods("DELETE")

	userAuthRouter.Path("/admin/policy/default").
		Queries("team", "{team}", "app", "{app}", "env", "{env}").
		HandlerFunc(router.userAuthHandler.AddDefaultPolicyAndRoles).Methods("POST")

	userAuthRouter.Path("/devtron/auth/verify").
		HandlerFunc(router.userAuthHandler.AuthVerification).Methods("GET")

	userAuthRouter.Path("/sso/create").
		HandlerFunc(router.userAuthHandler.CreateSSOLoginConfig).Methods("POST")
	userAuthRouter.Path("/sso/update").
		HandlerFunc(router.userAuthHandler.UpdateSSOLoginConfig).Methods("PUT")
	userAuthRouter.Path("/sso/list").
		HandlerFunc(router.userAuthHandler.GetAllSSOLoginConfig).Methods("GET")
	userAuthRouter.Path("/sso/{id}").
		HandlerFunc(router.userAuthHandler.GetSSOLoginConfig).Methods("GET")
}

func (router UserAuthRouterImpl) writeSuccess(message string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(message))
	if err != nil {
		router.logger.Error(err)
	}
}
