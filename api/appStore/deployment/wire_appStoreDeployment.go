package appStoreDeployment

import (
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	appStoreDeploymentTool "github.com/devtron-labs/devtron/pkg/appStore/deployment/tool"
	"github.com/google/wire"
)

var AppStoreDeploymentWireSet = wire.NewSet(
	repository.NewClusterInstalledAppsRepositoryImpl,
	wire.Bind(new(repository.ClusterInstalledAppsRepository), new(*repository.ClusterInstalledAppsRepositoryImpl)),
	appStoreDeploymentCommon.NewAppStoreDeploymentCommonServiceImpl,
	wire.Bind(new(appStoreDeploymentCommon.AppStoreDeploymentCommonService), new(*appStoreDeploymentCommon.AppStoreDeploymentCommonServiceImpl)),
	appStoreDeploymentTool.NewAppStoreDeploymentHelmServiceImpl,
	wire.Bind(new(appStoreDeploymentTool.AppStoreDeploymentHelmService), new(*appStoreDeploymentTool.AppStoreDeploymentHelmServiceImpl)),
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
)
