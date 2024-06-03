/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

import "github.com/devtron-labs/devtron/pkg/pipeline/types"

const GIT_MATERIAL_DELETE_SUCCESS_RESP = "Git material deleted successfully."

type BuildHistoryResponse struct {
	HideImageTaggingHardDelete bool                     `json:"hideImageTaggingHardDelete"`
	TagsEditable               bool                     `json:"tagsEditable"`
	AppReleaseTagNames         []string                 `json:"appReleaseTagNames"` // unique list of tags exists in the app
	CiWorkflows                []types.WorkflowResponse `json:"ciWorkflows"`
}

type ImageTagsQuery struct {
	AppNames []string `schema:"appName,required"`
}
