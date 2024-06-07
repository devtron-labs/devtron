/*
 * Copyright (c) 2024. Devtron Inc.
 */

package appStore

import (
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service"
	deployment3 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode/deployment"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/common"
	"github.com/google/wire"
)

var AppStoreWireSet = wire.NewSet(

	service.NewDeletePostProcessorEnterpriseImpl,
	wire.Bind(new(service.DeletePostProcessor), new(*service.DeletePostProcessorEnterpriseImpl)),

	service.NewAppStoreValidatorEnterpriseImpl,
	wire.Bind(new(service.AppStoreValidator), new(*service.AppStoreValidatorEnterpriseImpl)),

	NewAppStoreRouterEnterpriseImpl,
	wire.Bind(new(AppStoreRouterEnterprise), new(*AppStoreRouterEnterpriseImpl)),

	NewInstalledAppRestHandlerEnterpriseImpl,
	wire.Bind(new(InstalledAppRestHandlerEnterprise), new(*InstalledAppRestHandlerEnterpriseImpl)),

	deployment3.NewFullModeDeploymentServiceEnterpriseImpl,
	wire.Bind(new(deployment3.FullModeDeploymentServiceEnterprise), new(*deployment3.FullModeDeploymentServiceEnterpriseImpl)),

	appStoreDeploymentCommon.NewAppStoreDeploymentCommonServiceEnterpriseImpl,
	wire.Bind(new(appStoreDeploymentCommon.AppStoreDeploymentCommonServiceEnterprise), new(*appStoreDeploymentCommon.AppStoreDeploymentCommonServiceEnterpriseImpl)),
)
