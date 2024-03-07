package artifactPromotion

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/read"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/repository"
	"github.com/google/wire"
)

var ArtifactPromotionWireSet = wire.NewSet(

	read.NewArtifactPromotionDataReadServiceImpl,
	wire.Bind(new(read.ArtifactPromotionDataReadService), new(*read.ArtifactPromotionDataReadServiceImpl)),

	NewApprovalRequestServiceImpl,
	wire.Bind(new(ApprovalRequestService), new(*ApprovalRequestServiceImpl)),

	NewPromotionPolicyServiceImpl,
	wire.Bind(new(PolicyCUDService), new(*PromotionPolicyServiceImpl)),

	repository.NewRequestRepositoryImpl,
	wire.Bind(new(repository.RequestRepository), new(*repository.RequestRepositoryImpl)),
)
