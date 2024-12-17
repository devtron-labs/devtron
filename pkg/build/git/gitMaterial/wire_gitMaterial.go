package gitMaterial

import (
	"github.com/devtron-labs/devtron/pkg/build/git/gitMaterial/read"
	repository3 "github.com/devtron-labs/devtron/pkg/build/git/gitMaterial/repository"
	"github.com/google/wire"
)

var GitMaterialWireSet = wire.NewSet(
	read.NewGitMaterialReadServiceImpl,
	wire.Bind(new(read.GitMaterialReadService), new(*read.GitMaterialReadServiceImpl)),

	repository3.NewMaterialRepositoryImpl,
	wire.Bind(new(repository3.MaterialRepository), new(*repository3.MaterialRepositoryImpl)),
)
