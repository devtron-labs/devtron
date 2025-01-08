package status

import (
	"github.com/devtron-labs/devtron/pkg/deployment/deployedApp/status/resourceTree"
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	resourceTree.WireSet,
)
