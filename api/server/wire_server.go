package server

import (
	"github.com/devtron-labs/devtron/pkg/server"
	serverDataStore "github.com/devtron-labs/devtron/pkg/server/store"
	"github.com/google/wire"
)

var ServerWireSet = wire.NewSet(
	server.NewServerActionAuditLogRepositoryImpl,
	wire.Bind(new(server.ServerActionAuditLogRepository), new(*server.ServerActionAuditLogRepositoryImpl)),
	server.ParseServerEnvConfig,
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
