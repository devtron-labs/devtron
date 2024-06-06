package history

import (
	pipelineBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

//keeping these beans at separate dir because of resourceFilter package incorrect structure, TODO: pick this up in upcoming iterations

type CdPipelineDeploymentHistoryListReq struct {
	PipelineId        int
	AppId             int
	EnvId             int
	FilterByReleaseId int
	Offset            int
	Limit             int
}

type CdPipelineDeploymentHistoryConfigListReq struct {
	BaseConfigurationId  int
	PipelineId           int
	HistoryComponent     string
	HistoryComponentName string
}

type DeploymentHistoryResp struct {
	CdWorkflows                []pipelineBean.CdWorkflowWithArtifact `json:"cdWorkflows"`
	TagsEditable               bool                                  `json:"tagsEditable"`
	AppReleaseTagNames         []string                              `json:"appReleaseTagNames"` // unique list of tags exists in the app
	HideImageTaggingHardDelete bool                                  `json:"hideImageTaggingHardDelete"`
}
