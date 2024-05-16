package constants

const CDPipelineNotFoundErr = "cd pipeline not found"
const CiPipelineNotFoundErr = "ci pipeline not found"

type PipelineType string

const (
	CI_BUILD  PipelineType = "CI_BUILD"
	LINKED    PipelineType = "LINKED"
	EXTERNAL  PipelineType = "EXTERNAL"
	CI_JOB    PipelineType = "CI_JOB"
	LINKED_CD PipelineType = "LINKED_CD"
)

func (pType PipelineType) IsValidPipelineType() bool {
	switch pType {
	case CI_BUILD, LINKED, EXTERNAL, CI_JOB, LINKED_CD:
		return true
	default:
		return false
	}
}
