/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type WebhookListenerRouter interface {
	InitWebhookListenerRouter(gocdRouter *mux.Router)
}

type WebhookListenerRouterImpl struct {
	webhookEventHandler restHandler.WebhookEventHandler
}

func NewWebhookListenerRouterImpl(webhookEventHandler restHandler.WebhookEventHandler) *WebhookListenerRouterImpl {
	return &WebhookListenerRouterImpl{
		webhookEventHandler: webhookEventHandler,
	}
}

func (impl WebhookListenerRouterImpl) InitWebhookListenerRouter(configRouter *mux.Router) {
	configRouter.Path("/{gitHostId}").
		HandlerFunc(impl.webhookEventHandler.OnWebhookEvent).
		Methods("POST")
	configRouter.Path("/{gitHostId}/{secret}").
		HandlerFunc(impl.webhookEventHandler.OnWebhookEvent).
		Methods("POST")
}
