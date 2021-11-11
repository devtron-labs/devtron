package appbean

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
)

type AppDetail struct {
	Metadata                 *AppMetadata                    `json:"metadata,notnull" validate:"required"`
	GitMaterials             []*GitMaterial                  `json:"gitMaterials,notnull" validate:"required"`
	GlobalDeploymentTemplate *DeploymentTemplate             `json:"globalDeploymentTemplate,notnull" validate:"required"`
	GlobalConfigMaps         []*ConfigMap                    `json:"globalConfigMaps"`
	GlobalSecrets            []*Secret                       `json:"globalSecrets"`
	EnvironmentOverrides     map[string]*EnvironmentOverride `json:"environmentOverride"`
	AppWorkflows             []*AppWorkflow                  `json:"workflows"`
	DockerConfig             *DockerConfig                   `json:"dockerConfig:"`
}

type AppMetadata struct {
	AppName     string      `json:"appName" validate:"required"`
	ProjectName string      `json:"projectName" validate:"required"`
	Labels      []*AppLabel `json:"labels"`
}

type AppLabel struct {
	Key   string `json:"key,notnull" validate:"required"`
	Value string `json:"value,notnull" validate:"required"`
}

type GitMaterial struct {
	GitAccountUrl   string `json:"gitAccountUrl,notnull" validate:"required"`
	GitUrl          string `json:"gitUrl,notnull" validate:"required"`
	CheckoutPath    string `json:"checkoutPath,notnull" validate:"required"`
	FetchSubmodules bool   `json:"fetchSubmodules"`
}

type DeploymentTemplate struct {
	ChartRefId     int                    `json:"chartRefId,notnull" validate:"required"`
	Template       map[string]interface{} `json:"template,notnull" validate:"required"`
	ShowAppMetrics bool                   `json:"showAppMetrics"`
}

type ConfigMap struct {
	Name                  string                                `json:"name,notnull" validate:"required"`
	IsExternal            bool                                  `json:"isExternal"`
	UsageType             string                                `json:"usageType,omitempty" validate:"oneof=environment volume"`
	Data                  map[string]interface{}                `json:"data"`
	DataVolumeUsageConfig *ConfigMapSecretDataVolumeUsageConfig `json:"dataVolumeUsageConfig"`
}

type Secret struct {
	Name                  string                                `json:"name,notnull" validate:"required"`
	IsExternal            bool                                  `json:"isExternal"`
	ExternalType          string                                `json:"externalType,omitempty"`
	UsageType             string                                `json:"usageType,omitempty" validate:"oneof=environment volume"`
	Data                  map[string]interface{}                `json:"data"`
	DataVolumeUsageConfig *ConfigMapSecretDataVolumeUsageConfig `json:"dataVolumeUsageConfig"`
	RoleArn               string                                `json:"roleArn"`
	ExternalSecretData    []*ExternalSecret                     `json:"externalSecretData"`
}

type ConfigMapSecretDataVolumeUsageConfig struct {
	MountPath      string `json:"mountPath"`
	SubPath        bool   `json:"subPath"`
	FilePermission string `json:"filePermission"`
}

type ExternalSecret struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Property string `json:"property,omitempty"`
	IsBinary bool   `json:"isBinary"`
}

type EnvironmentOverride struct {
	DeploymentTemplate *DeploymentTemplate `json:"deploymentTemplate"`
	ConfigMaps         []*ConfigMap        `json:"configMaps"`
	Secrets            []*Secret           `json:"secrets"`
}

type AppWorkflow struct {
	Name       string               `json:"name"`
	CiPipeline *CiPipelineDetails   `json:"ciPipeline"`
	CdPipeline []*CdPipelineDetails `json:"cdPipeline"`
}

