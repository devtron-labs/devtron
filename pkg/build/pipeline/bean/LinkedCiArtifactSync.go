/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bean

type LinkedCiArtifactSyncResponse struct {
	CiPipelineId       int    `json:"ciPipelineId"`
	ParentCiPipelineId int    `json:"parentCiPipelineId"`
	AppId              int    `json:"appId"`
	AppName            string `json:"appName"`
	CopiedCount        int    `json:"copiedCount"`
	Error              string `json:"error,omitempty"`
}

type LinkedCiArtifactBulkSyncResponse struct {
	SourceCiPipelineId int                             `json:"sourceCiPipelineId"`
	Results            []*LinkedCiArtifactSyncResponse `json:"results"`
	UnauthorizedCount  int                             `json:"unauthorizedCount"`
}
