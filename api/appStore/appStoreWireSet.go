package appStore

import (
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/common"
	"github.com/google/wire"
)

var AppStoreCommonWireSetEA = wire.NewSet(
	service.NewAppStoreDeploymentDBServiceImpl,
	wire.Bind(new(service.AppStoreDeploymentDBService), new(*service.AppStoreDeploymentDBServiceImpl)),
	service.NewAppStoreDeploymentServiceImpl,
	wire.Bind(new(service.AppStoreDeploymentService), new(*service.AppStoreDeploymentServiceImpl)),
	service.NewDeletePostProcessorImpl,
	wire.Bind(new(service.DeletePostProcessor), new(*service.DeletePostProcessorImpl)),
	service.NewAppAppStoreValidatorImpl,
	wire.Bind(new(service.AppStoreValidator), new(*service.AppStoreValidatorImpl)),
)

var AppStoreCommonWireSet = wire.NewSet(
	AppStoreCommonWireSetEA,
	appStoreDeploymentCommon.NewAppStoreDeploymentCommonServiceImpl,
	wire.Bind(new(appStoreDeploymentCommon.AppStoreDeploymentCommonService), new(*appStoreDeploymentCommon.AppStoreDeploymentCommonServiceImpl)),
	EAMode.NewEAModeDeploymentServiceImpl,
	wire.Bind(new(EAMode.EAModeDeploymentService), new(*EAMode.EAModeDeploymentServiceImpl)),
)
