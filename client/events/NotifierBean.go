package client

import (
	"github.com/devtron-labs/devtron/pkg/bean"
)

type NotificationApprovalResponse struct {
	AppName    string `json:"appName"`
	EnvName    string `json:"envName"`
	ApprovedBy string `json:"approvedBy"`
	EventTime  string `json:"eventTime"`
}

type DraftApprovalResponse struct {
	ProtectConfigFileType        string                       `json:"protectConfigFileType"`
	ProtectConfigFileName        string                       `json:"protectConfigFileName"`
	ProtectConfigComment         string                       `json:"protectConfigComment"`
	ProtectConfigLink            string                       `json:"protectConfigLink"`
	NotificationApprovalResponse NotificationApprovalResponse `json:"notificationApprovalResponse"`
	DraftState                   uint8                        `json:"draftState"`
}
type DeploymentApprovalResponse struct {
	ImageTagNames                []string                     `json:"imageTagNames"`
	ImageComment                 string                       `json:"imageComment"`
	ImageApprovalLink            string                       `json:"imageApprovalLink"`
	DockerImageTag               string                       `json:"dockerImageUrl"`
	NotificationApprovalResponse NotificationApprovalResponse `json:"notificationApprovalResponse"`
	Status                       bean.ApprovalState           `json:"status"`
}
