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

package appStoreDiscover

import (
	"github.com/gorilla/mux"
)

type AppStoreDiscoverRouter interface {
	Init(configRouter *mux.Router)
}

type AppStoreDiscoverRouterImpl struct {
	appStoreRestHandler       AppStoreRestHandler
}

func NewAppStoreDiscoverRouterImpl(appStoreRestHandler AppStoreRestHandler) *AppStoreDiscoverRouterImpl {
	return &AppStoreDiscoverRouterImpl{
		appStoreRestHandler: appStoreRestHandler,
	}
}

func (router AppStoreDiscoverRouterImpl) Init(configRouter *mux.Router) {

	configRouter.Path("/").
		HandlerFunc(router.appStoreRestHandler.FindAllApps).Methods("GET")

	configRouter.Path("/application/{id}").
		HandlerFunc(router.appStoreRestHandler.GetChartDetailsForVersion).Methods("GET")

	configRouter.Path("/application/{appStoreId}/version/autocomplete").
		HandlerFunc(router.appStoreRestHandler.GetChartVersions).Methods("GET")

	configRouter.Path("/application/readme/{appStoreApplicationVersionId}").
		HandlerFunc(router.appStoreRestHandler.GetReadme).Methods("GET")

	configRouter.Path("/search").
		HandlerFunc(router.appStoreRestHandler.SearchAppStoreChartByName).Queries("chartName", "{chartName}").
		Methods("GET")

}
