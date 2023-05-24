package appStoreDiscover

import (
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/discover/service"
	"github.com/google/wire"
)

var AppStoreDiscoverWireSet = wire.NewSet(
	appStoreDiscoverRepository.NewAppStoreRepositoryImpl,
	wire.Bind(new(appStoreDiscoverRepository.AppStoreRepository), new(*appStoreDiscoverRepository.AppStoreRepositoryImpl)),
	appStoreDiscoverRepository.NewAppStoreApplicationVersionRepositoryImpl,
	wire.Bind(new(appStoreDiscoverRepository.AppStoreApplicationVersionRepository), new(*appStoreDiscoverRepository.AppStoreApplicationVersionRepositoryImpl)),
	service.NewAppStoreServiceImpl,
	wire.Bind(new(service.AppStoreService), new(*service.AppStoreServiceImpl)),
	NewAppStoreRestHandlerImpl,
	wire.Bind(new(AppStoreRestHandler), new(*AppStoreRestHandlerImpl)),
	NewAppStoreDiscoverRouterImpl,
	wire.Bind(new(AppStoreDiscoverRouter), new(*AppStoreDiscoverRouterImpl)),
)
