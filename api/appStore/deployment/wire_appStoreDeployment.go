package appStoreDeployment

import (
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/common"
	"github.com/google/wire"
)

var AppStoreDeploymentWireSet = wire.NewSet(
	service.GetDeploymentServiceTypeConfig,
	repository.NewClusterInstalledAppsRepositoryImpl,
	wire.Bind(new(repository.ClusterInstalledAppsRepository), new(*repository.ClusterInstalledAppsRepositoryImpl)),
	appStoreDeploymentCommon.NewAppStoreDeploymentCommonServiceImpl,
	wire.Bind(new(appStoreDeploymentCommon.AppStoreDeploymentCommonService), new(*appStoreDeploymentCommon.AppStoreDeploymentCommonServiceImpl)),
	EAMode.NewEAModeDeploymentServiceImpl,
	wire.Bind(new(EAMode.EAModeDeploymentService), new(*EAMode.EAModeDeploymentServiceImpl)),
	service.NewAppStoreDeploymentServiceImpl,
	wire.Bind(new(service.AppStoreDeploymentService), new(*service.AppStoreDeploymentServiceImpl)),
	NewAppStoreDeploymentRestHandlerImpl,
	wire.Bind(new(AppStoreDeploymentRestHandler), new(*AppStoreDeploymentRestHandlerImpl)),
	NewAppStoreDeploymentRouterImpl,
	wire.Bind(new(AppStoreDeploymentRouter), new(*AppStoreDeploymentRouterImpl)),
	repository.NewInstalledAppVersionHistoryRepositoryImpl,
	wire.Bind(new(repository.InstalledAppVersionHistoryRepository), new(*repository.InstalledAppVersionHistoryRepositoryImpl)),

	NewCommonDeploymentRestHandlerImpl,
	wire.Bind(new(CommonDeploymentRestHandler), new(*CommonDeploymentRestHandlerImpl)),
	NewCommonDeploymentRouterImpl,
	wire.Bind(new(CommonDeploymentRouter), new(*CommonDeploymentRouterImpl)),
	argocdServer.GetACDDeploymentConfig,

	EAMode.NewInstalledAppDBServiceImpl,
	wire.Bind(new(EAMode.InstalledAppDBService), new(*EAMode.InstalledAppDBServiceImpl)),
)
