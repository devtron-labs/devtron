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

type SourceType = int

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

func GetSourceType(sourceType int) string {
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
	SourceId           int      `json:"sourceId"`
	SourceType         string   `json:"sourceType"`
	Action             string   `json:"action"`
	PromotionRequestId int      `json:"promotionRequestId"`
	ArtifactId         int      `json:"artifactId"`
	AppName            string   `json:"appName"`
	EnvironmentNames   []string `json:"environmentNames"`
	UserId             int32    `json:"-"`
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

type PromotionPolicy struct {
	ApprovalCount                int  `json:"approvalCount"`
	AllowImageBuilderFromApprove bool `json:"AllowImageBuilderFromApprove"`
	AllowRequesterFromApprove    bool `json:"AllowRequesterFromApprove"`
	AllowApproverFromDeploy      bool `json:"AllowApproverFromDeploy"`
}
