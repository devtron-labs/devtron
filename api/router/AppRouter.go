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

type AppRouter interface {
	init(configRouter *mux.Router)
}

type AppRouterImpl struct {
	restHandler restHandler.AppRestHandler
}

func NewAppRouterImpl(restHandler restHandler.AppRestHandler) *AppRouterImpl {
	return &AppRouterImpl{restHandler: restHandler}

}

func (router AppRouterImpl) init(configRouter *mux.Router) {
	configRouter.Path("/").HandlerFunc(router.restHandler.CreateApp).Methods("POST")
	configRouter.Path("/{appId}").HandlerFunc(router.restHandler.GetAppAllDetail).Methods("GET")
}
