package config

type ConfigState int

const (
	PublishedConfigState ConfigState = 3
)

type ResourceType string

const (
	CM                 ResourceType = "ConfigMap"
	CS                 ResourceType = "Secret"
	DeploymentTemplate ResourceType = "Deployment Template"
)

type ConfigProperty struct {
	Name        string       `json:"name"`
	ConfigState ConfigState  `json:"configState"`
	Type        ResourceType `json:"type"`
}
type ConfigDataResponse struct {
	ResourceConfig []ConfigProperty `json:"resourceConfig"`
}
