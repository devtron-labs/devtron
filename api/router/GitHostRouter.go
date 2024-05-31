/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
