package out

import "github.com/google/wire"

var EventProcessorOutWireSet = wire.NewSet(
	NewWorkflowEventPublishServiceImpl,
	wire.Bind(new(WorkflowEventPublishService), new(*WorkflowEventPublishServiceImpl)),
)
