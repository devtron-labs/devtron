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

package history

import (
	"github.com/devtron-labs/devtron/api/restHandler/app/pipeline/history"
	"github.com/gorilla/mux"
)

type PipelineHistoryRouter interface {
	InitPipelineHistoryRouter(appRouter *mux.Router)
}
type PipelineHistoryRouterImpl struct {
	pipelineHistoryRestHandler history.PipelineHistoryRestHandler
}

func NewPipelineHistoryRouterImpl(pipelineHistoryRestHandler history.PipelineHistoryRestHandler) *PipelineHistoryRouterImpl {
	return &PipelineHistoryRouterImpl{
		pipelineHistoryRestHandler: pipelineHistoryRestHandler,
	}
}

func (router PipelineHistoryRouterImpl) InitPipelineHistoryRouter(appRouter *mux.Router) {
	appRouter.Path("/deployed-configuration/{appId}/{pipelineId}/{wfrId}").
		HandlerFunc(router.pipelineHistoryRestHandler.FetchDeployedConfigurationsForWorkflow).
		Methods("GET")

	appRouter.Path("/deployed-component/list/{appId}/{pipelineId}").
		HandlerFunc(router.pipelineHistoryRestHandler.FetchDeployedHistoryComponentList).
		Methods("GET")

	appRouter.Path("/deployed-component/detail/{appId}/{pipelineId}/{id}").
		HandlerFunc(router.pipelineHistoryRestHandler.FetchDeployedHistoryComponentDetail).
		Methods("GET")

	appRouter.Path("/deployed-configuration/latest/deployed/{appId}/{pipelineId}").
		HandlerFunc(router.pipelineHistoryRestHandler.GetAllDeployedConfigurationHistoryForLatestWfrIdForPipeline).
		Methods("GET")

	appRouter.Path("/deployed-configuration/all/{appId}/{pipelineId}/{wfrId}").
		HandlerFunc(router.pipelineHistoryRestHandler.GetAllDeployedConfigurationHistoryForSpecificWfrIdForPipeline).
		Methods("GET")
}
