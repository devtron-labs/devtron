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
	"github.com/gorilla/mux"
)

type GitHostRouter interface {
	InitGitHostRouter(gocdRouter *mux.Router)
}

type GitHostRouterImpl struct {
	gitHostRestHandler restHandler.GitHostRestHandler
}
func NewGitHostRouterImpl(gitHostRestHandler restHandler.GitHostRestHandler) *GitHostRouterImpl {
	return &GitHostRouterImpl{gitHostRestHandler: gitHostRestHandler}
}

func (impl GitHostRouterImpl) InitGitHostRouter(configRouter *mux.Router) {
	configRouter.Path("/host").
		HandlerFunc(impl.gitHostRestHandler.GetGitHosts).
		Methods("GET")
	configRouter.Path("/host").
		HandlerFunc(impl.gitHostRestHandler.CreateGitHost).
		Methods("POST")
	configRouter.Path("/host/{id}").
		HandlerFunc(impl.gitHostRestHandler.GetGitHostById).
		Methods("GET")
	configRouter.Path("/host/{id}/event").
		HandlerFunc(impl.gitHostRestHandler.GetAllWebhookEventConfig).
		Methods("GET")
	configRouter.Path("/host/event/{eventId}").
		HandlerFunc(impl.gitHostRestHandler.GetWebhookEventConfig).
		Methods("GET")
	configRouter.Path("/host/webhook-meta-config/{gitProviderId}").
		HandlerFunc(impl.gitHostRestHandler.GetWebhookDataMetaConfig).
		Methods("GET")
}
