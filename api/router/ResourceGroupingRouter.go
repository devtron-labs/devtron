package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/api/restHandler/app/pipeline/configure"
	"github.com/devtron-labs/devtron/api/restHandler/app/workflow"
	"github.com/gorilla/mux"
)

type ResourceGroupingRouter interface {
	InitResourceGroupingRouter(router *mux.Router)
}
type ResourceGroupingRouterImpl struct {
	pipelineConfigRestHandler configure.PipelineConfigRestHandler
	appWorkflowRestHandler    workflow.AppWorkflowRestHandler
	resourceGroupRestHandler  restHandler.ResourceGroupRestHandler
}

func NewResourceGroupingRouterImpl(restHandler configure.PipelineConfigRestHandler,
	appWorkflowRestHandler workflow.AppWorkflowRestHandler,
	resourceGroupRestHandler restHandler.ResourceGroupRestHandler) *ResourceGroupingRouterImpl {
	return &ResourceGroupingRouterImpl{
		pipelineConfigRestHandler: restHandler,
		appWorkflowRestHandler:    appWorkflowRestHandler,
		resourceGroupRestHandler:  resourceGroupRestHandler,
	}
}

func (router ResourceGroupingRouterImpl) InitResourceGroupingRouter(resourceGroupingRouter *mux.Router) {
	resourceGroupingRouter.Path("/{envId}/app-wf").
		HandlerFunc(router.appWorkflowRestHandler.FindAppWorkflowByEnvironment).Methods("GET")
	resourceGroupingRouter.Path("/{envId}/ci-pipeline").HandlerFunc(router.pipelineConfigRestHandler.GetCiPipelineByEnvironment).Methods("GET")
	resourceGroupingRouter.Path("/{envId}/cd-pipeline").HandlerFunc(router.pipelineConfigRestHandler.GetCdPipelinesByEnvironment).Methods("GET")
	resourceGroupingRouter.Path("/{envId}/external-ci").HandlerFunc(router.pipelineConfigRestHandler.GetExternalCiByEnvironment).Methods("GET")
	resourceGroupingRouter.Path("/{envId}/workflow/status").HandlerFunc(router.pipelineConfigRestHandler.FetchAppWorkflowStatusForTriggerViewByEnvironment).Methods("GET")
	resourceGroupingRouter.Path("/app-grouping").HandlerFunc(router.pipelineConfigRestHandler.GetEnvironmentListWithAppData).Methods("GET")
	resourceGroupingRouter.Path("/{envId}/applications").HandlerFunc(router.pipelineConfigRestHandler.GetApplicationsByEnvironment).Methods("GET")
	resourceGroupingRouter.Path("/{envId}/deployment/status").HandlerFunc(router.pipelineConfigRestHandler.FetchAppDeploymentStatusForEnvironments).Methods("GET")
	resourceGroupingRouter.Path("/{envId}/ci-pipeline/min").HandlerFunc(router.pipelineConfigRestHandler.GetCiPipelineByEnvironmentMin).Methods("GET")
	resourceGroupingRouter.Path("/{envId}/cd-pipeline/min").HandlerFunc(router.pipelineConfigRestHandler.GetCdPipelinesByEnvironmentMin).Methods("GET")

	resourceGroupingRouter.Path("/{resourceId}/group").HandlerFunc(router.resourceGroupRestHandler.CreateResourceGroup).Methods("POST")
	resourceGroupingRouter.Path("/{resourceId}/group").HandlerFunc(router.resourceGroupRestHandler.UpdateResourceGroup).Methods("PUT")
	//resourceGroupingRouter.Path("/{envId}/group/{resourceGroupId}").HandlerFunc(router.resourceGroupRestHandler.GetApplicationsForResourceGroup).Methods("GET")
	resourceGroupingRouter.Path("/{resourceId}/groups").Queries("groupType", "{groupType}").HandlerFunc(router.resourceGroupRestHandler.GetActiveResourceGroupList).Methods("GET")
	resourceGroupingRouter.Path("/{resourceId}/groups").HandlerFunc(router.resourceGroupRestHandler.GetActiveResourceGroupList).Methods("GET")

	resourceGroupingRouter.Path("/{resourceId}/group/{resourceGroupId}").Queries("groupType", "{groupType}").HandlerFunc(router.resourceGroupRestHandler.DeleteResourceGroup).Methods("DELETE")
	resourceGroupingRouter.Path("/{resourceId}/group/{resourceGroupId}").HandlerFunc(router.resourceGroupRestHandler.DeleteResourceGroup).Methods("DELETE")

	resourceGroupingRouter.Path("/{resourceId}/group/permission/check").HandlerFunc(router.resourceGroupRestHandler.CheckResourceGroupPermissions).Methods("POST")
}
