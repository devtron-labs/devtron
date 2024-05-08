package cluster

import (
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/genericNotes"
	repository2 "github.com/devtron-labs/devtron/pkg/genericNotes/repository"
	"github.com/google/wire"
)

//depends on sql,user,K8sUtil, logger, enforcer, TODO

var ClusterWireSet = wire.NewSet(
	repository.NewClusterRepositoryImpl,
	wire.Bind(new(repository.ClusterRepository), new(*repository.ClusterRepositoryImpl)),
	cluster.NewClusterServiceImplExtended,
	wire.Bind(new(cluster.ClusterService), new(*cluster.ClusterServiceImplExtended)),

	cluster.NewClusterRbacServiceImpl,
	wire.Bind(new(cluster.ClusterRbacService), new(*cluster.ClusterRbacServiceImpl)),

	repository.NewClusterDescriptionRepositoryImpl,
	wire.Bind(new(repository.ClusterDescriptionRepository), new(*repository.ClusterDescriptionRepositoryImpl)),
	repository2.NewGenericNoteHistoryRepositoryImpl,
	wire.Bind(new(repository2.GenericNoteHistoryRepository), new(*repository2.GenericNoteHistoryRepositoryImpl)),
	repository2.NewGenericNoteRepositoryImpl,
	wire.Bind(new(repository2.GenericNoteRepository), new(*repository2.GenericNoteRepositoryImpl)),
	genericNotes.NewGenericNoteHistoryServiceImpl,
	wire.Bind(new(genericNotes.GenericNoteHistoryService), new(*genericNotes.GenericNoteHistoryServiceImpl)),
	genericNotes.NewGenericNoteServiceImpl,
	wire.Bind(new(genericNotes.GenericNoteService), new(*genericNotes.GenericNoteServiceImpl)),
	cluster.NewClusterDescriptionServiceImpl,
	wire.Bind(new(cluster.ClusterDescriptionService), new(*cluster.ClusterDescriptionServiceImpl)),

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
	cluster.NewClusterRbacServiceImpl,
	wire.Bind(new(cluster.ClusterRbacService), new(*cluster.ClusterRbacServiceImpl)),
	cluster.NewClusterServiceImpl,
	wire.Bind(new(cluster.ClusterService), new(*cluster.ClusterServiceImpl)),

	repository.NewClusterDescriptionRepositoryImpl,
	wire.Bind(new(repository.ClusterDescriptionRepository), new(*repository.ClusterDescriptionRepositoryImpl)),
	repository2.NewGenericNoteHistoryRepositoryImpl,
	wire.Bind(new(repository2.GenericNoteHistoryRepository), new(*repository2.GenericNoteHistoryRepositoryImpl)),
	repository2.NewGenericNoteRepositoryImpl,
	wire.Bind(new(repository2.GenericNoteRepository), new(*repository2.GenericNoteRepositoryImpl)),
	genericNotes.NewGenericNoteHistoryServiceImpl,
	wire.Bind(new(genericNotes.GenericNoteHistoryService), new(*genericNotes.GenericNoteHistoryServiceImpl)),
	genericNotes.NewGenericNoteServiceImpl,
	wire.Bind(new(genericNotes.GenericNoteService), new(*genericNotes.GenericNoteServiceImpl)),
	cluster.NewClusterDescriptionServiceImpl,
	wire.Bind(new(cluster.ClusterDescriptionService), new(*cluster.ClusterDescriptionServiceImpl)),

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
