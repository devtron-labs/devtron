package bean

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository1 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/util"
	"log"
	"time"
)

type ArtifactPromotionRequestStatus int
type SearchableField string
type SortKey = string
type SortOrder = string

const (
	AWAITING_APPROVAL ArtifactPromotionRequestStatus = iota + 1
	CANCELED
	PROMOTED
	STALE
)

func (status ArtifactPromotionRequestStatus) Status() string {
	switch status {
	case PROMOTED:
		return "promoted"
	case CANCELED:
		return "cancelled"
	case STALE:
		return "stale"
	case AWAITING_APPROVAL:
		return "awaiting for approval"
	}
	return "deleted"
}

type SourceType int

const (
	CI SourceType = iota + 1
	WEBHOOK
	CD
)

type SourceTypeStr string

const (
	SOURCE_TYPE_CI                      SourceTypeStr = "CI"
	SOURCE_TYPE_WEBHOOK                 SourceTypeStr = "WEBHOOK"
	SOURCE_TYPE_CD                      SourceTypeStr = "ENVIRONMENT"
	ArtifactPromotionRequestNotFoundErr               = "artifact promotion request not found"
	ACTION_PROMOTE                                    = "PROMOTE"
	ACTION_CANCEL                                     = "CANCEL"
	ACTION_APPROVE                                    = "APPROVE"
	PROMOTION_APPROVAL_PENDING_NODE     SourceTypeStr = "PROMOTION_APPROVAL_PENDING_NODE"
)

func (sourceType SourceTypeStr) GetSourceType() SourceType {
	switch sourceType {
	case SOURCE_TYPE_CI:
		return CI
	case SOURCE_TYPE_WEBHOOK:
		return WEBHOOK
	case SOURCE_TYPE_CD:
		return CD
	}
	return CI
}

func (sourceType SourceType) GetSourceTypeStr() SourceTypeStr {
	switch sourceType {
	case CI:
		return SOURCE_TYPE_CI
	case WEBHOOK:
		return SOURCE_TYPE_WEBHOOK
	case CD:
		return SOURCE_TYPE_CD
	}
	return SOURCE_TYPE_CI
}

const (
	POLICY_NAME_SORT_KEY        SortKey         = "policyName"
	APPROVER_COUNT_SORT_KEY     SortKey         = "approverCount"
	ASC                         SortOrder       = "ASC"
	DESC                        SortOrder       = "DESC"
	APPROVER_COUNT_SEARCH_FIELD SearchableField = "approver_count"
)

type ArtifactPromotionRequest struct {
	SourceName         string                   `json:"sourceName"`
	SourceType         SourceTypeStr            `json:"sourceType"`
	Action             string                   `json:"action"`
	PromotionRequestId int                      `json:"promotionRequestId"`
	ArtifactId         int                      `json:"artifactId"`
	AppName            string                   `json:"appName"`
	EnvironmentNames   []string                 `json:"destinationObjectNames"`
	UserId             int32                    `json:"-"`
	WorkflowId         int                      `json:"workflowId"`
	AppId              int                      `json:"appId"`
	EnvNameIdMap       map[string]int           `json:"-"`
	EnvIdNameMap       map[int]string           `json:"-"`
	SourcePipelineId   int                      `json:"-"`
	SourceCdPipeline   *pipelineConfig.Pipeline `json:"-"`
}

type ArtifactPromotionApprovalResponse struct {
	Source          string        `json:"source"`
	SourceType      SourceTypeStr `json:"sourceType"`
	Destination     string        `json:"destination"`
	RequestedBy     string        `json:"requestedBy"`
	ApprovedUsers   []string      `json:"approvedUsers"`
	RequestedOn     time.Time     `json:"requestedOn"`
	PromotedOn      time.Time     `json:"promotedOn"`
	PromotionPolicy string        `json:"promotionPolicy"`
}

type PromotionApprovalMetaData struct {
	ApprovalRequestId    int                         `json:"approvalRequestId"`
	ApprovalRuntimeState string                      `json:"approvalRuntimeState"`
	ApprovalUsersData    []PromotionApprovalUserData `json:"approvalUsersData"`
	RequestedUserData    PromotionApprovalUserData   `json:"requestedUserData"`
	PromotedFrom         string                      `json:"promotedFrom"`
	PromotedFromType     string                      `json:"promotedFromType"`
	Policy               PromotionPolicy             `json:"policy" validate:"dive"`
}

