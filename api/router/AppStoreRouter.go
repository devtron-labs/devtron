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

type AppStoreRouter interface {
	initAppStoreRouter(configRouter *mux.Router)
}
type AppStoreRouterImpl struct {
	deployRestHandler         restHandler.InstalledAppRestHandler
	appStoreRestHandler       restHandler.AppStoreRestHandler
	appStoreValuesRestHandler restHandler.AppStoreValuesRestHandler
}

func NewAppStoreRouterImpl(appStoreRestHandler restHandler.AppStoreRestHandler, restHandler restHandler.InstalledAppRestHandler,
	appStoreValuesRestHandler restHandler.AppStoreValuesRestHandler) *AppStoreRouterImpl {
	return &AppStoreRouterImpl{deployRestHandler: restHandler, appStoreRestHandler: appStoreRestHandler,
		appStoreValuesRestHandler: appStoreValuesRestHandler}
}

func (router AppStoreRouterImpl) initAppStoreRouter(configRouter *mux.Router) {
	configRouter.Path("/application/install").
		HandlerFunc(router.deployRestHandler.CreateInstalledApp).Methods("POST")

	configRouter.Path("/application/exists").
		HandlerFunc(router.deployRestHandler.CheckAppExists).Methods("POST")
	configRouter.Path("/group/install").
		HandlerFunc(router.deployRestHandler.DeployBulk).Methods("POST")
	configRouter.Path("/application/update").
		HandlerFunc(router.deployRestHandler.UpdateInstalledApp).Methods("PUT")
	configRouter.Path("/").
		HandlerFunc(router.appStoreRestHandler.FindAllApps).Methods("GET")

	configRouter.Path("/application/{id}").
		HandlerFunc(router.appStoreRestHandler.GetChartDetailsForVersion).Methods("GET")

	configRouter.Path("/application/{appStoreId}/version/autocomplete").
		HandlerFunc(router.appStoreRestHandler.GetChartVersions).Methods("GET")
	configRouter.Path("/application/readme/{appStoreApplicationVersionId}").
		HandlerFunc(router.appStoreRestHandler.GetReadme).Methods("GET")

	configRouter.Path("/installed-app/detail").Queries("installed-app-id", "{installed-app-id}").Queries("env-id", "{env-id}").
		HandlerFunc(router.appStoreRestHandler.FetchAppDetailsForInstalledApp).
		Methods("GET")

	configRouter.Path("/installed-app").
		HandlerFunc(router.deployRestHandler.GetAllInstalledApp).Methods("GET")

	configRouter.Path("/installed-app/{appStoreId}").
		HandlerFunc(router.deployRestHandler.GetInstalledAppsByAppStoreId).Methods("GET")

	configRouter.Path("/application/version/{installedAppVersionId}").
		HandlerFunc(router.deployRestHandler.GetInstalledAppVersion).Methods("GET")
	configRouter.Path("/application/delete/{id}").
		HandlerFunc(router.deployRestHandler.DeleteInstalledApp).Methods("DELETE")

	configRouter.Path("/template/values").
		HandlerFunc(router.appStoreValuesRestHandler.CreateAppStoreVersionValues).Methods("POST")
	configRouter.Path("/template/values").
		HandlerFunc(router.appStoreValuesRestHandler.UpdateAppStoreVersionValues).Methods("PUT")
	configRouter.Path("/template/values").Queries("referenceId", "{referenceId}", "kind", "{kind}").
		HandlerFunc(router.appStoreValuesRestHandler.FindValuesById).Methods("GET")
	configRouter.Path("/template/values/{appStoreValueId}").
		HandlerFunc(router.appStoreValuesRestHandler.DeleteAppStoreVersionValues).Methods("DELETE")

	//used for manage api listing, will return only saved(template) values
	configRouter.Path("/template/values/list/{appStoreId}").
		HandlerFunc(router.appStoreValuesRestHandler.FindValuesByAppStoreIdAndReferenceType).Methods("GET")
	//used for all types of values category wise
	configRouter.Path("/application/values/list/{appStoreId}").
		HandlerFunc(router.appStoreValuesRestHandler.FetchTemplateValuesByAppStoreId).Methods("GET")
	configRouter.Path("/chart/selected/metadata").
		HandlerFunc(router.appStoreValuesRestHandler.GetSelectedChartMetadata).Methods("POST")

	configRouter.Path("/search").
		HandlerFunc(router.appStoreRestHandler.SearchAppStoreChartByName).Queries("chartName", "{chartName}").
		Methods("GET")

	configRouter.Path("/repo/create").
		HandlerFunc(router.appStoreRestHandler.CreateChartRepo).Methods("POST")

	configRouter.Path("/repo/update").
		HandlerFunc(router.appStoreRestHandler.UpdateChartRepo).Methods("POST")
}
