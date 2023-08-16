package repository

import "github.com/devtron-labs/devtron/pkg/sql"

type VariableEntityMapping struct {
	tableName  struct{} `sql:"variable_entity_mapping" pg:",discard_unknown_columns"`
	Id         int      `sql:"id,pk"`
	VariableId int      `sql:"variable_id"`
	IsDeleted  bool     `sql:"is_deleted"`
	Entity
	sql.AuditLog
}

type Entity struct {
	EntityType EntityType `sql:"entity_type"`
	EntityId   string     `sql:"entity_id"`
}

type EntityType string

const (
	EntityTypeDeploymentTemplateAppLevel EntityType = "DEPLOYMENT_TEMPLATE_APP_LEVEL"
	EntityTypeDeploymentTemplateEnvLevel EntityType = "DEPLOYMENT_TEMPLATE_ENV_LEVEL"
)

type Payload struct {
	ScopedVariables []*ScopedVariables `json:"ScopedVariables"`
}
type ScopedVariables struct {
	Definition   Definition     `json:"definition"`
	ScopedValues []ScopedValues `json:"scopedValues"`
}

type ScopedValues struct {
	ScopedVariableValue
	ApplicationName string `json:"applicationName"`
	EnvironmentName string `json:"environmentName"`
	ClusterName     string `json:"clusterName"`
}

type Definition struct {
	VarName     string `json:"varName"`
	DataType    string `json:"dataType" validate:"oneof=json yaml primitive"`
	VarType     string `json:"varType" validate:"oneof=private public"`
	Description string `json:"description"`
}

type ScopedVariableValue struct {
	Value string `json:"value"`
}
