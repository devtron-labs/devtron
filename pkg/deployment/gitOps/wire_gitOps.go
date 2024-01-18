package gitOps

import (
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/remote"
	"github.com/google/wire"
)

var GitOpsWireSet = wire.NewSet(
	config.NewGitOpsConfigReadServiceImpl,
	wire.Bind(new(config.GitOpsConfigReadService), new(*config.GitOpsConfigReadServiceImpl)),

	remote.NewGitOpsRemoteOperationServiceImpl,
	wire.Bind(new(remote.GitOpsRemoteOperationService), new(*remote.GitOpsRemoteOperationServiceImpl)),
)
