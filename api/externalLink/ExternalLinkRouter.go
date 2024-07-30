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

package externalLink

import (
	"github.com/gorilla/mux"
)

type ExternalLinkRouter interface {
	InitExternalLinkRouter(gocdRouter *mux.Router)
}
type ExternalLinkRouterImpl struct {
	externalLinkRestHandler ExternalLinkRestHandler
}

func NewExternalLinkRouterImpl(externalLinkRestHandler ExternalLinkRestHandler) *ExternalLinkRouterImpl {
	return &ExternalLinkRouterImpl{externalLinkRestHandler: externalLinkRestHandler}
}

func (impl ExternalLinkRouterImpl) InitExternalLinkRouter(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(impl.externalLinkRestHandler.CreateExternalLinks).Methods("POST")
	configRouter.Path("/tools").HandlerFunc(impl.externalLinkRestHandler.GetExternalLinkMonitoringTools).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.externalLinkRestHandler.GetExternalLinks).Methods("GET")
	configRouter.Path("/v2").HandlerFunc(impl.externalLinkRestHandler.GetExternalLinksV2).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.externalLinkRestHandler.UpdateExternalLink).Methods("PUT")
	configRouter.Path("").HandlerFunc(impl.externalLinkRestHandler.DeleteExternalLink).Queries("id", "{id}").Methods("DELETE")
}
