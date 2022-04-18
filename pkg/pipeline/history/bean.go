package history

import (
	"encoding/json"
	"time"
)

type ConfigMapAndSecretHistoryDto struct {
	Id         int           `json:"id"`
	PipelineId int           `json:"pipelineId"`
	AppId      int           `json:"appId"`
	DataType   string        `json:"dataType,omitempty"`
	ConfigData []*ConfigData `json:"configData,omitempty"`
	Deployed   bool          `json:"deployed"`
	DeployedOn time.Time     `json:"deployedOn"`
	DeployedBy int32         `json:"deployedBy"`
	EmailId    string        `json:"emailId"`
}

type PrePostCdScriptHistoryDto struct {
	Id                   int                              `json:"id"`
	PipelineId           int                              `json:"pipelineId"`
	Script               string                           `json:"script"`
	Stage                string                           `json:"stage"`
	ConfigMapSecretNames PrePostStageConfigMapSecretNames `json:"configmapSecretNames"`
	ConfigMapData        []*ConfigData                    `json:"configmapData"`
	SecretData           []*ConfigData                    `json:"secretData"`
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

type ConfigList struct {
	ConfigData []*ConfigData `json:"maps"`
}

type SecretList struct {
	ConfigData []*ConfigData `json:"secrets"`
}

type ConfigData struct {
	Name                  string           `json:"name"`
	Type                  string           `json:"type"`
	External              bool             `json:"external"`
	MountPath             string           `json:"mountPath,omitempty"`
	Data                  json.RawMessage  `json:"data"`
	DefaultData           json.RawMessage  `json:"defaultData,omitempty"`
	DefaultMountPath      string           `json:"defaultMountPath,omitempty"`
	Global                bool             `json:"global"`
	ExternalSecretType    string           `json:"externalType"`
	ExternalSecret        []ExternalSecret `json:"secretData"`
	DefaultExternalSecret []ExternalSecret `json:"defaultSecretData,omitempty"`
	RoleARN               string           `json:"roleARN"`
	SubPath               bool             `json:"subPath"`
	FilePermission        string           `json:"filePermission"`
}

type ExternalSecret struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Property string `json:"property,omitempty"`
	IsBinary bool   `json:"isBinary"`
}
