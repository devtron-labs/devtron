package constants

const CDPipelineNotFoundErr = "cd pipeline not found"
const CiPipelineNotFoundErr = "ci pipeline not found"

type PipelineType string

const (
	NORMAL    PipelineType = "NORMAL"
	LINKED    PipelineType = "LINKED"
	EXTERNAL  PipelineType = "EXTERNAL"
	CI_JOB    PipelineType = "CI_JOB"
	LINKED_CD PipelineType = "LINKED_CD"
)
