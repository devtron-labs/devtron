package globalTag

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/globalTag"
	"github.com/google/wire"
)

var GlobalTagWireSet = wire.NewSet(
	globalTag.NewGlobalTagRepositoryImpl,
	wire.Bind(new(globalTag.GlobalTagRepository), new(*globalTag.GlobalTagRepositoryImpl)),
	globalTag.NewGlobalTagServiceImpl,
	wire.Bind(new(globalTag.GlobalTagService), new(*globalTag.GlobalTagServiceImpl)),
	NewGlobalTagRestHandlerImpl,
	wire.Bind(new(GlobalTagRestHandler), new(*GlobalTagRestHandlerImpl)),
	NewGlobalTagRouterImpl,
	wire.Bind(new(GlobalTagRouter), new(*GlobalTagRouterImpl)),
)
