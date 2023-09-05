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

package chartProvider

import (
	"github.com/gorilla/mux"
)

type ChartProviderRouter interface {
	Init(configRouter *mux.Router)
}

type ChartProviderRouterImpl struct {
	chartProviderRestHandler ChartProviderRestHandler
}

func NewChartProviderRouterImpl(chartProviderRestHandler ChartProviderRestHandler) *ChartProviderRouterImpl {
	return &ChartProviderRouterImpl{
		chartProviderRestHandler: chartProviderRestHandler,
	}
}

func (router ChartProviderRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("/list").
		HandlerFunc(router.chartProviderRestHandler.GetChartProviderList).Methods("GET")
	configRouter.Path("/update").
		HandlerFunc(router.chartProviderRestHandler.ToggleChartProvider).Methods("POST")
	configRouter.Path("/sync-chart").
		HandlerFunc(router.chartProviderRestHandler.SyncChartProvider).Methods("POST")
}
