package models

type ScopedVariableData struct {
	VariableName     string         `json:"variableName"`
	ShortDescription string         `json:"shortDescription"`
	VariableValue    *VariableValue `json:"variableValue,omitempty"`
	IsRedacted       bool           `json:"isRedacted"`
}

type VariableScopeMapping struct {
	ScopeId int
}
