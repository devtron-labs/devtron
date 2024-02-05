package build

import (
	"github.com/devtron-labs/devtron/pkg/build/artifacts"
	"github.com/google/wire"
)

var BuildWireSet = wire.NewSet(
	artifacts.ArtifactsWireSet,
)
