package deployment

import (
	"github.com/devtron-labs/devtron/pkg/deployment/deployedApp"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger"
	"github.com/google/wire"
)

// TODO: add separate wire sets for full and ea mode when reached that level of transparency

var DeploymentWireSet = wire.NewSet(
	gitOps.GitOpsWireSet,
	manifest.DeploymentManifestWireSet,
	trigger.DeploymentTriggerWireSet,
	deployedApp.DeployedAppWireSet,
)
