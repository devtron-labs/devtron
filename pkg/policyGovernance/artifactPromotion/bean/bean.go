package bean

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository1 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	bean4 "github.com/devtron-labs/devtron/pkg/pipeline/constants"
	bean2 "github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/constants"
	"github.com/devtron-labs/devtron/util"
	errors2 "github.com/pkg/errors"
	"time"
)

type ArtifactPromotionRequest struct {
	SourceName         string              `json:"sourceName"`
	SourceType         bean2.SourceTypeStr `json:"sourceType"`
	Action             bean2.RequestAction `json:"action"`
	PromotionRequestId int                 `json:"promotionRequestId"`
	ArtifactId         int                 `json:"artifactId"`
	AppName            string              `json:"appName"`
	EnvironmentNames   []string            `json:"destinationObjectNames"`
	WorkflowId         int                 `json:"workflowId"`
	AppId              int                 `json:"appId"`
}

func GetSourceTypeFromPipelineType(pipelineType bean4.PipelineType) bean2.SourceTypeStr {
	switch pipelineType {
	case bean4.NORMAL:
		return bean2.SOURCE_TYPE_CI
	case bean4.LINKED:
		return bean2.SOURCE_TYPE_LINKED_CI
	case bean4.EXTERNAL:
		return bean2.SOURCE_TYPE_WEBHOOK
	case bean4.CI_JOB:
		return bean2.SOURCE_TYPE_JOB_CI
	case bean4.LINKED_CD:
		return bean2.SOURCE_TYPE_LINKED_CD
	}
	return bean2.SOURCE_TYPE_CI
}
func (r *ArtifactPromotionRequest) ValidateRequest() error {
	switch r.Action {
	case bean2.ACTION_APPROVE:
		if r.AppId <= 0 {
			return errors.New("appId is required")
		}
		if r.ArtifactId <= 0 {
			return errors.New("artifactId is required")
		}
		if len(r.EnvironmentNames) == 0 {
			return errors.New("destinationObjectNames are required")
		}
	case bean2.ACTION_PROMOTE:
		if r.SourceType != bean2.SOURCE_TYPE_CI &&
			r.SourceType != bean2.SOURCE_TYPE_WEBHOOK &&
			r.SourceType != bean2.SOURCE_TYPE_CD &&
			r.SourceType != bean2.PROMOTION_APPROVAL_PENDING_NODE &&
			r.SourceType != bean2.SOURCE_TYPE_LINKED_CI &&
			r.SourceType != bean2.SOURCE_TYPE_LINKED_CD &&
			r.SourceType != bean2.SOURCE_TYPE_JOB_CI {
			return errors.New("invalid sourceType")
		}
		if r.AppId <= 0 {
			return errors.New("appId is required")
		}
		if len(r.EnvironmentNames) == 0 {
			return errors.New("destinationObjectNames are required")
		}
	case bean2.ACTION_CANCEL:
		if r.PromotionRequestId <= 0 {
			return errors.New("promotionRequestId is required")
		}
	}
	return nil
}

type ArtifactPromotionApprovalResponse struct {
	Id                      int
	PolicyId                int
	PolicyEvaluationAuditId int
	ArtifactId              int
	SourceType              bean2.SourceType
	SourcePipelineId        int
	DestinationPipelineId   int
	Status                  bean2.ArtifactPromotionRequestStatus
}

type PromotionApprovalMetaData struct {
	ApprovalRequestId    int                                  `json:"approvalRequestId"`
	ApprovalRuntimeState bean2.ArtifactPromotionRequestStatus `json:"approvalRuntimeState"`
	ApprovalUsersData    []PromotionApprovalUserData          `json:"approvedUsersData"`
	RequestedUserData    PromotionApprovalUserData            `json:"requestedUserData"`
	PromotedFrom         string                               `json:"promotedFrom"`
	PromotedFromType     string                               `json:"promotedFromType"`
	Policy               PromotionPolicy                      `json:"policy" validate:"dive"`
}

func (p PromotionApprovalMetaData) GetApprovalUserIds() []int32 {
	approvalUserIds := make([]int32, len(p.ApprovalUsersData))
	for _, approvalUserData := range p.ApprovalUsersData {
		approvalUserIds = append(approvalUserIds, approvalUserData.UserId)
	}
	return approvalUserIds
}

func (p PromotionApprovalMetaData) IsUserApprover(userId int32) bool {
	approverUserIds := make([]int32, len(p.ApprovalUsersData))
	for _, approvalUserData := range p.ApprovalUsersData {
		approverUserIds = append(approverUserIds, approvalUserData.UserId)
	}
	for _, id := range approverUserIds {
		if id == userId {
			return true
		}
	}
	return false
}

