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

package cluster

import (
	"github.com/gorilla/mux"
)

type EnvironmentRouter interface {
	InitEnvironmentClusterMappingsRouter(clusterAccountsRouter *mux.Router)
}

type EnvironmentRouterImpl struct {
	environmentClusterMappingsRestHandler EnvironmentRestHandler
}

func NewEnvironmentRouterImpl(environmentClusterMappingsRestHandler EnvironmentRestHandler) *EnvironmentRouterImpl {
	return &EnvironmentRouterImpl{environmentClusterMappingsRestHandler: environmentClusterMappingsRestHandler}
}

func (impl EnvironmentRouterImpl) InitEnvironmentClusterMappingsRouter(environmentClusterMappingsRouter *mux.Router) {
	/*environmentClusterMappingsRouter.Path("/").
	Methods("GET").
	Queries("clusterName", "{clusterName}").
	HandlerFunc(impl.environmentClusterMappingsRestHandler.Get)*/
	environmentClusterMappingsRouter.Path("/name").
		Methods("GET").
		Queries("environment", "{environment}").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.Get)
	environmentClusterMappingsRouter.Path("").
		Methods("GET").
		Queries("id", "{id}").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.FindById)
	environmentClusterMappingsRouter.Path("").
		Methods("GET").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.GetAll)
	environmentClusterMappingsRouter.Path("/active").
		Methods("GET").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.GetAllActive)
	environmentClusterMappingsRouter.Path("").
		Methods("POST").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.Create)

	environmentClusterMappingsRouter.Path("").
		Methods("PUT").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.Update)
	environmentClusterMappingsRouter.Path("/autocomplete").
		Methods("GET").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.GetEnvironmentListForAutocomplete)
	environmentClusterMappingsRouter.Path("/autocomplete/helm").
		Methods("GET").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.GetCombinedEnvironmentListForDropDown)

}
