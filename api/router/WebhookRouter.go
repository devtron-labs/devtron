/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/api/restHandler/app/pipeline/configure"
	"github.com/gorilla/mux"
)

type WebhookRouter interface {
	intWebhookRouter(configRouter *mux.Router)
}

type WebhookRouterImpl struct {
	gitWebhookRestHandler   restHandler.GitWebhookRestHandler
	pipelineRestHandler     configure.PipelineConfigRestHandler
	externalCiRestHandler   restHandler.ExternalCiRestHandler
	pubSubClientRestHandler restHandler.PubSubClientRestHandler
}

func NewWebhookRouterImpl(gitWebhookRestHandler restHandler.GitWebhookRestHandler,
	pipelineRestHandler configure.PipelineConfigRestHandler, externalCiRestHandler restHandler.ExternalCiRestHandler,
	pubSubClientRestHandler restHandler.PubSubClientRestHandler) *WebhookRouterImpl {
	return &WebhookRouterImpl{
		gitWebhookRestHandler:   gitWebhookRestHandler,
		pipelineRestHandler:     pipelineRestHandler,
		externalCiRestHandler:   externalCiRestHandler,
		pubSubClientRestHandler: pubSubClientRestHandler,
	}
}

func (impl WebhookRouterImpl) intWebhookRouter(configRouter *mux.Router) {
	configRouter.Path("/git").HandlerFunc(impl.gitWebhookRestHandler.HandleGitWebhook).Methods("POST")
	configRouter.Path("/ci/workflow").HandlerFunc(impl.pipelineRestHandler.HandleWorkflowWebhook).Methods("POST")
	configRouter.Path("/msg/nats").HandlerFunc(impl.pubSubClientRestHandler.PublishEventsToNats).Methods("POST")
	configRouter.Path("/ext-ci/{externalCiId}").HandlerFunc(impl.externalCiRestHandler.HandleExternalCiWebhook).Methods("POST")
}