func (promotionApprovalMetaData PromotionApprovalMetaData) GetRequestedUserId() int32 {
	return promotionApprovalMetaData.RequestedUserData.UserId
}

type PromotionPolicyMetaRequest struct {
	Search    string `json:"search"`
	SortBy    string `json:"sortBy" validate:"oneof=ASC DESC"`
	SortOrder string `json:"sortOrder" validate:"oneof=policyName approverCount"`
}

type PromotionApprovalUserData struct {
	UserId         int32     `json:"userId"`
	UserEmail      string    `json:"userEmail"`
	UserActionTime time.Time `json:"userActionTime"`
}

type EnvironmentPromotionMetaData struct {
	Name                 string `json:"name"` // environment name
	IsVirtualEnvironment bool   `json:"isVirtualEnvironment"`

	ApprovalCount              int                            `json:"approvalCount,omitempty"`
	PromotionValidationState   bean2.PromotionValidationState `json:"promotionValidationState"`
	PromotionValidationMessage bean2.PromotionValidationMsg   `json:"promotionValidationMessage"`
	PromotionPossible          bool                           `json:"promotionPossible"`
}

func (o EnvironmentPromotionMetaData) NoValidationIssue() bool {
	return o.PromotionValidationMessage == bean2.EMPTY
}

type EnvironmentApprovalMetadata struct {
	Name            string   `json:"name"` // environment name
	ApprovalAllowed bool     `json:"approvalAllowed"`
	Reasons         []string `json:"reason"`
}

type PromotionPolicy struct {
	Id               int                      `json:"id" `
	Name             string                   `json:"name" devtronSearchableField:"name" validate:"min=3,max=50,global-entity-name"`
	Description      string                   `json:"description" validate:"max=300"`
	Conditions       []util.ResourceCondition `json:"conditions" validate:"min=1,dive"`
	ApprovalMetaData ApprovalMetaData         `json:"approvalMetadata" validate:"dive"`
	IdentifierCount  *int                     `json:"identifierCount,omitempty"`
}

func (p *PromotionPolicy) CanBePromoted(approvalsGot int) bool {
	return approvalsGot >= p.ApprovalMetaData.ApprovalCount
}
func (p *PromotionPolicy) CanImageBuilderApprove(imageBuiltByUserId, approvingUserId int32) bool {
	return !p.ApprovalMetaData.AllowImageBuilderFromApprove && imageBuiltByUserId == approvingUserId
}

func (p *PromotionPolicy) CanPromoteRequesterApprove(requestedUserId, approvingUserId int32) bool {
	return !p.ApprovalMetaData.AllowRequesterFromApprove && requestedUserId == approvingUserId
}

func (p *PromotionPolicy) BlockApproverFromDeploy() bool {
	return !p.ApprovalMetaData.AllowApproverFromDeploy
}

func (p *PromotionPolicy) CanApprove(requestedUserId, imageBuiltByUserId, approvingUserId int32) bool {
	return p.CanImageBuilderApprove(imageBuiltByUserId, approvingUserId) && p.CanPromoteRequesterApprove(requestedUserId, approvingUserId)
}

func (policy *PromotionPolicy) ConvertToGlobalPolicyBaseModal(userId int32) (*bean.GlobalPolicyBaseModel, error) {
	jsonPolicyBytes, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}
	return &bean.GlobalPolicyBaseModel{
		PolicyOf:      bean.GLOBAL_POLICY_TYPE_IMAGE_PROMOTION_POLICY,
		Name:          policy.Name,
		Description:   policy.Description,
		Enabled:       true, // all the policies are by default enabled
		PolicyVersion: bean.GLOBAL_POLICY_VERSION_V1,
		Active:        true,
		UserId:        userId,
		JsonData:      string(jsonPolicyBytes),
	}, nil
}

func (policy *PromotionPolicy) ConvertToGlobalPolicyDataModel(userId int32) (*bean.GlobalPolicyDataModel, error) {
	baseModel, err := policy.ConvertToGlobalPolicyBaseModal(userId)
	if err != nil {
		return nil, err
	}
	searchKeys := util.GetSearchableFields(*policy)
	approvalSearchKeys := util.GetSearchableFields(policy.ApprovalMetaData)
	searchKeys = append(searchKeys, approvalSearchKeys...)
	return &bean.GlobalPolicyDataModel{
		GlobalPolicyBaseModel: *baseModel,
		SearchableFields:      searchKeys,
	}, nil
}

