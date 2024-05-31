/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package cluster

import (
	"github.com/gorilla/mux"
)

type ClusterRouter interface {
	InitClusterRouter(clusterRouter *mux.Router)
}

type ClusterRouterImpl struct {
	clusterRestHandler ClusterRestHandler
}

func NewClusterRouterImpl(handler ClusterRestHandler) *ClusterRouterImpl {
	return &ClusterRouterImpl{
		clusterRestHandler: handler,
	}
}

func (impl ClusterRouterImpl) InitClusterRouter(clusterRouter *mux.Router) {
	clusterRouter.Path("").
		Methods("POST").
		HandlerFunc(impl.clusterRestHandler.Save)

	clusterRouter.Path("/virtual").
		Methods("POST").
		HandlerFunc(impl.clusterRestHandler.SaveVirtualCluster)

	clusterRouter.Path("/saveClusters").
		Methods("POST").
		HandlerFunc(impl.clusterRestHandler.SaveClusters)

	clusterRouter.Path("/validate").
		Methods("POST").
		HandlerFunc(impl.clusterRestHandler.ValidateKubeconfig)

	clusterRouter.Path("").
		Methods("GET").
		Queries("id", "{id}").
		HandlerFunc(impl.clusterRestHandler.FindById)

	clusterRouter.Path("/description").
		Methods("GET").
		Queries("id", "{id}").
		HandlerFunc(impl.clusterRestHandler.FindNoteByClusterId)

	clusterRouter.Path("").
		Methods("GET").
		HandlerFunc(impl.clusterRestHandler.FindAll)

	clusterRouter.Path("").
		Methods("PUT").
		HandlerFunc(impl.clusterRestHandler.Update)

	clusterRouter.Path("/note").
		Methods("PUT").
		HandlerFunc(impl.clusterRestHandler.UpdateClusterNote)

	clusterRouter.Path("/virtual").
		Methods("PUT").
		HandlerFunc(impl.clusterRestHandler.UpdateVirtualCluster)

	clusterRouter.Path("/autocomplete").
		Methods("GET").
		HandlerFunc(impl.clusterRestHandler.FindAllForAutoComplete)

	clusterRouter.Path("/namespaces/{clusterId}").
		Methods("GET").
		HandlerFunc(impl.clusterRestHandler.GetClusterNamespaces)

	clusterRouter.Path("/namespaces").
		Methods("GET").
		HandlerFunc(impl.clusterRestHandler.GetAllClusterNamespaces)

	clusterRouter.Path("").
		Methods("DELETE").
		HandlerFunc(impl.clusterRestHandler.DeleteCluster)

	clusterRouter.Path("/virtual").
		Methods("DELETE").
		HandlerFunc(impl.clusterRestHandler.DeleteVirtualCluster)

	clusterRouter.Path("/auth-list").
		Methods("GET").
		HandlerFunc(impl.clusterRestHandler.FindAllForClusterPermission)

	clusterRouter.Path("/description").
		Methods("PUT").
		HandlerFunc(impl.clusterRestHandler.UpdateClusterDescription)
}
