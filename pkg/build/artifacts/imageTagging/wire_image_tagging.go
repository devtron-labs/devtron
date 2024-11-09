package imageTagging

import (
	"github.com/devtron-labs/devtron/pkg/build/artifacts/imageTagging/read"
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	read.NewImageTaggingReadServiceImpl,
	wire.Bind(new(read.ImageTaggingReadService), new(*read.ImageTaggingReadServiceImpl)),
)
