package chart_repo

import (
	chart_repo "github.com/devtron-labs/devtron/pkg/chart-repo"
	chart_repo_repository "github.com/devtron-labs/devtron/pkg/chart-repo/repository"
	"github.com/google/wire"
)

var ChartRepositoryWireSet = wire.NewSet(
	chart_repo_repository.NewChartRepoRepositoryImpl,
	wire.Bind(new(chart_repo_repository.ChartRepoRepository), new(*chart_repo_repository.ChartRepoRepositoryImpl)),
	chart_repo_repository.NewChartRefRepositoryImpl,
	wire.Bind(new(chart_repo_repository.ChartRefRepository), new(*chart_repo_repository.ChartRefRepositoryImpl)),
	chart_repo_repository.NewChartRepository,
	wire.Bind(new(chart_repo_repository.ChartRepository), new(*chart_repo_repository.ChartRepositoryImpl)),
	chart_repo.NewChartRepositoryServiceImpl,
	wire.Bind(new(chart_repo.ChartRepositoryService), new(*chart_repo.ChartRepositoryServiceImpl)),
	NewChartRepositoryRestHandlerImpl,
	wire.Bind(new(ChartRepositoryRestHandler), new(*ChartRepositoryRestHandlerImpl)),
	NewChartRepositoryRouterImpl,
	wire.Bind(new(ChartRepositoryRouter), new(*ChartRepositoryRouterImpl)),
)
