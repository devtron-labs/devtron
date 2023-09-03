package models

type ScopedVariableData struct {
	VariableName  string        `json:"variableName"`
	VariableValue VariableValue `json:"variableValue,omitempty"`
}

const (
	YAML_TYPE = "yaml"
	JSON_TYPE = "json"
)

type VariableScopeMapping struct {
	ScopeId int
}
