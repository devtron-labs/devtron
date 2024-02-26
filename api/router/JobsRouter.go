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
