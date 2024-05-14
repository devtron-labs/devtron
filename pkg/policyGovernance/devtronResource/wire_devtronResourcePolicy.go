package devtronResource

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/devtronResource/release"
	"github.com/google/wire"
)

var PolicyWireSet = wire.NewSet(
	release.ReleasePolicyWireSet,
)
