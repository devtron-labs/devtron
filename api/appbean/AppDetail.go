package appbean

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
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

type AppWorkflowCloneDto struct {
	AppId                int                             `json:"appId"`
	AppName              string                          `json:"appName"`
	AppWorkflows         []*AppWorkflow                  `json:"workflows"`
	EnvironmentOverrides map[string]*EnvironmentOverride `json:"environmentOverride"`
}

type AppMetadata struct {
	AppName     string      `json:"appName" validate:"required"`
	ProjectName string      `json:"projectName" validate:"required"`
	Labels      []*AppLabel `json:"labels"`
}

type AppLabel struct {
	Key       string `json:"key,notnull" validate:"required"`
	Value     string `json:"value,notnull" validate:"required"`
	Propagate bool   `json:"propagate"`
}

type GitMaterial struct {
	GitProviderUrl  string `json:"gitProviderUrl,notnull" validate:"required"`
	GitRepoUrl      string `json:"gitRepoUrl,notnull" validate:"required"`
	CheckoutPath    string `json:"checkoutPath,notnull" validate:"required"`
	FetchSubmodules bool   `json:"fetchSubmodules"`
}

type DockerConfig struct {
	DockerRegistry    string                  `json:"dockerRegistry" validate:"required"`
	DockerRepository  string                  `json:"dockerRepository" validate:"required"`
	CiBuildConfig     *bean.CiBuildConfigBean `json:"ciBuildConfig"`
	DockerBuildConfig *DockerBuildConfig      `json:"dockerBuildConfig,omitempty"` // Deprecated, should use CiBuildConfig for development
	CheckoutPath      string                  `json:"checkoutPath"`
}

type DockerBuildConfig struct {
	GitCheckoutPath        string            `json:"gitCheckoutPath,omitempty" validate:"required"`
	DockerfileRelativePath string            `json:"dockerfileRelativePath,omitempty" validate:"required"`
	Args                   map[string]string `json:"args,omitempty"`
	TargetPlatform         string            `json:"targetPlatform"`
	DockerBuildOptions     map[string]string `json:"dockerBuildOptions,omitempty"`
}

type DeploymentTemplate struct {
	ChartRefId        int                         `json:"chartRefId,notnull" validate:"required"`
	Template          map[string]interface{}      `json:"template,notnull" validate:"required"`
	ShowAppMetrics    bool                        `json:"showAppMetrics"`
	IsOverride        bool                        `json:"isOverride"`
	IsBasicViewLocked bool                        `json:"isBasicViewLocked"`
	CurrentViewEditor models.ChartsViewEditorType `json:"currentViewEditor"` //default "UNDEFINED" in db
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
	PreBuildStage             *bean.PipelineStageDto      `json:"preBuildStage,omitempty"`
	PostBuildStage            *bean.PipelineStageDto      `json:"postBuildStage,omitempty"`
	IsExternal                bool                        `json:"isExternal"` // true for linked and external
	ParentCiPipeline          int                         `json:"parentCiPipeline,omitempty"`
	ParentAppId               int                         `json:"parentAppId,omitempty"`
	LinkedCount               int                         `json:"linkedCount,omitempty"`
}

type CiPipelineMaterialConfig struct {
	Type          pipelineConfig.SourceType `json:"type,omitempty" validate:"oneof=SOURCE_TYPE_BRANCH_FIXED WEBHOOK"`
	Value         string                    `json:"value,omitempty" `
	CheckoutPath  string                    `json:"checkoutPath"`
	GitMaterialId int                       `json:"gitMaterialId"`
}

type BuildScript struct {
	Index               int    `json:"index"`
	Name                string `json:"name"`
	Script              string `json:"script"`
	ReportDirectoryPath string `json:"reportDirectoryPath"`
}

type CdPipelineDetails struct {
	Name                          string                                 `json:"name"` //pipelineName
	EnvironmentName               string                                 `json:"environmentName" `
	TriggerType                   pipelineConfig.TriggerType             `json:"triggerType" validate:"required"`
	DeploymentAppType             string                                 `json:"deploymentAppType"`
	DeploymentStrategyType        chartRepoRepository.DeploymentStrategy `json:"deploymentType,omitempty"` //
	DeploymentStrategies          []*DeploymentStrategy                  `json:"deploymentStrategies"`
	PreStage                      *CdStage                               `json:"preStage"`
	PostStage                     *CdStage                               `json:"postStage"`
	PreStageConfigMapSecretNames  *CdStageConfigMapSecretNames           `json:"preStageConfigMapSecretNames"`
	PostStageConfigMapSecretNames *CdStageConfigMapSecretNames           `json:"postStageConfigMapSecretNames"`
	RunPreStageInEnv              bool                                   `json:"runPreStageInEnv"`
	RunPostStageInEnv             bool                                   `json:"runPostStageInEnv"`
	IsClusterCdActive             bool                                   `json:"isClusterCdActive"`
}

type DeploymentStrategy struct {
	DeploymentStrategyType chartRepoRepository.DeploymentStrategy `json:"deploymentType,omitempty"` //
	Config                 map[string]interface{}                 `json:"config,omitempty" validate:"string"`
	IsDefault              bool                                   `json:"isDefault" validate:"required"`
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
