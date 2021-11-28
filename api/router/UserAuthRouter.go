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
	"github.com/devtron-labs/authenticator/oidc"
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/pkg/dex"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

type UserAuthRouter interface {
	initUserAuthRouter(router *mux.Router)
}

type UserAuthRouterImpl struct {
	logger          *zap.SugaredLogger
	userAuthHandler restHandler.UserAuthHandler
	cdProxy         func(writer http.ResponseWriter, request *http.Request)
	dexProxy        func(writer http.ResponseWriter, request *http.Request)
}

func NewUserAuthRouterImpl(logger *zap.SugaredLogger, userAuthHandler restHandler.UserAuthHandler, cdCfg *argocdServer.Config, dexCfg *dex.Config, settings *settings.ArgoCDSettings, userService user.UserService) *UserAuthRouterImpl {
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
	dexClient := &http.Client{
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
	dexProxy := argocdServer.NewDexHTTPReverseProxy(fmt.Sprintf("%s:%s", dexCfg.Host, dexCfg.Port), dexClient.Transport)
	cdProxy := argocdServer.NewCDHTTPReverseProxy(fmt.Sprintf("https://%s:%s", cdCfg.Host, cdCfg.Port), client.Transport, userService.GetUserByToken)
	router := &UserAuthRouterImpl{
		userAuthHandler: userAuthHandler,
		cdProxy:         cdProxy,
		dexProxy:        dexProxy,
		logger:          logger,
	}
	return router
}

func (router UserAuthRouterImpl) initUserAuthRouter(userAuthRouter *mux.Router) {
	dexServerAddress := fmt.Sprintf("%s:%s", "http://127.0.0.1", "5556")
	settings := &oidc.Settings{
		URL: "https://127.0.0.1:8000/",
		OIDCConfig: oidc.OIDCConfig{CLIClientID: "argo-cd",
			ClientSecret: "RmCryx_nTtzUcKzp0Vg0Uh4XsyM3YBdagWMgzmNJ",
			Issuer:       "https://127.0.0.1:8000/api/dex"},
	}
	oidcClient, dexProxy, err := oidc.GetOidcClient(dexServerAddress, settings)
	if err != nil {
		fmt.Println(err)
		return
	}
	//sesionManager := middleware.NewSessionManager(settings, dexServerAddress)

	userAuthRouter.Path("/").
		HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			router.writeSuccess("Welcome @Devtron", writer)
		}).Methods("GET")

	userAuthRouter.PathPrefix("/api/dex").HandlerFunc(dexProxy)
	userAuthRouter.Path("/login").HandlerFunc(oidcClient.HandleLogin)
	userAuthRouter.Path("/auth/login").HandlerFunc(oidcClient.HandleCallback)
	userAuthRouter.PathPrefix("/auth/callback").HandlerFunc(router.cdProxy)

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
