/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package appStoreDeployment

import (
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/util"
	installedAppReader "github.com/devtron-labs/devtron/pkg/appStore/installedApp/read"
	repository3 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	deployment2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode/deployment"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode/deploymentTypeChange"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode/resource"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/common"
	"github.com/google/wire"
)

var EAModeWireSet = wire.NewSet(
	//util.GetDeploymentServiceTypeConfig,
	util.NewChartTemplateServiceImpl,
	wire.Bind(new(util.ChartTemplateService), new(*util.ChartTemplateServiceImpl)),
	appStoreDeploymentCommon.NewAppStoreDeploymentCommonServiceImpl,
	wire.Bind(new(appStoreDeploymentCommon.AppStoreDeploymentCommonService), new(*appStoreDeploymentCommon.AppStoreDeploymentCommonServiceImpl)),
	deployment2.NewEAModeDeploymentServiceImpl,
	wire.Bind(new(deployment2.EAModeDeploymentService), new(*deployment2.EAModeDeploymentServiceImpl)),
	service.NewAppStoreDeploymentServiceImpl,
	wire.Bind(new(service.AppStoreDeploymentService), new(*service.AppStoreDeploymentServiceImpl)),
	service.NewAppStoreDeploymentDBServiceImpl,
	wire.Bind(new(service.AppStoreDeploymentDBService), new(*service.AppStoreDeploymentDBServiceImpl)),
	NewAppStoreDeploymentRestHandlerImpl,
	wire.Bind(new(AppStoreDeploymentRestHandler), new(*AppStoreDeploymentRestHandlerImpl)),
	NewAppStoreDeploymentRouterImpl,
	wire.Bind(new(AppStoreDeploymentRouter), new(*AppStoreDeploymentRouterImpl)),
	repository3.NewInstalledAppVersionHistoryRepositoryImpl,
	wire.Bind(new(repository3.InstalledAppVersionHistoryRepository), new(*repository3.InstalledAppVersionHistoryRepositoryImpl)),

	NewCommonDeploymentRestHandlerImpl,
	wire.Bind(new(CommonDeploymentRestHandler), new(*CommonDeploymentRestHandlerImpl)),
	NewCommonDeploymentRouterImpl,
	wire.Bind(new(CommonDeploymentRouter), new(*CommonDeploymentRouterImpl)),
	argocdServer.GetACDDeploymentConfig,

	EAMode.NewInstalledAppDBServiceImpl,
	wire.Bind(new(EAMode.InstalledAppDBService), new(*EAMode.InstalledAppDBServiceImpl)),

	installedAppReader.EAWireSet,
)

var FullModeWireSet = wire.NewSet(

	EAModeWireSet,

	FullMode.NewInstalledAppDBExtendedServiceImpl,
	wire.Bind(new(FullMode.InstalledAppDBExtendedService), new(*FullMode.InstalledAppDBExtendedServiceImpl)),

	resource.NewInstalledAppResourceServiceImpl,
	wire.Bind(new(resource.InstalledAppResourceService), new(*resource.InstalledAppResourceServiceImpl)),

	deploymentTypeChange.NewInstalledAppDeploymentTypeChangeServiceImpl,
	wire.Bind(new(deploymentTypeChange.InstalledAppDeploymentTypeChangeService), new(*deploymentTypeChange.InstalledAppDeploymentTypeChangeServiceImpl)),

	installedAppReader.WireSet,
)
