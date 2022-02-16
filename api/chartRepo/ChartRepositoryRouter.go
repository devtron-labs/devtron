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

package chartRepo

import (
	"github.com/gorilla/mux"
)

type ChartRepositoryRouter interface {
	Init(configRouter *mux.Router)
}

type ChartRepositoryRouterImpl struct {
	chartRepositoryRestHandler ChartRepositoryRestHandler
}

func NewChartRepositoryRouterImpl(chartRepositoryRestHandler ChartRepositoryRestHandler) *ChartRepositoryRouterImpl {
	return &ChartRepositoryRouterImpl{chartRepositoryRestHandler: chartRepositoryRestHandler}
}

func (router ChartRepositoryRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("/sync-charts").
		HandlerFunc(router.chartRepositoryRestHandler.TriggerChartSyncManual).Methods("POST")
	configRouter.Path("/list").
		HandlerFunc(router.chartRepositoryRestHandler.GetChartRepoList).Methods("GET")
	configRouter.Path("/{id}").
		HandlerFunc(router.chartRepositoryRestHandler.GetChartRepoById).Methods("GET")
	configRouter.Path("/create").
		HandlerFunc(router.chartRepositoryRestHandler.CreateChartRepo).Methods("POST")
	configRouter.Path("/update").
		HandlerFunc(router.chartRepositoryRestHandler.UpdateChartRepo).Methods("POST")
	configRouter.Path("/validate").
		HandlerFunc(router.chartRepositoryRestHandler.ValidateChartRepo).Methods("POST")
	configRouter.Path("/").
		HandlerFunc(router.chartRepositoryRestHandler.DeleteChartRepo).Methods("DELETE")
}
