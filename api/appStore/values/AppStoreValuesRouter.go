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

package appStoreValues

import (
	"github.com/gorilla/mux"
)

type AppStoreValuesRouter interface {
	Init(configRouter *mux.Router)
}

type AppStoreValuesRouterImpl struct {
	appStoreValuesRestHandler AppStoreValuesRestHandler
}

func NewAppStoreValuesRouterImpl(appStoreValuesRestHandler AppStoreValuesRestHandler) *AppStoreValuesRouterImpl {
	return &AppStoreValuesRouterImpl{
		appStoreValuesRestHandler: appStoreValuesRestHandler,
	}
}

func (router AppStoreValuesRouterImpl) Init(configRouter *mux.Router) {

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
}
