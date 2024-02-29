//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/devtron/cmd/external-app/commonWireset"
	"github.com/devtron-labs/devtron/internals/util"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		commonWireset.CommonWireSet,
		NewApp,
		NewMuxRouter,
		util.NewHttpClient,
		util.NewSugardLogger,
		util.IntValidator,
		service.NewAppStoreDeploymentDBServiceImpl,
		wire.Bind(new(service.AppStoreDeploymentDBService), new(*service.AppStoreDeploymentDBServiceImpl)),
		service.NewAppStoreDeploymentServiceImpl,
		wire.Bind(new(service.AppStoreDeploymentService), new(*service.AppStoreDeploymentServiceImpl)),
		service.NewDeletePostProcessorImpl,
		wire.Bind(new(service.DeletePostProcessor), new(*service.DeletePostProcessorImpl)),
		service.NewAppAppStoreValidatorImpl,
		wire.Bind(new(service.AppStoreValidator), new(*service.AppStoreValidatorImpl)),
	)
	return &App{}, nil
}
