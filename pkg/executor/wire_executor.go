package executor

import "github.com/google/wire"

var ExecutorWireSet = wire.NewSet(
	NewWorkflowServiceImpl,
	wire.Bind(new(WorkflowService), new(*WorkflowServiceImpl)),
)
