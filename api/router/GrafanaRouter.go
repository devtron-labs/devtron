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

package router

import (
	"fmt"
	"github.com/devtron-labs/devtron/client/grafana"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

type GrafanaRouter interface {
	initGrafanaRouter(router *mux.Router)
}

type GrafanaRouterImpl struct {
	logger       *zap.SugaredLogger
	grafanaProxy func(writer http.ResponseWriter, request *http.Request)
}

func NewGrafanaRouterImpl(logger *zap.SugaredLogger, grafanaCfg *grafana.Config) (*GrafanaRouterImpl, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	grafanaProxy, err := grafana.NewGrafanaHTTPReverseProxy(fmt.Sprintf("http://%s:%s", grafanaCfg.Host, grafanaCfg.Port), client.Transport)
	if err != nil {
		return nil, err
	}
	router := &GrafanaRouterImpl{
		grafanaProxy: grafanaProxy,
		logger:       logger,
	}
	return router, nil
}

func (router GrafanaRouterImpl) initGrafanaRouter(grafanaRouter *mux.Router) {
	grafanaRouter.PathPrefix("").HandlerFunc(router.grafanaProxy)
}
