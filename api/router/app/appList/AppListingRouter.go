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

package appList

import (
	"github.com/devtron-labs/devtron/api/restHandler/app/appList"
	"github.com/gorilla/mux"
)

type AppListingRouter interface {
	InitAppListingRouter(helmRouter *mux.Router)
}

type AppListingRouterImpl struct {
	appListingRestHandler appList.AppListingRestHandler
}

func NewAppListingRouterImpl(appListingRestHandler appList.AppListingRestHandler) *AppListingRouterImpl {
	router := &AppListingRouterImpl{
		appListingRestHandler: appListingRestHandler,
	}
	return router
}

func (router AppListingRouterImpl) InitAppListingRouter(appListingRouter *mux.Router) {
	appListingRouter.Path("").
		HandlerFunc(router.appListingRestHandler.FetchAppsByEnvironmentV2).
		Methods("POST")

	appListingRouter.Path("/v2").
		HandlerFunc(router.appListingRestHandler.FetchAppsByEnvironmentV2).
		Methods("POST")

	appListingRouter.Path("/group/{env-id}").
		HandlerFunc(router.appListingRestHandler.FetchOverviewAppsByEnvironment).
		Methods("GET")
}
