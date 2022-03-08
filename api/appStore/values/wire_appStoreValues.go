package appStoreValues

import (
	appStoreRepository "github.com/devtron-labs/devtron/pkg/appStore/repository"
	appStoreValues "github.com/devtron-labs/devtron/pkg/appStore/values"
	appStoreValuesRepository "github.com/devtron-labs/devtron/pkg/appStore/values/repository"
	"github.com/google/wire"
)

var AppStoreValuesWireSet = wire.NewSet(
	NewAppStoreValuesRouterImpl,
	wire.Bind(new(AppStoreValuesRouter), new(*AppStoreValuesRouterImpl)),
	NewAppStoreValuesRestHandlerImpl,
	wire.Bind(new(AppStoreValuesRestHandler), new(*AppStoreValuesRestHandlerImpl)),
	appStoreValues.NewAppStoreValuesServiceImpl,
	wire.Bind(new(appStoreValues.AppStoreValuesService), new(*appStoreValues.AppStoreValuesServiceImpl)),
	appStoreValuesRepository.NewAppStoreVersionValuesRepositoryImpl,
	wire.Bind(new(appStoreValuesRepository.AppStoreVersionValuesRepository), new(*appStoreValuesRepository.AppStoreVersionValuesRepositoryImpl)),
	appStoreRepository.NewInstalledAppRepositoryImpl,
	wire.Bind(new(appStoreRepository.InstalledAppRepository), new(*appStoreRepository.InstalledAppRepositoryImpl)),
)
