/*
 * Copyright (c) 2024. Devtron Inc.
 */

package client

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/bean"
	"time"
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

type InterceptEventNotificationData struct {
	Heading                  string    `json:"heading"`
	Kind                     string    `json:"kind"`
	Name                     string    `json:"name"`
	Action                   string    `json:"action"`
	ClusterName              string    `json:"clusterName"`
	Namespace                string    `json:"namespace"`
	WatcherName              string    `json:"watcherName"`
	PipelineName             string    `json:"pipelineName"`
	ViewResourceManifestLink string    `json:"viewResourceManifestLink"`
	InterceptedAt            time.Time `json:"interceptedAt"`
	// Color
	// created : #E9FBF4
	// updated : #FFF5E5
	// deleted: #FDE7E7
	Color Color `json:"color"`
}

type Color string

const Green Color = "#E9FBF4"
const Orange Color = "#FFF5E5"
const Red Color = "#FDE7E7"

func NewInterceptEventNotificationData(kind, name, action, clusterName, namespace, watcherName, hostUrl, pipelineName string, interceptedAt time.Time, interceptEventId int) *InterceptEventNotificationData {
	color := Green
	if action == "updated" {
		color = Orange
	} else if action == "deleted" {
		color = Red
	}

	// heading := fmt.Sprintf("Change: Resource %s", action)
	return &InterceptEventNotificationData{
		// not setting heading here as this is causing some parsing error in curl request in plugin because of :(colon)
		// Heading:                  heading,
		Color:                    color,
		Kind:                     kind,
		Name:                     name,
		Namespace:                namespace,
		Action:                   action,
		ClusterName:              clusterName,
		WatcherName:              watcherName,
		PipelineName:             pipelineName,
		InterceptedAt:            interceptedAt,
		ViewResourceManifestLink: hostUrl + fmt.Sprintf("/dashboard/resource-watcher/intercepted-changes?id=%d", interceptEventId),
	}
}
