package casbin

import (
	"github.com/devtron-labs/devtron/pkg/user/casbin/client"
	"github.com/google/wire"
)

var CasbinWireSet = wire.NewSet(
	client.NewCasbinClientImpl,
	wire.Bind(new(client.CasbinClient), new(*client.CasbinClientImpl)),
	NewCasbinServiceImpl,
	wire.Bind(new(CasbinService), new(*CasbinServiceImpl)),
	client.GetConfig,
)
