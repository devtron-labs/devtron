package bean

import (
	"encoding/json"
	"errors"
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
	AWAITING_APPROVAL ArtifactPromotionRequestStatus = iota
	CANCELED
	PROMOTED
	STALE = 3
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
	CI SourceType = iota
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
	PROMOTION_APPROVAL_PENDING_NODE                   = "PROMOTION_APPROVAL_PENDING_NODE"
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
	APPROVER_COUNT_SEARCH_FIELD SearchableField = "approverCount"
)

type ArtifactPromotionRequest struct {
	SourceName         string         `json:"sourceName"`
	SourceType         SourceTypeStr  `json:"sourceType"`
	Action             string         `json:"action"`
	PromotionRequestId int            `json:"promotionRequestId"`
	ArtifactId         int            `json:"artifactId"`
	AppName            string         `json:"appName"`
	EnvironmentNames   []string       `json:"environmentNames"`
	UserId             int32          `json:"-"`
	WorkflowId         int            `json:"workflowId"`
	AppId              int            `json:"appId"`
	EnvNameIdMap       map[string]int `json:"-"`
	EnvIdNameMap       map[int]string `json:"-"`
	SourcePipelineId   int            `json:"-"`
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

type EnvironmentResponse struct {
	Name                       string                   `json:"name"` // environment name
	ApprovalCount              int                      `json:"approvalCount,omitempty"`
	PromotionPossible          *bool                    `json:"promotionPossible"`
	PromotionValidationMessage string                   `json:"promotionEvaluationMessage"`
	PromotionValidationState   PromotionValidationState `json:"promotionEvaluationState"`
	IsVirtualEnvironment       *bool                    `json:"isVirtualEnvironment,omitempty"`
}

type EnvironmentApprovalMetadata struct {
	Name            string   `json:"name"` // environment name
	ApprovalAllowed bool     `json:"approvalAllowed"`
	Reasons         []string `json:"reason"`
}

type PromotionPolicy struct {
	Id                 int                      `json:"id" `
	Name               string                   `json:"name" isSearchField:"true" validate:"min=3,max=50,global-entity-name"`
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
	return &bean.GlobalPolicyDataModel{
		GlobalPolicyBaseModel: *baseModel,
		SearchableFields:      util.GetSearchableFields(policy),
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
	ApprovalCount                int  `json:"approverCount" isSearchField:"true" validate:"min=0"`
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
	CiSource     CiSourceMetaData      `json:"ciSource"`
	Environments []EnvironmentResponse `json:"environments"`
}

type CiSourceMetaData struct {
	Id   int           `json:"id"`
	Name string        `json:"name"`
	Type SourceTypeStr `json:"type"`
}

type WorkflowMetaData struct {
	WorkflowId   int
	AppName      string
	AppId        int
	EnvMap       map[string]repository1.Environment
	CiSourceData CiSourceMetaData
}
