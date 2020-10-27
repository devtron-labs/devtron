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

type ClusterRouter interface {
	InitClusterRouter(clusterRouter *mux.Router)
}

type ClusterRouterImpl struct {
	clusterRestHandler restHandler.ClusterRestHandler
}

func NewClusterRouterImpl(handler restHandler.ClusterRestHandler) *ClusterRouterImpl {
	return &ClusterRouterImpl{
		clusterRestHandler: handler,
	}
}

func (impl ClusterRouterImpl) InitClusterRouter(clusterRouter *mux.Router) {
	clusterRouter.Path("").
		Methods("POST").
		HandlerFunc(impl.clusterRestHandler.Save)

	clusterRouter.Path("").
		Methods("GET").
		Queries("id", "{id}").
		HandlerFunc(impl.clusterRestHandler.FindById)

	clusterRouter.Path("").
		Methods("GET").
		HandlerFunc(impl.clusterRestHandler.FindAll)

	clusterRouter.Path("/env").
		Methods("GET").
		Queries("id", "{id}").
		HandlerFunc(impl.clusterRestHandler.FindByEnvId)

	clusterRouter.Path("").
		Methods("PUT").
		HandlerFunc(impl.clusterRestHandler.Update)

	clusterRouter.Path("/acd/").
		Methods("GET").
		HandlerFunc(impl.clusterRestHandler.ClusterListFromACD)

	clusterRouter.Path("/acd/delete").
		Methods("DELETE").
		HandlerFunc(impl.clusterRestHandler.DeleteClusterFromACD)

	clusterRouter.Path("/autocomplete").
		Methods("GET").
		HandlerFunc(impl.clusterRestHandler.FindAllForAutoComplete)

	clusterRouter.Path("/component/install/{clusterId}").
		Methods("POST").
		HandlerFunc(impl.clusterRestHandler.DefaultComponentInstallation)
}
