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

package status

import (
	"github.com/devtron-labs/devtron/api/restHandler/app/pipeline/status"
	"github.com/gorilla/mux"
)

type PipelineStatusRouter interface {
	InitPipelineStatusRouter(appRouter *mux.Router)
}
type PipelineStatusRouterImpl struct {
	pipelineStatusTimelineRestHandler status.PipelineStatusTimelineRestHandler
}

func NewPipelineStatusRouterImpl(pipelineStatusTimelineRestHandler status.PipelineStatusTimelineRestHandler) *PipelineStatusRouterImpl {
	return &PipelineStatusRouterImpl{
		pipelineStatusTimelineRestHandler: pipelineStatusTimelineRestHandler,
	}
}

func (router PipelineStatusRouterImpl) InitPipelineStatusRouter(appRouter *mux.Router) {
	appRouter.Path("/timeline/{appId}/{envId}").
		HandlerFunc(router.pipelineStatusTimelineRestHandler.FetchTimelines).
		Methods("GET")

	appRouter.Path("/deployment-status/manual-sync/{appId}/{envId}").
		HandlerFunc(router.pipelineStatusTimelineRestHandler.ManualSyncAcdPipelineDeploymentStatus).
		Methods("GET")
}
