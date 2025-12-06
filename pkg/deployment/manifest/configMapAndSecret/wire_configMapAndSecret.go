package configMapAndSecret

import (
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/configMapAndSecret/read"
	"github.com/google/wire"
)

var ConfigMapAndSecretWireSet = wire.NewSet(
	read.NewConfigMapHistoryReadService,
	wire.Bind(new(read.ConfigMapHistoryReadService), new(*read.ConfigMapHistoryReadServiceImpl)),
)
