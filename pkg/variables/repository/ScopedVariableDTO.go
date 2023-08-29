package repository

type Payload struct {
	SpecVersion string       `json:"specVersion"`
	Variables   []*Variables `json:"variables" validate:"required,dive"`
	UserId      int32        `json:"-"`
}
type Variables struct {
	Definition      Definition       `json:"definition" validate:"required,dive"`
	AttributeValues []AttributeValue `json:"attributeValue" validate:"dive" `
}
type AttributeValue struct {
	VariableValue   VariableValue             `json:"variableValue" validate:"required,dive"`
	AttributeType   AttributeType             `json:"attributeType" validate:"oneof=ApplicationEnv Application Env Cluster Global"`
	AttributeParams map[IdentifierType]string `json:"attributeParams"`
}

type Definition struct {
	VarName     string `json:"varName" validate:"required"`
	DataType    string `json:"dataType" validate:"oneof=json yaml primitive"`
	VarType     string `json:"varType" validate:"oneof=private public"`
	Description string `json:"description" validate:"max=300"`
}

type AttributeType string

const (
	ApplicationEnv AttributeType = "ApplicationEnv"
	Application    AttributeType = "Application"
	Env            AttributeType = "Env"
	Cluster        AttributeType = "Cluster"
	Global         AttributeType = "Global"
)

type IdentifierType string

const (
	ApplicationName IdentifierType = "ApplicationName"
	EnvName         IdentifierType = "EnvName"
	ClusterName     IdentifierType = "ClusterName"
)

var IdentifiersList = []IdentifierType{ApplicationName, EnvName, ClusterName}

type VariableValue struct {
	Value interface{} `json:"value" validate:"required"`
}
