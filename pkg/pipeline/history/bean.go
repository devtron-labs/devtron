package history

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

type ConfigMapAndSecretHistoryDto struct {
	Id         int           `json:"id"`
	PipelineId int           `json:"pipelineId"`
	DataType   string        `json:"dataType"`
	ConfigData []*ConfigData `json:"configData"`
	Deployed   bool          `json:"deployed"`
	DeployedOn time.Time     `json:"deployedOn"`
	DeployedBy int32         `json:"deployedBy"`
	EmailId    string        `json:"emailId"`
	sql.AuditLog
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
	sql.AuditLog
}

type PrePostStageConfigMapSecretNames struct {
	ConfigMaps []string `json:"configMaps"`
	Secrets    []string `json:"secrets"`
}

type DeploymentTemplateHistoryDto struct {
	Id                      int       `json:"id"`
	PipelineId              int       `json:"pipelineId,"`
	ImageDescriptorTemplate string    `json:"imageDescriptorTemplate"`
	Template                string    `json:"template"`
	TemplateName            string    `json:"templateName"`
	TemplateVersion         string    `json:"templateVersion"`
	IsAppMetricsEnabled     bool      `json:"isAppMetricsEnabled"`
	TargetEnvironment       int       `json:"targetEnvironment"`
	Deployed                bool      `json:"deployed"`
	DeployedOn              time.Time `json:"deployedOn"`
	DeployedBy              int32     `json:"deployedBy"`
	EmailId                 string    `json:"emailId"`
	sql.AuditLog
}

type PipelineStrategyHistoryDto struct {
	Id         int       `json:"id"`
	PipelineId int       `json:"pipelineId"`
	Strategy   string    `json:"strategy"`
	Config     string    `json:"config"`
	Default    bool      `json:"default"`
	Deployed   bool      `json:"deployed"`
	DeployedOn time.Time `json:"deployedOn"`
	DeployedBy int32     `json:"deployedBy"`
	EmailId    string    `json:"emailId"`
	sql.AuditLog
}

// duplicate structs below, because importing from pkg/pipeline was resulting in circular dependency

type ConfigList struct {
	ConfigData []*ConfigData `json:"maps"`
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
