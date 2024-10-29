package history

import (
	"github.com/devtron-labs/devtron/pkg/pipeline/history/read"
	"github.com/google/wire"
)

var AllHistoryWireSet = wire.NewSet(
	read.NewDeploymentTemplateHistoryReadServiceImpl,
	wire.Bind(new(read.DeploymentTemplateHistoryReadService), new(*read.DeploymentTemplateHistoryReadServiceImpl)),
)
