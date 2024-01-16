package manifest

import (
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	"github.com/google/wire"
)

var DeploymentManifestWireSet = wire.NewSet(
	repository.NewAppLevelMetricsRepositoryImpl,
	wire.Bind(new(repository.AppLevelMetricsRepository), new(*repository.AppLevelMetricsRepositoryImpl)),
	repository.NewEnvLevelAppMetricsRepositoryImpl,
	wire.Bind(new(repository.EnvLevelAppMetricsRepository), new(*repository.EnvLevelAppMetricsRepositoryImpl)),

	deployedAppMetrics.NewDeployedAppMetricsServiceImpl,
	wire.Bind(new(deployedAppMetrics.DeployedAppMetricsService), new(*deployedAppMetrics.DeployedAppMetricsServiceImpl)),
	deploymentTemplate.NewDeploymentTemplateServiceImpl,
	wire.Bind(new(deploymentTemplate.DeploymentTemplateService), new(*deploymentTemplate.DeploymentTemplateServiceImpl)),
	chartRef.NewChartRefServiceImpl,
	wire.Bind(new(chartRef.ChartRefService), new(*chartRef.ChartRefServiceImpl)),
)
