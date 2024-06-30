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

package chartRepo

import (
	chartRepo "github.com/devtron-labs/devtron/pkg/chartRepo"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/google/wire"
)

var ChartRepositoryWireSet = wire.NewSet(
	chartRepoRepository.NewChartRepoRepositoryImpl,
	wire.Bind(new(chartRepoRepository.ChartRepoRepository), new(*chartRepoRepository.ChartRepoRepositoryImpl)),
	chartRepoRepository.NewChartRefRepositoryImpl,
	wire.Bind(new(chartRepoRepository.ChartRefRepository), new(*chartRepoRepository.ChartRefRepositoryImpl)),
	chartRepoRepository.NewChartRepository,
	wire.Bind(new(chartRepoRepository.ChartRepository), new(*chartRepoRepository.ChartRepositoryImpl)),
	chartRepo.NewChartRepositoryServiceImpl,
	wire.Bind(new(chartRepo.ChartRepositoryService), new(*chartRepo.ChartRepositoryServiceImpl)),
	NewChartRepositoryRestHandlerImpl,
	wire.Bind(new(ChartRepositoryRestHandler), new(*ChartRepositoryRestHandlerImpl)),
	NewChartRepositoryRouterImpl,
	wire.Bind(new(ChartRepositoryRouter), new(*ChartRepositoryRouterImpl)),
)
