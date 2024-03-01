package in

import "github.com/google/wire"

var EventProcessorInWireSet = wire.NewSet(
	NewCIPipelineEventProcessorImpl,
	NewWorkflowEventProcessorImpl,
	NewDeployedApplicationEventProcessorImpl,
	NewCDPipelineEventProcessorImpl,
)
