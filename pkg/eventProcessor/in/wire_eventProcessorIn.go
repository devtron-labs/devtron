/*
 * Copyright (c) 2024. Devtron Inc.
 */

package in

import "github.com/google/wire"

var EventProcessorInWireSet = wire.NewSet(
	NewCIPipelineEventProcessorImpl,
	NewWorkflowEventProcessorImpl,
	NewDeployedApplicationEventProcessorImpl,
	NewCDPipelineEventProcessorImpl,
	NewAppStoreAppsEventProcessorImpl,
	NewChartScanEventProcessorImpl,
)
