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

	repository.NewClusterNoteHistoryRepositoryImpl,
	wire.Bind(new(repository.ClusterNoteHistoryRepository), new(*repository.ClusterNoteHistoryRepositoryImpl)),
	repository.NewClusterNoteRepositoryImpl,
	wire.Bind(new(repository.ClusterNoteRepository), new(*repository.ClusterNoteRepositoryImpl)),
	cluster.NewClusterNoteServiceImpl,
	wire.Bind(new(cluster.ClusterNoteService), new(*cluster.ClusterNoteServiceImpl)),

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

// minimal wire to be used with EA
var ClusterWireSetEa = wire.NewSet(
	repository.NewClusterRepositoryImpl,
	wire.Bind(new(repository.ClusterRepository), new(*repository.ClusterRepositoryImpl)),
	cluster.NewClusterServiceImpl,
	wire.Bind(new(cluster.ClusterService), new(*cluster.ClusterServiceImpl)),

	repository.NewClusterNoteRepositoryImpl,
	wire.Bind(new(repository.ClusterNoteRepository), new(*repository.ClusterNoteRepositoryImpl)),
	repository.NewClusterNoteHistoryRepositoryImpl,
	wire.Bind(new(repository.ClusterNoteHistoryRepository), new(*repository.ClusterNoteHistoryRepositoryImpl)),
	cluster.NewClusterNoteServiceImpl,
	wire.Bind(new(cluster.ClusterNoteService), new(*cluster.ClusterNoteServiceImpl)),

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
