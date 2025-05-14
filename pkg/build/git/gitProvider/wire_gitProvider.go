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

package gitProvider

import (
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider/read"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider/repository"
	"github.com/google/wire"
)

var GitProviderWireSet = wire.NewSet(
	read.NewGitProviderReadService,
	wire.Bind(new(read.GitProviderReadService), new(*read.GitProviderReadServiceImpl)),

	repository.NewGitProviderRepositoryImpl,
	wire.Bind(new(repository.GitProviderRepository), new(*repository.GitProviderRepositoryImpl)),
	NewGitRegistryConfigImpl,
	wire.Bind(new(GitRegistryConfig), new(*GitRegistryConfigImpl)),
)
