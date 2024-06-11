/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package application

import (
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/gorilla/mux"
)

type K8sApplicationRouter interface {
	InitK8sApplicationRouter(helmRouter *mux.Router)
}
type K8sApplicationRouterImpl struct {
	k8sApplicationRestHandler K8sApplicationRestHandler
}

func NewK8sApplicationRouterImpl(k8sApplicationRestHandler K8sApplicationRestHandler) *K8sApplicationRouterImpl {
	return &K8sApplicationRouterImpl{
		k8sApplicationRestHandler: k8sApplicationRestHandler,
	}
}

func (impl *K8sApplicationRouterImpl) InitK8sApplicationRouter(k8sAppRouter *mux.Router) {
	k8sAppRouter.Path("/resource/rotate").Queries("appId", "{appId}").
		HandlerFunc(impl.k8sApplicationRestHandler.RotatePod).Methods("POST")

	k8sAppRouter.Path("/resource/urls").Queries("appId", "{appId}").
		HandlerFunc(impl.k8sApplicationRestHandler.GetHostUrlsByBatch).Methods("GET")

	k8sAppRouter.Path("/resource").
		HandlerFunc(impl.k8sApplicationRestHandler.GetResource).Methods("POST")

	k8sAppRouter.Path("/resource/create").
		HandlerFunc(impl.k8sApplicationRestHandler.CreateResource).Methods("POST")

	k8sAppRouter.Path("/resource").
		HandlerFunc(impl.k8sApplicationRestHandler.UpdateResource).Methods("PUT")

	k8sAppRouter.Path("/resource/delete").
		HandlerFunc(impl.k8sApplicationRestHandler.DeleteResource).Methods("POST")

	k8sAppRouter.Path("/events").
		HandlerFunc(impl.k8sApplicationRestHandler.ListEvents).Methods("POST")

	k8sAppRouter.Path("/pods/logs/{podName}").
		Queries("containerName", "{containerName}").
		Queries("follow", "{follow}").
		HandlerFunc(impl.k8sApplicationRestHandler.GetPodLogs).Methods("GET")

	k8sAppRouter.Path("/pods/logs/download/{podName}").
		Queries("containerName", "{containerName}").
		HandlerFunc(impl.k8sApplicationRestHandler.DownloadPodLogs).Methods("GET")

	k8sAppRouter.Path("/pod/exec/session/{identifier}/{namespace}/{pod}/{shell}/{container}").
		HandlerFunc(impl.k8sApplicationRestHandler.GetTerminalSession).Methods("GET")
	k8sAppRouter.PathPrefix("/pod/exec/sockjs/ws").Handler(terminal.CreateAttachHandler("/pod/exec/sockjs/ws"))

	/*k8sAppRouter.Path("/pod/exec/sockjs/ws/").
	Handler(terminal.CreateAttachHandler("/api/v1/applications/pod/exec/sockjs/ws/"))*/

	k8sAppRouter.Path("/resource/inception/info").
		HandlerFunc(impl.k8sApplicationRestHandler.GetResourceInfo).Methods("GET")

	k8sAppRouter.Path("/api-resources/{clusterId}").
		HandlerFunc(impl.k8sApplicationRestHandler.GetAllApiResources).Methods("GET")

	k8sAppRouter.Path("/resource/list").
		HandlerFunc(impl.k8sApplicationRestHandler.GetResourceList).Methods("POST")

	k8sAppRouter.Path("/resources/apply").
		HandlerFunc(impl.k8sApplicationRestHandler.ApplyResources).Methods("POST")

	//create/delete ephemeral containers API's
	k8sAppRouter.Path("/resources/ephemeralContainers").
		Queries("identifier", "{identifier}").
		HandlerFunc(impl.k8sApplicationRestHandler.CreateEphemeralContainer).Methods("POST")
	k8sAppRouter.Path("/resources/ephemeralContainers").
		Queries("identifier", "{identifier}").
		HandlerFunc(impl.k8sApplicationRestHandler.DeleteEphemeralContainer).Methods("DELETE")

	k8sAppRouter.Path("/api-resources/gvk/{clusterId}").
		HandlerFunc(impl.k8sApplicationRestHandler.GetAllApiResourceGVKWithoutAuthorization).Methods("GET")

}
