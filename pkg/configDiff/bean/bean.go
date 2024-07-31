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

func NewConfigProperty() *ConfigProperty {
	return &ConfigProperty{}
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

type DeploymentAndCmCsConfig struct {
	ResourceType bean.ResourceType `json:"resourceType"`
	Data         json.RawMessage   `json:"data"`
}

func NewDeploymentAndCmCsConfig() *DeploymentAndCmCsConfig {
	return &DeploymentAndCmCsConfig{}
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
	DeploymentTemplate *DeploymentAndCmCsConfig `json:"deploymentTemplate"`
	ConfigMapsData     *DeploymentAndCmCsConfig `json:"configMapData"`
	SecretsData        *DeploymentAndCmCsConfig `json:"secretsData"`
}

func NewDeploymentAndCmCsConfigDto() *DeploymentAndCmCsConfigDto {
	return &DeploymentAndCmCsConfigDto{}
}

func (r *DeploymentAndCmCsConfigDto) WithDeploymentTemplateData(data *DeploymentAndCmCsConfig) *DeploymentAndCmCsConfigDto {
	r.DeploymentTemplate = data
	return r
}
func (r *DeploymentAndCmCsConfigDto) WithConfigMapData(data *DeploymentAndCmCsConfig) *DeploymentAndCmCsConfigDto {
	r.ConfigMapsData = data
	return r
}
func (r *DeploymentAndCmCsConfigDto) WithSecretData(data *DeploymentAndCmCsConfig) *DeploymentAndCmCsConfigDto {
	r.SecretsData = data
	return r
}

type ConfigDataQueryParams struct {
	AppName      string
	EnvName      string
	ConfigType   string
	IdentifierId int
	PipelineId   int // req for fetching previous deployments data
	ResourceName string
	ResourceType string
	ResourceId   int
	UserId       int32
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

const (
	InvalidConfigTypeErr = "invalid config type provided, please send a valid config type"
)
