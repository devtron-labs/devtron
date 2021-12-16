/*
 * Copyright (c) 2021 Devtron Labs
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

package client

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/authenticator/oidc"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"
)

func GetOidcClient(conf *DexConfig, userVerifier oidc.UserVerifier, RedirectUrlSanitiser oidc.RedirectUrlSanitiser) (*oidc.ClientApp, func(writer http.ResponseWriter, request *http.Request), error) {
	settings, err := GetSettings(conf)
	oidcClient, dexProxy, err := getOidcClient(conf.DexServerAddress, settings, userVerifier, RedirectUrlSanitiser)
	return oidcClient, dexProxy, err
}

func GetSettings(conf *DexConfig) (*oidc.Settings, error) {
	proxyUrl, err := conf.getDexProxyUrl()
	if err != nil {
		return nil, err
	}
	settings := &oidc.Settings{
		URL: conf.Url,
		OIDCConfig: oidc.OIDCConfig{CLIClientID: conf.DexClientID,
			ClientSecret: conf.DexClientSecret,
			Issuer:       proxyUrl,
			ServerSecret: conf.ServerSecret},
		UserSessionDuration: time.Duration(conf.UserSessionDurationSeconds) * time.Second,
		AdminPasswordMtime:  conf.AdminPasswordMtime,
	}
	return settings, nil
}
func getOidcClient(dexServerAddress string, settings *oidc.Settings, userVerifier oidc.UserVerifier, RedirectUrlSanitiser oidc.RedirectUrlSanitiser) (*oidc.ClientApp, func(writer http.ResponseWriter, request *http.Request), error) {
	dexClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: nil,
			Proxy:           http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	dexProxy := oidc.NewDexHTTPReverseProxy(dexServerAddress, dexClient.Transport)
	cahecStore := &oidc.Cache{OidcState: map[string]*oidc.OIDCState{}}
	oidcClient, err := oidc.NewClientApp(settings, cahecStore, "/", userVerifier, RedirectUrlSanitiser)
	if err != nil {
		return nil, nil, err
	}
	return oidcClient, dexProxy, err
}

const dexProxyUri = "api/dex"

type DexConfig struct {
	DexHost          string `env:"DEX_HOST" envDefault:"http://localhost"`
	DexPort          string `env:"DEX_PORT" envDefault:"5556"`
	DexClientID      string `env:"DEX_CLIENT_ID" envDefault:"argo-cd"`
	DexServerAddress string
	Url              string
	DexClientSecret  string
	ServerSecret     string
	// Specifies token expiration duration
	UserSessionDurationSeconds int       `env:"USER_SESSION_DURATION_SECONDS" envDefault:"86400"`
	AdminPasswordMtime         time.Time `json:"ADMIN_PASSWORD_MTIME"`
}

func (c *DexConfig) getDexProxyUrl() (string, error) {
	u, err := url.Parse(c.Url)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, dexProxyUri)
	s := u.String()
	return s, nil
}

func BuildDexConfig(k8sClient *K8sClient) (*DexConfig, error) {
	dexConfig, err := dexConfigConfigFromEnv()
	if err != nil {
		return nil, err
	}
	settings, err := k8sClient.GetServerSettings()
	if err != nil {
		return nil, err
	}
	dexConfig.Url = settings.Url
	dexConfig.ServerSecret = settings.ServerSecret
	clientSecret, err := generateDexClientSecret(dexConfig.ServerSecret)
	if err != nil {
		return nil, err
	}
	dexConfig.DexClientSecret = clientSecret
	dexConfig.DexServerAddress = fmt.Sprintf("%s:%s", dexConfig.DexHost, dexConfig.DexPort)
	return dexConfig, nil
}
func generateDexClientSecret(serverSecret string) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(serverSecret))
	if err != nil {
		return "", err
	}
	sha := h.Sum(nil)
	s := base64.URLEncoding.EncodeToString(sha)[:40]
	return s, nil
}

func dexConfigConfigFromEnv() (*DexConfig, error) {
	cfg := &DexConfig{}
	err := env.Parse(cfg)
	return cfg, err
}
