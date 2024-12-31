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

package appStoreDiscover

import (
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/discover/service"
	"github.com/google/wire"
)

var AppStoreDiscoverWireSet = wire.NewSet(
	appStoreDiscoverRepository.NewAppStoreRepositoryImpl,
	wire.Bind(new(appStoreDiscoverRepository.AppStoreRepository), new(*appStoreDiscoverRepository.AppStoreRepositoryImpl)),
	appStoreDiscoverRepository.NewAppStoreApplicationVersionRepositoryImpl,
	wire.Bind(new(appStoreDiscoverRepository.AppStoreApplicationVersionRepository), new(*appStoreDiscoverRepository.AppStoreApplicationVersionRepositoryImpl)),
	service.NewAppStoreServiceImpl,
	wire.Bind(new(service.AppStoreService), new(*service.AppStoreServiceImpl)),
	NewAppStoreRestHandlerImpl,
	wire.Bind(new(AppStoreRestHandler), new(*AppStoreRestHandlerImpl)),
	NewAppStoreDiscoverRouterImpl,
	wire.Bind(new(AppStoreDiscoverRouter), new(*AppStoreDiscoverRouterImpl)),
)
