package oidc

import (
	"net"
	"net/http"
	"time"
)

func GetOidcClient(dexServerAddress string, settings *Settings) (*ClientApp, func(writer http.ResponseWriter, request *http.Request), error) {
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
	oidcClient, err := NewClientApp(settings, cahecStore, "/")
	if err != nil {
		return nil, nil, err
	}
	return oidcClient, dexProxy, err
}
