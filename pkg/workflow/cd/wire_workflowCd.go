package cd

import (
	"github.com/devtron-labs/devtron/pkg/workflow/cd/configHistory"
	"github.com/google/wire"
)

var CdWorkflowWireSet = wire.NewSet(
	configHistory.NewPipelineConfigOverrideReadServiceImpl,
	wire.Bind(new(configHistory.PipelineConfigOverrideReadService), new(*configHistory.PipelineConfigOverrideReadServiceImpl)),
	NewCdWorkflowCommonServiceImpl,
	wire.Bind(new(CdWorkflowCommonService), new(*CdWorkflowCommonServiceImpl)),
	NewCdWorkflowServiceImpl,
	wire.Bind(new(CdWorkflowService), new(*CdWorkflowServiceImpl)),
	NewCdWorkflowRunnerServiceImpl,
	wire.Bind(new(CdWorkflowRunnerService), new(*CdWorkflowRunnerServiceImpl)),
)
