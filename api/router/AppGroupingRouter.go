package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/api/restHandler/app"
	"github.com/gorilla/mux"
)

type AppGroupingRouter interface {
	InitAppGroupingRouter(router *mux.Router)
}
type AppGroupingRouterImpl struct {
	restHandler            app.PipelineConfigRestHandler
	appWorkflowRestHandler restHandler.AppWorkflowRestHandler
	appGroupRestHandler    restHandler.AppGroupRestHandler
}

func NewAppGroupingRouterImpl(restHandler app.PipelineConfigRestHandler,
	appWorkflowRestHandler restHandler.AppWorkflowRestHandler,
	appGroupRestHandler restHandler.AppGroupRestHandler) *AppGroupingRouterImpl {
	return &AppGroupingRouterImpl{
		restHandler:            restHandler,
		appWorkflowRestHandler: appWorkflowRestHandler,
		appGroupRestHandler:    appGroupRestHandler,
	}
}

func (router AppGroupingRouterImpl) InitAppGroupingRouter(appGroupingRouter *mux.Router) {
	appGroupingRouter.Path("/{envId}/app-wf").
		HandlerFunc(router.appWorkflowRestHandler.FindAppWorkflowByEnvironment).Methods("GET")
	appGroupingRouter.Path("/{envId}/ci-pipeline").HandlerFunc(router.restHandler.GetCiPipelineByEnvironment).Methods("GET")
	appGroupingRouter.Path("/{envId}/cd-pipeline").HandlerFunc(router.restHandler.GetCdPipelinesByEnvironment).Methods("GET")
	appGroupingRouter.Path("/{envId}/cd-pipeline/v2").HandlerFunc(router.restHandler.GetCdPipelinesByEnvironment).Methods("GET")
	appGroupingRouter.Path("/{envId}/external-ci").HandlerFunc(router.restHandler.GetExternalCiByEnvironment).Methods("GET")
	appGroupingRouter.Path("/{envId}/workflow/status").HandlerFunc(router.restHandler.FetchAppWorkflowStatusForTriggerViewByEnvironment).Methods("GET")
	appGroupingRouter.Path("/app-grouping").HandlerFunc(router.restHandler.GetEnvironmentListWithAppData).Methods("GET")
	appGroupingRouter.Path("/{envId}/applications").HandlerFunc(router.restHandler.GetApplicationsByEnvironment).Methods("GET")
	appGroupingRouter.Path("/{envId}/deployment/status").HandlerFunc(router.restHandler.FetchAppDeploymentStatusForEnvironments).Methods("GET")

	appGroupingRouter.Path("/{envId}/group").HandlerFunc(router.appGroupRestHandler.CreateAppGroup).Methods("POST")
	appGroupingRouter.Path("/{envId}/group").HandlerFunc(router.appGroupRestHandler.UpdateAppGroup).Methods("PUT")
	appGroupingRouter.Path("/{envId}/group/{appGroupId}").HandlerFunc(router.appGroupRestHandler.GetApplicationsForAppGroup).Methods("GET")
	appGroupingRouter.Path("/{envId}/groups").HandlerFunc(router.appGroupRestHandler.GetActiveAppGroupList).Methods("GET")
	appGroupingRouter.Path("/{envId}/group/{appGroupId}").HandlerFunc(router.appGroupRestHandler.DeleteAppGroup).Methods("DELETE")

	appGroupingRouter.Path("/{envId}/ci-pipeline/min").HandlerFunc(router.restHandler.GetCiPipelineByEnvironmentMin).Methods("GET")
	appGroupingRouter.Path("/{envId}/cd-pipeline/min").HandlerFunc(router.restHandler.GetCdPipelinesByEnvironmentMin).Methods("GET")

}
