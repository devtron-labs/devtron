package gocd

// GetStages from the pipeline template
func (pt PipelineTemplate) GetStages() []*Stage {
	return pt.Stages
}

// GetStage from the pipeline template
func (pt PipelineTemplate) GetStage(stageName string) *Stage {
	for _, stage := range pt.Stages {
		if stage.Name == stageName {
			return stage
		}
	}
	return nil
}

// GetName of the pipeline template
func (pt PipelineTemplate) GetName() string {
	return pt.Name
}

// SetStages overwrites any existing stages
func (pt *PipelineTemplate) SetStages(stages []*Stage) {
	pt.Stages = stages
}

// AddStage appends a stage to this pipeline
func (pt *PipelineTemplate) AddStage(stage *Stage) {
	pt.Stages = append(pt.Stages, stage)
}

// RemoveLinks gets the PipelineTemplate ready to be submitted to the GoCD API.
func (pt *PipelineTemplate) RemoveLinks() {
	pt.Links = nil
}

// Pipelines returns a list of Pipelines attached to this PipelineTemplate object.
func (pt PipelineTemplate) Pipelines() []*Pipeline {
	return pt.Embedded.Pipelines
}

// SetStage replaces a stage if it already exists
func (pt *PipelineTemplate) SetStage(newStage *Stage) {
	for i, stage := range pt.Stages {
		if stage.Name == newStage.Name {
			pt.Stages[i] = newStage
			return
		}
	}
	pt.AddStage(newStage)
}

// SetVersion sets a version string for this pipeline
func (pt *PipelineTemplate) SetVersion(version string) {
	pt.Version = version
}

// GetVersion retrieves a version string for this pipeline
func (pt PipelineTemplate) GetVersion() (version string) {
	return pt.Version
}

// SetVersion sets a version string for this pipeline
func (pt *PipelineTemplateRequest) SetVersion(version string) {
	pt.Version = version
}

// GetVersion retrieves a version string for this pipeline
func (pt PipelineTemplateRequest) GetVersion() (version string) {
	return pt.Version
}
