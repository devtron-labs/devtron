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

type GitOpsConfigRouter interface {
	InitGitOpsConfigRouter(gocdRouter *mux.Router)
}
type GitOpsConfigRouterImpl struct {
	gitOpsConfigRestHandler restHandler.GitOpsConfigRestHandler
}

func NewGitOpsConfigRouterImpl(gitOpsConfigRestHandler restHandler.GitOpsConfigRestHandler) *GitOpsConfigRouterImpl {
	return &GitOpsConfigRouterImpl{gitOpsConfigRestHandler: gitOpsConfigRestHandler}
}
func (impl GitOpsConfigRouterImpl) InitGitOpsConfigRouter(configRouter *mux.Router) {
	configRouter.Path("/config").
		HandlerFunc(impl.gitOpsConfigRestHandler.CreateGitOpsConfig).
		Methods("POST")
	configRouter.Path("/config").
		HandlerFunc(impl.gitOpsConfigRestHandler.UpdateGitOpsConfig).
		Methods("PUT")
	configRouter.Path("/config/{id}").
		HandlerFunc(impl.gitOpsConfigRestHandler.GetGitOpsConfigById).
		Methods("GET")
	configRouter.Path("/config").
		HandlerFunc(impl.gitOpsConfigRestHandler.GetAllGitOpsConfig).
		Methods("GET")
	configRouter.Path("/config").
		HandlerFunc(impl.gitOpsConfigRestHandler.GetGitOpsConfigByProvider).Queries("provider", "{provider}").
		Methods("GET")

}