func (policy *PromotionPolicy) UpdateWithGlobalPolicy(rawPolicy *bean.GlobalPolicyBaseModel) error {
	err := json.Unmarshal([]byte(rawPolicy.JsonData), policy)
	if err != nil {
		return errors2.Wrap(err, "unable to extract promotion policies")
	}
	policy.Name = rawPolicy.Name
	policy.Id = rawPolicy.Id
	policy.Description = rawPolicy.Description
	return nil
}

type ApprovalMetaData struct {
	ApprovalCount                int  `json:"approverCount" devtronSearchableField:"approver_count" validate:"min=0"`
	AllowImageBuilderFromApprove bool `json:"allowImageBuilderFromApprove"`
	AllowRequesterFromApprove    bool `json:"allowRequesterFromApprove"`
	AllowApproverFromDeploy      bool `json:"allowApproverFromDeploy"`
}

type EnvironmentListingResponse struct {
	CiSource     CiSourceMetaData               `json:"ciSource"`
	Environments []EnvironmentPromotionMetaData `json:"environments"`
}

type CiSourceMetaData struct {
	Id   int                 `json:"id"`
	Name string              `json:"name"`
	Type bean2.SourceTypeStr `json:"type"`
}

// rename to appworkflow metadata
type WorkflowMetaData struct {
	WorkflowId   int
	AppName      string
	AppId        int
	EnvMap       map[string]repository1.Environment
	CiSourceData CiSourceMetaData
}

type pipelinesMetaData struct {
	activeAuthorisedPipelineIds            []int
	activeAuthorisedPipelineIdVsEnvNameMap map[int]string
	activeAuthorisedEnvNameVsPipelineIdMap map[string]int
	activeAuthorisedPipelineIdDaoMap       map[int]*pipelineConfig.Pipeline
	pipelineEnvIds                         []int
	promotableEnvs                         []string
	promotablePipelineIds                  []int
}

type SourceMetaData struct {
	id               int
	typeStr          bean2.SourceTypeStr
	name             string
	sourceWorkflowId int
	cdPipeline       *pipelineConfig.Pipeline
}

func (s *SourceMetaData) WithSourceWorkflowId(sourceWorkflowId int) *SourceMetaData {
	s.sourceWorkflowId = sourceWorkflowId
	return s
}

func (s *SourceMetaData) WithId(id int) *SourceMetaData {
	s.id = id
	return s
}

func (s *SourceMetaData) WithType(typeStr bean2.SourceTypeStr) *SourceMetaData {
	s.typeStr = typeStr
	return s
}
func (s *SourceMetaData) WithName(name string) *SourceMetaData {
	s.name = name
	return s
}
func (s *SourceMetaData) WithCdPipeline(cdPipeline *pipelineConfig.Pipeline) *SourceMetaData {
	s.cdPipeline = cdPipeline
	return s
}

func (s *SourceMetaData) GetCiSourceMeta() CiSourceMetaData {
	return CiSourceMetaData{
		Id:   s.id,
		Type: s.typeStr,
		Name: s.name,
	}
}

type RequestMetaData struct {
	activeEnvIdNameMap          map[int]string
	activeEnvNameIdMap          map[string]int
	userEnvNames                []string
	authorisedEnvMap            map[string]bool
	activeEnvironments          []*cluster.EnvironmentBean
	activeEnvironmentsMap       map[string]*cluster.EnvironmentBean
	destinationPipelineMetaData *pipelinesMetaData
	activeEnvIds                []int
	activeEnvNames              []string
	activeAuthorisedEnvNames    []string
	activeAuthorisedEnvIds      []int
	sourceMetaData              *SourceMetaData
	appId                       int
	workflowId                  int
	imageTags                   []string
	imageComment                string
	ciArtifactId                int
	ciArtifact                  *repository.CiArtifact
}

func (r *RequestMetaData) GetDefaultEnvironmentPromotionMetaDataResponseMap() map[string]EnvironmentPromotionMetaData {
	response := make(map[string]EnvironmentPromotionMetaData)
	for _, env := range r.GetUserGivenEnvNames() {
		envResponse := EnvironmentPromotionMetaData{
			Name:                       env,
			PromotionValidationMessage: bean2.PIPELINE_NOT_FOUND,
		}
		if !r.GetAuthorisedEnvMap()[env] {
			envResponse.PromotionValidationMessage = bean2.NO_PERMISSION
		}
		response[env] = envResponse
	}
	for _, pipelineId := range r.GetActiveAuthorisedPipelineIds() {
		envName := r.GetActiveAuthorisedPipelineIdEnvMap()[pipelineId]
		resp := response[envName]
		resp.PromotionValidationMessage = bean2.EMPTY
		response[envName] = resp
	}
	return response
}

