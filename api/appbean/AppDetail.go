package appbean

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
)

type AppDetail struct {
	Metadata                 *AppMetadata                    `json:"metadata,notnull" validate:"required"`
	GitMaterials             []*GitMaterial                  `json:"gitMaterials,notnull"`
	DockerConfig             *DockerConfig                   `json:"dockerConfig"`
	GlobalDeploymentTemplate *DeploymentTemplate             `json:"globalDeploymentTemplate,notnull"`
	AppWorkflows             []*AppWorkflow                  `json:"workflows"`
	GlobalConfigMaps         []*ConfigMap                    `json:"globalConfigMaps"`
	GlobalSecrets            []*Secret                       `json:"globalSecrets"`
	EnvironmentOverrides     map[string]*EnvironmentOverride `json:"environmentOverride"`
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
	GitProviderUrl  string `json:"gitProviderUrl,notnull" validate:"required"`
	GitRepoUrl      string `json:"gitRepoUrl,notnull" validate:"required"`
	CheckoutPath    string `json:"checkoutPath,notnull" validate:"required"`
	FetchSubmodules bool   `json:"fetchSubmodules"`
}

type DockerConfig struct {
	DockerRegistry   string             `json:"dockerRegistry" validate:"required"`
	DockerRepository string             `json:"dockerRepository" validate:"required"`
	BuildConfig      *DockerBuildConfig `json:"dockerBuildConfig"`
}

type DockerBuildConfig struct {
	GitCheckoutPath        string            `json:"gitCheckoutPath,omitempty" validate:"required"`
	DockerfileRelativePath string            `json:"dockerfileRelativePath,omitempty" validate:"required"`
	Args                   map[string]string `json:"args,omitempty"`
}

type DeploymentTemplate struct {
	ChartRefId     int                    `json:"chartRefId,notnull" validate:"required"`
	Template       map[string]interface{} `json:"template,notnull" validate:"required"`
	ShowAppMetrics bool                   `json:"showAppMetrics"`
	IsOverride     bool                   `json:"isOverride"`
}

type AppWorkflow struct {
	Name        string               `json:"name"`
	CiPipeline  *CiPipelineDetails   `json:"ciPipeline"`
	CdPipelines []*CdPipelineDetails `json:"cdPipelines"`
}

type CiPipelineDetails struct {
	Name                      string                      `json:"name" validate:"required"` //name suffix of corresponding pipeline
	IsManual                  bool                        `json:"isManual" validate:"required"`
	CiPipelineMaterialsConfig []*CiPipelineMaterialConfig `json:"ciPipelineMaterialsConfig"`
	DockerBuildArgs           map[string]string           `json:"dockerBuildArgs"`
	BeforeDockerBuildScripts  []*BuildScript              `json:"beforeDockerBuildScripts"`
	AfterDockerBuildScripts   []*BuildScript              `json:"afterDockerBuildScripts"`
	VulnerabilityScanEnabled  bool                        `json:"vulnerabilitiesScanEnabled"`
	IsExternal                bool                        `json:"isExternal"` // true for linked and external
}

type CiPipelineMaterialConfig struct {
	Type         pipelineConfig.SourceType `json:"type,omitempty" validate:"oneof=SOURCE_TYPE_BRANCH_FIXED WEBHOOK"`
	Value        string                    `json:"value,omitempty" `
	CheckoutPath string                    `json:"checkoutPath"`
}

type BuildScript struct {
	Index               int    `json:"index"`
	Name                string `json:"name"`
	Script              string `json:"script"`
	ReportDirectoryPath string `json:"reportDirectoryPath"`
}

type CdPipelineDetails struct {
	Name                          string                            `json:"name"` //pipelineName
	EnvironmentName               string                            `json:"environmentName" `
	TriggerType                   pipelineConfig.TriggerType        `json:"triggerType" validate:"required"`
	DeploymentType                pipelineConfig.DeploymentTemplate `json:"deploymentType,omitempty" validate:"oneof=BLUE-GREEN ROLLING CANARY RECREATE"` //
	DeploymentStrategies          []*DeploymentStrategy             `json:"deploymentStrategies"`
	PreStage                      *CdStage                          `json:"preStage"`
	PostStage                     *CdStage                          `json:"postStage"`
	PreStageConfigMapSecretNames  *CdStageConfigMapSecretNames      `json:"preStageConfigMapSecretNames"`
	PostStageConfigMapSecretNames *CdStageConfigMapSecretNames      `json:"postStageConfigMapSecretNames"`
	RunPreStageInEnv              bool                              `json:"runPreStageInEnv"`
	RunPostStageInEnv             bool                              `json:"runPostStageInEnv"`
	IsClusterCdActive             bool                              `json:"isClusterCdActive"`
}

type DeploymentStrategy struct {
	DeploymentType pipelineConfig.DeploymentTemplate `json:"deploymentType,omitempty" validate:"oneof=BLUE-GREEN ROLLING CANARY RECREATE"` //
	Config         map[string]interface{}            `json:"config,omitempty" validate:"string"`
	IsDefault      bool                              `json:"isDefault" validate:"required"`
}

type CdStage struct {
	Name        string                     `json:"name,omitempty"`
	TriggerType pipelineConfig.TriggerType `json:"triggerType,omitempty"`
	Config      string                     `json:"config,omitempty"`
}

type CdStageConfigMapSecretNames struct {
	ConfigMaps []string `json:"configMaps"`
	Secrets    []string `json:"secrets"`
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
