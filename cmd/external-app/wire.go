//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/devtron/internals/util"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		CommonWireSet,
		NewApp,
		NewMuxRouter,
		util.NewHttpClient,
		util.NewSugardLogger,
		util.IntValidator,
	)
	return &App{}, nil
}
