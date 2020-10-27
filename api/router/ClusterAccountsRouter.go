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

type ClusterAccountsRouter interface {
	InitClusterAccountsRouter(clusterAccountsRouter *mux.Router)
}

type ClusterAccountsRouterImpl struct {
	clusterAccountsRestHandler restHandler.ClusterAccountsRestHandler
}

func NewClusterAccountsRouterImpl(clusterAccountsRestHandler restHandler.ClusterAccountsRestHandler) *ClusterAccountsRouterImpl {
	return &ClusterAccountsRouterImpl{clusterAccountsRestHandler: clusterAccountsRestHandler}
}

func (impl ClusterAccountsRouterImpl) InitClusterAccountsRouter(clusterAccountsRouter *mux.Router) {
	clusterAccountsRouter.Path("/").
		Methods("GET").
		Queries("clusterName", "{clusterName}").
		HandlerFunc(impl.clusterAccountsRestHandler.Get)
	clusterAccountsRouter.Path("/").
		Methods("GET").
		Queries("environment", "{environment}").
		HandlerFunc(impl.clusterAccountsRestHandler.GetByEnvironment)
	clusterAccountsRouter.Path("/").
		Methods("POST").
		HandlerFunc(impl.clusterAccountsRestHandler.Save)

	clusterAccountsRouter.Path("/").
		Methods("PUT").
		HandlerFunc(impl.clusterAccountsRestHandler.Update)

	clusterAccountsRouter.Path("/").
		Methods("GET").
		Queries("id", "{id}").
		HandlerFunc(impl.clusterAccountsRestHandler.FindById)
	clusterAccountsRouter.Path("/all").
		Methods("GET").
		HandlerFunc(impl.clusterAccountsRestHandler.FindAll)
}
