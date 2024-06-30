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

package client

import (
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/google/wire"
)

var HelmAppWireSet = wire.NewSet(
	gRPC.NewHelmAppClientImpl,
	wire.Bind(new(gRPC.HelmAppClient), new(*gRPC.HelmAppClientImpl)),
	service.GetHelmReleaseConfig,
	service.NewHelmAppServiceImpl,
	wire.Bind(new(service.HelmAppService), new(*service.HelmAppServiceImpl)),
	NewHelmAppRestHandlerImpl,
	wire.Bind(new(HelmAppRestHandler), new(*HelmAppRestHandlerImpl)),
	NewHelmAppRouterImpl,
	wire.Bind(new(HelmAppRouter), new(*HelmAppRouterImpl)),
	gRPC.GetConfig,
	rbac.NewEnforcerUtilHelmImpl,
	wire.Bind(new(rbac.EnforcerUtilHelm), new(*rbac.EnforcerUtilHelmImpl)),
)
