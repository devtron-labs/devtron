/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type DeploymentGroupRouter interface {
	initDeploymentGroupRouter(configRouter *mux.Router)
}
type DeploymentGroupRouterImpl struct {
	restHandler restHandler.DeploymentGroupRestHandler
}

func NewDeploymentGroupRouterImpl(restHandler restHandler.DeploymentGroupRestHandler) *DeploymentGroupRouterImpl {
	return &DeploymentGroupRouterImpl{restHandler: restHandler}

}

func (router DeploymentGroupRouterImpl) initDeploymentGroupRouter(configRouter *mux.Router) {
	configRouter.Path("/dg/create").HandlerFunc(router.restHandler.CreateDeploymentGroup).Methods("POST")
	configRouter.Path("/dg/fetch/ci/{deploymentGroupId}").HandlerFunc(router.restHandler.FetchParentCiForDG).Methods("GET")
	configRouter.Path("/dg/fetch/env/apps/{ciPipelineId}").HandlerFunc(router.restHandler.FetchEnvApplicationsForDG).Methods("GET")
	configRouter.Path("/dg/fetch/all").HandlerFunc(router.restHandler.FetchAllDeploymentGroups).Methods("GET")
	configRouter.Path("/dg/delete/{id}").HandlerFunc(router.restHandler.DeleteDeploymentGroup).Methods("DELETE")
	configRouter.Path("/release/trigger").HandlerFunc(router.restHandler.TriggerReleaseForDeploymentGroup).Methods("POST")
	configRouter.Path("/dg/update").HandlerFunc(router.restHandler.UpdateDeploymentGroup).Methods("PUT")
	configRouter.Path("/dg/material/{deploymentGroupId}").HandlerFunc(router.restHandler.GetArtifactsByCiPipeline).Methods("GET")
	configRouter.Path("/dg/{deploymentGroupId}").HandlerFunc(router.restHandler.GetDeploymentGroupById).Methods("GET")

}
