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

package argoApplication

import "github.com/gorilla/mux"

type ArgoApplicationRouter interface {
	InitArgoApplicationRouter(argoApplicationRouter *mux.Router)
}

type ArgoApplicationRouterImpl struct {
	argoApplicationRestHandler ArgoApplicationRestHandler
}

func NewArgoApplicationRouterImpl(argoApplicationRestHandler ArgoApplicationRestHandler) *ArgoApplicationRouterImpl {
	return &ArgoApplicationRouterImpl{
		argoApplicationRestHandler: argoApplicationRestHandler,
	}
}

func (impl *ArgoApplicationRouterImpl) InitArgoApplicationRouter(argoApplicationRouter *mux.Router) {
	argoApplicationRouter.Path("").
		Methods("GET").
		HandlerFunc(impl.argoApplicationRestHandler.ListApplications)

	argoApplicationRouter.Path("/detail").
		Methods("GET").
		HandlerFunc(impl.argoApplicationRestHandler.GetApplicationDetail)
}
