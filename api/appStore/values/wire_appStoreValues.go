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

package appStoreValues

import (
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
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
