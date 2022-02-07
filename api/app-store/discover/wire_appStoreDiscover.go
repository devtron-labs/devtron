package app_store_discover

import (
	app_store_discover "github.com/devtron-labs/devtron/pkg/app-store/discover"
	app_store_discover_repository "github.com/devtron-labs/devtron/pkg/app-store/discover/repository"
	"github.com/google/wire"
)

var AppStoreDiscoverWireSet = wire.NewSet(
	app_store_discover_repository.NewAppStoreRepositoryImpl,
	wire.Bind(new(app_store_discover_repository.AppStoreRepository), new(*app_store_discover_repository.AppStoreRepositoryImpl)),
	app_store_discover_repository.NewAppStoreApplicationVersionRepositoryImpl,
	wire.Bind(new(app_store_discover_repository.AppStoreApplicationVersionRepository), new(*app_store_discover_repository.AppStoreApplicationVersionRepositoryImpl)),
	app_store_discover.NewAppStoreServiceImpl,
	wire.Bind(new(app_store_discover.AppStoreService), new(*app_store_discover.AppStoreServiceImpl)),
	NewAppStoreRestHandlerImpl,
	wire.Bind(new(AppStoreRestHandler), new(*AppStoreRestHandlerImpl)),
	NewAppStoreDiscoverRouterImpl,
	wire.Bind(new(AppStoreDiscoverRouter), new(*AppStoreDiscoverRouterImpl)),
)
