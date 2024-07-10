package bean

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

type ConfigState int

const (
	PublishedConfigState ConfigState = 3
)

type ConfigProperty struct {
	Id          int               `json:"id"`
	Name        string            `json:"name"`
	ConfigState ConfigState       `json:"configState"`
	Type        bean.ResourceType `json:"type"`
	Overridden  bool              `json:"overridden"`
	Global      bool              `json:"global"`
}

func NewConfigProperty() *ConfigProperty {
	return &ConfigProperty{}
}

func (r *ConfigProperty) IsConfigPropertyGlobal() bool {
	return r.Global
}

func (r *ConfigProperty) SetConfigProperty(Name string, ConfigState ConfigState, Type bean.ResourceType, Overridden bool, Global bool) *ConfigProperty {
	r.Name = Name
	r.ConfigState = ConfigState
	r.Type = Type
	r.Overridden = Overridden
	r.Global = Global
	return r
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

type DeploymentAndCmCsConfig struct {
	Id          int         `json:"id ,omitempty"`
	Name        string      `json:"name"`
	CmCsType    string      `json:"cmcsType ,omitempty"` //example:- environment or volume, in case of cm cs
	Data        string      `json:"data"`
	ConfigState ConfigState `json:"configState"`
}

type DeploymentAndCmCsResponseDto struct {
	AppId         int                        `json:"appId"`
	EnvironmentId int                        `json:"environmentId"`
	ConfigData    []*DeploymentAndCmCsConfig `json:"configData"`
}
