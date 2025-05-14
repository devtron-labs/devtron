/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package history

import (
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

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
	CdWorkflows                []bean.CdWorkflowWithArtifact `json:"cdWorkflows"`
	TagsEditable               bool                          `json:"tagsEditable"`
	AppReleaseTagNames         []string                      `json:"appReleaseTagNames"` // unique list of tags exists in the app
	HideImageTaggingHardDelete bool                          `json:"hideImageTaggingHardDelete"`
}
