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

package auth

import (
	"github.com/devtron-labs/authenticator/client"
	authMiddleware "github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/authenticator/oidc"
	"github.com/devtron-labs/devtron/client/argocdServer/connection"
	"github.com/devtron-labs/devtron/pkg/user"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type UserAuthOidcHelper interface {
	GetClientApp() *oidc.ClientApp
	GetDexProxy() func(writer http.ResponseWriter, request *http.Request)
	UpdateInMemoryDataOnSsoAddUpdate(ssoUrl string) error
}

type UserAuthOidcHelperImpl struct {
	logger                       *zap.SugaredLogger
	selfRegistrationRolesService user.SelfRegistrationRolesService
	dexProxy                     func(writer http.ResponseWriter, request *http.Request)
	clientApp                    *oidc.ClientApp
	dexConfig                    *client.DexConfig
	settings                     *oidc.Settings
	sessionManager               *authMiddleware.SessionManager
}

func NewUserAuthOidcHelperImpl(logger *zap.SugaredLogger, selfRegistrationRolesService user.SelfRegistrationRolesService, dexConfig *client.DexConfig,
	settings *oidc.Settings, sessionManager *authMiddleware.SessionManager) (*UserAuthOidcHelperImpl, error) {
	impl := &UserAuthOidcHelperImpl{
		logger:                       logger,
		settings:                     settings,
		dexConfig:                    dexConfig,
		sessionManager:               sessionManager,
		selfRegistrationRolesService: selfRegistrationRolesService,
	}
	logger.Infow("auth starting with dex conf", "conf", dexConfig)
	oidcClient, dexProxy, err := client.GetOidcClient(dexConfig, selfRegistrationRolesService.CheckAndCreateUserIfConfigured, impl.sanitiseRedirectUrl)
	if err != nil {
		logger.Errorw("error in getting oidc client", "err", err)
		return nil, err
	}
	impl.dexProxy = dexProxy
	impl.clientApp = oidcClient
	return impl, nil
}

// SanitiseRedirectUrl replaces initial "/orchestrator" from url
func (impl UserAuthOidcHelperImpl) sanitiseRedirectUrl(redirectUrl string) string {
	if strings.Contains(redirectUrl, connection.Dashboard) {
		redirectUrl = strings.ReplaceAll(redirectUrl, connection.Orchestrator, "")
	}
	return redirectUrl
}

func (impl UserAuthOidcHelperImpl) GetClientApp() *oidc.ClientApp {
	return impl.clientApp
}

func (impl UserAuthOidcHelperImpl) GetDexProxy() func(writer http.ResponseWriter, request *http.Request) {
	return impl.dexProxy
}

func (impl UserAuthOidcHelperImpl) UpdateInMemoryDataOnSsoAddUpdate(ssoUrl string) error {
	impl.logger.Infow("updating in memory data on sso update", "ssoUrl", ssoUrl)

	// set url in dexConfig
	impl.dexConfig.Url = ssoUrl
	proxyUrl, err := impl.dexConfig.GetDexProxyUrl()
	if err != nil {
		impl.logger.Errorw("error in getting proxy url from ssoUrl", "err", err, "ssoUrl", ssoUrl)
		return err
	}

	// update url in oidc settings
	impl.settings.URL = ssoUrl
	impl.settings.OIDCConfig.Issuer = proxyUrl

	// update session manager oidc settings
	impl.sessionManager.UpdateSettings(impl.settings, impl.dexConfig)

	// get oidc client
	oidcClient, _, err := client.GetOidcClient(impl.dexConfig, impl.selfRegistrationRolesService.CheckAndCreateUserIfConfigured, impl.sanitiseRedirectUrl)
	if err != nil {
		impl.logger.Errorw("error in getting oidc client", "err", err, "ssoUrl", ssoUrl)
		return err
	}

	// update client app config
	impl.clientApp.UpdateConfig(oidcClient)
	return nil
}
