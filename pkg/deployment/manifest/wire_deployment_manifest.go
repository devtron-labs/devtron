package manifest

import (
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	"github.com/google/wire"
)

var DeploymentManifestWireSet = wire.NewSet(
	deployedAppMetrics.NewDeployedAppMetricsServiceImpl,
	wire.Bind(new(deployedAppMetrics.DeployedAppMetricsService), new(*deployedAppMetrics.DeployedAppMetricsServiceImpl)),
	deploymentTemplate.NewDeploymentTemplateServiceImpl,
	wire.Bind(new(deploymentTemplate.DeploymentTemplateService), new(*deploymentTemplate.DeploymentTemplateServiceImpl)),
	deploymentTemplate.NewDeploymentTemplateValidationServiceImpl,
	wire.Bind(new(deploymentTemplate.DeploymentTemplateValidationService), new(*deploymentTemplate.DeploymentTemplateValidationServiceImpl)),
	chartRef.NewChartRefServiceImpl,
	wire.Bind(new(chartRef.ChartRefService), new(*chartRef.ChartRefServiceImpl)),
)
