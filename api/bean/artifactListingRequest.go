package bean

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/constants"
	"github.com/devtron-labs/devtron/util"
)

type ArtifactsListFilterOptions struct {
	//list filter data
	Limit        int
	Offset       int
	SearchString string
	Order        string

	//self stage data
	PipelineId int
	StageType  WorkflowType

	// CiPipelineId is id of ci-pipeline present in the same app-workflow of PipelineId
	CiPipelineId int

	//parent satge data
	ParentCdId      int
	ParentId        int
	ParentStageType WorkflowType

	//excludeArtifactIds
	ExcludeArtifactIds []int

	//excludeWfRunners
	ExcludeWfrIds []int

	//ApprovalNode data
	ApprovalNodeConfigured bool
	ApproversCount         int

	//pluginStage
	PluginStage string

	// UseCdStageQueryV2 is to set query version
	UseCdStageQueryV2 bool
}

type CiNodeMaterialRequest struct {
	ciPipelineId   int
	listingOptions util.ListingFilterOptions
}

func (r *CiNodeMaterialRequest) WithCiPipelineId(id int) *CiNodeMaterialRequest {
	r.ciPipelineId = id
	return r
}

func (r *CiNodeMaterialRequest) WithListingFilterOptions(listingOptions util.ListingFilterOptions) *CiNodeMaterialRequest {
	r.listingOptions = listingOptions
	return r
}

func (r *CiNodeMaterialRequest) GetCiPipelineId() int {
	return r.ciPipelineId
}

func (r *CiNodeMaterialRequest) GetListingOptions() util.ListingFilterOptions {
	return r.listingOptions
}

type ExtCiNodeMaterialRequest struct {
	externalCiPipelineId int
	listingOptions       util.ListingFilterOptions
}

func (r *ExtCiNodeMaterialRequest) WithExtCiPipelineId(id int) *ExtCiNodeMaterialRequest {
	r.externalCiPipelineId = id
	return r
}

func (r *ExtCiNodeMaterialRequest) WithListingFilterOptions(listingOptions util.ListingFilterOptions) *ExtCiNodeMaterialRequest {
	r.listingOptions = listingOptions
	return r
}

func (r *ExtCiNodeMaterialRequest) GetExtCiPipelineId() int {
	return r.externalCiPipelineId
}

func (r *ExtCiNodeMaterialRequest) GetListingOptions() util.ListingFilterOptions {
	return r.listingOptions
}

type CdNodeMaterialRequest struct {
	resourceCdPipelineId int
	listingOptions       util.ListingFilterOptions
}

func (r *CdNodeMaterialRequest) WithCDPipelineId(id int) *CdNodeMaterialRequest {
	r.resourceCdPipelineId = id
	return r
}

func (r *CdNodeMaterialRequest) WithListingFilterOptions(listingOptions util.ListingFilterOptions) *CdNodeMaterialRequest {
	r.listingOptions = listingOptions
	return r
}

func (r *CdNodeMaterialRequest) GetCDPipelineId() int {
	return r.resourceCdPipelineId
}

func (r *CdNodeMaterialRequest) GetListingOptions() util.ListingFilterOptions {
	return r.listingOptions
}

type PromotionPendingNodeMaterialRequest struct {
	resourceCdPipelineId []int
	listingOptions       util.ListingFilterOptions
}

func (r *PromotionPendingNodeMaterialRequest) WithCDPipelineIds(ids []int) *PromotionPendingNodeMaterialRequest {
	r.resourceCdPipelineId = ids
	return r
}

func (r *PromotionPendingNodeMaterialRequest) WithListingFilterOptions(listingOptions util.ListingFilterOptions) *PromotionPendingNodeMaterialRequest {
	r.listingOptions = listingOptions
	return r
}

func (r *PromotionPendingNodeMaterialRequest) GetCDPipelineIds() []int {
	return r.resourceCdPipelineId
}

func (r *PromotionPendingNodeMaterialRequest) GetListingOptions() util.ListingFilterOptions {
	return r.listingOptions
}

