/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
	configRouter.Path("/config-by-provider").
		HandlerFunc(impl.gitOpsConfigRestHandler.GetGitOpsConfigByProvider).Queries("provider", "{provider}").
		Methods("GET")
	configRouter.Path("/configured").
		HandlerFunc(impl.gitOpsConfigRestHandler.GitOpsConfigured).
		Methods("GET")
	configRouter.Path("/validate").
		HandlerFunc(impl.gitOpsConfigRestHandler.GitOpsValidator).
		Methods("POST")
}
