/*
 * Copyright (c) 2026. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

import (
	"testing"

	"github.com/devtron-labs/devtron/pkg/plugin/repository"
)

func TestBuildPluginStepsByPluginId(t *testing.T) {
	pluginSteps := []*repository.PluginStep{
		{Id: 1, PluginId: 10, Name: "inline", Index: 1, StepType: repository.PLUGIN_STEP_TYPE_INLINE, ScriptId: 100},
		{Id: 2, PluginId: 11, Name: "reference", Index: 1, StepType: repository.PLUGIN_STEP_TYPE_REF_PLUGIN, RefPluginId: 20},
	}
	pluginStepVariables := []*repository.PluginStepVariable{
		{Id: 101, PluginStepId: 1, Name: "input", VariableType: repository.PLUGIN_VARIABLE_TYPE_INPUT},
		{Id: 102, PluginStepId: 1, Name: "output", VariableType: repository.PLUGIN_VARIABLE_TYPE_OUTPUT},
	}
	pluginStepConditions := []*repository.PluginStepCondition{
		{Id: 201, PluginStepId: 1, ConditionVariableId: 102, ConditionalOperator: "==", ConditionalValue: "success"},
	}
	pluginScripts := []*repository.PluginPipelineScript{
		{Id: 100, Script: "echo hello", Type: repository.SCRIPT_TYPE_SHELL},
	}
	scriptMappings := []*repository.ScriptPathArgPortMapping{
		{Id: 301, ScriptId: 100, TypeOfMapping: repository.SCRIPT_MAPPING_TYPE_PORT, PortOnLocal: 8080, PortOnContainer: 80},
	}

	stepsByPluginId := buildPluginStepsByPluginId(pluginSteps, pluginStepVariables, pluginStepConditions, pluginScripts, scriptMappings)

	customPluginSteps := stepsByPluginId[10]
	if len(customPluginSteps) != 1 {
		t.Fatalf("expected one step for plugin 10, got %d", len(customPluginSteps))
	}
	step := customPluginSteps[0]
	if step.Name != "inline" || step.PluginPipelineScript == nil || step.PluginPipelineScript.Script != "echo hello" {
		t.Fatalf("unexpected hydrated step: %#v", step)
	}
	if len(step.PluginStepVariable) != 2 {
		t.Fatalf("expected two variables, got %d", len(step.PluginStepVariable))
	}
	if len(step.PluginStepVariable[0].PluginStepCondition) != 0 {
		t.Fatal("did not expect conditions on the input variable")
	}
	if len(step.PluginStepVariable[1].PluginStepCondition) != 1 || step.PluginStepVariable[1].PluginStepCondition[0].Id != 201 {
		t.Fatalf("expected the output condition to be preserved: %#v", step.PluginStepVariable[1].PluginStepCondition)
	}
	if len(step.PluginPipelineScript.PathArgPortMapping) != 1 || step.PluginPipelineScript.PathArgPortMapping[0].Id != 301 {
		t.Fatalf("expected the script mapping to be preserved: %#v", step.PluginPipelineScript.PathArgPortMapping)
	}

	referencePluginSteps := stepsByPluginId[11]
	if len(referencePluginSteps) != 1 || referencePluginSteps[0].PluginPipelineScript != nil {
		t.Fatalf("expected a reference step without an inline script: %#v", referencePluginSteps)
	}
}

func TestBuildPluginStepsByPluginIdEmpty(t *testing.T) {
	stepsByPluginId := buildPluginStepsByPluginId(nil, nil, nil, nil, nil)
	if len(stepsByPluginId) != 0 {
		t.Fatalf("expected no plugin steps, got %#v", stepsByPluginId)
	}
}
