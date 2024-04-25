package scoop

import (
	"github.com/devtron-labs/devtron/api/restHandler/autoRemediation"
	autoRemediation2 "github.com/devtron-labs/devtron/pkg/autoRemediation"
	"github.com/devtron-labs/devtron/pkg/autoRemediation/repository"
	"github.com/google/wire"
)

var ScoopWireSet = wire.NewSet(

	repository.NewInterceptedEventsRepositoryImpl,
	wire.Bind(new(repository.InterceptedEventsRepository), new(*repository.InterceptedEventsRepositoryImpl)),

	repository.NewTriggerRepositoryImpl,
	wire.Bind(new(repository.TriggerRepository), new(*repository.TriggerRepositoryImpl)),

	repository.NewWatcherRepositoryImpl,
	wire.Bind(new(repository.WatcherRepository), new(*repository.WatcherRepositoryImpl)),

	autoRemediation2.NewWatcherServiceImpl,
	wire.Bind(new(autoRemediation2.WatcherService), new(*autoRemediation2.WatcherServiceImpl)),

	autoRemediation.NewWatcherRestHandlerImpl,
	wire.Bind(new(autoRemediation.WatcherRestHandler), new(*autoRemediation.WatcherRestHandlerImpl)),

	NewServiceImpl,
	wire.Bind(new(Service), new(*ServiceImpl)),

	NewRestHandler,
	wire.Bind(new(RestHandler), new(*RestHandlerImpl)),

	NewRouterImpl,
	wire.Bind(new(Router), new(*RouterImpl)),
)
