/*
 * Copyright (c) 2024. Devtron Inc.
 */

package scanTool

import (
	"github.com/google/wire"
)

var ScanToolMetadataWireSet = wire.NewSet(
	NewScanToolRouterImpl,
	wire.Bind(new(ScanToolRouter), new(*ScanToolRouterImpl)),
	NewScanToolRestHandlerImpl,
	wire.Bind(new(ScanToolRestHandler), new(*ScanToolRestHandlerImpl)),
)