func (promotionApprovalMetaData PromotionApprovalMetaData) GetApprovalUserIds() []int32 {
	approvalUserIds := make([]int32, len(promotionApprovalMetaData.ApprovalUsersData))
	for _, approvalUserData := range promotionApprovalMetaData.ApprovalUsersData {
		approvalUserIds = append(approvalUserIds, approvalUserData.UserId)
	}
	return approvalUserIds
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

// todo: change it to EnvironmentPromotionMetaData
type EnvironmentPromotionMetaData struct {
	Name                       string                   `json:"name"` // environment name
	ApprovalCount              int                      `json:"approvalCount,omitempty"`
	PromotionValidationMessage string                   `json:"promotionEvaluationMessage"`
	PromotionValidationState   PromotionValidationState `json:"promotionEvaluationState"`
	PromotionPossible          bool                     `json:"promotionPossible"`
	IsVirtualEnvironment       bool                     `json:"isVirtualEnvironment"`
}

type EnvironmentApprovalMetadata struct {
	Name            string   `json:"name"` // environment name
	ApprovalAllowed bool     `json:"approvalAllowed"`
	Reasons         []string `json:"reason"`
}

type PromotionPolicy struct {
	Id                 int                      `json:"id" `
	Name               string                   `json:"name" devtronSearchableField:"name" validate:"min=3,max=50,global-entity-name"`
	Description        string                   `json:"description" validate:"max=300"`
	PolicyEvaluationId int                      `json:"-"`
	Conditions         []util.ResourceCondition `json:"conditions" validate:"omitempty,min=1"`
	ApprovalMetaData   ApprovalMetaData         `json:"approvalMetadata" validate:"dive"`
	IdentifierCount    *int                     `json:"identifierCount,omitempty"`
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
		log.Printf("error in unmarshalling global policy json into promotionPolicy object, globalPolicy:%v,  err:%v", rawPolicy, err)
		return errors.New("unable to extract promotion policies")
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

type PromotionValidationState string

const ARTIFACT_ALREADY_PROMOTED PromotionValidationState = "already promoted"
const ALREADY_REQUEST_RAISED PromotionValidationState = "promotion request already raised"
const ERRORED PromotionValidationState = "error occurred"
const EMPTY PromotionValidationState = ""
const PIPELINE_NOT_FOUND PromotionValidationState = "pipeline Not Found"
const POLICY_NOT_CONFIGURED PromotionValidationState = "policy not configured"
const NO_PERMISSION PromotionValidationState = "no permission"
const PROMOTION_SUCCESSFUL PromotionValidationState = "image promoted"
const SENT_FOR_APPROVAL PromotionValidationState = "sent for approval"
const SOURCE_AND_DESTINATION_PIPELINE_MISMATCH PromotionValidationState = "source and destination pipeline order mismatch"
const POLICY_EVALUATION_ERRORED PromotionValidationState = "server unable to evaluate the policy"
const BLOCKED_BY_POLICY PromotionValidationState = "blocked by the policy "
const APPROVED PromotionValidationState = "approved"

type EnvironmentListingResponse struct {
	CiSource     CiSourceMetaData               `json:"ciSource"`
	Environments []EnvironmentPromotionMetaData `json:"environments"`
}

type CiSourceMetaData struct {
	Id   int           `json:"id"`
	Name string        `json:"name"`
	Type SourceTypeStr `json:"type"`
}

// rename to appworkflow metadata
type WorkflowMetaData struct {
	WorkflowId   int
	AppName      string
	AppId        int
	EnvMap       map[string]repository1.Environment
	CiSourceData CiSourceMetaData
}

type PipelinesMetaData struct {
	PipelineIds            []int
	PipelineIdVsEnvNameMap map[int]string
	PipelineIdDaoMap       map[int]*pipelineConfig.Pipeline
	EnvIds                 []int
}

type RequestMetaData struct {
	activeEnvIdNameMap map[int]string
	activeEnvNameIdMap map[string]int
	UserEnvNames       []string
	AuthorisedEnvMap   map[string]bool
	activeEnvironments []*cluster.EnvironmentBean
	PipelinesMetaData
	activeEnvIds             []int
	activeEnvNames           []string
	activeAuthorisedEnvNames []string
	activeAuthorisedEnvIds   []int
}

func (r *RequestMetaData) SetActiveEnvironments(activeEnvs []*cluster.EnvironmentBean) {
	r.activeEnvironments = activeEnvs
	activeEnvNames := make([]string, 0, len(r.activeEnvironments))
	authorisedEnvNames := make([]string, 0, len(r.AuthorisedEnvMap))
	activeAuthorisedEnvIds := make([]int, 0, len(r.AuthorisedEnvMap))
	activeEnvIds := make([]int, 0, len(r.activeEnvironments))
	activeEnvIdNameMap := make(map[int]string)
	activeEnvNameIdMap := make(map[string]int)
	for _, env := range r.activeEnvironments {
		activeEnvNames = append(activeEnvNames, env.Environment)
		activeEnvIds = append(activeEnvIds, env.Id)
		activeEnvIdNameMap[env.Id] = env.Environment
		activeEnvNameIdMap[env.Environment] = env.Id
		if r.AuthorisedEnvMap[env.Environment] {
			authorisedEnvNames = append(authorisedEnvNames, env.Environment)
			activeAuthorisedEnvIds = append(activeAuthorisedEnvIds, env.Id)
		}
	}
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
	return r.PipelinesMetaData.PipelineIdDaoMap[id]
}

// todo: first extract metadata
// db env names
// db pipelines
// user given envnames
// authorised envs names map
// authorised envs
// authorised envIds
// appId
// workflowId
// PipelinesMetaData
