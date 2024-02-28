package artifactPromotion

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/read"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/repository"
	"github.com/google/wire"
)

var ArtifactPromotionWireSet = wire.NewSet(

	read.NewArtifactPromotionDataReadServiceImpl,
	wire.Bind(new(read.ArtifactPromotionDataReadService), new(*read.ArtifactPromotionDataReadServiceImpl)),

	NewArtifactPromotionApprovalServiceImpl,
	wire.Bind(new(ArtifactPromotionApprovalService), new(*ArtifactPromotionApprovalServiceImpl)),

	NewPromotionPolicyServiceImpl,
	wire.Bind(new(PromotionPolicyCUDService), new(*PromotionPolicyServiceImpl)),

	repository.NewArtifactPromotionApprovalRequestImpl,
	wire.Bind(new(repository.ArtifactPromotionApprovalRequestRepository), new(*repository.ArtifactPromotionApprovalRequestRepoImpl)),
)
