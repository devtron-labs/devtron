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
