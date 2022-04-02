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

package appStoreDeployment

import (
	"github.com/gorilla/mux"
)

type AppStoreDeploymentRouter interface {
	Init(configRouter *mux.Router)
}

type AppStoreDeploymentRouterImpl struct {
	appStoreDeploymentRestHandler AppStoreDeploymentRestHandler
}

func NewAppStoreDeploymentRouterImpl(appStoreDeploymentRestHandler AppStoreDeploymentRestHandler) *AppStoreDeploymentRouterImpl {
	return &AppStoreDeploymentRouterImpl{
		appStoreDeploymentRestHandler: appStoreDeploymentRestHandler,
	}
}

func (router AppStoreDeploymentRouterImpl) Init(configRouter *mux.Router) {

	configRouter.Path("/application/install").
		HandlerFunc(router.appStoreDeploymentRestHandler.InstallApp).Methods("POST")

	configRouter.Path("/installed-app/{appStoreId}").
		HandlerFunc(router.appStoreDeploymentRestHandler.GetInstalledAppsByAppStoreId).Methods("GET")

	configRouter.Path("/application/delete/{id}").
		HandlerFunc(router.appStoreDeploymentRestHandler.DeleteInstalledApp).Methods("DELETE")

	configRouter.Path("/application/helm/link-to-chart-store").
		HandlerFunc(router.appStoreDeploymentRestHandler.LinkHelmApplicationToChartStore).Methods("PUT")
}
