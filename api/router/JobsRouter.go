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

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler/app/appList"
	"github.com/devtron-labs/devtron/api/restHandler/app/pipeline/configure"
	"github.com/gorilla/mux"
)

type JobRouter interface {
	InitJobRouter(router *mux.Router)
}
type JobRouterImpl struct {
	pipelineConfigRestHandler configure.PipelineConfigRestHandler
	appListingRestHandler     appList.AppListingRestHandler
}

func NewJobRouterImpl(pipelineConfigRestHandler configure.PipelineConfigRestHandler, appListingRestHandler appList.AppListingRestHandler) *JobRouterImpl {
	return &JobRouterImpl{
		appListingRestHandler:     appListingRestHandler,
		pipelineConfigRestHandler: pipelineConfigRestHandler,
	}
	//return router
}
func (router JobRouterImpl) InitJobRouter(jobRouter *mux.Router) {
	jobRouter.Path("").HandlerFunc(router.pipelineConfigRestHandler.CreateApp).Methods("POST")
	jobRouter.Path("/list").HandlerFunc(router.appListingRestHandler.FetchJobs).Methods("POST")
	jobRouter.Path("/ci-pipeline/list/{jobId}").HandlerFunc(router.appListingRestHandler.FetchJobOverviewCiPipelines).Methods("GET")
}
