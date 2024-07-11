package bean

import (
	"encoding/json"
	"fmt"
	v1 "github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

type ConfigState string

const (
	PublishedConfigState ConfigState = "PublishedOnly"
)

func (r ConfigState) ToString() string {
	return string(r)
}

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
	Id           int               `json:"id ,omitempty"`
	Name         string            `json:"name"`
	ResourceType bean.ResourceType `json:"resourceType"`
	Data         json.RawMessage   `json:"data"`
	ConfigState  ConfigState       `json:"configState"`
}

func NewDeploymentAndCmCsConfig() *DeploymentAndCmCsConfig {
	return &DeploymentAndCmCsConfig{}
}

func (r *DeploymentAndCmCsConfig) WithIdAndName(id int, name string) *DeploymentAndCmCsConfig {
	r.Id = id
	r.Name = name
	return r
}

func (r *DeploymentAndCmCsConfig) WithConfigState(configState ConfigState) *DeploymentAndCmCsConfig {
	r.ConfigState = configState
	return r
}

func (r *DeploymentAndCmCsConfig) WithResourceType(resourceType bean.ResourceType) *DeploymentAndCmCsConfig {
	r.ResourceType = resourceType
	return r
}

func (r *DeploymentAndCmCsConfig) WithConfigData(data json.RawMessage) *DeploymentAndCmCsConfig {
	r.Data = data
	return r
}

type DeploymentAndCmCsConfigDto struct {
	AppId         int                        `json:"appId"`
	EnvironmentId int                        `json:"environmentId"`
	ConfigData    []*DeploymentAndCmCsConfig `json:"configData"`
}

func NewDeploymentAndCmCsConfigDto() *DeploymentAndCmCsConfigDto {
	return &DeploymentAndCmCsConfigDto{ConfigData: make([]*DeploymentAndCmCsConfig, 0)}
}

func (r *DeploymentAndCmCsConfigDto) WithConfigData(configData []*DeploymentAndCmCsConfig) *DeploymentAndCmCsConfigDto {
	r.ConfigData = configData
	return r
}

func (r *DeploymentAndCmCsConfigDto) WithAppAndEnvIdId(appId, envId int) *DeploymentAndCmCsConfigDto {
	r.AppId = appId
	r.EnvironmentId = envId
	return r
}

type ConfigDataQueryParams struct {
	AppName      string
	EnvName      string
	ConfigType   string
	IdentifierId int
	ResourceName string
	ResourceType string
}

func (r *ConfigDataQueryParams) IsResourceTypeSecret() bool {
	return r.ResourceType == v1.Secret
}

func (r *ConfigDataQueryParams) IsResourceTypeConfigMap() bool {
	return r.ResourceType == v1.ConfigMap
}

func (r *ConfigDataQueryParams) IsEnvNameProvided() bool {
	return len(r.EnvName) > 0
}
func (r *ConfigDataQueryParams) IsValidConfigType() bool {
	return r.ConfigType == PublishedConfigState.ToString()
}

func (r *ConfigDataQueryParams) IsRequestMadeForOneResource() bool {
	return len(r.ResourceName) > 0 && len(r.ResourceType) > 0
}

func (r *ConfigDataQueryParams) WithConfigDataQueryParams(appName, envName, configType, resourceName, resourceType string, identifierId int) {
	r.AppName = appName
	r.EnvName = envName
	r.ConfigType = configType
	r.ResourceName = resourceName
	r.IdentifierId = identifierId
	r.ResourceType = resourceType
}

const (
	InvalidConfigTypeErr = "invalid config type provided, please send a valid config type"
)