type CiPipelineDetails struct {
	Name                     string            `json:"name"` //name suffix of corresponding pipeline
	IsManual                 bool              `json:"isManual"`
	DockerArgs               map[string]string `json:"dockerArgs"`
	CiMaterials              []*CiMaterial     `json:"ciMaterials"`
	BeforeDockerBuild        []*Task           `json:"beforeDockerBuild"`
	AfterDockerBuild         []*Task           `json:"afterDockerBuild"`
	BeforeDockerBuildScripts []*CiScript       `json:"beforeDockerBuildScripts"`
	AfterDockerBuildScripts  []*CiScript       `json:"afterDockerBuildScripts"`
	LinkedCount              int               `json:"linkedCount"`
	ScanEnabled              bool              `json:"scanEnabled"`
}

type Task struct {
	Name string   `json:"name"`
	Type string   `json:"type"` //for now ignore this input
	Cmd  string   `json:"cmd"`
	Args []string `json:"args"`
}

//type ExternalCiConfiguration struct {
//	WebhookUrl string `json:"webhookUrl"`
//	Payload    string `json:"payload"`
//	AccessKey  string `json:"accessKey"`
//}

type CiMaterial struct {
	Source          *SourceTypeConfig `json:"source"`
	Path            string            `json:"path"`
	CheckoutPath    string            `json:"checkoutPath"`
	GitMaterialName string            `json:"gitMaterialName"`
}

type SourceTypeConfig struct {
	Type  pipelineConfig.SourceType `json:"type,omitempty" validate:"oneof=SOURCE_TYPE_BRANCH_FIXED SOURCE_TYPE_BRANCH_REGEX SOURCE_TYPE_TAG_ANY WEBHOOK"`
	Value string                    `json:"value,omitempty" `
}

type CiScript struct {
	Index          int    `json:"index"`
	Name           string `json:"name"`
	Script         string `json:"script"`
	OutputLocation string `json:"outputLocation"`
}

type CdPipelineDetails struct {
	EnvironmentName               string                            `json:"environmentName" `
	TriggerType                   pipelineConfig.TriggerType        `json:"triggerType"`
	Name                          string                            `json:"name"` //pipelineName
	Strategies                    []Strategy                        `json:"strategies"`
	Namespace                     string                            `json:"namespace"` //namespace
	DeploymentTemplate            pipelineConfig.DeploymentTemplate `json:"deploymentTemplate"`
	PreStage                      CdStage                           `json:"preStage"`
	PostStage                     CdStage                           `json:"postStage"`
	PreStageConfigMapSecretNames  PreStageConfigMapSecretNames      `json:"preStageConfigMapSecretNames"`
	PostStageConfigMapSecretNames PostStageConfigMapSecretNames     `json:"postStageConfigMapSecretNames"`
	RunPreStageInEnv              bool                              `json:"runPreStageInEnv"`
	RunPostStageInEnv             bool                              `json:"runPostStageInEnv"`
	CdArgoSetup                   bool                              `json:"isClusterCdActive"`
}

type Strategy struct {
	DeploymentTemplate pipelineConfig.DeploymentTemplate `json:"deploymentTemplate,omitempty" validate:"oneof=BLUE-GREEN ROLLING CANARY RECREATE"` //
	Config             json.RawMessage                   `json:"config,omitempty" validate:"string"`
	Default            bool                              `json:"default"`
}

type CdStage struct {
	TriggerType pipelineConfig.TriggerType `json:"triggerType,omitempty"`
	Name        string                     `json:"name,omitempty"`
	Status      string                     `json:"status,omitempty"`
	Config      string                     `json:"config,omitempty"`
}

type PreStageConfigMapSecretNames struct {
	ConfigMaps []string `json:"configMaps"`
	Secrets    []string `json:"secrets"`
}

type PostStageConfigMapSecretNames struct {
	ConfigMaps []string `json:"configMaps"`
	Secrets    []string `json:"secrets"`
}

type DockerConfig struct {
	DockerRegistry   string             `json:"dockerRegistry"`
	DockerRepository string             `json:"dockerRepository"`
	BuildConfig      *DockerBuildConfig `json:"dockerBuildConfig"`
}

type DockerBuildConfig struct {
	GitMaterialUrl string            `json:"gitMaterialUrl,omitempty" validate:"required"`
	DockerfilePath string            `json:"dockerfileRelativePath,omitempty" validate:"required"`
	Args           map[string]string `json:"args,omitempty"`
}
