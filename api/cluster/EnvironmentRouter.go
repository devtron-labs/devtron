/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
	environmentClusterMappingsRouter.Path("/virtual").
		Methods("POST").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.CreateVirtualEnvironment)
	environmentClusterMappingsRouter.Path("").
		Methods("PUT").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.Update)
	environmentClusterMappingsRouter.Path("/virtual").
		Methods("PUT").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.UpdateVirtualEnvironment)
	environmentClusterMappingsRouter.Path("/autocomplete").
		Methods("GET").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.GetEnvironmentListForAutocomplete)
	environmentClusterMappingsRouter.Path("/autocomplete/helm").
		Methods("GET").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.GetCombinedEnvironmentListForDropDown)
	environmentClusterMappingsRouter.Path("").
		Methods("DELETE").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.DeleteEnvironment)
	environmentClusterMappingsRouter.Path("/namespace/autocomplete").
		Methods("GET").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.GetCombinedEnvironmentListForDropDownByClusterIds)
	environmentClusterMappingsRouter.Path("/{envId}/connection").
		Methods("GET").
		HandlerFunc(impl.environmentClusterMappingsRestHandler.GetEnvironmentConnection)
}
