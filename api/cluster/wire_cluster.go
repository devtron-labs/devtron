package cluster

import (
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/google/wire"
)

//depends on sql,user,K8sUtil, logger, enforcer, TODO

var ClusterWireSet = wire.NewSet(
	repository.NewClusterRepositoryImpl,
	wire.Bind(new(repository.ClusterRepository), new(*repository.ClusterRepositoryImpl)),
	cluster.NewClusterServiceImplExtended,
	wire.Bind(new(cluster.ClusterService), new(*cluster.ClusterServiceImplExtended)),
	NewClusterRestHandlerImpl,
	wire.Bind(new(ClusterRestHandler), new(*ClusterRestHandlerImpl)),
	NewClusterRouterImpl,
	wire.Bind(new(ClusterRouter), new(*ClusterRouterImpl)),

	repository.NewEnvironmentRepositoryImpl,
	wire.Bind(new(repository.EnvironmentRepository), new(*repository.EnvironmentRepositoryImpl)),
	cluster.NewEnvironmentServiceImpl,
	wire.Bind(new(cluster.EnvironmentService), new(*cluster.EnvironmentServiceImpl)),
	NewEnvironmentRestHandlerImpl,
	wire.Bind(new(EnvironmentRestHandler), new(*EnvironmentRestHandlerImpl)),
	NewEnvironmentRouterImpl,
	wire.Bind(new(EnvironmentRouter), new(*EnvironmentRouterImpl)),
)

//minimal wire to be used with EA
var ClusterWireSetEa = wire.NewSet(
	repository.NewClusterRepositoryImpl,
	wire.Bind(new(repository.ClusterRepository), new(*repository.ClusterRepositoryImpl)),
	cluster.NewClusterServiceImpl,
	wire.Bind(new(cluster.ClusterService), new(*cluster.ClusterServiceImpl)),
	NewClusterRestHandlerImpl,
	wire.Bind(new(ClusterRestHandler), new(*ClusterRestHandlerImpl)),
	NewClusterRouterImpl,
	wire.Bind(new(ClusterRouter), new(*ClusterRouterImpl)),
	repository.NewEnvironmentRepositoryImpl,
	wire.Bind(new(repository.EnvironmentRepository), new(*repository.EnvironmentRepositoryImpl)),
	cluster.NewEnvironmentServiceImpl,
	wire.Bind(new(cluster.EnvironmentService), new(*cluster.EnvironmentServiceImpl)),
	NewEnvironmentRestHandlerImpl,
	wire.Bind(new(EnvironmentRestHandler), new(*EnvironmentRestHandlerImpl)),
	NewEnvironmentRouterImpl,
	wire.Bind(new(EnvironmentRouter), new(*EnvironmentRouterImpl)),
)
