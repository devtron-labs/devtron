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

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type ApplicationRouter interface {
	initApplicationRouter(router *mux.Router)
}

type ApplicationRouterImpl struct {
	handler restHandler.ArgoApplicationRestHandler
	logger  *zap.SugaredLogger
}

func NewApplicationRouterImpl(handler restHandler.ArgoApplicationRestHandler, logger *zap.SugaredLogger) *ApplicationRouterImpl {
	return &ApplicationRouterImpl{
		handler: handler,
		logger:  logger,
	}
}

func (router *ApplicationRouterImpl) initApplicationRouter(applicationRouter *mux.Router) {

	applicationRouter.Path("/stream").
		Queries("name", "{name}").
		Methods("GET").
		HandlerFunc(router.handler.Watch)

	applicationRouter.Path("/{name}/pods/{podName}/logs").
		Queries("container", "{container}", "namespace", "{namespace}").
		Queries("follow", "{follow}").
		Queries("sinceSeconds", "{sinceSeconds}").
		Queries("sinceTime.seconds", "{sinceTime.seconds}").
		Queries("sinceTime.nanos", "{sinceTime.nanos}").
		Queries("tailLines", "{tailLines}").
		Methods("GET").
		HandlerFunc(router.handler.GetPodLogs)
	applicationRouter.Path("/{name}/pods/{podName}/logs").
		Methods("GET").
		HandlerFunc(router.handler.GetPodLogs)
	applicationRouter.Path("/{name}/resource-tree").
		Methods("GET").
		HandlerFunc(router.handler.GetResourceTree)
	applicationRouter.Path("/{name}/resource").
		Queries("version", "{version}", "namespace", "{namespace}", "group", "{group}", "kind", "{kind}", "resourceName", "{resourceName}").
		Methods("GET").
		HandlerFunc(router.handler.GetResource)
	applicationRouter.Path("/{name}/events").
		Queries("resourceNamespace", "{resourceNamespace}", "resourceUID", "{resourceUID}", "resourceName", "{resourceName}").
		Methods("GET").
		HandlerFunc(router.handler.ListResourceEvents)
	applicationRouter.Path("/{name}/events").
		Methods("GET").
		HandlerFunc(router.handler.ListResourceEvents)
	applicationRouter.Path("/").
		Queries("name", "{name}", "refresh", "{refresh}", "project", "{project}").
		Methods("GET").
		HandlerFunc(router.handler.List)
	applicationRouter.Path("/{applicationName}/managed-resources").
		Methods("GET").
		HandlerFunc(router.handler.ManagedResources)
	applicationRouter.Path("/{name}/rollback").
		Methods("GET").
		HandlerFunc(router.handler.Rollback)

	applicationRouter.Path("/{name}/manifests").
		Methods("GET").
		HandlerFunc(router.handler.GetManifests)
	applicationRouter.Path("/{name}").
		Methods("GET").
		HandlerFunc(router.handler.Get)
	applicationRouter.Path("/{appName}/operation").
		Methods("DELETE").
		HandlerFunc(router.handler.TerminateOperation)
	applicationRouter.Path("/{name}/resource").
		Methods("POST").
		HandlerFunc(router.handler.PatchResource)
	applicationRouter.Path("/{appNameACD}/resource").
		Queries("name", "{name}", "namespace", "{namespace}", "resourceName", "{resourceName}", "version", "{version}",
			"force", "{force}", "appId", "{appId}", "envId", "{envId}", "group", "{group}", "kind", "{kind}").
		Methods("DELETE").
		HandlerFunc(router.handler.DeleteResource)

	applicationRouter.Path("/{name}/service-link").
		Methods("GET").
		HandlerFunc(router.handler.GetServiceLink)
	applicationRouter.Path("/pod/exec/session/{appId}/{environmentId}/{namespace}/{pod}/{shell}/{container}").
		Methods("GET").
		HandlerFunc(router.handler.GetTerminalSession)
	applicationRouter.Path("/pod/exec/sockjs/ws/").Handler(terminal.CreateAttachHandler("/api/v1/applications/pod/exec/sockjs/ws/"))
}
