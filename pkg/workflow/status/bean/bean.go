package bean

type ArgoPipelineStatusSyncEvent struct {
	PipelineId            int   `json:"pipelineId"`
	InstalledAppVersionId int   `json:"installedAppVersionId"`
	UserId                int32 `json:"userId"`
	IsAppStoreApplication bool  `json:"isAppStoreApplication"`
}
