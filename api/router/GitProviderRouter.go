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

type GitProviderRouter interface {
	InitGitProviderRouter(gocdRouter *mux.Router)
}
type GitProviderRouterImpl struct {
	gitRestHandler restHandler.GitProviderRestHandler
}

func NewGitProviderRouterImpl(gitRestHandler restHandler.GitProviderRestHandler) *GitProviderRouterImpl {
	return &GitProviderRouterImpl{gitRestHandler: gitRestHandler}
}
func (impl GitProviderRouterImpl) InitGitProviderRouter(configRouter *mux.Router) {
	configRouter.Path("/provider").
		HandlerFunc(impl.gitRestHandler.SaveGitRepoConfig).
		Methods("POST")
	configRouter.Path("/provider/autocomplete").
		HandlerFunc(impl.gitRestHandler.GetGitProviders).
		Methods("GET")
	configRouter.Path("/provider").
		HandlerFunc(impl.gitRestHandler.FetchAllGitProviders).
		Methods("GET")
	configRouter.Path("/provider/{id}").
		HandlerFunc(impl.gitRestHandler.FetchOneGitProviders).
		Methods("GET")
	configRouter.Path("/provider").
		HandlerFunc(impl.gitRestHandler.UpdateGitRepoConfig).
		Methods("PUT")
}
