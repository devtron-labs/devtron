package gocd

import "net/url"

// StageContainer describes structs which contain stages, eg Pipelines and PipelineTemplates
type StageContainer interface {
	GetName() string
	SetStage(stage *Stage)
	GetStage(string) *Stage
	SetStages(stages []*Stage)
	GetStages() []*Stage
	AddStage(stage *Stage)
	Versioned
}

// HALContainer represents objects with HAL _link and _embedded resources.
type HALContainer interface {
	RemoveLinks()
	GetLinks() *HALLinks
}

// Versioned describes resources which can get and set versions
type Versioned interface {
	GetVersion() string
	SetVersion(version string)
}

// HALLink describes a HAL link
type HALLink struct {
	Name string
	URL  *url.URL
}
