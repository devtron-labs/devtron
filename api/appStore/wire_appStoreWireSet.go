/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
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
