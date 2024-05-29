/*
 * Copyright (c) 2024. Devtron Inc.
 */

package artifacts

import "github.com/google/wire"

var ArtifactsWireSet = wire.NewSet(
	NewCommonArtifactServiceImpl,
	wire.Bind(new(CommonArtifactService), new(*CommonArtifactServiceImpl)),
)
