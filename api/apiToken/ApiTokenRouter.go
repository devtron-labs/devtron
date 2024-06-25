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

package apiToken

import (
	"github.com/gorilla/mux"
)

type ApiTokenRouter interface {
	InitApiTokenRouter(configRouter *mux.Router)
}

type ApiTokenRouterImpl struct {
	apiTokenRestHandler ApiTokenRestHandler
}

func NewApiTokenRouterImpl(apiTokenRestHandler ApiTokenRestHandler) *ApiTokenRouterImpl {
	return &ApiTokenRouterImpl{apiTokenRestHandler: apiTokenRestHandler}
}

func (impl ApiTokenRouterImpl) InitApiTokenRouter(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(impl.apiTokenRestHandler.GetAllApiTokens).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.apiTokenRestHandler.CreateApiToken).Methods("POST")
	configRouter.Path("/{id}").HandlerFunc(impl.apiTokenRestHandler.UpdateApiToken).Methods("PUT")
	configRouter.Path("/{id}").HandlerFunc(impl.apiTokenRestHandler.DeleteApiToken).Methods("DELETE")
	configRouter.Path("/webhook").HandlerFunc(impl.apiTokenRestHandler.GetAllApiTokensForWebhook).Methods("GET")
}
