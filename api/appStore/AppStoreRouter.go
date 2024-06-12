/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package appStore

import (
	"github.com/devtron-labs/devtron/api/appStore/chartProvider"
	appStoreDeployment "github.com/devtron-labs/devtron/api/appStore/deployment"
	appStoreDiscover "github.com/devtron-labs/devtron/api/appStore/discover"
	appStoreValues "github.com/devtron-labs/devtron/api/appStore/values"
	"github.com/gorilla/mux"
)

type AppStoreRouter interface {
	Init(configRouter *mux.Router)
}

type AppStoreRouterImpl struct {
	deployRestHandler                 InstalledAppRestHandler
	appStoreValuesRouter              appStoreValues.AppStoreValuesRouter
	appStoreDiscoverRouter            appStoreDiscover.AppStoreDiscoverRouter
	appStoreDeploymentRouter          appStoreDeployment.AppStoreDeploymentRouter
	chartProviderRouter               chartProvider.ChartProviderRouter
	appStoreStatusTimelineRestHandler AppStoreStatusTimelineRestHandler
}

func NewAppStoreRouterImpl(restHandler InstalledAppRestHandler,
	appStoreValuesRouter appStoreValues.AppStoreValuesRouter,
	appStoreDiscoverRouter appStoreDiscover.AppStoreDiscoverRouter,
	chartProviderRouter chartProvider.ChartProviderRouter,
	appStoreDeploymentRouter appStoreDeployment.AppStoreDeploymentRouter,
	appStoreStatusTimelineRestHandler AppStoreStatusTimelineRestHandler) *AppStoreRouterImpl {
	return &AppStoreRouterImpl{
		deployRestHandler:                 restHandler,
		appStoreValuesRouter:              appStoreValuesRouter,
		appStoreDiscoverRouter:            appStoreDiscoverRouter,
		chartProviderRouter:               chartProviderRouter,
		appStoreDeploymentRouter:          appStoreDeploymentRouter,
		appStoreStatusTimelineRestHandler: appStoreStatusTimelineRestHandler,
	}
}

func (router AppStoreRouterImpl) Init(configRouter *mux.Router) {
	// deployment router starts
	appStoreDeploymentSubRouter := configRouter.PathPrefix("/deployment").Subrouter()
	router.appStoreDeploymentRouter.Init(appStoreDeploymentSubRouter)

	configRouter.Path("/deployment-status/timeline/{installedAppId}/{envId}").
		HandlerFunc(router.appStoreStatusTimelineRestHandler.FetchTimelinesForAppStore).Methods("GET")

	// deployment router ends

	// values router starts
	appStoreValuesSubRouter := configRouter.PathPrefix("/values").Subrouter()
	router.appStoreValuesRouter.Init(appStoreValuesSubRouter)
	// values router ends

	// discover router starts
	appStoreDiscoverSubRouter := configRouter.PathPrefix("/discover").Subrouter()
	router.appStoreDiscoverRouter.Init(appStoreDiscoverSubRouter)
	// discover router ends

	// chart provider router starts
	chartProviderSubRouter := configRouter.PathPrefix("/chart-provider").Subrouter()
	router.chartProviderRouter.Init(chartProviderSubRouter)
	// chart provider router ends
	configRouter.Path("/overview").Queries("installedAppId", "{installedAppId}").
		HandlerFunc(router.deployRestHandler.FetchAppOverview).Methods("GET")
	configRouter.Path("/application/exists").
		HandlerFunc(router.deployRestHandler.CheckAppExists).Methods("POST")
	configRouter.Path("/group/install").
		HandlerFunc(router.deployRestHandler.DeployBulk).Methods("POST")
	configRouter.Path("/installed-app/detail").Queries("installed-app-id", "{installed-app-id}").Queries("env-id", "{env-id}").
		HandlerFunc(router.deployRestHandler.FetchAppDetailsForInstalledApp).
		Methods("GET")
	configRouter.Path("/installed-app/delete/{installedAppId}/non-cascade").
		HandlerFunc(router.deployRestHandler.DeleteArgoInstalledAppWithNonCascade).
		Methods("DELETE")
	configRouter.Path("/installed-app/detail/v2").Queries("installed-app-id", "{installed-app-id}").Queries("env-id", "{env-id}").
		HandlerFunc(router.deployRestHandler.FetchAppDetailsForInstalledAppV2).
		Methods("GET")
	configRouter.Path("/installed-app/detail/resource-tree").Queries("installed-app-id", "{installed-app-id}").Queries("env-id", "{env-id}").
		HandlerFunc(router.deployRestHandler.FetchResourceTree).
		Methods("GET")
	configRouter.Path("/installed-app/resource/hibernate").Queries("installed-app-id", "{installed-app-id}").Queries("env-id", "{env-id}").
		HandlerFunc(router.deployRestHandler.FetchResourceTreeForACDApp).
		Methods("GET")
	configRouter.Path("/installed-app/notes").Queries("installed-app-id", "{installed-app-id}").Queries("env-id", "{env-id}").
		HandlerFunc(router.deployRestHandler.FetchNotesForArgoInstalledApp).
		Methods("GET")
	configRouter.Path("/installed-app").
		HandlerFunc(router.deployRestHandler.GetAllInstalledApp).Methods("GET")
	configRouter.Path("/cluster-component/install/{clusterId}").
		HandlerFunc(router.deployRestHandler.DefaultComponentInstallation).Methods("POST")

	configRouter.Path("/installed-app/migrate").
		HandlerFunc(router.deployRestHandler.MigrateDeploymentTypeForChartStore).Methods("POST")
	configRouter.Path("/installed-app/trigger").
		HandlerFunc(router.deployRestHandler.TriggerChartStoreAppAfterMigration).Methods("POST")
}
