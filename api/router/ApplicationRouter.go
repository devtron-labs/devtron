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

func (r ApplicationRouterImpl) initApplicationRouter(router *mux.Router) {

	router.Path("/stream").
		Queries("name", "{name}").
		Methods("GET").
		HandlerFunc(r.handler.Watch)

	router.Path("/{name}/pods/{podName}/logs").
		Queries("container", "{container}", "namespace", "{namespace}").
		Queries("follow", "{follow}").
		Queries("sinceSeconds", "{sinceSeconds}").
		Queries("sinceTime.seconds", "{sinceTime.seconds}").
		Queries("sinceTime.nanos", "{sinceTime.nanos}").
		Queries("tailLines", "{tailLines}").
		Methods("GET").
		HandlerFunc(r.handler.GetPodLogs)
	router.Path("/{name}/pods/{podName}/logs").
		Methods("GET").
		HandlerFunc(r.handler.GetPodLogs)
	router.Path("/{name}/resource-tree").
		Methods("GET").
		HandlerFunc(r.handler.GetResourceTree)
	router.Path("/{name}/resource").
		Queries("version", "{version}", "namespace", "{namespace}", "group", "{group}", "kind", "{kind}", "resourceName", "{resourceName}").
		Methods("GET").
		HandlerFunc(r.handler.GetResource)
	router.Path("/{name}/events").
		Queries("resourceNamespace", "{resourceNamespace}", "resourceUID", "{resourceUID}", "resourceName", "{resourceName}").
		Methods("GET").
		HandlerFunc(r.handler.ListResourceEvents)
	router.Path("/{name}/events").
		Methods("GET").
		HandlerFunc(r.handler.ListResourceEvents)
	router.Path("/").
		Queries("name", "{name}", "refresh", "{refresh}", "project", "{project}").
		Methods("GET").
		HandlerFunc(r.handler.List)
	router.Path("/{applicationName}/managed-resources").
		Methods("GET").
		HandlerFunc(r.handler.ManagedResources)
	router.Path("/{name}/rollback").
		Methods("GET").
		HandlerFunc(r.handler.Rollback)

	router.Path("/{name}/manifests").
		Methods("GET").
		HandlerFunc(r.handler.GetManifests)
	router.Path("/{name}").
		Methods("GET").
		HandlerFunc(r.handler.Get)
	router.Path("/{name}/sync").
		Methods("POST").
		HandlerFunc(r.handler.Sync)
	router.Path("/{appName}/operation").
		Methods("DELETE").
		HandlerFunc(r.handler.TerminateOperation)
	router.Path("/{name}/resource").
		Methods("POST").
		HandlerFunc(r.handler.PatchResource)
	router.Path("/{appNameACD}/resource").
		Queries("name", "{name}", "namespace", "{namespace}", "resourceName", "{resourceName}", "version", "{version}",
			"force", "{force}", "appId", "{appId}", "envId", "{envId}", "group", "{group}", "kind", "{kind}").
		Methods("DELETE").
		HandlerFunc(r.handler.DeleteResource)

	router.Path("/{name}/service-link").
		Methods("GET").
		HandlerFunc(r.handler.GetServiceLink)
	router.Path("/pod/exec/session/{appId}/{environmentId}/{namespace}/{pod}/{shell}/{container}").
		Methods("GET").
		HandlerFunc(r.handler.GetTerminalSession)
	router.Path("/pod/exec/sockjs/ws/").Handler(terminal.CreateAttachHandler("/api/v1/applications/pod/exec/sockjs/ws/"))
}
