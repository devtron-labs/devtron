package common

import (
	commonBean "github.com/devtron-labs/common-lib/ci-runner/bean"
)

// RuntimeParameters holds values that needed to be injected/used in ci build process.
type RuntimeParameters struct {
	RuntimePluginVariables []*RuntimePluginVariableDto `json:"runtimePluginVariables,omitempty" validate:"dive"`
}

func (r *RuntimeParameters) GetGlobalRuntimeVariables() map[string]string {
	response := make(map[string]string)
	for _, variable := range r.RuntimePluginVariables {
		if variable.IsGlobalVariableScope() {
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

// NewRuntimeGlobalVariableDto returns a new instance of RuntimePluginVariableDto with global variable scope.
func NewRuntimeGlobalVariableDto(name, value string) *RuntimePluginVariableDto {
	return NewRuntimePluginVariableDto(name, value, commonBean.FormatTypeString, GlobalVariableScope)
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

// IsGlobalVariableScope returns true if the runtime plugin variable is of global variable scope.
// If the variable is nil, it returns false.
func (r *RuntimePluginVariableDto) IsGlobalVariableScope() bool {
	if r == nil {
		return false
	}
	return r.VariableStepScope == GlobalVariableScope
}

// VariableStepScope is used to define the scope of the runtime plugin variable.
type VariableStepScope string

const (
	// GlobalVariableScope is used to define the global variable scope.
	GlobalVariableScope VariableStepScope = "GLOBAL"
)
