package workflow

import (
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	"github.com/devtron-labs/devtron/pkg/workflow/status"
	"github.com/google/wire"
)

var WorkflowWireSet = wire.NewSet(
	cd.CdWorkflowWireSet,
	status.WorkflowStatusWireSet,
)
