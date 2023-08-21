package repository

type Payload struct {
	Variables []*Variables `json:"variables"`
	UserId    int32        `json:"-"`
}
type Variables struct {
	Definition      Definition       `json:"definition"`
	AttributeValues []AttributeValue `json:"attributeValue"`
}
type AttributeValue struct {
	VariableValue   VariableValue             `json:"variableValue"`
	AttributeType   AttributeType             `json:"attributeType"`
	AttributeParams map[IdentifierType]string `json:"attributeParams"`
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
	Cluster        AttributeType = "Cluster"
	Global         AttributeType = "Global"
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
