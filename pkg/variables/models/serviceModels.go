/*
 * Copyright (c) 2024. Devtron Inc.
 */

package models

import (
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
)

type ScopedVariableData struct {
	VariableName     string         `json:"variableName"`
	ShortDescription string         `json:"shortDescription"`
	VariableValue    *VariableValue `json:"variableValue,omitempty"`
	IsRedacted       bool           `json:"isRedacted"`
}

type VariableScopeMapping struct {
	ScopeId int
}

type VariableScope struct {
	id int
	*resourceQualifiers.ResourceMappingSelection
	Data string
}
