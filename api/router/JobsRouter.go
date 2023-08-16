package router

import (
	"github.com/devtron-labs/devtron/api/logger"
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
	userAuth                  logger.UserAuth
}

func NewJobRouterImpl(pipelineConfigRestHandler app.PipelineConfigRestHandler, appListingRestHandler restHandler.AppListingRestHandler, userAuth logger.UserAuth) *JobRouterImpl {
	return &JobRouterImpl{
		appListingRestHandler:     appListingRestHandler,
		pipelineConfigRestHandler: pipelineConfigRestHandler,
		userAuth:                  userAuth,
	}
	//return router
}
func (router JobRouterImpl) InitJobRouter(jobRouter *mux.Router) {
	jobRouter.Use(router.userAuth.LoggingMiddleware)
	jobRouter.Path("").HandlerFunc(router.pipelineConfigRestHandler.CreateApp).Methods("POST")
	jobRouter.Path("/list").HandlerFunc(router.appListingRestHandler.FetchJobs).Methods("POST")
	jobRouter.Path("/ci-pipeline/list/{jobId}").HandlerFunc(router.appListingRestHandler.FetchJobOverviewCiPipelines).Methods("GET")
}
