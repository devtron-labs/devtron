package git

import (
	"github.com/devtron-labs/devtron/pkg/build/git/gitHost"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider"
	"github.com/devtron-labs/devtron/pkg/build/git/gitWebhook"
	"github.com/devtron-labs/devtron/pkg/build/git/gitWebhook/repository"
	"github.com/google/wire"
)

var GitWireSet = wire.NewSet(
	gitProvider.GitProviderWireSet,
	gitHost.GitHostWireSet,

	gitWebhook.NewWebhookSecretValidatorImpl,
	wire.Bind(new(gitWebhook.WebhookSecretValidator), new(*gitWebhook.WebhookSecretValidatorImpl)),

	gitWebhook.NewGitWebhookServiceImpl,
	wire.Bind(new(gitWebhook.GitWebhookService), new(*gitWebhook.GitWebhookServiceImpl)),

	repository.NewGitWebhookRepositoryImpl,
	wire.Bind(new(repository.GitWebhookRepository), new(*repository.GitWebhookRepositoryImpl)))
