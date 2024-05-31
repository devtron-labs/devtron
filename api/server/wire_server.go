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

package server

import (
	"github.com/devtron-labs/devtron/pkg/server"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	serverDataStore "github.com/devtron-labs/devtron/pkg/server/store"
	"github.com/google/wire"
)

var ServerWireSet = wire.NewSet(
	server.NewServerActionAuditLogRepositoryImpl,
	wire.Bind(new(server.ServerActionAuditLogRepository), new(*server.ServerActionAuditLogRepositoryImpl)),
	serverEnvConfig.ParseServerEnvConfig,
	serverDataStore.InitServerDataStore,
	server.NewServerServiceImpl,
	wire.Bind(new(server.ServerService), new(*server.ServerServiceImpl)),
	server.NewServerCacheServiceImpl,
	wire.Bind(new(server.ServerCacheService), new(*server.ServerCacheServiceImpl)),
	NewServerRestHandlerImpl,
	wire.Bind(new(ServerRestHandler), new(*ServerRestHandlerImpl)),
	NewServerRouterImpl,
	wire.Bind(new(ServerRouter), new(*ServerRouterImpl)),
)
