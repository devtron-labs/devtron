/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package fluxApplication

import (
	"github.com/gorilla/mux"
)

type FluxApplicationRouter interface {
	InitFluxApplicationRouter(fluxApplicationRouter *mux.Router)
}

type FluxApplicationRouterImpl struct {
	fluxApplicationRestHandler FluxApplicationRestHandler
}

func NewFluxApplicationRouterImpl(fluxApplicationRestHandler FluxApplicationRestHandler) *FluxApplicationRouterImpl {
	return &FluxApplicationRouterImpl{
		fluxApplicationRestHandler: fluxApplicationRestHandler,
	}
}

func (impl *FluxApplicationRouterImpl) InitFluxApplicationRouter(fluxApplicationRouter *mux.Router) {
	fluxApplicationRouter.Path("").
		Methods("GET").
		HandlerFunc(impl.fluxApplicationRestHandler.ListFluxApplications)
	fluxApplicationRouter.Path("/app").Queries("appId", "{appId}").
		HandlerFunc(impl.fluxApplicationRestHandler.GetApplicationDetail).Methods("GET")
}
