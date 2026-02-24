/*
 * Copyright (c) 2024. Devtron Inc.
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

package globalConfig

import "github.com/gorilla/mux"

type AuthorisationConfigRouter interface {
	InitAuthorisationConfigRouter(router *mux.Router)
}

type AuthorisationConfigRouterImpl struct {
	handler AuthorisationConfigRestHandler
}

func NewGlobalConfigAuthorisationRouterImpl(handler AuthorisationConfigRestHandler) *AuthorisationConfigRouterImpl {
	return &AuthorisationConfigRouterImpl{handler: handler}
}

func (router *AuthorisationConfigRouterImpl) InitAuthorisationConfigRouter(authorisationConfigRouter *mux.Router) {
	authorisationConfigRouter.Path("/global-config").
		HandlerFunc(router.handler.CreateOrUpdateAuthorisationConfig).Methods("POST")
	authorisationConfigRouter.Path("/global-config").
		HandlerFunc(router.handler.GetAllActiveAuthorisationConfig).Methods("GET")
}
