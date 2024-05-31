/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
	configRouter.Path("/provider").
		HandlerFunc(impl.gitRestHandler.DeleteGitRepoConfig).
		Methods("DELETE")
}
