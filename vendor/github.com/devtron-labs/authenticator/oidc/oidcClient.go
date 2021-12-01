package oidc

import (
	"github.com/caarlos0/env/v6"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"
)

func GetOidcClient(conf *DexConfig, userVerifier UserVerifier, RedirectUrlSanitiser RedirectUrlSanitiser) (*ClientApp, func(writer http.ResponseWriter, request *http.Request), error) {
	settings, err := GetSettings(conf)
	oidcClient, dexProxy, err := getOidcClient(conf.DexServerAddress, settings, userVerifier, RedirectUrlSanitiser)
	return oidcClient, dexProxy, err
}

func GetSettings(conf *DexConfig) (*Settings, error) {
	proxyUrl, err := conf.getDexProxyUrl()
	if err != nil {
		return nil, err
	}
	settings := &Settings{
		URL: conf.Url,
		OIDCConfig: OIDCConfig{CLIClientID: conf.DexClientID,
			ClientSecret: conf.DexClientSecret,
			Issuer:       proxyUrl},
		UserSessionDuration: time.Duration(conf.UserSessionDurationSeconds) * time.Second,
	}
	return settings, nil
}
func getOidcClient(dexServerAddress string, settings *Settings, userVerifier UserVerifier, RedirectUrlSanitiser RedirectUrlSanitiser) (*ClientApp, func(writer http.ResponseWriter, request *http.Request), error) {
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
	dexProxy := NewDexHTTPReverseProxy(dexServerAddress, dexClient.Transport)
	cahecStore := &Cache{OidcState: map[string]*OIDCState{}}
	oidcClient, err := NewClientApp(settings, cahecStore, "/", userVerifier, RedirectUrlSanitiser)
	if err != nil {
		return nil, nil, err
	}
	return oidcClient, dexProxy, err
}

const dexProxyUri = "api/dex"

type DexConfig struct {
	DexServerAddress string `env:"DEX_SERVER_ADDRESS" envDefault:"http://argocd-dex-server.devtroncd:5556/authenticator"`
	Url              string `env:"AUTHENTICATOR_URL" envDefault:"https://demo.devtron.info:32443/authenticator/"`
	DexClientSecret  string `env:"DEX_CLIENT_SECRET" envDefault:""`
	DexClientID      string `env:"DEX_CLIENT_ID" envDefault:"argo-cd"`
	// Specifies token expiration duration
	UserSessionDurationSeconds int `env:"USER_SESSION_DURATION_SECONDS" envDefault:"300"`
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

func DexConfigConfigFromEnv() (*DexConfig, error) {
	cfg := &DexConfig{}
	err := env.Parse(cfg)
	return cfg, err
}
