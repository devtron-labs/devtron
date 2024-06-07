/*
 * Copyright (c) 2024. Devtron Inc.
 */

package scanningResultsParser

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/scanningResultsParser"
	"github.com/google/wire"
)

var ScanningResultWireSet = wire.NewSet(
	NewScanningResultRouterImpl,
	wire.Bind(new(ScanningResultRouter), new(*ScanningResultRouterImpl)),
	NewScanningResultRestHandlerImpl,
	wire.Bind(new(ScanningResultRestHandler), new(*ScanningResultRestHandlerImpl)),
	scanningResultsParser.NewServiceImpl,
	wire.Bind(new(scanningResultsParser.Service), new(*scanningResultsParser.ServiceImpl)),
)
