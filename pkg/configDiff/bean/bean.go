package bean

import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
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

type ConfigArea string

const (
	AppConfiguration  ConfigArea = "AppConfiguration"
	DeploymentHistory ConfigArea = "DeploymentHistory"
	CdRollback        ConfigArea = "CdRollback"
	ResolveData       ConfigArea = "ResolveData"
)

func (r ConfigArea) ToString() string {
	return string(r)
}

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
	ResourceType     bean.ResourceType            `json:"resourceType"`
	Data             json.RawMessage              `json:"data"`
	VariableSnapshot map[string]map[string]string `json:"variableSnapshot"` // for deployment->{Deployment Template: resolvedValuesMap}, for cm->{cmComponentName: resolvedValuesMap}
	ResolvedValue    json.RawMessage              `json:"resolvedValue"`
	// for deployment template
	TemplateVersion     string `json:"templateVersion,omitempty"`
	IsAppMetricsEnabled bool   `json:"isAppMetricsEnabled,omitempty"`
	//for pipeline strategy
	PipelineTriggerType pipelineConfig.TriggerType `json:"pipelineTriggerType,omitempty"`
	Strategy            string                     `json:"strategy,omitempty"`
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

func (r *DeploymentAndCmCsConfig) WithVariableSnapshot(snapshot map[string]map[string]string) *DeploymentAndCmCsConfig {
	r.VariableSnapshot = snapshot
	return r
}

func (r *DeploymentAndCmCsConfig) WithResolvedValue(resolvedValue json.RawMessage) *DeploymentAndCmCsConfig {
	r.ResolvedValue = resolvedValue
	return r
}

func (r *DeploymentAndCmCsConfig) WithDeploymentConfigMetadata(templateVersion string, isAppMetricsEnabled bool) *DeploymentAndCmCsConfig {
	r.TemplateVersion = templateVersion
	r.IsAppMetricsEnabled = isAppMetricsEnabled
	return r
}

func (r *DeploymentAndCmCsConfig) WithPipelineStrategyMetadata(pipelineTriggerType pipelineConfig.TriggerType, strategy string) *DeploymentAndCmCsConfig {
	r.PipelineTriggerType = pipelineTriggerType
	r.Strategy = strategy
	return r
}

type DeploymentAndCmCsConfigDto struct {
	DeploymentTemplate *DeploymentAndCmCsConfig `json:"deploymentTemplate"`
	ConfigMapsData     *DeploymentAndCmCsConfig `json:"configMapData"`
	SecretsData        *DeploymentAndCmCsConfig `json:"secretsData"`
	PipelineConfigData *DeploymentAndCmCsConfig `json:"pipelineConfigData,omitempty"`
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
func (r *DeploymentAndCmCsConfigDto) WithPipelineConfigData(data *DeploymentAndCmCsConfig) *DeploymentAndCmCsConfigDto {
	r.PipelineConfigData = data
	return r
}

type ConfigDataQueryParams struct {
	AppName      string `schema:"appName"`
	EnvName      string `schema:"envName"`
	ConfigType   string `schema:"configType"`
	IdentifierId int    `schema:"identifierId"`
	PipelineId   int    `schema:"pipelineId"`   // req for fetching previous deployments data
	ResourceName string `schema:"resourceName"` // used in case of cm and cs
	ResourceType string `schema:"resourceType"` // used in case of cm and cs
	ResourceId   int    `schema:"resourceId"`   // used in case of cm and cs
	UserId       int32  `schema:"-"`
	WfrId        int    `schema:"wfrId"`
	ConfigArea   string `schema:"configArea"`
	Values       string `schema:"values"`
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

type CmCsMetadataDto struct {
	CmMap            map[string]*bean3.ConfigData
	SecretMap        map[string]*bean3.ConfigData
	ConfigAppLevelId int
	ConfigEnvLevelId int
}

type ResolvedCmCsMetadataDto struct {
	ResolvedConfigMapData string
	ResolvedSecretData    string
	VariableMapCM         map[string]map[string]string
	VariableMapCS         map[string]map[string]string
}

type ValuesDto struct {
	Values string `json:"values"`
}

type DeploymentTemplateMetadata struct {
	DeploymentTemplateJson json.RawMessage
	TemplateVersion        string
	IsAppMetricsEnabled    bool
}

const (
	NoDeploymentDoneForSelectedImage = "there were no deployments done for the selected image"
)
