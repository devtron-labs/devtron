package client

import (
	"github.com/devtron-labs/devtron/pkg/bean"
)

type NotificationMetaData struct {
	AppName    string `json:"appName"`
	EnvName    string `json:"envName"`
	ApprovedBy string `json:"approvedBy"`
	EventTime  string `json:"eventTime"`
}

type ImageMetadata struct {
	ImageTagNames  []string `json:"imageTagNames"`
	ImageComment   string   `json:"imageComment"`
	DockerImageUrl string   `json:"dockerImageUrl"`
}

type DraftApprovalResponse struct {
	ProtectConfigFileType string                `json:"protectConfigFileType"`
	ProtectConfigFileName string                `json:"protectConfigFileName"`
	ProtectConfigComment  string                `json:"protectConfigComment"`
	NotificationMetaData  *NotificationMetaData `json:"notificationMetaData"`
	DraftState            uint8                 `json:"draftState"`
}
type DeploymentApprovalResponse struct {
	NotificationMetaData *NotificationMetaData `json:"notificationMetaData"`
	Status               bean.ApprovalState    `json:"status"`
	ImageMetadata
}

type PromotionApprovalResponse struct {
	SourceInfo           string                `json:"sourceInfo"`
	NotificationMetaData *NotificationMetaData `json:"notificationMetaData"`
	Status               bean.ApprovalState    `json:"status"`
	ImageMetadata
}
