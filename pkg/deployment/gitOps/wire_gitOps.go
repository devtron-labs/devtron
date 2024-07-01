/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gitOps

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/validation"
	"github.com/google/wire"
)

var GitOpsWireSet = wire.NewSet(
	repository.NewGitOpsConfigRepositoryImpl,
	wire.Bind(new(repository.GitOpsConfigRepository), new(*repository.GitOpsConfigRepositoryImpl)),

	config.NewGitOpsConfigReadServiceImpl,
	wire.Bind(new(config.GitOpsConfigReadService), new(*config.GitOpsConfigReadServiceImpl)),

	git.NewGitOperationServiceImpl,
	wire.Bind(new(git.GitOperationService), new(*git.GitOperationServiceImpl)),

	validation.NewGitOpsValidationServiceImpl,
	wire.Bind(new(validation.GitOpsValidationService), new(*validation.GitOpsValidationServiceImpl)),
)

var GitOpsEAWireSet = wire.NewSet(
	repository.NewGitOpsConfigRepositoryImpl,
	wire.Bind(new(repository.GitOpsConfigRepository), new(*repository.GitOpsConfigRepositoryImpl)),

	config.NewGitOpsConfigReadServiceImpl,
	wire.Bind(new(config.GitOpsConfigReadService), new(*config.GitOpsConfigReadServiceImpl)),
)
