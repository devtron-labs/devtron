/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package workflow

import (
	"github.com/devtron-labs/devtron/api/restHandler/app/workflow"
	"github.com/gorilla/mux"
)

type AppWorkflowRouter interface {
	InitAppWorkflowRouter(appRouter *mux.Router)
}
type AppWorkflowRouterImpl struct {
	appWorkflowRestHandler workflow.AppWorkflowRestHandler
}

func NewAppWorkflowRouterImpl(appWorkflowRestHandler workflow.AppWorkflowRestHandler) *AppWorkflowRouterImpl {
	return &AppWorkflowRouterImpl{
		appWorkflowRestHandler: appWorkflowRestHandler,
	}
}

func (router AppWorkflowRouterImpl) InitAppWorkflowRouter(appRouter *mux.Router) {
	appRouter.Path("").
		HandlerFunc(router.appWorkflowRestHandler.CreateAppWorkflow).Methods("POST")

	appRouter.Path("/{app-id}").
		HandlerFunc(router.appWorkflowRestHandler.FindAppWorkflow).Methods("GET")

	appRouter.Path("/view/{app-id}").
		HandlerFunc(router.appWorkflowRestHandler.GetWorkflowsViewData).Methods("GET")

	appRouter.Path("/{app-id}/{app-wf-id}").
		HandlerFunc(router.appWorkflowRestHandler.DeleteAppWorkflow).Methods("DELETE")

	appRouter.Path("/all").
		HandlerFunc(router.appWorkflowRestHandler.FindAllWorkflowsForApps).Methods("POST")

	appRouter.Path("/all/component-names/{appId}").
		HandlerFunc(router.appWorkflowRestHandler.FindAllWorkflows).Methods("GET")

	appRouter.Path("/list-components/{appName}").
		HandlerFunc(router.appWorkflowRestHandler.FindAllComponentsByAppName).Methods("GET")
}
