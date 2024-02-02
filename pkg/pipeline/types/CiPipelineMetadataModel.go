package types

import "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"

type CiPipelineMetadata struct {
	CiPipelineId             int
	CiPipelineName           string
	IsDockerConfigOverridden bool
	ScanEnabled              bool
	IsExternal               bool
	AppId                    int
	DockerArgs               string
	PipelineType             string
}

func InitFromPipelineEntity(pipeline *pipelineConfig.CiPipeline) *CiPipelineMetadata {
	pipelineMetadata := &CiPipelineMetadata{}
	pipelineMetadata.CiPipelineId = pipeline.Id
	pipelineMetadata.CiPipelineName = pipeline.Name
	pipelineMetadata.IsExternal = pipeline.IsExternal
	pipelineMetadata.IsDockerConfigOverridden = pipeline.IsDockerConfigOverridden
	pipelineMetadata.ScanEnabled = pipeline.ScanEnabled
	pipelineMetadata.AppId = pipeline.AppId
	pipelineMetadata.DockerArgs = pipeline.DockerArgs
	pipelineMetadata.PipelineType = pipeline.PipelineType
	return pipelineMetadata
}
