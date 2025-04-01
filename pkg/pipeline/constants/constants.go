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

package constants

const CDPipelineNotFoundErr = "cd pipeline not found"
const CiPipelineNotFoundErr = "ci pipeline not found"

type PipelineType string

// default PipelineType
const DefaultPipelineType = CI_BUILD

const (
	CI_BUILD  PipelineType = "CI_BUILD"
	LINKED    PipelineType = "LINKED"
	EXTERNAL  PipelineType = "EXTERNAL"
	CI_JOB    PipelineType = "CI_JOB"
	LINKED_CD PipelineType = "LINKED_CD"
)

type PatchPipelineActionResponse string

const (
	PATCH_PIPELINE_ACTION_CREATED        PatchPipelineActionResponse = "created"
	PATCH_PIPELINE_ACTION_UPDATED        PatchPipelineActionResponse = "updated"
	PATCH_PIPELINE_ACTION_DELETED        PatchPipelineActionResponse = "deleted"
	PATCH_PIPELINE_ACTION_IGNORED        PatchPipelineActionResponse = "ignored"
	PATCH_PIPELINE_ACTION_ERRORED        PatchPipelineActionResponse = "errored"
	PATCH_PIPELINE_ACTION_NOT_APPLICABLE PatchPipelineActionResponse = "N/A"
)

func (r PipelineType) ToString() string {
	return string(r)
}

func (pType PipelineType) IsValidPipelineType() bool {
	switch pType {
	case CI_BUILD, LINKED, EXTERNAL, CI_JOB, LINKED_CD:
		return true
	default:
		return false
	}
}

const (
	UNIQUE_DEPLOYMENT_APP_NAME = "unique_deployment_app_name"
)

const (
	ExtraEnvVarExternalCiArtifactKey = "externalCiArtifact"
	ExtraEnvVarImageDigestKey        = "imageDigest"
)

const DefaultCiWorkflowNamespace = "devtron-ci"
const Running = "Running"
const Starting = "Starting"
const TERMINATE_MESSAGE = "workflow shutdown with strategy: Terminate"
const FORCE_ABORT_MESSAGE_AFTER_STARTING_STAGE = "workflow shutdown with strategy: Force Abort"
const POD_TIMEOUT_MESSAGE = "Pod was active on the node longer than the specified deadline"
