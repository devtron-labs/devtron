/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package proxy

import (
	"encoding/json"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type ProxyRouter interface {
	InitProxyRouter(router *mux.Router)
}

type ProxyConnection struct {
	Host        string `json:"host"`
	Port        string `json:"port"`
	PassKey     string `json:"passKey"`
	ServiceName ProxyServiceName
}

type ProxyServiceName string

const (
	IMAGE_SCANNER ProxyServiceName = "image-scanner"
	KUBELINK                       = "kubelink"
	GIT_SENSOR                     = "gitsensor"
	KUBEWATCH                      = "kubewatch"
	SCOOP                          = "scoop"
)

type Config struct {
	ProxyServiceConfig string `env:"PROXY_SERVICE_CONFIG" envDefault:"{}"`
}

func GetProxyConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}

type ProxyRouterImpl struct {
	logger *zap.SugaredLogger
	proxy  map[ProxyServiceName]func(writer http.ResponseWriter, request *http.Request)
}

func NewProxyRouterImpl(logger *zap.SugaredLogger, proxyCfg *Config, enforcer casbin.Enforcer) (*ProxyRouterImpl, error) {
	client := &http.Client{
		Transport: NewProxyTransport(),
	}
	proxyConnection := make(map[ProxyServiceName]ProxyConnection)
	err := json.Unmarshal([]byte(proxyCfg.ProxyServiceConfig), &proxyConnection)
	if err != nil {
		logger.Warnw("bad env value for PROXY_SERVICE_CONFIG", "err", err)
	}

	proxy := make(map[ProxyServiceName]func(writer http.ResponseWriter, request *http.Request))
	for serviceName, connection := range proxyConnection {
		connection.ServiceName = serviceName
		proxy[serviceName], err = NewHTTPReverseProxy(connection, client.Transport, enforcer)
		if err != nil {
			return nil, err
		}
	}


	router := &ProxyRouterImpl{
		proxy:  proxy,
		logger: logger,
	}
	return router, nil
}

func (router ProxyRouterImpl) InitProxyRouter(ProxyRouter *mux.Router) {

	ProxyRouter.PathPrefix("/kubelink").HandlerFunc(router.proxy[KUBELINK])
	ProxyRouter.PathPrefix("/gitsensor").HandlerFunc(router.proxy[GIT_SENSOR])
	ProxyRouter.PathPrefix("/kubewatch").HandlerFunc(router.proxy[KUBEWATCH])
	ProxyRouter.PathPrefix("/image-scanner").HandlerFunc(router.proxy[IMAGE_SCANNER])
	ProxyRouter.PathPrefix("/scoop").HandlerFunc(router.proxy[SCOOP])
}
