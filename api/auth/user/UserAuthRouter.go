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
	"net/http"

	"github.com/devtron-labs/devtron/pkg/auth/authentication"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type UserAuthRouter interface {
	InitUserAuthRouter(router *mux.Router)
}

type UserAuthRouterImpl struct {
	logger             *zap.SugaredLogger
	userAuthHandler    UserAuthHandler
	userAuthOidcHelper authentication.UserAuthOidcHelper
}

func NewUserAuthRouterImpl(logger *zap.SugaredLogger, userAuthHandler UserAuthHandler, userAuthOidcHelper authentication.UserAuthOidcHelper) *UserAuthRouterImpl {
	router := &UserAuthRouterImpl{
		logger:             logger,
		userAuthHandler:    userAuthHandler,
		userAuthOidcHelper: userAuthOidcHelper,
	}
	return router
}

func (router UserAuthRouterImpl) InitUserAuthRouter(userAuthRouter *mux.Router) {
	userAuthRouter.Path("/").
		HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			router.writeSuccess("Welcome @Devtron", writer)
		}).Methods("GET")

	userAuthRouter.PathPrefix("/api/dex").HandlerFunc(router.userAuthOidcHelper.GetDexProxy())
	userAuthRouter.Path("/login").HandlerFunc(router.userAuthOidcHelper.GetClientApp().HandleLogin)
	userAuthRouter.Path("/auth/login").HandlerFunc(router.userAuthOidcHelper.GetClientApp().HandleLogin)
	userAuthRouter.Path("/auth/callback").HandlerFunc(router.userAuthOidcHelper.GetClientApp().HandleCallback)

	userAuthRouter.Path("/api/v1/session").HandlerFunc(router.userAuthHandler.LoginHandler)
	userAuthRouter.Path("/refresh").HandlerFunc(router.userAuthHandler.RefreshTokenHandler)
	// Policies mapping in orchestrator
	userAuthRouter.Path("/admin/policy/default").
		Queries("team", "{team}", "app", "{app}", "env", "{env}").
		HandlerFunc(router.userAuthHandler.AddDefaultPolicyAndRoles).Methods("POST")
	userAuthRouter.Path("/devtron/auth/verify").
		HandlerFunc(router.userAuthHandler.AuthVerification).Methods("GET")
	userAuthRouter.Path("/devtron/auth/verify/v2").
		HandlerFunc(router.userAuthHandler.AuthVerificationV2).Methods("GET")
}

func (router UserAuthRouterImpl) writeSuccess(message string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(message))
	if err != nil {
		router.logger.Error(err)
	}
}
