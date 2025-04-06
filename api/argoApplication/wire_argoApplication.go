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

package argoApplication

import (
	argoApplication2 "github.com/devtron-labs/devtron/pkg/nucleus/argoApplication"
	"github.com/devtron-labs/devtron/pkg/nucleus/argoApplication/read"
	"github.com/devtron-labs/devtron/pkg/nucleus/argoApplication/read/config"
	"github.com/google/wire"
)

var ArgoApplicationWireSetFull = wire.NewSet(
	read.NewArgoApplicationReadServiceImpl,
	wire.Bind(new(read.ArgoApplicationReadService), new(*read.ArgoApplicationReadServiceImpl)),

	config.NewArgoApplicationConfigServiceImpl,
	wire.Bind(new(config.ArgoApplicationConfigService), new(*config.ArgoApplicationConfigServiceImpl)),

	argoApplication2.NewArgoApplicationServiceImpl,
	argoApplication2.NewArgoApplicationServiceExtendedServiceImpl,
	wire.Bind(new(argoApplication2.ArgoApplicationService), new(*argoApplication2.ArgoApplicationServiceExtendedImpl)),

	NewArgoApplicationRestHandlerImpl,
	wire.Bind(new(ArgoApplicationRestHandler), new(*ArgoApplicationRestHandlerImpl)),

	NewArgoApplicationRouterImpl,
	wire.Bind(new(ArgoApplicationRouter), new(*ArgoApplicationRouterImpl)),
)

var ArgoApplicationWireSetEA = wire.NewSet(
	read.NewArgoApplicationReadServiceImpl,
	wire.Bind(new(read.ArgoApplicationReadService), new(*read.ArgoApplicationReadServiceImpl)),

	config.NewArgoApplicationConfigServiceImpl,
	wire.Bind(new(config.ArgoApplicationConfigService), new(*config.ArgoApplicationConfigServiceImpl)),

	argoApplication2.NewArgoApplicationServiceImpl,
	wire.Bind(new(argoApplication2.ArgoApplicationService), new(*argoApplication2.ArgoApplicationServiceImpl)),

	NewArgoApplicationRestHandlerImpl,
	wire.Bind(new(ArgoApplicationRestHandler), new(*ArgoApplicationRestHandlerImpl)),

	NewArgoApplicationRouterImpl,
	wire.Bind(new(ArgoApplicationRouter), new(*ArgoApplicationRouterImpl)),
)
