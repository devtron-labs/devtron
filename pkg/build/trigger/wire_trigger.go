package trigger

import (
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	NewHandlerServiceImpl,
	wire.Bind(new(HandlerService), new(*HandlerServiceImpl)),
)
