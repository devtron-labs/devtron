package common

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

func (pType PipelineType) ToString() string {
	return string(pType)
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
	BEFORE_DOCKER_BUILD = "BEFORE_DOCKER_BUILD"
	AFTER_DOCKER_BUILD  = "AFTER_DOCKER_BUILD"
)

type RefPluginName = string

const (
	COPY_CONTAINER_IMAGE            RefPluginName = "Copy container image"
	COPY_CONTAINER_IMAGE_VERSION_V1               = "1.0.0"
	COPY_CONTAINER_IMAGE_VERSION_V2               = "2.0.0"
)
