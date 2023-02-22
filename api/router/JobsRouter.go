package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/api/restHandler/app"
	"github.com/gorilla/mux"
)

type JobRouter interface {
	InitJobRouter(router *mux.Router)
}
type JobRouterImpl struct {
	pipelineConfigRestHandler app.PipelineConfigRestHandler
	appListingRestHandler     restHandler.AppListingRestHandler
}

func NewJobRouterImpl(pipelineConfigRestHandler app.PipelineConfigRestHandler, appListingRestHandler restHandler.AppListingRestHandler) *JobRouterImpl {
	return &JobRouterImpl{
		appListingRestHandler:     appListingRestHandler,
		pipelineConfigRestHandler: pipelineConfigRestHandler,
	}
	//return router
}
func (router JobRouterImpl) InitJobRouter(jobRouter *mux.Router) {
	jobRouter.Path("").HandlerFunc(router.pipelineConfigRestHandler.CreateJob).Methods("POST")
	jobRouter.Path("/ci-pipeline/patch").HandlerFunc(router.pipelineConfigRestHandler.PatchCiPipelines).Methods("POST")
	jobRouter.Path("/list").HandlerFunc(router.appListingRestHandler.FetchJobs).Methods("POST")
	jobRouter.Path("/ci-pipeline/list/{jobId}").HandlerFunc(router.appListingRestHandler.FetchOverviewCiPipelines).Methods("GET")
}
