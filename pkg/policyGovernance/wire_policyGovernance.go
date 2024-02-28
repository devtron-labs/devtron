package policyGovernance

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactApproval"
	artifactPromotion2 "github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion"
	"github.com/google/wire"
)

var PolicyGovernanceWireSet = wire.NewSet(
	artifactApproval.ArtifactApprovalWireSet,
	artifactPromotion2.ArtifactPromotionWireSet,
)
