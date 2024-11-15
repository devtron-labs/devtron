package gitProvider

import (
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider/read"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider/repository"
	"github.com/google/wire"
)

var GitProviderWireSet = wire.NewSet(
	read.NewGitProviderReadService,
	wire.Bind(new(read.GitProviderReadService), new(*read.GitProviderReadServiceImpl)),

	repository.NewGitProviderRepositoryImpl,
	wire.Bind(new(repository.GitProviderRepository), new(*repository.GitProviderRepositoryImpl)),
	NewGitRegistryConfigImpl,
	wire.Bind(new(GitRegistryConfig), new(*GitRegistryConfigImpl)),
)
