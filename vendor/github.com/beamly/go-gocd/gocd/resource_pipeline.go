package gocd

// GetStages from the pipeline
func (p *Pipeline) GetStages() []*Stage {
	return p.Stages
}

// GetStage from the pipeline template
func (p *Pipeline) GetStage(stageName string) (stage *Stage) {
	for _, s := range p.Stages {
		if s.Name == stageName {
			stage = s
			break
		}
	}
	return
}

// RemoveLinks from the pipeline object for json marshalling.
func (p *Pipeline) RemoveLinks() {
	p.Links = nil
}

// GetLinks from pipeline
func (p *Pipeline) GetLinks() *HALLinks {
	return p.Links
}

// GetName of the pipeline
func (p *Pipeline) GetName() string {
	return p.Name
}

// SetStages overwrites any existing stages
func (p *Pipeline) SetStages(stages []*Stage) {
	p.Stages = stages
}

// AddStage appends a stage to this pipeline
func (p *Pipeline) AddStage(stage *Stage) {
	p.Stages = append(p.Stages, stage)
}

// SetStage replaces a stage if it already exists
func (p *Pipeline) SetStage(newStage *Stage) {
	for i, stage := range p.Stages {
		if stage.Name == newStage.Name {
			p.Stages[i] = newStage
			return
		}
	}
	p.AddStage(newStage)
}

// SetVersion sets a version string for this pipeline
func (p *Pipeline) SetVersion(version string) {
	p.Version = version
}

// GetVersion retrieves a version string for this pipeline
func (p *Pipeline) GetVersion() (version string) {
	return p.Version
}

// GetVersion of pipeline config
func (pr *PipelineConfigRequest) GetVersion() (version string) {
	return pr.Pipeline.GetVersion()
}

// SetVersion of pipeline config
func (pr *PipelineConfigRequest) SetVersion(version string) {
	pr.Pipeline.SetVersion(version)
}
