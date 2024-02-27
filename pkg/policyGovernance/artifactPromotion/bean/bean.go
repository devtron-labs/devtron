package bean

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"time"
)

type ArtifactPromotionRequestStatus int

const (
	PROMOTED ArtifactPromotionRequestStatus = iota
	CANCELED
	AWAITING_APPROVAL
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

const (
	SOURCE_TYPE_CI                      string = "CI"
	SOURCE_TYPE_WEBHOOK                 string = "WEBHOOK"
	SOURCE_TYPE_CD                      string = "CD"
	ArtifactPromotionRequestNotFoundErr        = "artifact promotion request not found"
	ACTION_PROMOTE                             = "PROMOTE"
	ACTION_CANCEL                              = "CANCEL"
	ACTION_APPROVE                             = "APPROVE"
	PROMOTION_APPROVAL_PENDING_NODE            = "PROMOTION_APPROVAL_PENDING_NODE"
)

func (sourceType SourceType) GetSourceType() string {
	switch sourceType {
	case CI:
		return SOURCE_TYPE_CI
	case WEBHOOK:
		return SOURCE_TYPE_WEBHOOK
	case CD:
		return SOURCE_TYPE_CD
	}
	return ""
}

type ArtifactPromotionRequest struct {
	SourceName         string         `json:"sourceName"`
	SourceType         string         `json:"sourceType"`
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
	Source          string    `json:"source"`
	SourceType      string    `json:"sourceType"`
	Destination     string    `json:"destination"`
	RequestedBy     string    `json:"requestedBy"`
	ApprovedUsers   []string  `json:"approvedUsers"`
	RequestedOn     time.Time `json:"requestedOn"`
	PromotedOn      time.Time `json:"promotedOn"`
	PromotionPolicy string    `json:"promotionPolicy"`
}

type PromotionApprovalMetaData struct {
	ApprovalRequestId    int                         `json:"approvalRequestId"`
	ApprovalRuntimeState string                      `json:"approvalRuntimeState"`
	ApprovalUsersData    []PromotionApprovalUserData `json:"approvalUsersData"`
	RequestedUserData    PromotionApprovalUserData   `json:"requestedUserData"`
	PromotedFrom         string                      `json:"promotedFrom"`
	PromotedFromType     string                      `json:"promotedFromType"`
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
	Id                 int                                `json:"-"`
	Name               string                             `json:"-"`
	PolicyEvaluationId int                                `json:"-"`
	Conditions         []resourceFilter.ResourceCondition `json:"conditions"`
	ApprovalMetaData   ApprovalMetaData                   `json:"approvalMetadata"`
}

type ApprovalMetaData struct {
	ApprovalCount                int  `json:"approverCount"`
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
