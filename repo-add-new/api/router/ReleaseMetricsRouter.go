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
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type ReleaseMetricsRouter interface {
	initReleaseMetricsRouter(router *mux.Router)
}

type ReleaseMetricsRouterImpl struct {
	logger                    *zap.SugaredLogger
	releaseMetricsRestHandler restHandler.ReleaseMetricsRestHandler
}

func NewReleaseMetricsRouterImpl(logger *zap.SugaredLogger,
	releaseMetricsRestHandler restHandler.ReleaseMetricsRestHandler) *ReleaseMetricsRouterImpl {
	return &ReleaseMetricsRouterImpl{
		logger:                    logger,
		releaseMetricsRestHandler: releaseMetricsRestHandler,
	}
}

func (impl ReleaseMetricsRouterImpl) initReleaseMetricsRouter(router *mux.Router) {
	router.Path("/reset-app-environment").
		HandlerFunc(impl.releaseMetricsRestHandler.ResetDataForAppEnvironment).
		Methods("POST")
	router.Path("/reset-all-app-environment").
		HandlerFunc(impl.releaseMetricsRestHandler.ResetDataForAllAppEnvironment).
		Methods("POST")
	router.Path("/").
		HandlerFunc(impl.releaseMetricsRestHandler.GetDeploymentMetrics).
		Methods("GET")
}
