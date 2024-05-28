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

package dashboardEvent

import (
	"github.com/gorilla/mux"
)

type DashboardTelemetryRouter interface {
	Init(configRouter *mux.Router)
}

type DashboardTelemetryRouterImpl struct {
	deploymentRestHandler DashboardTelemetryRestHandler
}

func NewDashboardTelemetryRouterImpl(deploymentRestHandler DashboardTelemetryRestHandler) *DashboardTelemetryRouterImpl {
	return &DashboardTelemetryRouterImpl{
		deploymentRestHandler: deploymentRestHandler,
	}
}

func (router DashboardTelemetryRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("/dashboardAccessed").
		HandlerFunc(router.deploymentRestHandler.SendDashboardAccessedEvent).Methods("GET")
	configRouter.Path("/dashboardLoggedIn").
		HandlerFunc(router.deploymentRestHandler.SendDashboardLoggedInEvent).Methods("GET")
}
