package policyGovernance

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning"
	"github.com/google/wire"
)

var PolicyGoveranceWireSet = wire.NewSet(
	imageScanning.ImageScanningWireSet,
)
