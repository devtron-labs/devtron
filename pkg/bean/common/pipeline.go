package common

import (
	commonBean "github.com/devtron-labs/common-lib/workflow"
)

// RuntimeParameters holds values that needed to be injected/used in ci build process.
type RuntimeParameters struct {
	RuntimePluginVariables []*RuntimePluginVariableDto `json:"runtimePluginVariables,omitempty" validate:"dive"`
}

func (r *RuntimeParameters) AddSystemVariables(name, value string) *RuntimeParameters {
	r.RuntimePluginVariables = append(r.RuntimePluginVariables, NewRuntimeSystemVariableDto(name, value))
	return r
}

func (r *RuntimeParameters) GetSystemVariables() map[string]string {
	response := make(map[string]string)
	for _, variable := range r.RuntimePluginVariables {
		if variable.IsSystemVariableScope() {
			response[variable.Name] = variable.Value
		}
	}
	return response
}

// RuntimePluginVariableDto is used to define the runtime plugin variables.
type RuntimePluginVariableDto struct {
	Name              string            `json:"name" validate:"required"`
	Value             string            `json:"value"`
	Format            commonBean.Format `json:"format" validate:"required"`
	VariableStepScope VariableStepScope `json:"variableStepScope" validate:"oneof=GLOBAL PIPELINE_STAGE"`
}

// NewRuntimeParameters returns a new instance of RuntimeParameters.
func NewRuntimeParameters() *RuntimeParameters {
	return &RuntimeParameters{}
}

// NewRuntimeSystemVariableDto returns a new instance of RuntimePluginVariableDto with system variable.
func NewRuntimeSystemVariableDto(name, value string) *RuntimePluginVariableDto {
	return NewRuntimePluginVariableDto(name, value, commonBean.FormatTypeString, SystemVariableScope)
}

// NewRuntimePluginVariableDto returns a new instance of RuntimePluginVariableDto.
func NewRuntimePluginVariableDto(name, value string, format commonBean.Format, variableStepScope VariableStepScope) *RuntimePluginVariableDto {
	return &RuntimePluginVariableDto{
		Name:              name,
		Value:             value,
		Format:            format,
		VariableStepScope: variableStepScope,
	}
}

// IsSystemVariableScope returns true if the variable is of SYSTEM variable type.
// If the variable is nil, it returns false.
func (r *RuntimePluginVariableDto) IsSystemVariableScope() bool {
	if r == nil {
		return false
	}
	return r.VariableStepScope == SystemVariableScope
}

// VariableStepScope is used to define the scope of the runtime plugin variable.
type VariableStepScope string

const (
	// SystemVariableScope is used to define the global variable scope.
	SystemVariableScope VariableStepScope = "SYSTEM"
)
