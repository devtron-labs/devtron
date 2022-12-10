package webhookHelm

import (
	"github.com/gorilla/mux"
)

type WebhookHelmRouter interface {
	InitWebhookHelmRouter(configRouter *mux.Router)
}

type WebhookHelmRouterImpl struct {
	webhookHelmRestHandler WebhookHelmRestHandler
}

func NewWebhookHelmRouterImpl(webhookHelmRestHandler WebhookHelmRestHandler) *WebhookHelmRouterImpl {
	return &WebhookHelmRouterImpl{webhookHelmRestHandler: webhookHelmRestHandler}
}

func (router *WebhookHelmRouterImpl) InitWebhookHelmRouter(configRouter *mux.Router) {
	configRouter.Path("/app").
		HandlerFunc(router.webhookHelmRestHandler.InstallOrUpdateApplication).
		Methods("POST")
}
