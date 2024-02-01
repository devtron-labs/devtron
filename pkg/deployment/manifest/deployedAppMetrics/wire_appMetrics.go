package deployedAppMetrics

import (
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics/repository"
	"github.com/google/wire"
)

var AppMetricsWireSet = wire.NewSet(
	repository.NewAppLevelMetricsRepositoryImpl,
	wire.Bind(new(repository.AppLevelMetricsRepository), new(*repository.AppLevelMetricsRepositoryImpl)),
	repository.NewEnvLevelAppMetricsRepositoryImpl,
	wire.Bind(new(repository.EnvLevelAppMetricsRepository), new(*repository.EnvLevelAppMetricsRepositoryImpl)),

	NewDeployedAppMetricsServiceImpl,
	wire.Bind(new(DeployedAppMetricsService), new(*DeployedAppMetricsServiceImpl)),
)
