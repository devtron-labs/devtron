package gitOps

import (
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"github.com/google/wire"
)

var GitOpsWireSet = wire.NewSet(
	config.NewGitOpsConfigReadServiceImpl,
	wire.Bind(new(config.GitOpsConfigReadService), new(*config.GitOpsConfigReadServiceImpl)),

	git.NewGitOpsRemoteOperationServiceImpl,
	wire.Bind(new(git.GitOpsRemoteOperationService), new(*git.GitOpsRemoteOperationServiceImpl)),
)
