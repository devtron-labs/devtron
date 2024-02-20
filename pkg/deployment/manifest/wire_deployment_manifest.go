package manifest

import (
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate"
	"github.com/google/wire"
)

var DeploymentManifestWireSet = wire.NewSet(
	deployedAppMetrics.AppMetricsWireSet,
	deploymentTemplate.DeploymentTemplateWireSet,

	NewManifestCreationServiceImpl,
	wire.Bind(new(ManifestCreationService), new(*ManifestCreationServiceImpl)),
)
