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

package dashboard

import (
	"fmt"
	"github.com/devtron-labs/devtron/client/proxy"
	"github.com/google/wire"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

type DashboardRouter interface {
	InitDashboardRouter(router *mux.Router)
}

type DashboardRouterImpl struct {
	logger         *zap.SugaredLogger
	dashboardProxy func(writer http.ResponseWriter, request *http.Request)
}

func NewDashboardRouterImpl(logger *zap.SugaredLogger, dashboardCfg *Config) (*DashboardRouterImpl, error) {
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
	dashboardProxy, err := proxy.NewDashboardHTTPReverseProxy(fmt.Sprintf("http://%s:%s", dashboardCfg.Host, dashboardCfg.Port), client.Transport)
	if err != nil {
		return nil, err
	}
	router := &DashboardRouterImpl{
		dashboardProxy: dashboardProxy,
		logger:         logger,
	}
	return router, nil
}

func (router DashboardRouterImpl) InitDashboardRouter(dashboardRouter *mux.Router) {
	dashboardRouter.PathPrefix("").HandlerFunc(router.dashboardProxy)
}

var DashboardWireSet = wire.NewSet(
	GetConfig,
	NewDashboardRouterImpl,
	wire.Bind(new(DashboardRouter), new(*DashboardRouterImpl)),
)
