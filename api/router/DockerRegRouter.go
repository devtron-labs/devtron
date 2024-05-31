/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
	configRouter.Path("/registry/validate").
		HandlerFunc(impl.dockerRestHandler.ValidateDockerRegistryConfig).
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
