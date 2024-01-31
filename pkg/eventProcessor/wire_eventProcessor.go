package eventProcessor

import (
	"github.com/devtron-labs/devtron/pkg/eventProcessor/in"
	"github.com/google/wire"
)

var EventProcessorWireSet = wire.NewSet(
	NewCentralEventProcessor,
	in.EventProcessorInWireSet,
)
