package bean

import "C"
import (
	"encoding/json"
	"fmt"
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
	IsAppAdmin         bool                     `json:"isAppAdmin"`
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
	AppName      string `schema:"appName"`
	EnvName      string `schema:"envName"`
	ConfigType   string `schema:"configType"`
	IdentifierId int    `schema:"identifierId"`
	PipelineId   int    `schema:"pipelineId"` // req for fetching previous deployments data
	ResourceName string `schema:"resourceName"`
	ResourceType string `schema:"resourceType"`
	ResourceId   int    `schema:"resourceId"`
	UserId       int32  `schema:"-"`
}

// FilterCriteria []string `schema:"filterCriteria"`
// OffSet         int      `schema:"offSet"`
// Limit          int      `schema:"limit"`
func (r *ConfigDataQueryParams) IsResourceTypeSecret() bool {
	return r.ResourceType == bean.CS.ToString()
}

func (r *ConfigDataQueryParams) IsResourceTypeConfigMap() bool {
	return r.ResourceType == bean.CM.ToString()
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
