package webhookHelm

import (
	"github.com/devtron-labs/devtron/api/logger"
	"github.com/gorilla/mux"
)

type WebhookHelmRouter interface {
	InitWebhookHelmRouter(configRouter *mux.Router)
}

type WebhookHelmRouterImpl struct {
	webhookHelmRestHandler WebhookHelmRestHandler
	userAuth               logger.UserAuth
}

func NewWebhookHelmRouterImpl(webhookHelmRestHandler WebhookHelmRestHandler, userAuth logger.UserAuth) *WebhookHelmRouterImpl {
	return &WebhookHelmRouterImpl{webhookHelmRestHandler: webhookHelmRestHandler, userAuth: userAuth}
}

func (impl WebhookHelmRouterImpl) InitWebhookHelmRouter(configRouter *mux.Router) {
	configRouter.Use(impl.userAuth.LoggingMiddleware)
	configRouter.Path("/app").
		HandlerFunc(impl.webhookHelmRestHandler.InstallOrUpdateApplication).
		Methods("POST")
}
