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

package appStore

import (
	appStoreDiscover "github.com/devtron-labs/devtron/api/appStore/discover"
	appStoreValues "github.com/devtron-labs/devtron/api/appStore/values"
	"github.com/gorilla/mux"
)

type AppStoreRouter interface {
	Init(configRouter *mux.Router)
}

type AppStoreRouterImpl struct {
	deployRestHandler         InstalledAppRestHandler
	appStoreValuesRouter appStoreValues.AppStoreValuesRouter
	appStoreDiscoverRouter    appStoreDiscover.AppStoreDiscoverRouter
}

func NewAppStoreRouterImpl(restHandler InstalledAppRestHandler,
	appStoreValuesRouter appStoreValues.AppStoreValuesRouter, appStoreDiscoverRouter appStoreDiscover.AppStoreDiscoverRouter) *AppStoreRouterImpl {
	return &AppStoreRouterImpl{
		deployRestHandler:         restHandler,
		appStoreValuesRouter: appStoreValuesRouter,
		appStoreDiscoverRouter:    appStoreDiscoverRouter,
	}
}

func (router AppStoreRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("/application/install").
		HandlerFunc(router.deployRestHandler.CreateInstalledApp).Methods("POST")

	configRouter.Path("/application/exists").
		HandlerFunc(router.deployRestHandler.CheckAppExists).Methods("POST")
	configRouter.Path("/group/install").
		HandlerFunc(router.deployRestHandler.DeployBulk).Methods("POST")
	configRouter.Path("/application/update").
		HandlerFunc(router.deployRestHandler.UpdateInstalledApp).Methods("PUT")

	// discover router starts
	appStoreDiscoverSubRouter := configRouter.PathPrefix("/discover").Subrouter()
	router.appStoreDiscoverRouter.Init(appStoreDiscoverSubRouter)
	// discover router ends

	configRouter.Path("/installed-app/detail").Queries("installed-app-id", "{installed-app-id}").Queries("env-id", "{env-id}").
		HandlerFunc(router.deployRestHandler.FetchAppDetailsForInstalledApp).
		Methods("GET")

	configRouter.Path("/installed-app").
		HandlerFunc(router.deployRestHandler.GetAllInstalledApp).Methods("GET")

	configRouter.Path("/installed-app/{appStoreId}").
		HandlerFunc(router.deployRestHandler.GetInstalledAppsByAppStoreId).Methods("GET")

	configRouter.Path("/application/version/{installedAppVersionId}").
		HandlerFunc(router.deployRestHandler.GetInstalledAppVersion).Methods("GET")
	configRouter.Path("/application/delete/{id}").
		HandlerFunc(router.deployRestHandler.DeleteInstalledApp).Methods("DELETE")

	// values router starts
	appStoreValuesSubRouter := configRouter.PathPrefix("/values").Subrouter()
	router.appStoreValuesRouter.Init(appStoreValuesSubRouter)
	// values router ends

	configRouter.Path("/cluster-component/install/{clusterId}").
		HandlerFunc(router.deployRestHandler.DefaultComponentInstallation).Methods("POST")
}
