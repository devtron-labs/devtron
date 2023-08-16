package repository

import "github.com/devtron-labs/devtron/pkg/sql"

type VariableEntityMapping struct {
	tableName    struct{} `sql:"variable_entity_mapping" pg:",discard_unknown_columns"`
	Id           int      `sql:"id,pk"`
	VariableName string   `sql:"variable_name"`
	IsDeleted    bool     `sql:"is_deleted"`
	Entity
	sql.AuditLog
}

type Entity struct {
	EntityType EntityType `sql:"entity_type"`
	EntityId   int        `sql:"entity_id"`
}

type EntityType string

const (
	EntityTypeDeploymentTemplateAppLevel EntityType = "DEPLOYMENT_TEMPLATE_APP_LEVEL"
	EntityTypeDeploymentTemplateEnvLevel EntityType = "DEPLOYMENT_TEMPLATE_ENV_LEVEL"
)

//type Payload struct {
//	Variables []*Variables `json:"Variables"`
//}
//type Variables struct {
//	Definition      Definition       `json:"definition"`
//	AttributeValues []AttributeValue `json:"attributeValue"`
//}
//
//// Based on provided values, system will deduce the type and store the // value mapping
//type AttributeValue struct {
//	VariableValue
//	AttributeType   AttributeType
//	AttributeParams []VariableValue
//}
//
//type Definition struct {
//	VarName     string `json:"varName"`
//	DataType    string `json:"dataType" validate:"oneof=json yaml primitive"`
//	VarType     string `json:"varType" validate:"oneof=private public"`
//	Description string `json:"description"`
//}
//
//type AttributeType string
//
//const (
//	ApplicationEnv AttributeType = "ApplicationEnv"
//	Application    AttributeType = "Application"
//	Env            AttributeType = "Env"
//)
//
//type VariableValue struct {
//	Value string `json:"value"`
//}

type Payload struct {
	Variables []*Variables `json:"Variables"`
}
type Variables struct {
	Definition      Definition       `json:"definition"`
	AttributeValues []AttributeValue `json:"attributeValue"`
}

type AttributeValue struct {
	VariableValue
	AttributeType   AttributeType
	AttributeParams map[IdentifierType]string
}

type Definition struct {
	VarName     string `json:"varName"`
	DataType    string `json:"dataType" validate:"oneof=json yaml primitive"`
	VarType     string `json:"varType" validate:"oneof=private public"`
	Description string `json:"description"`
}

type AttributeType string

const (
	ApplicationEnv AttributeType = "ApplicationEnv"
	Application    AttributeType = "Application"
	Env            AttributeType = "Env"
)

type IdentifierType string

const (
	ApplicationName IdentifierType = "ApplicationName"
	EnvName         IdentifierType = "EnvName"
	ClusterName     IdentifierType = "ClusterName"
)

type VariableValue struct {
	Value string `json:"value"`
}
