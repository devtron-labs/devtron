package config

import (
	"github.com/devtron-labs/devtron/pkg/config/read"
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	read.NewConfigReadServiceImpl,
	wire.Bind(new(read.ConfigReadService), new(*read.ConfigReadServiceImpl)),
)
