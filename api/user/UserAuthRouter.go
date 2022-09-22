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

package user

import (
	"github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/authenticator/oidc"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type UserAuthRouter interface {
	InitUserAuthRouter(router *mux.Router)
}

type UserAuthRouterImpl struct {
	logger          *zap.SugaredLogger
	userAuthHandler UserAuthHandler
	dexProxy        func(writer http.ResponseWriter, request *http.Request)
	clientApp       *oidc.ClientApp
}

func NewUserAuthRouterImpl(logger *zap.SugaredLogger, userAuthHandler UserAuthHandler, selfRegistrationRolesService user.SelfRegistrationRolesService, dexConfig *client.DexConfig) (*UserAuthRouterImpl, error) {
	router := &UserAuthRouterImpl{
		userAuthHandler: userAuthHandler,
		logger:          logger,
	}
	logger.Infow("auth starting with dex conf", "conf", dexConfig)
	oidcClient, dexProxy, err := client.GetOidcClient(dexConfig, selfRegistrationRolesService.CheckAndCreateUserIfConfigured, router.RedirectUrlSanitiser)
	if err != nil {
		return nil, err
	}
	router.dexProxy = dexProxy
	router.clientApp = oidcClient
	return router, nil
}

// RedirectUrlSanitiser replaces initial "/orchestrator" from url
func (router UserAuthRouterImpl) RedirectUrlSanitiser(redirectUrl string) string {
	if strings.Contains(redirectUrl, argocdServer.Dashboard) {
		redirectUrl = strings.ReplaceAll(redirectUrl, argocdServer.Orchestrator, "")
	}
	return redirectUrl
}

func (router UserAuthRouterImpl) InitUserAuthRouter(userAuthRouter *mux.Router) {
	userAuthRouter.Path("/").
		HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			router.writeSuccess("Welcome @Devtron", writer)
		}).Methods("GET")

	userAuthRouter.PathPrefix("/api/dex").HandlerFunc(router.dexProxy)
	userAuthRouter.Path("/login").HandlerFunc(router.clientApp.HandleLogin)
	userAuthRouter.Path("/auth/login").HandlerFunc(router.clientApp.HandleLogin)
	userAuthRouter.Path("/auth/callback").HandlerFunc(router.clientApp.HandleCallback)
	userAuthRouter.Path("/api/v1/session").HandlerFunc(router.userAuthHandler.LoginHandler)
	userAuthRouter.Path("/refresh").HandlerFunc(router.userAuthHandler.RefreshTokenHandler)
	// Policies mapping in orchestrator
	userAuthRouter.Path("/admin/policy/default").
		Queries("team", "{team}", "app", "{app}", "env", "{env}").
		HandlerFunc(router.userAuthHandler.AddDefaultPolicyAndRoles).Methods("POST")
	userAuthRouter.Path("/devtron/auth/verify").
		HandlerFunc(router.userAuthHandler.AuthVerification).Methods("GET")
}

func (router UserAuthRouterImpl) writeSuccess(message string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(message))
	if err != nil {
		router.logger.Error(err)
	}
}
