package out

import "github.com/google/wire"

var EventProcessorOutWireSet = wire.NewSet(
	NewWorkflowEventPublishServiceImpl,
	wire.Bind(new(WorkflowEventPublishService), new(*WorkflowEventPublishServiceImpl)),

	NewPipelineConfigEventPublishServiceImpl,
	wire.Bind(new(PipelineConfigEventPublishService), new(*PipelineConfigEventPublishServiceImpl)),

	NewCDPipelineEventPublishServiceImpl,
	wire.Bind(new(CDPipelineEventPublishService), new(*CDPipelineEventPublishServiceImpl)),

	NewAppStoreAppsEventPublishServiceImpl,
	wire.Bind(new(AppStoreAppsEventPublishService), new(*AppStoreAppsEventPublishServiceImpl)),
)
