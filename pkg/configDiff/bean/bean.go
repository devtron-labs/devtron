package bean

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

type ConfigState string

const (
	PublishedConfigState ConfigState = "Published"
)

type ConfigStage string

const (
	Env        ConfigStage = "Env"
	Inheriting ConfigStage = "Inheriting"
	Overridden ConfigStage = "Overridden"
)

type ConfigProperty struct {
	Id          int               `json:"id"`
	Name        string            `json:"name"`
	ConfigState ConfigState       `json:"configState"`
	Type        bean.ResourceType `json:"type"`
	ConfigStage ConfigStage       `json:"configStage"`
}

func (r *ConfigProperty) IsConfigPropertyGlobal() bool {
	return r.ConfigStage == Inheriting
}

type ConfigDataResponse struct {
	ResourceConfig []*ConfigProperty `json:"resourceConfig"`
}

func NewConfigDataResponse() *ConfigDataResponse {
	return &ConfigDataResponse{}
}

func (r *ConfigDataResponse) WithResourceConfig(resourceConfig []*ConfigProperty) *ConfigDataResponse {
	r.ResourceConfig = resourceConfig
	return r
}

func (r *ConfigProperty) GetKey() string {
	return fmt.Sprintf("%s-%s", string(r.Type), r.Name)
}

type ConfigPropertyIdentifier struct {
	Name string            `json:"name"`
	Type bean.ResourceType `json:"type"`
}

func (r *ConfigProperty) GetIdentifier() ConfigPropertyIdentifier {
	return ConfigPropertyIdentifier{
		Name: r.Name,
		Type: r.Type,
	}
}
