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

type DockerRegRouter interface {
	InitDockerRegRouter(gocdRouter *mux.Router)
}
type DockerRegRouterImpl struct {
	dockerRestHandler restHandler.DockerRegRestHandler
}

func NewDockerRegRouterImpl(dockerRestHandler restHandler.DockerRegRestHandler) *DockerRegRouterImpl {
	return &DockerRegRouterImpl{dockerRestHandler: dockerRestHandler}
}
func (impl DockerRegRouterImpl) InitDockerRegRouter(configRouter *mux.Router) {
	configRouter.Path("/registry").
		HandlerFunc(impl.dockerRestHandler.SaveDockerRegistryConfig).
		Methods("POST")
	configRouter.Path("/registry/active").
		HandlerFunc(impl.dockerRestHandler.GetDockerArtifactStore).
		Methods("GET")
	configRouter.Path("/registry").
		HandlerFunc(impl.dockerRestHandler.FetchAllDockerAccounts).
		Methods("GET")
	configRouter.Path("/registry/autocomplete").
		HandlerFunc(impl.dockerRestHandler.FetchAllDockerRegistryForAutocomplete).
		Methods("GET")
	configRouter.Path("/registry/{id}").
		HandlerFunc(impl.dockerRestHandler.FetchOneDockerAccounts).
		Methods("GET")
	configRouter.Path("/registry").
		HandlerFunc(impl.dockerRestHandler.UpdateDockerRegistryConfig).
		Methods("PUT")
	configRouter.Path("/registry/configure/status").
		HandlerFunc(impl.dockerRestHandler.IsDockerRegConfigured).
		Methods("GET")
	configRouter.Path("/registry").
		HandlerFunc(impl.dockerRestHandler.DeleteDockerRegistryConfig).
		Methods("DELETE")
}
