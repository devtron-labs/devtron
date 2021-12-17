package main

import (
	"github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/google/wire"
)

// AuthWireSet:	 set of components used to initialise authentication with dex
var AuthWireSet = wire.NewSet(
	wire.Value(client.LocalDevMode(false)),
	client.NewK8sClient,
	client.BuildDexConfig,
	client.GetSettings,
	middleware.NewSessionManager,
	middleware.NewUserLogin,
)
