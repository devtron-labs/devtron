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

type ClusterHelmConfigRouter interface {
	InitClusterHelmConfigRouter(clusterHelmConfigRouter *mux.Router)
}

type ClusterHelmConfigRouterImpl struct {
	clusterHelmConfigRestHandler restHandler.ClusterHelmConfigRestHandler
}

func NewClusterHelmConfigRouterImpl(handler restHandler.ClusterHelmConfigRestHandler) *ClusterHelmConfigRouterImpl {
	return &ClusterHelmConfigRouterImpl{
		clusterHelmConfigRestHandler: handler,
	}
}

func (impl ClusterHelmConfigRouterImpl) InitClusterHelmConfigRouter(clusterHelmConfigRouter *mux.Router) {
	clusterHelmConfigRouter.PathPrefix("/").
		Methods("POST").
		HandlerFunc(impl.clusterHelmConfigRestHandler.Save)
	clusterHelmConfigRouter.PathPrefix("/").
		Methods("GET").
		Queries("environment", "{environment}").
		HandlerFunc(impl.clusterHelmConfigRestHandler.GetByEnvironment)
}
