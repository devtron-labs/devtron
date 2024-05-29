/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type ConfigMapRouter interface {
	initConfigMapRouter(configRouter *mux.Router)
}
type ConfigMapRouterImpl struct {
	restHandler restHandler.ConfigMapRestHandler
}

func NewConfigMapRouterImpl(restHandler restHandler.ConfigMapRestHandler) *ConfigMapRouterImpl {
	return &ConfigMapRouterImpl{restHandler: restHandler}

}

func (router ConfigMapRouterImpl) initConfigMapRouter(configRouter *mux.Router) {
	configRouter.Path("/global/cm").
		HandlerFunc(router.restHandler.CMGlobalAddUpdate).Methods("POST")
	configRouter.Path("/environment/cm").
		HandlerFunc(router.restHandler.CMEnvironmentAddUpdate).Methods("POST")

	configRouter.Path("/global/cm/{appId}").
		HandlerFunc(router.restHandler.CMGlobalFetch).Methods("GET")
	configRouter.Path("/environment/cm/{appId}/{envId}").
		HandlerFunc(router.restHandler.CMEnvironmentFetch).Methods("GET")
	configRouter.Path("/global/cm/edit/{appId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CMGlobalFetchForEdit).Methods("GET")
	configRouter.Path("/environment/cm/edit/{appId}/{envId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CMEnvironmentFetchForEdit).Methods("GET")

	configRouter.Path("/global/cs").
		HandlerFunc(router.restHandler.CSGlobalAddUpdate).Methods("POST")
	configRouter.Path("/environment/cs").
		HandlerFunc(router.restHandler.CSEnvironmentAddUpdate).Methods("POST")

	configRouter.Path("/global/cs/{appId}").
		HandlerFunc(router.restHandler.CSGlobalFetch).Methods("GET")
	configRouter.Path("/environment/cs/{appId}/{envId}").
		HandlerFunc(router.restHandler.CSEnvironmentFetch).Methods("GET")

	configRouter.Path("/global/cm/{appId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CMGlobalDelete).Methods("DELETE")
	configRouter.Path("/environment/cm/{appId}/{envId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CMEnvironmentDelete).Methods("DELETE")
	configRouter.Path("/global/cs/{appId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CSGlobalDelete).Methods("DELETE")
	configRouter.Path("/environment/cs/{appId}/{envId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CSEnvironmentDelete).Methods("DELETE")

	configRouter.Path("/global/cs/edit/{appId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CSGlobalFetchForEdit).Methods("GET")
	configRouter.Path("/environment/cs/edit/{appId}/{envId}/{id}").
		Queries("name", "{name}").
		HandlerFunc(router.restHandler.CSEnvironmentFetchForEdit).Methods("GET")

	configRouter.Path("/bulk/patch").HandlerFunc(router.restHandler.ConfigSecretBulkPatch).Methods("POST")
	configRouter.Path("/environment").
		HandlerFunc(router.restHandler.AddEnvironmentToJob).Methods("POST")
	configRouter.Path("/environment").
		HandlerFunc(router.restHandler.RemoveEnvironmentFromJob).Methods("DELETE")
	configRouter.Path("/environment/{appId}").
		HandlerFunc(router.restHandler.GetEnvironmentsForJob).Methods("GET")

}
