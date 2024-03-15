package bean

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/constants"
	"github.com/devtron-labs/devtron/util"
)

type PromotionMaterialRequest struct {
	Resource              string // CI, CD, WEBHOOK, PROMOTION_APPROVAL_PENDING_NODE
	ResourceName          string
	ResourceId            int
	AppId                 int
	WorkflowId            int
	PendingForCurrentUser bool
	CiPipelineId          int
	CdPipelineId          int
	ExternalCiPipelineId  int
	util.ListingFilterOptions
}

func (p PromotionMaterialRequest) IsCINode() bool {
	return p.Resource == string(constants.SOURCE_TYPE_CI)
}

func (p PromotionMaterialRequest) IsCDNode() bool {
	return p.Resource == string(constants.SOURCE_TYPE_CD)
}

func (p PromotionMaterialRequest) IsWebhookNode() bool {
	return p.Resource == string(constants.SOURCE_TYPE_WEBHOOK)
}

func (p PromotionMaterialRequest) IsPromotionApprovalPendingNode() bool {
	return p.Resource == string(constants.PROMOTION_APPROVAL_PENDING_NODE)
}

func (p PromotionMaterialRequest) IsPendingForUserRequest() bool {
	return p.Resource == string(constants.PROMOTION_APPROVAL_PENDING_NODE) && p.PendingForCurrentUser
}

type PromotionRequestWfMetadata struct {
	cdPipelineIds     []int
	authCdPipelineIds []int
}

func (p *PromotionRequestWfMetadata) SetCdPipelineIds(cdPipelineIds []int) {
	p.cdPipelineIds = cdPipelineIds
}

func (p *PromotionRequestWfMetadata) SetAuthCdPipelineIds(authCdPipelineIds []int) {
	p.authCdPipelineIds = authCdPipelineIds
}

func (p PromotionRequestWfMetadata) GetCdPipelineIds() []int {
	return p.cdPipelineIds
}

func (p PromotionRequestWfMetadata) GetAuthCdPipelineIds() []int {
	return p.authCdPipelineIds
}
