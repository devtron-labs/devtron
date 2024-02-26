package policyGovernance

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactApproval"
	"github.com/google/wire"
)

var PolicyGovernanceWireSet = wire.NewSet(
	artifactApproval.ArtifactApprovalWireSet,
)
