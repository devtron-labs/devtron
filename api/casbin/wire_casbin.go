package client

import (
	"github.com/google/wire"
)

var CasbinWireSet = wire.NewSet(
	NewCasbinClientImpl,
	wire.Bind(new(CasbinClient), new(*CasbinClientImpl)),
	NewCasbinServiceImpl,
	wire.Bind(new(CasbinService), new(*CasbinServiceImpl)),
	GetConfig,
)
