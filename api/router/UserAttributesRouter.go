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
	user "github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type UserAttributesRouter interface {
	initAttributesRouter(helmRouter *mux.Router)
}

type UserAttributesRouterImpl struct {
	userAttributesRestHandler user.UserAttributesRestHandler
}

func NewUserAttributesRouterImpl(userAttributesRestHandler user.UserAttributesRestHandler) *UserAttributesRouterImpl {
	router := &UserAttributesRouterImpl{
		userAttributesRestHandler: userAttributesRestHandler,
	}
	return router
}

func (router UserAttributesRouterImpl) initAttributesRouter(attributesRouter *mux.Router) {
	attributesRouter.Path("/create").
		HandlerFunc(router.userAttributesRestHandler.AddUserAttributes).Methods("POST")
	attributesRouter.Path("/update").
		HandlerFunc(router.userAttributesRestHandler.UpdateUserAttributes).Methods("PUT")
	attributesRouter.Path("/get").
		HandlerFunc(router.userAttributesRestHandler.GetUserAttribute).Methods("POST")
}
