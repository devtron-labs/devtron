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

package insightgrid

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

type InsightGridRouter interface {
	InitInsightGridRouter(router *mux.Router)
}

type InsightGridRouterImpl struct {
	logger           *zap.SugaredLogger
	insightGridProxy func(writer http.ResponseWriter, request *http.Request)
}

func NewInsightGridRouterImpl(logger *zap.SugaredLogger, insightGridCfg *Config) (*InsightGridRouterImpl, error) {
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
	insightGridProxy, err := proxy.NewDashboardHTTPReverseProxy(fmt.Sprintf("http://%s:%s", insightGridCfg.Host, insightGridCfg.Port), client.Transport)
	if err != nil {
		return nil, err
	}
	router := &InsightGridRouterImpl{
		insightGridProxy: insightGridProxy,
		logger:           logger,
	}
	return router, nil
}

func (router InsightGridRouterImpl) InitInsightGridRouter(insightGridRouter *mux.Router) {
	insightGridRouter.PathPrefix("").HandlerFunc(router.insightGridProxy)
}

var InsightGridWireSet = wire.NewSet(
	GetConfig,
	NewInsightGridRouterImpl,
	wire.Bind(new(InsightGridRouter), new(*InsightGridRouterImpl)),
)
