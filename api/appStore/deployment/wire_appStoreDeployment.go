package appStoreDeployment

import (
	appStoreDeployment "github.com/devtron-labs/devtron/pkg/appStore/deployment"
	appStoreRepository "github.com/devtron-labs/devtron/pkg/appStore/repository"
	"github.com/google/wire"
)

var AppStoreDeploymentWireSet = wire.NewSet(
	appStoreRepository.NewClusterInstalledAppsRepositoryImpl,
	wire.Bind(new(appStoreRepository.ClusterInstalledAppsRepository), new(*appStoreRepository.ClusterInstalledAppsRepositoryImpl)),
	appStoreDeployment.NewAppStoreDeploymentServiceImpl,
	wire.Bind(new(appStoreDeployment.AppStoreDeploymentService), new(*appStoreDeployment.AppStoreDeploymentServiceImpl)),
	NewAppStoreDeploymentRestHandlerImpl,
	wire.Bind(new(AppStoreDeploymentRestHandler), new(*AppStoreDeploymentRestHandlerImpl)),
	NewAppStoreDeploymentRouterImpl,
	wire.Bind(new(AppStoreDeploymentRouter), new(*AppStoreDeploymentRouterImpl)),
)
