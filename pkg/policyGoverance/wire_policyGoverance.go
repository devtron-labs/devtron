package policyGoverance

import (
	"github.com/devtron-labs/devtron/pkg/policyGoverance/security/imageScanning"
	"github.com/google/wire"
)

var PolicyGoveranceWireSet = wire.NewSet(
	imageScanning.ImageScanningWireSet,
)
