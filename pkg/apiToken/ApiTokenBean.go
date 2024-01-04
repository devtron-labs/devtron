package apiToken

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/golang-jwt/jwt/v4"
)

type ApiTokenCustomClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}
type TokenCustomClaimsForNotification struct {
	DraftId           int                         `json:"draftId"`
	DraftVersionId    int                         `json:"draftVersionId"`
	ApprovalRequestId int                         `json:"approvalRequestId"`
	ArtifactId        int                         `json:"artifactId"`
	PipelineId        int                         `json:"pipelineId"`
	ActionType        bean.UserApprovalActionType `json:"actionType" validate:"required"`
	AppId             int                         `json:"appId" validate:"required"`
	EnvId             int                         `json:"envId"`
	ApprovalType      string                      `json:"approvalType"`
	ApiTokenCustomClaims
}
type DraftApprovalRequest struct {
	DraftId        int `json:"draftId"`
	DraftVersionId int `json:"draftVersionId"`
	NotificationApprovalRequest
}

type DeploymentApprovalRequest struct {
	ApprovalRequestId int `json:"approvalRequestId"`
	ArtifactId        int `json:"artifactId"`
	PipelineId        int `json:"pipelineId"`
	NotificationApprovalRequest
}

type NotificationApprovalRequest struct {
	AppId int `json:"appId" validate:"required"`
	EnvId int `json:"envId"`
	//ApprovalType string `json:"approvalType"`
	EmailId string `json:"emailId"`
	UserId  int32  `json:"userId"`
}

func (draftReq *DraftApprovalRequest) CreateDraftApprovalRequest(jsonStr []byte) error {
	err := json.Unmarshal(jsonStr, draftReq)
	return err
}

func (depReq *DeploymentApprovalRequest) CreateDeploymentApprovalRequest(jsonStr []byte) error {
	err := json.Unmarshal(jsonStr, depReq)
	return err
}
