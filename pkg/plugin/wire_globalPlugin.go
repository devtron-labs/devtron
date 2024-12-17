package plugin

import (
	repository6 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	repository6.NewGlobalPluginRepository,
	wire.Bind(new(repository6.GlobalPluginRepository), new(*repository6.GlobalPluginRepositoryImpl)),

	NewGlobalPluginService,
	wire.Bind(new(GlobalPluginService), new(*GlobalPluginServiceImpl)),
)
