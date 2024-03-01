//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/devtron/api/appStore"
	"github.com/devtron-labs/devtron/cmd/external-app/commonWireset"
	"github.com/devtron-labs/devtron/internals/util"
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
		appStore.AppStoreCommonWireSetEA,
	)
	return &App{}, nil
}
