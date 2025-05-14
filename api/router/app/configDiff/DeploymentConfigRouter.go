/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package configDiff

import (
	"github.com/devtron-labs/devtron/api/restHandler/app/configDiff"
	"github.com/gorilla/mux"
)

type DeploymentConfigurationRouter interface {
	InitDeploymentConfigurationRouter(configRouter *mux.Router)
}

type DeploymentConfigurationRouterImpl struct {
	deploymentGroupRestHandler configDiff.DeploymentConfigurationRestHandler
}

func NewDeploymentConfigurationRouter(deploymentGroupRestHandler configDiff.DeploymentConfigurationRestHandler) *DeploymentConfigurationRouterImpl {
	router := &DeploymentConfigurationRouterImpl{
		deploymentGroupRestHandler: deploymentGroupRestHandler,
	}
	return router
}

func (router DeploymentConfigurationRouterImpl) InitDeploymentConfigurationRouter(configRouter *mux.Router) {
	configRouter.Path("/autocomplete").
		HandlerFunc(router.deploymentGroupRestHandler.ConfigAutoComplete).
		Methods("GET")
	configRouter.Path("/data").
		HandlerFunc(router.deploymentGroupRestHandler.GetConfigData).
		Methods("GET")
	configRouter.Path("/compare/{resource}").
		HandlerFunc(router.deploymentGroupRestHandler.CompareCategoryWiseConfigData).
		Methods("GET")

	configRouter.Path("/manifest").
		HandlerFunc(router.deploymentGroupRestHandler.GetManifest).
		Methods("POST")
}
