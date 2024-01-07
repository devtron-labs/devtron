package client

import (
	"github.com/devtron-labs/devtron/pkg/bean"
)

const TOKEN = "token"

type NotificationMetaData struct {
	AppName    string `json:"appName"`
	EnvName    string `json:"envName"`
	ApprovedBy string `json:"approvedBy"`
	EventTime  string `json:"eventTime"`
}

type DraftApprovalResponse struct {
	ProtectConfigFileType string               `json:"protectConfigFileType"`
	ProtectConfigFileName string               `json:"protectConfigFileName"`
	ProtectConfigComment  string               `json:"protectConfigComment"`
	NotificationMetaData  NotificationMetaData `json:"notificationMetaData"`
	DraftState            uint8                `json:"draftState"`
}
type DeploymentApprovalResponse struct {
	ImageTagNames        []string             `json:"imageTagNames"`
	ImageComment         string               `json:"imageComment"`
	DockerImageUrl       string               `json:"dockerImageUrl"`
	NotificationMetaData NotificationMetaData `json:"notificationMetaData"`
	Status               bean.ApprovalState   `json:"status"`
}
