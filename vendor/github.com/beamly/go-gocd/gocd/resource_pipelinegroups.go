package gocd

// GetGroupByPipelineName finds the pipeline group for the name of the pipeline supplied
func (pg *PipelineGroups) GetGroupByPipelineName(pipelineName string) *PipelineGroup {
	for _, pipelineGroup := range *pg {
		for _, pipeline := range pipelineGroup.Pipelines {
			if pipeline.Name == pipelineName {
				return pipelineGroup
			}
		}
	}
	return nil
}

// GetGroupByPipeline finds the pipeline group for the pipeline supplied
func (pg *PipelineGroups) GetGroupByPipeline(pipeline *Pipeline) *PipelineGroup {
	return pg.GetGroupByPipelineName(pipeline.Name)
}
