package models

type ScopedVariableData struct {
	VariableName  string        `json:"variableName"`
	Description   string        `json:"description"`
	VariableValue VariableValue `json:"variableValue,omitempty"`
}

const (
	YAML_TYPE = "yaml"
	JSON_TYPE = "json"
)

type VariableScopeMapping struct {
	ScopeId int
}
