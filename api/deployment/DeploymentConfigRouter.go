/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package deployment

import (
	"github.com/gorilla/mux"
)

type DeploymentConfigRouter interface {
	Init(configRouter *mux.Router)
}

type DeploymentConfigRouterImpl struct {
	deploymentRestHandler DeploymentConfigRestHandler
}

func NewDeploymentRouterImpl(deploymentRestHandler DeploymentConfigRestHandler) *DeploymentConfigRouterImpl {
	return &DeploymentConfigRouterImpl{
		deploymentRestHandler: deploymentRestHandler,
	}
}

func (router DeploymentConfigRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("/validate").
		HandlerFunc(router.deploymentRestHandler.CreateChartFromFile).Methods("POST")
	configRouter.Path("/upload").
		HandlerFunc(router.deploymentRestHandler.SaveChart).Methods("PUT")
	configRouter.Path("/download/{chartRefId}").
		HandlerFunc(router.deploymentRestHandler.DownloadChart).Methods("GET")
	configRouter.Path("/fetch").
		HandlerFunc(router.deploymentRestHandler.GetUploadedCharts).Methods("GET")
}
