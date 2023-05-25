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
