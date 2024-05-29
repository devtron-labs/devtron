/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

	appRouter.Path("/manual-sync/{appId}/{envId}").
		HandlerFunc(router.pipelineStatusTimelineRestHandler.ManualSyncAcdPipelineDeploymentStatus).
		Methods("GET")
}
