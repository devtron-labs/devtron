//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/devtron/pkg/sourceController/internal/logger"
	"github.com/devtron-labs/devtron/pkg/sourceController/sql"
	repository "github.com/devtron-labs/devtron/pkg/sourceController/sql/repo"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		NewApp,
		logger.NewSugardLogger(),
		logger.NewHttpClient,
		sql.GetConfig(),
		sql.NewDbConnection,
		GetSourceControllerConfig(),

		repository.NewCiArtifactRepositoryImpl,
		wire.Bind(new(repository.CiArtifactRepository), new(*repository.CiArtifactRepositoryImpl)),

		sourceController.NewSourceControllerServiceImpl,
		wire.Bind(new(sourceController.SourceControllerService), new(*sourceController.SourceControllerServiceImpl)),

		sourceController.NewSourceControllerCronServiceImpl,
		wire.Bind(new(sourceController.SourceControllerCronService), new(*sourceController.SourceControllerCronServiceImpl)),
	)
	return &App{}, nil
}
