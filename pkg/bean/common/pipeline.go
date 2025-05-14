/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	commonBean "github.com/devtron-labs/common-lib/workflow"
)

// RuntimeParameters holds values that needed to be injected/used in ci build process.
type RuntimeParameters struct {
	RuntimePluginVariables []*RuntimePluginVariableDto `json:"runtimePluginVariables,omitempty" validate:"dive"`
}

func (r *RuntimeParameters) AddSystemVariable(name, value string) *RuntimeParameters {
	r.RuntimePluginVariables = append(r.RuntimePluginVariables, NewRuntimeSystemVariableDto(name, value))
	return r
}

func (r *RuntimeParameters) AddRuntimeGlobalVariable(name, value string) *RuntimeParameters {
	r.RuntimePluginVariables = append(r.RuntimePluginVariables, NewRuntimeGlobalVariableDto(name, value))
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

// NewRuntimeSystemVariableDto returns a new instance of RuntimePluginVariableDto with system variable.
func NewRuntimeSystemVariableDto(name, value string) *RuntimePluginVariableDto {
	return NewRuntimePluginVariableDto(name, value, commonBean.FormatTypeString, SystemVariableScope)
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

// IsSystemVariableScope returns true if the variable is of SYSTEM variable type.
// If the variable is nil, it returns false.
func (r *RuntimePluginVariableDto) IsSystemVariableScope() bool {
	if r == nil {
		return false
	}
	return r.VariableStepScope == SystemVariableScope
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
	// SystemVariableScope is used to define the global variable scope.
	SystemVariableScope VariableStepScope = "SYSTEM"
)

type WorkflowCacheConfigType string

const (
	WorkflowCacheConfigInherit WorkflowCacheConfigType = "INHERIT"
)
