package out

import (
	"github.com/devtron-labs/devtron/pkg/eventProcessor/celEvaluator"
	"github.com/google/wire"
)

var EventProcessorOutWireSet = wire.NewSet(
	celEvaluator.NewTriggerEventEvaluatorImpl,
	wire.Bind(new(celEvaluator.TriggerEventEvaluator), new(*celEvaluator.TriggerEventEvaluatorImpl)),

	NewWorkflowEventPublishServiceImpl,
	wire.Bind(new(WorkflowEventPublishService), new(*WorkflowEventPublishServiceImpl)),

	NewPipelineConfigEventPublishServiceImpl,
	wire.Bind(new(PipelineConfigEventPublishService), new(*PipelineConfigEventPublishServiceImpl)),

	NewCDPipelineEventPublishServiceImpl,
	wire.Bind(new(CDPipelineEventPublishService), new(*CDPipelineEventPublishServiceImpl)),

	NewAppStoreAppsEventPublishServiceImpl,
	wire.Bind(new(AppStoreAppsEventPublishService), new(*AppStoreAppsEventPublishServiceImpl)),

	NewCIPipelineEventPublishServiceImpl,
	wire.Bind(new(CIPipelineEventPublishService), new(*CIPipelineEventPublishServiceImpl)),
)
