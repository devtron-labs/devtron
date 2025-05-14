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

package infraConfig

import (
	"github.com/devtron-labs/devtron/api/infraConfig"
	"github.com/devtron-labs/devtron/pkg/infraConfig/config"
	infraRepository "github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	auditRepo "github.com/devtron-labs/devtron/pkg/infraConfig/repository/audit"
	infraConfigService "github.com/devtron-labs/devtron/pkg/infraConfig/service"
	"github.com/devtron-labs/devtron/pkg/infraConfig/service/audit"
	"github.com/devtron-labs/devtron/pkg/pipeline/infraProviders"
	"github.com/devtron-labs/devtron/pkg/pipeline/infraProviders/infraGetters/ci"
	"github.com/devtron-labs/devtron/pkg/pipeline/infraProviders/infraGetters/job"
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	job.NewJobInfraGetter,

	ci.NewCiInfraGetter,

	infraRepository.NewInfraProfileRepositoryImpl,
	wire.Bind(new(infraRepository.InfraConfigRepository), new(*infraRepository.InfraConfigRepositoryImpl)),

	auditRepo.NewInfraConfigAuditRepositoryImpl,
	wire.Bind(new(auditRepo.InfraConfigAuditRepository), new(*auditRepo.InfraConfigAuditRepositoryImpl)),

	audit.NewInfraConfigAuditServiceImpl,
	wire.Bind(new(audit.InfraConfigAuditService), new(*audit.InfraConfigAuditServiceImpl)),

	config.NewInfraConfigClient,
	wire.Bind(new(config.InfraConfigClient), new(*config.InfraConfigClientImpl)),

	infraConfigService.NewInfraConfigServiceImpl,
	wire.Bind(new(infraConfigService.InfraConfigService), new(*infraConfigService.InfraConfigServiceImpl)),

	infraProviders.NewInfraProviderImpl,
	wire.Bind(new(infraProviders.InfraProvider), new(*infraProviders.InfraProviderImpl)),

	infraConfig.NewInfraConfigRestHandlerImpl,
	wire.Bind(new(infraConfig.InfraConfigRestHandler), new(*infraConfig.InfraConfigRestHandlerImpl)),

	infraConfig.NewInfraProfileRouterImpl,
	wire.Bind(new(infraConfig.InfraConfigRouter), new(*infraConfig.InfraConfigRouterImpl)),
)
