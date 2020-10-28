package gocd

// GetLinks returns HAL links for agent
func (a *Agent) GetLinks() *HALLinks {
	return a.Links
}

// RemoveLinks sets the `Link` attribute as `nil`. Used when rendering an `Agent` struct to JSON.
func (a *Agent) RemoveLinks() {
	a.Links = nil
}
