package git

import (
	"github.com/devtron-labs/devtron/pkg/build/git/gitHost"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider"
	"github.com/devtron-labs/devtron/pkg/build/git/gitWebhook"
	"github.com/google/wire"
)

var GitWireSet = wire.NewSet(
	gitProvider.GitProviderWireSet,

	gitWebhook.NewWebhookSecretValidatorImpl,
	wire.Bind(new(gitWebhook.WebhookSecretValidator), new(*gitWebhook.WebhookSecretValidatorImpl)),

	gitWebhook.NewGitWebhookServiceImpl,
	wire.Bind(new(gitWebhook.GitWebhookService), new(*gitWebhook.GitWebhookServiceImpl)),

	gitHost.NewGitHostConfigImpl,
	wire.Bind(new(gitHost.GitHostConfig), new(*gitHost.GitHostConfigImpl)))
