package deploymentTemplate

import (
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	"github.com/google/wire"
)

var DeploymentTemplateWireSet = wire.NewSet(
	NewDeploymentTemplateServiceImpl,
	wire.Bind(new(DeploymentTemplateService), new(*DeploymentTemplateServiceImpl)),
	NewDeploymentTemplateValidationServiceImpl,
	wire.Bind(new(DeploymentTemplateValidationService), new(*DeploymentTemplateValidationServiceImpl)),
	chartRef.NewChartRefServiceImpl,
	wire.Bind(new(chartRef.ChartRefService), new(*chartRef.ChartRefServiceImpl)),
)
