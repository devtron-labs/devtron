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
}
