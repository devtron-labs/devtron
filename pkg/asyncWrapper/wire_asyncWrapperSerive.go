package asyncWrapper

import (
	"github.com/google/wire"
)

var ServiceWire = wire.NewSet(
	NewAsyncGoFuncServiceImpl,
	wire.Bind(new(AsyncGoFuncService), new(*AsyncGoFuncServiceImpl)),
)
