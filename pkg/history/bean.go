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
	DeployedOn time.Time     `json:"deployed_on"`
	DeployedBy int32         `json:"deployed_by"`
	sql.AuditLog
}

// duplicate structs below, because importing from pkg/pipeline was resulting in circular dependency

type ConfigsList struct {
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
