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

type ExternalAppsRouter interface {
	initExternalAppsRouterImpl(helmRouter *mux.Router)
}

type ExternalAppsRouterImpl struct {
	externalAppsRestHandler restHandler.ExternalAppsRestHandler
}

func NewExternalAppsRouterImpl(externalAppsRestHandler restHandler.ExternalAppsRestHandler) *ExternalAppsRouterImpl {
	router := &ExternalAppsRouterImpl{
		externalAppsRestHandler: externalAppsRestHandler,
	}
	return router
}

func (router ExternalAppsRouterImpl) initExternalAppsRouterImpl(externalAppsRouter *mux.Router) {

	externalAppsRouter.Path("/external-apps/{id}").
		HandlerFunc(router.externalAppsRestHandler.FindById).
		Methods("GET")

	externalAppsRouter.Path("/external-apps/all").
		HandlerFunc(router.externalAppsRestHandler.FindAll).
		Methods("GET")

}
