package webhookHelm

import (
	webhookHelm "github.com/devtron-labs/devtron/pkg/webhook/helm"
	"github.com/google/wire"
)

var WebhookHelmWireSet = wire.NewSet(
	webhookHelm.NewWebhookHelmServiceImpl,
	wire.Bind(new(webhookHelm.WebhookHelmService), new(*webhookHelm.WebhookHelmServiceImpl)),
	NewWebhookHelmRestHandlerImpl,
	wire.Bind(new(WebhookHelmRestHandler), new(*WebhookHelmRestHandlerImpl)),
	NewWebhookHelmRouterImpl,
	wire.Bind(new(WebhookHelmRouter), new(*WebhookHelmRouterImpl)),
)
