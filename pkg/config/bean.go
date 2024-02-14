package config

import "github.com/devtron-labs/devtron/pkg/pipeline/bean"

type ConfigState int

const (
	PublishedConfigState ConfigState = 3
)

type ConfigProperty struct {
	Name        string            `json:"name"`
	ConfigState ConfigState       `json:"configState"`
	Type        bean.ResourceType `json:"type"`
}
type ConfigDataResponse struct {
	ResourceConfig []*ConfigProperty `json:"resourceConfig"`
}

func (config ConfigProperty) getKey() string {
	return string(config.Type) + config.Name
}
