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
)

type AppListingRouter interface {
	initAppListingRouter(helmRouter *mux.Router)
}

type AppListingRouterImpl struct {
	appListingRestHandler restHandler.AppListingRestHandler
}

func NewAppListingRouterImpl(appListingRestHandler restHandler.AppListingRestHandler) *AppListingRouterImpl {
	router := &AppListingRouterImpl{
		appListingRestHandler: appListingRestHandler,
	}
	return router
}

func (router AppListingRouterImpl) initAppListingRouter(appListingRouter *mux.Router) {
	appListingRouter.Path("/allApps").HandlerFunc(router.appListingRestHandler.FetchAllDevtronManagedApps).
		Methods("GET")

	appListingRouter.Path("/resource/urls").Queries("envId", "{envId}").
		HandlerFunc(router.appListingRestHandler.GetHostUrlsByBatch).Methods("GET")

	appListingRouter.Path("/list").
		HandlerFunc(router.appListingRestHandler.FetchAppsByEnvironment).
		Methods("POST")

	appListingRouter.Path("/list/{version}").
		HandlerFunc(router.appListingRestHandler.FetchAppsByEnvironmentVersioned).
		Methods("POST")

	appListingRouter.Path("/list/group/{env-id}").
		HandlerFunc(router.appListingRestHandler.FetchOverviewAppsByEnvironment).
		Methods("GET")

	appListingRouter.Path("/list/group/{env-id}").
		Queries("size", "{size}", "offset", "{offset}").
		HandlerFunc(router.appListingRestHandler.FetchOverviewAppsByEnvironment).
		Methods("GET")
	//This API used for fetch app details, not deployment details
	appListingRouter.Path("/detail").Queries("app-id", "{app-id}").Queries("env-id", "{env-id}").
		HandlerFunc(router.appListingRestHandler.FetchAppDetails).
		Methods("GET")

	appListingRouter.Path("/detail/v2").Queries("app-id", "{app-id}").Queries("env-id", "{env-id}").
		HandlerFunc(router.appListingRestHandler.FetchAppDetailsV2).
		Methods("GET")

	appListingRouter.Path("/detail/resource-tree").Queries("app-id", "{app-id}").Queries("env-id", "{env-id}").
		HandlerFunc(router.appListingRestHandler.FetchResourceTree).
		Methods("GET")

	appListingRouter.Path("/stage/status").Queries("app-id", "{app-id}").
		HandlerFunc(router.appListingRestHandler.FetchAppStageStatus).
		Methods("GET")

	appListingRouter.Path("/other-env").Queries("app-id", "{app-id}").
		HandlerFunc(router.appListingRestHandler.FetchOtherEnvironment).
		Methods("GET")

	appListingRouter.Path("/other-env/min").Queries("app-id", "{app-id}").
		HandlerFunc(router.appListingRestHandler.FetchMinDetailOtherEnvironment).Methods("GET")

	appListingRouter.Path("/linkouts/{Id}/{appId}/{envId}").Queries("podName", "{podName}").
		Queries("containerName", "{containerName}").
		HandlerFunc(router.appListingRestHandler.RedirectToLinkouts).
		Methods("GET")

	appListingRouter.Path("/deployment-status/manual-sync/{appId}/{envId}").
		HandlerFunc(router.appListingRestHandler.ManualSyncAcdPipelineDeploymentStatus).
		Methods("GET")

	appListingRouter.Path("/app-listing/autocomplete").
		HandlerFunc(router.appListingRestHandler.GetClusterTeamAndEnvListForAutocomplete).Methods("GET")

}
