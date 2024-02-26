package bean

import (
	"time"
)

type ArtifactPromotionRequestStatus = int

const (
	PROMOTED ArtifactPromotionRequestStatus = iota
	CANCELED
	AWAITING_APPROVAL
)

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

type EnvironmentResponse struct {
	Name                       string                   `json:"name"` // environment name
	ApprovalCount              int                      `json:"approvalCount"`
	PromotionPossible          bool                     `json:"promotionPossible"`
	PromotionEvaluationMessage string                   `json:"promotionEvaluationMessage""`
	PromotionEvaluationState   PromotionEvaluationState `json:"promotionEvaluationState"`
	IsVirtualEnvironment       bool                     `json:"isVirtualEnvironment"`
}

type PromotionEvaluationState string

const ARTIFACT_ALREADY_PROMOTED = "already promoted"
const ALREADY_REQUEST_RAISED = "promotion request already raised"
const ERRORED PromotionEvaluationState = "error occurred"
const EMPTY PromotionEvaluationState = ""
const PIPELINE_NOT_FOUND PromotionEvaluationState = "pipeline Not Found"
const POLICY_NOT_CONFIGURED PromotionEvaluationState = "policy not configured"
const NO_PERMISSION PromotionEvaluationState = "no permission"
const PROMOTION_SUCCESSFUL PromotionEvaluationState = "image promoted"
const SENT_FOR_APPROVAL PromotionEvaluationState = "sent for approval"
const SOURCE_AND_DESTINATION_PIPELINE_MISMATCH PromotionEvaluationState = "source and destination pipeline order mismatch"
