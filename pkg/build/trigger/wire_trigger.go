package trigger

import (
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	NewServiceImpl,
	wire.Bind(new(Service), new(*ServiceImpl)),
)
