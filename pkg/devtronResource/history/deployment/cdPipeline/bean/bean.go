package bean

import (
	pipelineBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

type DeploymentHistoryResp struct {
	CdWorkflows                []pipelineBean.CdWorkflowWithArtifact `json:"cdWorkflows"`
	TagsEditable               bool                                  `json:"tagsEditable"`
	AppReleaseTagNames         []string                              `json:"appReleaseTagNames"` // unique list of tags exists in the app
	HideImageTaggingHardDelete bool                                  `json:"hideImageTaggingHardDelete"`
}
