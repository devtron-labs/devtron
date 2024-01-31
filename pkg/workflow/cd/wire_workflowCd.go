package cd

import "github.com/google/wire"

var CdWorkflowWireSet = wire.NewSet(
	NewCdWorkflowCommonServiceImpl,
	wire.Bind(new(CdWorkflowCommonService), new(*CdWorkflowCommonServiceImpl)),
	NewCdWorkflowServiceImpl,
	wire.Bind(new(CdWorkflowService), new(*CdWorkflowServiceImpl)),
	NewCdWorkflowRunnerServiceImpl,
	wire.Bind(new(CdWorkflowRunnerService), new(*CdWorkflowRunnerServiceImpl)),
)
