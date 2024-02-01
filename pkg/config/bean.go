package config

type ConfigState int

const (
	PublishedConfigState ConfigState = iota + 2
)

type ResourceType string

const (
	CM                 ResourceType = "ConfigMap"
	CS                 ResourceType = "Secret"
	DeploymentTemplate ResourceType = "Deployment Template"
)

type ConfigDefinition struct {
	Name        string       `json:"name"`
	ConfigState ConfigState  `json:"draftState"`
	Type        ResourceType `json:"type"`
}
type ConfigDataResponse struct {
	ResourceConfig []ConfigDefinition `json:"resourceConfig"`
}
