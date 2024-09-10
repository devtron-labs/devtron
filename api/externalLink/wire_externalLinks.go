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

package externalLink

import (
	"github.com/devtron-labs/devtron/pkg/externalLink"
	"github.com/google/wire"
)

//depends on sql,
//TODO integrate user auth module

var ExternalLinkWireSet = wire.NewSet(
	externalLink.NewExternalLinkMonitoringToolRepositoryImpl,
	wire.Bind(new(externalLink.ExternalLinkMonitoringToolRepository), new(*externalLink.ExternalLinkMonitoringToolRepositoryImpl)),
	externalLink.NewExternalLinkIdentifierMappingRepositoryImpl,
	wire.Bind(new(externalLink.ExternalLinkIdentifierMappingRepository), new(*externalLink.ExternalLinkIdentifierMappingRepositoryImpl)),
	externalLink.NewExternalLinkRepositoryImpl,
	wire.Bind(new(externalLink.ExternalLinkRepository), new(*externalLink.ExternalLinkRepositoryImpl)),

	externalLink.NewExternalLinkServiceImpl,
	wire.Bind(new(externalLink.ExternalLinkService), new(*externalLink.ExternalLinkServiceImpl)),
	NewExternalLinkRestHandlerImpl,
	wire.Bind(new(ExternalLinkRestHandler), new(*ExternalLinkRestHandlerImpl)),
	NewExternalLinkRouterImpl,
	wire.Bind(new(ExternalLinkRouter), new(*ExternalLinkRouterImpl)),
)