func (r *RequestMetaData) GetSourceMetaData() *SourceMetaData {
	return r.sourceMetaData
}

func (r *RequestMetaData) WithCiArtifact(ciArtifact *repository.CiArtifact) *RequestMetaData {
	r.ciArtifact = ciArtifact
	r.ciArtifactId = ciArtifact.Id
	return r
}

func (r *RequestMetaData) WithAppId(appId int) *RequestMetaData {
	r.appId = appId
	return r
}

func (r *RequestMetaData) WithWorkflowId(workflowId int) *RequestMetaData {
	r.workflowId = workflowId
	return r
}

func (r *RequestMetaData) WithPromotableEnvs(promotableEnvs []string) *RequestMetaData {
	r.destinationPipelineMetaData.promotableEnvs = promotableEnvs
	promotablePipelineIds := make([]int, 0, len(promotableEnvs))
	for _, promotableEnv := range promotableEnvs {
		promotablePipelineIds = append(promotablePipelineIds, r.destinationPipelineMetaData.activeAuthorisedEnvNameVsPipelineIdMap[promotableEnv])
	}
	r.destinationPipelineMetaData.promotablePipelineIds = promotablePipelineIds
	return r
}

func (r *RequestMetaData) SetSourceMetaData(sourceMetaData *SourceMetaData) {
	r.sourceMetaData = sourceMetaData
}

func (r *RequestMetaData) WithImageTags(imageTags []string) *RequestMetaData {
	r.imageTags = imageTags
	return r
}

func (r *RequestMetaData) WithImageComment(imageComment string) *RequestMetaData {
	r.imageComment = imageComment
	return r
}

func (r *RequestMetaData) GetImageTags() []string {
	return r.imageTags
}

func (r *RequestMetaData) GetImageComment() string {
	return r.imageComment
}

func (r *RequestMetaData) SetDestinationPipelineMetaData(activeAuthorisedPipelines []*pipelineConfig.Pipeline) {
	pipelineIds := make([]int, 0, len(activeAuthorisedPipelines))
	pipelineIdEnvNameMap := make(map[int]string)
	pipelineIdPipelineDaoMap := make(map[int]*pipelineConfig.Pipeline)
	pipelineEnvIds := make([]int, 0, len(activeAuthorisedPipelines))
	activeAuthorisedEnvNameVsPipelineIdMap := make(map[string]int)
	activeAuthorisedEnvIds := make([]int, 0, len(activeAuthorisedPipelines))
	for _, pipeline := range activeAuthorisedPipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
		activeAuthorisedEnvIds = append(activeAuthorisedEnvIds, pipeline.EnvironmentId)
		pipelineIdEnvNameMap[pipeline.Id] = pipeline.Environment.Name
		activeAuthorisedEnvNameVsPipelineIdMap[pipeline.Environment.Name] = pipeline.Id
		pipelineIdPipelineDaoMap[pipeline.Id] = pipeline
		pipelineEnvIds = append(pipelineEnvIds, pipeline.EnvironmentId)
	}

	pipelineMetaData := &pipelinesMetaData{
		pipelineEnvIds:                         activeAuthorisedEnvIds,
		activeAuthorisedPipelineIds:            pipelineIds,
		activeAuthorisedPipelineIdDaoMap:       pipelineIdPipelineDaoMap,
		activeAuthorisedPipelineIdVsEnvNameMap: pipelineIdEnvNameMap,
		activeAuthorisedEnvNameVsPipelineIdMap: activeAuthorisedEnvNameVsPipelineIdMap,
	}
	r.destinationPipelineMetaData = pipelineMetaData
}

