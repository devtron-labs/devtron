package chart_repository

import (
	chart_repository_service "github.com/devtron-labs/devtron/pkg/chart-repository"
	chart_repository_repo "github.com/devtron-labs/devtron/pkg/chart-repository/repository"
	"github.com/google/wire"
)

var ChartRepositoryWireSet = wire.NewSet(
	chart_repository_repo.NewChartRepoRepositoryImpl,
	wire.Bind(new(chart_repository_repo.ChartRepoRepository), new(*chart_repository_repo.ChartRepoRepositoryImpl)),
	chart_repository_repo.NewChartRefRepositoryImpl,
	wire.Bind(new(chart_repository_repo.ChartRefRepository), new(*chart_repository_repo.ChartRefRepositoryImpl)),
	chart_repository_service.NewChartRepositoryServiceImpl,
	wire.Bind(new(chart_repository_service.ChartRepositoryService), new(*chart_repository_service.ChartRepositoryServiceImpl)),
	NewChartRepositoryRestHandlerImpl,
	wire.Bind(new(ChartRepositoryRestHandler), new(*ChartRepositoryRestHandlerImpl)),
	NewChartRepositoryRouterImpl,
	wire.Bind(new(ChartRepositoryRouter), new(*ChartRepositoryRouterImpl)),
)
