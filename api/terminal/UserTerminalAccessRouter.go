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

package terminal

import (
	"github.com/gorilla/mux"
)

type UserTerminalAccessRouter interface {
	InitTerminalAccessRouter(userTerminalAccessRouter *mux.Router)
}

type UserTerminalAccessRouterImpl struct {
	userTerminalAccessRestHandler UserTerminalAccessRestHandler
}

func NewUserTerminalAccessRouterImpl(userTerminalAccessRestHandler UserTerminalAccessRestHandler) *UserTerminalAccessRouterImpl {
	return &UserTerminalAccessRouterImpl{
		userTerminalAccessRestHandler: userTerminalAccessRestHandler,
	}
}

func (router UserTerminalAccessRouterImpl) InitTerminalAccessRouter(userTerminalAccessRouter *mux.Router) {
	userTerminalAccessRouter.Path("/update").
		HandlerFunc(router.userTerminalAccessRestHandler.UpdateTerminalSession).Methods("PUT")
	userTerminalAccessRouter.Path("/update/shell").
		HandlerFunc(router.userTerminalAccessRestHandler.UpdateTerminalShellSession).Methods("PUT")
	userTerminalAccessRouter.Path("/start").
		HandlerFunc(router.userTerminalAccessRestHandler.StartTerminalSession).Methods("POST")
	userTerminalAccessRouter.Path("/get").
		HandlerFunc(router.userTerminalAccessRestHandler.FetchTerminalStatus).Queries("terminalAccessId", "{terminalAccessId}", "namespace", "{namespace}", "shellName", "{shellName}", "containerName", "{containerName}").Methods("GET")
	userTerminalAccessRouter.Path("/pod/events").
		HandlerFunc(router.userTerminalAccessRestHandler.FetchTerminalPodEvents).Queries("terminalAccessId", "{terminalAccessId}").Methods("GET")
	userTerminalAccessRouter.Path("/pod/manifest").
		HandlerFunc(router.userTerminalAccessRestHandler.FetchTerminalPodManifest).Queries("terminalAccessId", "{terminalAccessId}").Methods("GET")
	userTerminalAccessRouter.Path("/disconnect").
		HandlerFunc(router.userTerminalAccessRestHandler.DisconnectTerminalSession).Queries("terminalAccessId", "{terminalAccessId}").Methods("POST")
	userTerminalAccessRouter.Path("/stop").
		HandlerFunc(router.userTerminalAccessRestHandler.StopTerminalSession).Queries("terminalAccessId", "{terminalAccessId}").Methods("PUT")
	userTerminalAccessRouter.Path("/disconnectAndRetry").
		HandlerFunc(router.userTerminalAccessRestHandler.DisconnectAllTerminalSessionAndRetry).Methods("POST")
	userTerminalAccessRouter.Path("/validateShell").Queries("podName", "{podName}", "namespace", "{namespace}", "shellName", "{shellName}", "clusterId", "{clusterId}").
		HandlerFunc(router.userTerminalAccessRestHandler.ValidateShell)
	userTerminalAccessRouter.Path("/edit").
		HandlerFunc(router.userTerminalAccessRestHandler.EditPodManifest).Methods("PUT")
	//TODO fetch all user running/starting pods
	//TODO fetch all running/starting pods also include sessionIds if session exists
	//TODO terminate all Sessions
	//TODO delete all terminal-pods from k8s directly
}
