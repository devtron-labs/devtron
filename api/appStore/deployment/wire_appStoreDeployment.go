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
	//util.GetDeploymentServiceTypeConfig,
	repository.NewClusterInstalledAppsRepositoryImpl,
	wire.Bind(new(repository.ClusterInstalledAppsRepository), new(*repository.ClusterInstalledAppsRepositoryImpl)),
	appStoreDeploymentCommon.NewAppStoreDeploymentCommonServiceImpl,
	wire.Bind(new(appStoreDeploymentCommon.AppStoreDeploymentCommonService), new(*appStoreDeploymentCommon.AppStoreDeploymentCommonServiceImpl)),
	EAMode.NewEAModeDeploymentServiceImpl,
	wire.Bind(new(EAMode.EAModeDeploymentService), new(*EAMode.EAModeDeploymentServiceImpl)),
	service.NewAppStoreDeploymentServiceImpl,
	wire.Bind(new(service.AppStoreDeploymentService), new(*service.AppStoreDeploymentServiceImpl)),
	service.NewAppStoreDeploymentDBServiceImpl,
	wire.Bind(new(service.AppStoreDeploymentDBService), new(*service.AppStoreDeploymentDBServiceImpl)),
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
