//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/devtron/pkg/sourceController/api"
	"github.com/devtron-labs/devtron/pkg/sourceController/internal/logger"
	"github.com/devtron-labs/devtron/pkg/sourceController/sql"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		NewApp,
		logger.NewSugardLogger,
		api.NewRouter,
		sql.GetConfig,
		sql.NewDbConnection,
		//GetSourceControllerConfig,

		//repository.NewCiArtifactRepositoryImpl,
		//wire.Bind(new(repository.CiArtifactRepository), new(*repository.CiArtifactRepositoryImpl)),

		//NewSourceControllerServiceImpl,
		//wire.Bind(new(SourceControllerService), new(*SourceControllerServiceImpl)),
		//
		//NewSourceControllerCronServiceImpl,
		//wire.Bind(new(SourceControllerCronService), new(*SourceControllerCronServiceImpl)),
	)
	return &App{}, nil
}
