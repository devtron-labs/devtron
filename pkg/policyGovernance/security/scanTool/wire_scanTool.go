package scanTool

import "github.com/google/wire"

var ScanToolWireSet = wire.NewSet(
	NewScanToolMetadataServiceImpl,
	wire.Bind(new(ScanToolMetadataService), new(*ScanToolMetadataServiceImpl)),
)
