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

package devtronResource

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/history/deployment/cdPipeline"
	"github.com/devtron-labs/devtron/pkg/devtronResource/read"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/google/wire"
)

var DevtronResourceWireSet = wire.NewSet(
	//old bindings, migrated from wire.go
	read.NewDevtronResourceSearchableKeyServiceImpl,
	wire.Bind(new(read.DevtronResourceSearchableKeyService), new(*read.DevtronResourceSearchableKeyServiceImpl)),
	repository.NewDevtronResourceSearchableKeyRepositoryImpl,
	wire.Bind(new(repository.DevtronResourceSearchableKeyRepository), new(*repository.DevtronResourceSearchableKeyRepositoryImpl)),

	NewDevtronResourceRouterImpl,
	wire.Bind(new(DevtronResourceRouter), new(*DevtronResourceRouterImpl)),

	NewHistoryRouterImpl,
	wire.Bind(new(HistoryRouter), new(*HistoryRouterImpl)),
	NewHistoryRestHandlerImpl,
	wire.Bind(new(HistoryRestHandler), new(*HistoryRestHandlerImpl)),
	cdPipeline.NewDeploymentHistoryServiceImpl,
	wire.Bind(new(cdPipeline.DeploymentHistoryService), new(*cdPipeline.DeploymentHistoryServiceImpl)),

	devtronResource.NewAPIReqDecoderServiceImpl,
	wire.Bind(new(devtronResource.APIReqDecoderService), new(*devtronResource.APIReqDecoderServiceImpl)),
)