func (r *RequestMetaData) SetActiveEnvironments(userGivenEnvNames []string, authorizedEnvironmentsMap map[string]bool, activeEnvs []*cluster.EnvironmentBean) {
	r.userEnvNames = userGivenEnvNames
	r.authorisedEnvMap = authorizedEnvironmentsMap
	r.activeEnvironments = activeEnvs
	activeEnvironmentsMap := make(map[string]*cluster.EnvironmentBean)
	activeEnvNames := make([]string, 0, len(r.activeEnvironments))
	authorisedEnvNames := make([]string, 0, len(r.authorisedEnvMap))
	activeAuthorisedEnvIds := make([]int, 0, len(r.authorisedEnvMap))
	activeEnvIds := make([]int, 0, len(r.activeEnvironments))
	activeEnvIdNameMap := make(map[int]string)
	activeEnvNameIdMap := make(map[string]int)
	for _, env := range r.activeEnvironments {
		activeEnvNames = append(activeEnvNames, env.Environment)
		activeEnvIds = append(activeEnvIds, env.Id)
		activeEnvironmentsMap[env.Environment] = env
		activeEnvIdNameMap[env.Id] = env.Environment
		activeEnvNameIdMap[env.Environment] = env.Id
		if r.authorisedEnvMap[env.Environment] {
			authorisedEnvNames = append(authorisedEnvNames, env.Environment)
			activeAuthorisedEnvIds = append(activeAuthorisedEnvIds, env.Id)
		}
	}

	r.activeEnvironmentsMap = activeEnvironmentsMap
	r.activeEnvNames = activeEnvNames
	r.activeAuthorisedEnvNames = authorisedEnvNames
	r.activeAuthorisedEnvIds = activeAuthorisedEnvIds
	r.activeEnvIds = activeAuthorisedEnvIds
	r.activeEnvIdNameMap = activeEnvIdNameMap
	r.activeEnvNameIdMap = activeEnvNameIdMap
	r.activeEnvNames = activeEnvNames
}

func (r *RequestMetaData) GetActiveEnvNames() []string {
	return r.activeEnvNames
}

func (r *RequestMetaData) GetActiveAuthorisedEnvNames() []string {
	return r.activeAuthorisedEnvNames
}

func (r *RequestMetaData) GetActiveAuthorisedEnvIds() []int {
	return r.activeAuthorisedEnvIds
}

func (r *RequestMetaData) GetPipelineById(id int) *pipelineConfig.Pipeline {
	return r.destinationPipelineMetaData.activeAuthorisedPipelineIdDaoMap[id]
}

func (r *RequestMetaData) GetWorkflowId() int {
	return r.sourceMetaData.sourceWorkflowId
}

func (r *RequestMetaData) GetSourceTypeStr() bean2.SourceTypeStr {
	return r.sourceMetaData.typeStr
}

func (r *RequestMetaData) GetSourceDisplayName() string {
	// for source type cd env name is displayed and for other sources source type is displayed
	if r.GetSourceTypeStr() == bean2.SOURCE_TYPE_CD {
		return r.sourceMetaData.name
	}
	return string(r.GetSourceTypeStr())
}

func (r *RequestMetaData) GetSourcePipelineId() int {
	return r.sourceMetaData.id
}

func (r *RequestMetaData) GetSourceName() string {
	return r.sourceMetaData.name
}

func (r *RequestMetaData) GetSourceCdPipeline() *pipelineConfig.Pipeline {
	pipeline := *r.sourceMetaData.cdPipeline
	return &pipeline
}

func (r *RequestMetaData) GetActiveAuthorisedPipelineIds() []int {
	return r.destinationPipelineMetaData.activeAuthorisedPipelineIds
}

func (r *RequestMetaData) GetActiveAuthorisedPipelineIdEnvMap() map[int]string {
	return r.destinationPipelineMetaData.activeAuthorisedPipelineIdVsEnvNameMap
}

func (r *RequestMetaData) GetActiveAuthorisedPipelineDaoMap() map[int]*pipelineConfig.Pipeline {
	return r.destinationPipelineMetaData.activeAuthorisedPipelineIdDaoMap
}

func (r *RequestMetaData) GetActiveAuthorisedPipelineEnvIds() []int {
	return r.destinationPipelineMetaData.pipelineEnvIds
}

func (r *RequestMetaData) GetUserGivenEnvNames() []string {
	return r.userEnvNames
}

func (r *RequestMetaData) GetAuthorisedEnvMap() map[string]bool {
	return r.authorisedEnvMap
}

func (r *RequestMetaData) GetCiArtifact() *repository.CiArtifact {
	artifact := *r.ciArtifact
	return &artifact
}

func (r *RequestMetaData) GetCiArtifactId() int {
	return r.ciArtifactId
}

func (r *RequestMetaData) GetActiveEnvironmentsMap() map[string]*cluster.EnvironmentBean {
	return r.activeEnvironmentsMap
}

func (r *RequestMetaData) GetAppId() int {
	return r.appId
}

func (r *RequestMetaData) GetPromotablePipelineIds() []int {
	return r.destinationPipelineMetaData.promotablePipelineIds
}
