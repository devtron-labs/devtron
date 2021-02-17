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

type AttributesRouter interface {
	initAttributesRouter(helmRouter *mux.Router)
}

type AttributesRouterImpl struct {
	attributesRestHandler restHandler.AttributesRestHandlerImpl
}

func NewAttributesRouterImpl(attributesRestHandler restHandler.AttributesRestHandlerImpl) *AttributesRouterImpl {
	router := &AttributesRouterImpl{
		attributesRestHandler: attributesRestHandler,
	}
	return router
}

func (router AttributesRouterImpl) initAttributesRouter(attributesRouter *mux.Router) {
	attributesRouter.Path("/create").
		HandlerFunc(router.attributesRestHandler.AddAttributes).Methods("POST")
	attributesRouter.Path("/update").
		HandlerFunc(router.attributesRestHandler.UpdateAttributes).Methods("PUT")
	attributesRouter.Path("").Queries("key", "{key}").
		HandlerFunc(router.attributesRestHandler.GetAttributesByKey).Methods("GET")
	attributesRouter.Path("/{id}").
		HandlerFunc(router.attributesRestHandler.GetAttributesById).Methods("GET")
	attributesRouter.Path("/active/list").
		HandlerFunc(router.attributesRestHandler.GetAttributesActiveList).Methods("GET")
}
