package gocd

// SetVersion sets a version string for this role
func (r *Role) SetVersion(version string) {
	r.Version = version
}

// GetVersion retrieves a version string for this role
func (r Role) GetVersion() (version string) {
	return r.Version
}

// RemoveLinks from the pipeline object for json marshalling.
func (r *Role) RemoveLinks() {
	r.Links = nil
}

// GetLinks from pipeline
func (r *Role) GetLinks() *HALLinks {
	return r.Links
}
