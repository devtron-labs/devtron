/*
 * Copyright (c) 2024. Devtron Inc.
 */

package resourceScan

import (
	"github.com/google/wire"
)

var ScanningResultWireSet = wire.NewSet(
	NewScanningResultRouterImpl,
	wire.Bind(new(ScanningResultRouter), new(*ScanningResultRouterImpl)),
	NewScanningResultRestHandlerImpl,
	wire.Bind(new(ScanningResultRestHandler), new(*ScanningResultRestHandlerImpl)),
)
