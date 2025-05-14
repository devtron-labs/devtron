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

package bean

import "github.com/devtron-labs/devtron/pkg/pipeline/types"

const GIT_MATERIAL_DELETE_SUCCESS_RESP = "Git material deleted successfully."

type BuildHistoryResponse struct {
	HideImageTaggingHardDelete bool                     `json:"hideImageTaggingHardDelete"`
	TagsEditable               bool                     `json:"tagsEditable"`
	AppReleaseTagNames         []string                 `json:"appReleaseTagNames"` //unique list of tags exists in the app
	CiWorkflows                []types.WorkflowResponse `json:"ciWorkflows"`
}
