package gocd

// RemoveLinks gets the EnvironmentsResponse ready to be submitted to the GoCD API.
func (er *EnvironmentsResponse) RemoveLinks() {
	er.Links = nil
	for _, env := range er.Embedded.Environments {
		env.RemoveLinks()
	}
}

// GetLinks from the EnvironmentResponse
func (er *EnvironmentsResponse) GetLinks() *HALLinks {
	return er.Links
}

// RemoveLinks gets the Environment ready to be submitted to the GoCD API.
func (env *Environment) RemoveLinks() {
	env.Links = nil
	for _, p := range env.Pipelines {
		p.RemoveLinks()
	}
	for _, a := range env.Agents {
		a.RemoveLinks()
	}
}

// GetLinks from the Environment
func (env *Environment) GetLinks() *HALLinks {
	return env.Links
}

// SetVersion sets a version string for this pipeline
func (env *Environment) SetVersion(version string) {
	env.Version = version
}

// GetVersion retrieves a version string for this pipeline
func (env *Environment) GetVersion() (version string) {
	return env.Version
}
