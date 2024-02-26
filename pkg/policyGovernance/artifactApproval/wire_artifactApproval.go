package artifactApproval

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactApproval/action"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactApproval/read"
	"github.com/google/wire"
)

var ArtifactApprovalWireSet = wire.NewSet(
	read.NewArtifactApprovalDataReadServiceImpl,
	wire.Bind(new(read.ArtifactApprovalDataReadService), new(*read.ArtifactApprovalDataReadServiceImpl)),

	action.NewArtifactApprovalActionServiceImpl,
	wire.Bind(new(action.ArtifactApprovalActionService), new(*action.ArtifactApprovalActionServiceImpl)),
)
