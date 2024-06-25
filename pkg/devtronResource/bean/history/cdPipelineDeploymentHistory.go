package history

import "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"

type CdPipelineDeploymentHistoryListReq struct {
	PipelineId int
	AppId      int
	EnvId      int
	Offset     int
	Limit      int
}

type CdPipelineDeploymentHistoryConfigListReq struct {
	BaseConfigurationId  int
	PipelineId           int
	HistoryComponent     string
	HistoryComponentName string
}

type DeploymentHistoryResp struct {
	CdWorkflows                []pipelineConfig.CdWorkflowWithArtifact `json:"cdWorkflows"`
	TagsEditable               bool                                    `json:"tagsEditable"`
	AppReleaseTagNames         []string                                `json:"appReleaseTagNames"` // unique list of tags exists in the app
	HideImageTaggingHardDelete bool                                    `json:"hideImageTaggingHardDelete"`
}
