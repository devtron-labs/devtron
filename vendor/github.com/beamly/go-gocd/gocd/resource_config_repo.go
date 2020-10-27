package gocd

// SetVersion sets a version string for this config repo
func (c *ConfigRepo) SetVersion(version string) {
	c.Version = version
}

// GetVersion retrieves a version string for this config repo
func (c *ConfigRepo) GetVersion() (version string) {
	return c.Version
}
