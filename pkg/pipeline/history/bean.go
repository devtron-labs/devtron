package history

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	"time"
)

type HistoryComponent string

const (
	DEPLOYMENT_TEMPLATE_TYPE_HISTORY_COMPONENT HistoryComponent = "DEPLOYMENT_TEMPLATE"
	CONFIGMAP_TYPE_HISTORY_COMPONENT           HistoryComponent = "CONFIGMAP"
	SECRET_TYPE_HISTORY_COMPONENT              HistoryComponent = "SECRET"
	PIPELINE_STRATEGY_TYPE_HISTORY_COMPONENT   HistoryComponent = "PIPELINE_STRATEGY"
)

type ComponentLevelHistoryDetailDto struct {
	ComponentName string            `json:"componentName"`
	HistoryConfig *HistoryDetailDto `json:"config"`
}

type AllDeploymentConfigurationDetail struct {
	DeploymentTemplateConfig *HistoryDetailDto                 `json:"deploymentTemplate"`
	ConfigMapConfig          []*ComponentLevelHistoryDetailDto `json:"configMap"`
	SecretConfig             []*ComponentLevelHistoryDetailDto `json:"secret"`
	StrategyConfig           *HistoryDetailDto                 `json:"pipelineStrategy"`
	WfrId                    int                               `json:"wfrId"`
}

type DeploymentConfigurationDto struct {
	Id                  int              `json:"id,omitempty"`
	Name                HistoryComponent `json:"name"`
	ChildComponentNames []string         `json:"childList,omitempty"`
}

type DeployedHistoryComponentMetadataDto struct {
	Id               int       `json:"id"`
	DeployedOn       time.Time `json:"deployedOn"`
	DeployedBy       string    `json:"deployedBy"` //emailId of user
	DeploymentStatus string    `json:"deploymentStatus"`
}

type HistoryDetailDto struct {
	//for deployment template
	TemplateName        string `json:"templateName,omitempty"`
	TemplateVersion     string `json:"templateVersion,omitempty"`
	IsAppMetricsEnabled *bool  `json:"isAppMetricsEnabled,omitempty"`
	//for pipeline strategy
	PipelineTriggerType pipelineConfig.TriggerType `json:"pipelineTriggerType,omitempty"`
	Strategy            string                     `json:"strategy,omitempty"`
	//for configmap and secret
	Type               string               `json:"type,omitempty"`
	External           *bool                `json:"external,omitempty"`
	MountPath          string               `json:"mountPath,omitempty"`
	ExternalSecretType string               `json:"externalType,omitempty"`
	RoleARN            string               `json:"roleARN,omitempty"`
	SubPath            *bool                `json:"subPath,omitempty"`
	FilePermission     string               `json:"filePermission,omitempty"`
	CodeEditorValue    *HistoryDetailConfig `json:"codeEditorValue"`
	SecretViewAccess   bool                 `json:"secretViewAccess"` // this is being used to check whether a user can see obscured secret values or not.
}

type HistoryDetailConfig struct {
	DisplayName      string            `json:"displayName"`
	Value            string            `json:"value"`
	VariableSnapshot map[string]string `json:"variableSnapshot"`
	ResolvedValue    string            `json:"resolvedValue"`
}

//history components(deployment template, configMaps, secrets, pipeline strategy) components below

type ConfigMapAndSecretHistoryDto struct {
	Id         int                `json:"id"`
	PipelineId int                `json:"pipelineId"`
	AppId      int                `json:"appId"`
	DataType   string             `json:"dataType,omitempty"`
	ConfigData []*bean.ConfigData `json:"configData,omitempty"`
	Deployed   bool               `json:"deployed"`
	DeployedOn time.Time          `json:"deployedOn"`
	DeployedBy int32              `json:"deployedBy"`
	EmailId    string             `json:"emailId"`
}

type PrePostCdScriptHistoryDto struct {
	Id                   int                              `json:"id"`
	PipelineId           int                              `json:"pipelineId"`
	Script               string                           `json:"script"`
	Stage                string                           `json:"stage"`
	ConfigMapSecretNames PrePostStageConfigMapSecretNames `json:"configmapSecretNames"`
	ConfigMapData        []*bean.ConfigData               `json:"configmapData"`
	SecretData           []*bean.ConfigData               `json:"secretData"`
	TriggerType          string                           `json:"triggerType"`
	ExecInEnv            bool                             `json:"execInEnv"`
	Deployed             bool                             `json:"deployed"`
	DeployedOn           time.Time                        `json:"deployedOn"`
	DeployedBy           int32                            `json:"deployedBy"`
}

type PrePostStageConfigMapSecretNames struct {
	ConfigMaps []string `json:"configMaps"`
	Secrets    []string `json:"secrets"`
}

type DeploymentTemplateHistoryDto struct {
	Id                      int       `json:"id"`
	PipelineId              int       `json:"pipelineId"`
	AppId                   int       `json:"appId"`
	ImageDescriptorTemplate string    `json:"imageDescriptorTemplate,omitempty"`
	Template                string    `json:"template,omitempty"`
	TemplateName            string    `json:"templateName,omitempty"`
	TemplateVersion         string    `json:"templateVersion,omitempty"`
	IsAppMetricsEnabled     bool      `json:"isAppMetricsEnabled"`
	TargetEnvironment       int       `json:"targetEnvironment,omitempty"`
	Deployed                bool      `json:"deployed"`
	DeployedOn              time.Time `json:"deployedOn"`
	DeployedBy              int32     `json:"deployedBy"`
	EmailId                 string    `json:"emailId"`
	DeploymentStatus        string    `json:"deploymentStatus,omitempty"`
	WfrId                   int       `json:"wfrId,omitempty"`
	WorkflowType            string    `json:"workflowType,omitempty"`
}

type PipelineStrategyHistoryDto struct {
	Id         int       `json:"id"`
	PipelineId int       `json:"pipelineId"`
	Strategy   string    `json:"strategy,omitempty"`
	Config     string    `json:"config,omitempty"`
	Default    bool      `json:"default,omitempty"`
	Deployed   bool      `json:"deployed"`
	DeployedOn time.Time `json:"deployedOn"`
	DeployedBy int32     `json:"deployedBy"`
	EmailId    string    `json:"emailId"`
}

// duplicate structs below, because importing from pkg/pipeline was resulting in circular dependency
