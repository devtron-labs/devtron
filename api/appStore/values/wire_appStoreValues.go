package appStoreValues

import (
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreValuesRepository "github.com/devtron-labs/devtron/pkg/appStore/values/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/values/service"
	"github.com/google/wire"
)

var AppStoreValuesWireSet = wire.NewSet(
	NewAppStoreValuesRouterImpl,
	wire.Bind(new(AppStoreValuesRouter), new(*AppStoreValuesRouterImpl)),
	NewAppStoreValuesRestHandlerImpl,
	wire.Bind(new(AppStoreValuesRestHandler), new(*AppStoreValuesRestHandlerImpl)),
	service.NewAppStoreValuesServiceImpl,
	wire.Bind(new(service.AppStoreValuesService), new(*service.AppStoreValuesServiceImpl)),
	appStoreValuesRepository.NewAppStoreVersionValuesRepositoryImpl,
	wire.Bind(new(appStoreValuesRepository.AppStoreVersionValuesRepository), new(*appStoreValuesRepository.AppStoreVersionValuesRepositoryImpl)),
	repository.NewInstalledAppRepositoryImpl,
	wire.Bind(new(repository.InstalledAppRepository), new(*repository.InstalledAppRepositoryImpl)),
)
