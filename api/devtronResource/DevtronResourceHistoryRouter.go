/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package devtronResource

import "github.com/gorilla/mux"

type HistoryRouter interface {
	InitDtResourceHistoryRouter(devtronResourceRouter *mux.Router)
}

type HistoryRouterImpl struct {
	dtResourceHistoryRestHandler HistoryRestHandler
}

func NewHistoryRouterImpl(dtResourceHistoryRestHandler HistoryRestHandler) *HistoryRouterImpl {
	return &HistoryRouterImpl{dtResourceHistoryRestHandler: dtResourceHistoryRestHandler}
}

func (router *HistoryRouterImpl) InitDtResourceHistoryRouter(historyRouter *mux.Router) {
	historyRouter.Path("/deployment/config/{kind:[a-zA-Z0-9/-]+}/{version:[a-zA-Z0-9]+}").
		HandlerFunc(router.dtResourceHistoryRestHandler.GetDeploymentHistoryConfigList).Methods("GET")

	historyRouter.Path("/deployment/{kind:[a-zA-Z0-9/-]+}/{version:[a-zA-Z0-9]+}").
		HandlerFunc(router.dtResourceHistoryRestHandler.GetDeploymentHistory).Methods("GET")
}
