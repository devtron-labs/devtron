package appStoreDiscover

import (
	appStoreDiscover "github.com/devtron-labs/devtron/pkg/appStore/discover"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/google/wire"
)

var AppStoreDiscoverWireSet = wire.NewSet(
	appStoreDiscoverRepository.NewAppStoreRepositoryImpl,
	wire.Bind(new(appStoreDiscoverRepository.AppStoreRepository), new(*appStoreDiscoverRepository.AppStoreRepositoryImpl)),
	appStoreDiscoverRepository.NewAppStoreApplicationVersionRepositoryImpl,
	wire.Bind(new(appStoreDiscoverRepository.AppStoreApplicationVersionRepository), new(*appStoreDiscoverRepository.AppStoreApplicationVersionRepositoryImpl)),
	appStoreDiscover.NewAppStoreServiceImpl,
	wire.Bind(new(appStoreDiscover.AppStoreService), new(*appStoreDiscover.AppStoreServiceImpl)),
	NewAppStoreRestHandlerImpl,
	wire.Bind(new(AppStoreRestHandler), new(*AppStoreRestHandlerImpl)),
	NewAppStoreDiscoverRouterImpl,
	wire.Bind(new(AppStoreDiscoverRouter), new(*AppStoreDiscoverRouterImpl)),
)
