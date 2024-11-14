/*
 * Copyright (c) 2024. Devtron Inc.
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
