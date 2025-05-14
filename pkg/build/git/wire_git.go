/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package git

import (
	"github.com/devtron-labs/devtron/pkg/build/git/gitHost"
	"github.com/devtron-labs/devtron/pkg/build/git/gitMaterial"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider"
	"github.com/devtron-labs/devtron/pkg/build/git/gitWebhook"
	"github.com/devtron-labs/devtron/pkg/build/git/gitWebhook/repository"
	"github.com/google/wire"
)

var GitWireSet = wire.NewSet(
	gitProvider.GitProviderWireSet,
	gitHost.GitHostWireSet,
	gitMaterial.GitMaterialWireSet,

	gitWebhook.NewWebhookSecretValidatorImpl,
	wire.Bind(new(gitWebhook.WebhookSecretValidator), new(*gitWebhook.WebhookSecretValidatorImpl)),

	gitWebhook.NewGitWebhookServiceImpl,
	wire.Bind(new(gitWebhook.GitWebhookService), new(*gitWebhook.GitWebhookServiceImpl)),

	repository.NewGitWebhookRepositoryImpl,
	wire.Bind(new(repository.GitWebhookRepository), new(*repository.GitWebhookRepositoryImpl)))
