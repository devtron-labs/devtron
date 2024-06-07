/*
 * Copyright (c) 2024. Devtron Inc.
 */

package status

import "github.com/google/wire"

var WorkflowStatusWireSet = wire.NewSet(
	NewWorkflowStatusServiceImpl,
	wire.Bind(new(WorkflowStatusService), new(*WorkflowStatusServiceImpl)),
)
