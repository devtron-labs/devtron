package gitHost

import (
	"github.com/devtron-labs/devtron/pkg/build/git/gitHost/read"
	"github.com/devtron-labs/devtron/pkg/build/git/gitHost/repository"
	"github.com/google/wire"
)

var GitHostWireSet = wire.NewSet(
	NewGitHostConfigImpl,
	wire.Bind(new(GitHostConfig), new(*GitHostConfigImpl)),

	read.NewGitHostReadServiceImpl,
	wire.Bind(new(read.GitHostReadService), new(*read.GitHostReadServiceImpl)),

	repository.NewGitHostRepositoryImpl,
	wire.Bind(new(repository.GitHostRepository), new(*repository.GitHostRepositoryImpl)))
