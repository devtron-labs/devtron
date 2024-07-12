package asyncProvider

import (
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	NewAsyncRunnable,
)
