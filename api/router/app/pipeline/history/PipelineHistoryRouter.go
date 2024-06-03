/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
