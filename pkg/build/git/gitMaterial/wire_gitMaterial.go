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

package gitMaterial

import (
	"github.com/devtron-labs/devtron/pkg/build/git/gitMaterial/read"
	repository3 "github.com/devtron-labs/devtron/pkg/build/git/gitMaterial/repository"
	"github.com/google/wire"
)

var GitMaterialWireSet = wire.NewSet(
	read.NewGitMaterialReadServiceImpl,
	wire.Bind(new(read.GitMaterialReadService), new(*read.GitMaterialReadServiceImpl)),

	repository3.NewMaterialRepositoryImpl,
	wire.Bind(new(repository3.MaterialRepository), new(*repository3.MaterialRepositoryImpl)),
)
