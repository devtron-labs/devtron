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

package app

import (
	"github.com/devtron-labs/devtron/api/restHandler/app/appList"
	workflow2 "github.com/devtron-labs/devtron/api/restHandler/app/workflow"
	"github.com/devtron-labs/devtron/api/router/app/appInfo"
	appList2 "github.com/devtron-labs/devtron/api/router/app/appList"
	pipeline2 "github.com/devtron-labs/devtron/api/router/app/pipeline"
	"github.com/devtron-labs/devtron/api/router/app/pipeline/configure"
	"github.com/devtron-labs/devtron/api/router/app/pipeline/history"
	"github.com/devtron-labs/devtron/api/router/app/pipeline/status"
	"github.com/devtron-labs/devtron/api/router/app/pipeline/trigger"
	"github.com/devtron-labs/devtron/api/router/app/workflow"
	"github.com/gorilla/mux"
)

type AppRouter interface {
	InitAppRouter(helmRouter *mux.Router)
}

type AppRouterImpl struct {
	appFilteringRouter           appList2.AppFilteringRouter
	appListingRouter             appList2.AppListingRouter
	appInfoRouter                appInfo.AppInfoRouter
	helmRouter                   trigger.PipelineTriggerRouter
	pipelineConfigRouter         configure.PipelineConfigRouter
	pipelineHistoryRouter        history.PipelineHistoryRouter
	pipelineStatusRouter         status.PipelineStatusRouter
	appWorkflowRouter            workflow.AppWorkflowRouter
	devtronAppAutoCompleteRouter pipeline2.DevtronAppAutoCompleteRouter

	// TODO remove these dependencies after migration
	appWorkflowRestHandler  workflow2.AppWorkflowRestHandler
	appListingRestHandler   appList.AppListingRestHandler
	appFilteringRestHandler appList.AppFilteringRestHandler
}

func NewAppRouterImpl(appFilteringRouter appList2.AppFilteringRouter,
	appListingRouter appList2.AppListingRouter,
	appInfoRouter appInfo.AppInfoRouter,
	helmRouter trigger.PipelineTriggerRouter,
	pipelineConfigRouter configure.PipelineConfigRouter,
	pipelineHistoryRouter history.PipelineHistoryRouter,
	pipelineStatusRouter status.PipelineStatusRouter,
	appWorkflowRouter workflow.AppWorkflowRouter,
	devtronAppAutoCompleteRouter pipeline2.DevtronAppAutoCompleteRouter,
	appWorkflowRestHandler workflow2.AppWorkflowRestHandler,
	appListingRestHandler appList.AppListingRestHandler,
	appFilteringRestHandler appList.AppFilteringRestHandler) *AppRouterImpl {
	router := &AppRouterImpl{
		appInfoRouter:                appInfoRouter,
		helmRouter:                   helmRouter,
		appFilteringRouter:           appFilteringRouter,
		appListingRouter:             appListingRouter,
		pipelineConfigRouter:         pipelineConfigRouter,
		pipelineHistoryRouter:        pipelineHistoryRouter,
		pipelineStatusRouter:         pipelineStatusRouter,
		appWorkflowRouter:            appWorkflowRouter,
		devtronAppAutoCompleteRouter: devtronAppAutoCompleteRouter,
		appWorkflowRestHandler:       appWorkflowRestHandler,
		appListingRestHandler:        appListingRestHandler,
		appFilteringRestHandler:      appFilteringRestHandler,
	}
	return router
}

func (router AppRouterImpl) InitAppRouter(AppRouter *mux.Router) {
	router.appInfoRouter.InitAppInfoRouter(AppRouter)
	router.pipelineConfigRouter.InitPipelineConfigRouter(AppRouter)
	router.helmRouter.InitPipelineTriggerRouter(AppRouter)
	router.devtronAppAutoCompleteRouter.InitDevtronAppAutoCompleteRouter(AppRouter)

	appFilterRouter := AppRouter.PathPrefix("/filter").Subrouter()
	router.appFilteringRouter.InitAppFilteringRouter(appFilterRouter)

	appListRouter := AppRouter.PathPrefix("/list").Subrouter()
	router.appListingRouter.InitAppListingRouter(appListRouter)

	pipelineHistoryRouter := AppRouter.PathPrefix("/history").Subrouter()
	router.pipelineHistoryRouter.InitPipelineHistoryRouter(pipelineHistoryRouter)

	deploymentStatusRouter := AppRouter.PathPrefix("/deployment-status").Subrouter()
	router.pipelineStatusRouter.InitPipelineStatusRouter(deploymentStatusRouter)

	appWorkflowRouter := AppRouter.PathPrefix("/app-wf").Subrouter()
	router.appWorkflowRouter.InitAppWorkflowRouter(appWorkflowRouter)

	// TODO refactoring: categorise and move to respective folders
	AppRouter.Path("/allApps").
		HandlerFunc(router.appListingRestHandler.FetchAllDevtronManagedApps).
		Methods("GET")

	AppRouter.Path("/resource/urls").
		Queries("envId", "{envId}").
		HandlerFunc(router.appListingRestHandler.GetHostUrlsByBatch).
		Methods("GET")

	//This API used for fetch app details, not deployment details
	AppRouter.Path("/detail").
		Queries("app-id", "{app-id}").
		Queries("env-id", "{env-id}").
		HandlerFunc(router.appListingRestHandler.FetchAppDetails).
		Methods("GET")

	AppRouter.Path("/detail/v2").Queries("app-id", "{app-id}").
		Queries("env-id", "{env-id}").
		HandlerFunc(router.appListingRestHandler.FetchAppDetailsV2).
		Methods("GET")

	AppRouter.Path("/detail/resource-tree").Queries("app-id", "{app-id}").
		Queries("env-id", "{env-id}").
		HandlerFunc(router.appListingRestHandler.FetchResourceTree).
		Methods("GET")

	AppRouter.Path("/stage/status").Queries("app-id", "{app-id}").
		HandlerFunc(router.appListingRestHandler.FetchAppStageStatus).
		Methods("GET")

	AppRouter.Path("/other-env").Queries("app-id", "{app-id}").
		HandlerFunc(router.appListingRestHandler.FetchOtherEnvironment).
		Methods("GET")

	AppRouter.Path("/other-env/min").Queries("app-id", "{app-id}").
		HandlerFunc(router.appListingRestHandler.FetchMinDetailOtherEnvironment).
		Methods("GET")

	AppRouter.Path("/linkouts/{Id}/{appId}/{envId}").Queries("podName", "{podName}").
		Queries("containerName", "{containerName}").
		HandlerFunc(router.appListingRestHandler.RedirectToLinkouts).
		Methods("GET")

	// TODO refactoring: migrate
	AppRouter.Path("/app-listing/autocomplete").
		HandlerFunc(router.appFilteringRestHandler.GetClusterTeamAndEnvListForAutocomplete).
		Methods("GET") // deprecated; use filter/autocomplete instead.

	// TODO refactoring: migrate
	AppRouter.Path("/wf/all/component-names/{appId}").
		HandlerFunc(router.appWorkflowRestHandler.FindAllWorkflows).
		Methods("GET") // deprecated; use /app-wf/all/component-names/{appId} instead.
}
