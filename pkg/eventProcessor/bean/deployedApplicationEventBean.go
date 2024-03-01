package bean

import (
	v1alpha12 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"time"
)

type ApplicationDetail struct {
	Application *v1alpha12.Application `json:"application"`
	StatusTime  time.Time              `json:"statusTime"`
}

type ArgoPipelineStatusSyncEvent struct {
	PipelineId            int   `json:"pipelineId"`
	InstalledAppVersionId int   `json:"installedAppVersionId"`
	UserId                int32 `json:"userId"`
	IsAppStoreApplication bool  `json:"isAppStoreApplication"`
}