type PromotionMaterialRequest struct {
	resource              string // CI, CD, WEBHOOK, PROMOTION_APPROVAL_PENDING_NODE
	resourceName          string
	resourceId            int
	appId                 int
	workflowId            int
	pendingForCurrentUser bool
	ciPipelineId          int
	cdPipelineId          int
	externalCiPipelineId  int
	util.ListingFilterOptions
}

func (p *PromotionMaterialRequest) WithResource(resource string) *PromotionMaterialRequest {
	p.resource = resource
	return p
}

func (p *PromotionMaterialRequest) GetResource() string {
	return p.resource
}

func (p *PromotionMaterialRequest) WithResourceName(resourceName string) *PromotionMaterialRequest {
	p.resourceName = resourceName
	return p
}

func (p *PromotionMaterialRequest) GetResourceName() string {
	return p.resourceName
}

func (p *PromotionMaterialRequest) WithResourceId(resourceId int) *PromotionMaterialRequest {
	p.resourceId = resourceId
	return p
}

func (p *PromotionMaterialRequest) GetResourceId() int {
	return p.resourceId
}

func (p *PromotionMaterialRequest) WithAppId(appId int) *PromotionMaterialRequest {
	p.appId = appId
	return p
}

func (p *PromotionMaterialRequest) GetAppId() int {
	return p.appId
}

func (p *PromotionMaterialRequest) WithWorkflowId(workflowId int) *PromotionMaterialRequest {
	p.workflowId = workflowId
	return p
}

func (p *PromotionMaterialRequest) GetWorkflowId() int {
	return p.workflowId
}

func (p *PromotionMaterialRequest) WithPendingForCurrentUser(pendingForCurrentUser bool) *PromotionMaterialRequest {
	p.pendingForCurrentUser = pendingForCurrentUser
	return p
}

func (p *PromotionMaterialRequest) GetPendingForCurrentUser() bool {
	return p.pendingForCurrentUser
}

func (p *PromotionMaterialRequest) WithCiPipelineId(id int) *PromotionMaterialRequest {
	p.ciPipelineId = id
	return p
}

func (p *PromotionMaterialRequest) GetCiPipelineId() int {
	return p.ciPipelineId
}

func (p *PromotionMaterialRequest) WithCdPipelineId(id int) *PromotionMaterialRequest {
	p.cdPipelineId = id
	return p
}

func (p *PromotionMaterialRequest) GetCdPipelineId() int {
	return p.cdPipelineId
}

func (p *PromotionMaterialRequest) WithExternalCiPipelineId(id int) *PromotionMaterialRequest {
	p.externalCiPipelineId = id
	return p
}

func (p *PromotionMaterialRequest) GetExtCiPipelineId() int {
	return p.externalCiPipelineId
}

func (p *PromotionMaterialRequest) WithListingOptions(listingOptions util.ListingFilterOptions) *PromotionMaterialRequest {
	p.ListingFilterOptions = listingOptions
	return p
}

func (p *PromotionMaterialRequest) GetListingOptions() util.ListingFilterOptions {
	return p.ListingFilterOptions
}

func (p *PromotionMaterialRequest) IsCINode() bool {
	return p.resource == string(constants.SOURCE_TYPE_CI) ||
		p.resource == string(constants.SOURCE_TYPE_JOB_CI) ||
		p.resource == string(constants.SOURCE_TYPE_LINKED_CD) ||
		p.resource == string(constants.SOURCE_TYPE_LINKED_CI)
}

func (p *PromotionMaterialRequest) IsCDNode() bool {
	return p.resource == string(constants.SOURCE_TYPE_CD)
}

func (p *PromotionMaterialRequest) IsWebhookNode() bool {
	return p.resource == string(constants.SOURCE_TYPE_WEBHOOK)
}

func (p *PromotionMaterialRequest) IsPromotionApprovalPendingNode() bool {
	return p.resource == string(constants.PROMOTION_APPROVAL_PENDING_NODE)
}

func (p *PromotionMaterialRequest) IsPendingForUserRequest() bool {
	return p.resource == string(constants.PROMOTION_APPROVAL_PENDING_NODE) && p.pendingForCurrentUser
}
