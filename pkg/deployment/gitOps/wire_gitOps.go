package gitOps

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"github.com/google/wire"
)

var GitOpsWireSet = wire.NewSet(
	repository.NewGitOpsConfigRepositoryImpl,
	wire.Bind(new(repository.GitOpsConfigRepository), new(*repository.GitOpsConfigRepositoryImpl)),

	config.NewGitOpsConfigReadServiceImpl,
	wire.Bind(new(config.GitOpsConfigReadService), new(*config.GitOpsConfigReadServiceImpl)),

	git.NewGitOperationServiceImpl,
	wire.Bind(new(git.GitOperationService), new(*git.GitOperationServiceImpl)),
)

var GitOpsEAWireSet = wire.NewSet(
	repository.NewGitOpsConfigRepositoryImpl,
	wire.Bind(new(repository.GitOpsConfigRepository), new(*repository.GitOpsConfigRepositoryImpl)),

	config.NewGitOpsConfigReadServiceImpl,
	wire.Bind(new(config.GitOpsConfigReadService), new(*config.GitOpsConfigReadServiceImpl)),
)
