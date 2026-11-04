/*
 * Copyright (c) 2024. Devtron Inc.
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

package pipeline

import (
	"errors"
	"testing"

	commonBean "github.com/devtron-labs/common-lib/workflow"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// fakePluginRepoForWfRequest serves a plugin definition (steps, script, variables, conditions) to the workflow request
// builders. Any repository method not overridden here panics if called, which is intended.
type fakePluginRepoForWfRequest struct {
	repository2.GlobalPluginRepository
	steps          []*repository2.PluginStep
	scripts        map[int]*repository2.PluginPipelineScript
	variables      map[int][]*repository2.PluginStepVariable
	variablesErr   error
	conditions     map[int][]*repository2.PluginStepCondition
	conditionCalls int
}

func (f *fakePluginRepoForWfRequest) GetStepsByPluginIds(pluginIds []int) ([]*repository2.PluginStep, error) {
	return f.steps, nil
}

func (f *fakePluginRepoForWfRequest) GetScriptDetailById(id int) (*repository2.PluginPipelineScript, error) {
	if script, ok := f.scripts[id]; ok {
		return script, nil
	}
	return nil, pg.ErrNoRows
}

func (f *fakePluginRepoForWfRequest) GetScriptMappingDetailByScriptId(scriptId int) ([]*repository2.ScriptPathArgPortMapping, error) {
	return nil, nil
}

func (f *fakePluginRepoForWfRequest) GetVariablesByStepId(stepId int) ([]*repository2.PluginStepVariable, error) {
	if f.variablesErr != nil {
		return nil, f.variablesErr
	}
	return f.variables[stepId], nil
}

func (f *fakePluginRepoForWfRequest) GetConditionsByStepId(stepId int) ([]*repository2.PluginStepCondition, error) {
	f.conditionCalls++
	return f.conditions[stepId], nil
}

func newPluginConditionsTestService(repo *fakePluginRepoForWfRequest) *PipelineStageServiceImpl {
	return &PipelineStageServiceImpl{logger: zap.NewNop().Sugar(), globalPluginRepository: repo}
}

func pluginStepVariable(name string, variableType repository2.PluginStepVariableType, exposed bool, value, defaultValue string) *repository2.PluginStepVariable {
	return &repository2.PluginStepVariable{
		Id:                1,
		Name:              name,
		Format:            repository2.PLUGIN_VARIABLE_FORMAT_TYPE_STRING,
		VariableType:      variableType,
		ValueType:         repository2.PLUGIN_VARIABLE_VALUE_TYPE_NEW,
		IsExposed:         exposed,
		Value:             value,
		DefaultValue:      defaultValue,
		VariableStepIndex: 1,
	}
}

// pluginWithConditions models a custom plugin saved from the dashboard: one shell step, an exposed input variable
// FILE_PATH with a TRIGGER condition on it and an exposed output variable CONTENT with a PASS condition on it.
func pluginWithConditions() *fakePluginRepoForWfRequest {
	input := pluginStepVariable("FILE_PATH", repository2.PLUGIN_VARIABLE_TYPE_INPUT, true, "/tmp/shivam", "")
	input.Id = 386
	output := pluginStepVariable("CONTENT", repository2.PLUGIN_VARIABLE_TYPE_OUTPUT, true, "", "")
	output.Id = 387
	return &fakePluginRepoForWfRequest{
		steps: []*repository2.PluginStep{
			{Id: 69, PluginId: 71, Name: "Task 1", Index: 1, StepType: repository2.PLUGIN_STEP_TYPE_INLINE, ScriptId: 5},
		},
		scripts:   map[int]*repository2.PluginPipelineScript{5: {Type: repository2.SCRIPT_TYPE_SHELL, Script: "echo hi > $FILE_PATH"}},
		variables: map[int][]*repository2.PluginStepVariable{69: {input, output}},
		conditions: map[int][]*repository2.PluginStepCondition{69: {
			{Id: 23, PluginStepId: 69, ConditionVariableId: 386, ConditionType: repository2.PLUGIN_CONDITION_TYPE_TRIGGER, ConditionalOperator: "==", ConditionalValue: "/tmp/shivam"},
			{Id: 24, PluginStepId: 69, ConditionVariableId: 387, ConditionType: repository2.PLUGIN_CONDITION_TYPE_SUCCESS, ConditionalOperator: "==", ConditionalValue: "ok"},
		}},
	}
}

func TestBuildPluginStepDataForWfRequest_PluginConditionsAreNotSentToCiRunner(t *testing.T) {
	t.Run("inline plugin step: variables and script are sent, plugin-defined conditions are not", func(t *testing.T) {
		repo := pluginWithConditions()
		stepData, err := newPluginConditionsTestService(repo).BuildPluginStepDataForWfRequest(repo.steps[0])
		assert.NoError(t, err)
		assert.Equal(t, "Task 1", stepData.Name)
		assert.Equal(t, 1, stepData.Index)
		assert.Equal(t, string(repository2.PLUGIN_STEP_TYPE_INLINE), stepData.StepType)
		assert.Equal(t, string(repository2.SCRIPT_TYPE_SHELL), stepData.ExecutorType)
		assert.Equal(t, "echo hi > $FILE_PATH", stepData.Script)
		if assert.Len(t, stepData.InputVars, 1) {
			assert.Equal(t, "FILE_PATH", stepData.InputVars[0].Name)
			assert.Equal(t, "/tmp/shivam", stepData.InputVars[0].Value)
		}
		if assert.Len(t, stepData.OutputVars, 1) {
			assert.Equal(t, "CONTENT", stepData.OutputVars[0].Name)
		}
		assert.Nil(t, stepData.TriggerSkipConditions, "plugin-level trigger/skip conditions must not reach ci-runner")
		assert.Nil(t, stepData.SuccessFailureConditions, "plugin-level pass/fail conditions must not reach ci-runner")
		assert.Equal(t, 0, repo.conditionCalls, "plugin_step_condition must not even be read for the workflow request")
	})
	t.Run("plugin step referencing another plugin: no conditions either", func(t *testing.T) {
		repo := pluginWithConditions()
		nested := &repository2.PluginStep{Id: 69, PluginId: 71, Name: "nested", Index: 1, StepType: repository2.PLUGIN_STEP_TYPE_REF_PLUGIN, RefPluginId: 42}
		stepData, err := newPluginConditionsTestService(repo).BuildPluginStepDataForWfRequest(nested)
		assert.NoError(t, err)
		assert.Equal(t, "PLUGIN", stepData.ExecutorType)
		assert.Equal(t, 42, stepData.RefPluginId)
		assert.Len(t, stepData.InputVars, 1)
		assert.Nil(t, stepData.TriggerSkipConditions)
		assert.Nil(t, stepData.SuccessFailureConditions)
		assert.Equal(t, 0, repo.conditionCalls)
	})
	t.Run("variable lookup error is propagated", func(t *testing.T) {
		repo := pluginWithConditions()
		repo.variablesErr = errors.New("db down")
		stepData, err := newPluginConditionsTestService(repo).BuildPluginStepDataForWfRequest(repo.steps[0])
		assert.EqualError(t, err, "db down")
		assert.Nil(t, stepData)
	})
}

func TestGetRefPluginStepsByIds_NoPluginConditionOnAnyStep(t *testing.T) {
	repo := pluginWithConditions()
	pluginIdStepsMap, err := newPluginConditionsTestService(repo).GetRefPluginStepsByIds([]int{71}, map[int]bool{71: true})
	assert.NoError(t, err)
	if assert.Len(t, pluginIdStepsMap[71], 1) {
		assert.Nil(t, pluginIdStepsMap[71][0].TriggerSkipConditions)
		assert.Nil(t, pluginIdStepsMap[71][0].SuccessFailureConditions)
		assert.Len(t, pluginIdStepsMap[71][0].InputVars, 1)
	}
	assert.Equal(t, 0, repo.conditionCalls)
}

func TestBuildPluginVariableDataForWfRequest(t *testing.T) {
	newRepo := func(vars ...*repository2.PluginStepVariable) *fakePluginRepoForWfRequest {
		return &fakePluginRepoForWfRequest{variables: map[int][]*repository2.PluginStepVariable{69: vars}}
	}
	t.Run("exposed input without default uses the stored value, non-exposed input uses its default", func(t *testing.T) {
		exposed := pluginStepVariable("A", repository2.PLUGIN_VARIABLE_TYPE_INPUT, true, "user", "")
		exposed.VariableStepIndexInPlugin = 3 // only meaningful for plugins nested in plugins; must be passed through untouched
		internal := pluginStepVariable("B", repository2.PLUGIN_VARIABLE_TYPE_INPUT, false, "ignored", "default")
		exposedWithDefault := pluginStepVariable("C", repository2.PLUGIN_VARIABLE_TYPE_INPUT, true, "", "default")
		inputs, outputs, err := newPluginConditionsTestService(newRepo(exposed, internal, exposedWithDefault)).BuildPluginVariableDataForWfRequest(69)
		assert.NoError(t, err)
		assert.Empty(t, outputs)
		if assert.Len(t, inputs, 3) {
			assert.Equal(t, "user", inputs[0].Value)
			assert.Equal(t, "default", inputs[1].Value)
			assert.Equal(t, "", inputs[2].Value, "exposed variable value is filled from the pipeline step by ci-runner")
			assert.Equal(t, commonBean.VariableTypeValue, inputs[0].VariableType)
			assert.Equal(t, commonBean.FormatTypeString, inputs[0].Format)
			assert.Equal(t, 3, inputs[0].VariableStepIndexInPlugin, "nested-plugin step index is passed through")
			assert.Equal(t, 0, inputs[1].VariableStepIndexInPlugin)
		}
	})
	t.Run("value types map to ci-runner variable types", func(t *testing.T) {
		global := pluginStepVariable("G", repository2.PLUGIN_VARIABLE_TYPE_INPUT, false, "", "")
		global.ValueType = repository2.PLUGIN_VARIABLE_VALUE_TYPE_GLOBAL
		global.ReferenceVariableName = "DOCKER_IMAGE"
		previous := pluginStepVariable("P", repository2.PLUGIN_VARIABLE_TYPE_INPUT, false, "", "")
		previous.ValueType = repository2.PLUGIN_VARIABLE_VALUE_TYPE_PREVIOUS
		previous.PreviousStepIndex = 1
		previous.ReferenceVariableName = "OUT"
		exposedPrevious := pluginStepVariable("E", repository2.PLUGIN_VARIABLE_TYPE_INPUT, true, "", "")
		exposedPrevious.ValueType = repository2.PLUGIN_VARIABLE_VALUE_TYPE_PREVIOUS
		inputs, _, err := newPluginConditionsTestService(newRepo(global, previous, exposedPrevious)).BuildPluginVariableDataForWfRequest(69)
		assert.NoError(t, err)
		if assert.Len(t, inputs, 3) {
			assert.Equal(t, commonBean.VariableTypeRefGlobal, inputs[0].VariableType)
			assert.Equal(t, "DOCKER_IMAGE", inputs[0].ReferenceVariableName)
			assert.Equal(t, commonBean.VariableTypeRefPlugin, inputs[1].VariableType)
			assert.Equal(t, 1, inputs[1].ReferenceVariableStepIndex)
			assert.Equal(t, commonBean.VariableTypeValue, inputs[2].VariableType, "exposed previous-step variables are resolved at pipeline level")
		}
	})
	t.Run("output variables are returned separately", func(t *testing.T) {
		out := pluginStepVariable("OUT", repository2.PLUGIN_VARIABLE_TYPE_OUTPUT, true, "", "")
		inputs, outputs, err := newPluginConditionsTestService(newRepo(out)).BuildPluginVariableDataForWfRequest(69)
		assert.NoError(t, err)
		assert.Empty(t, inputs)
		if assert.Len(t, outputs, 1) {
			assert.Equal(t, "OUT", outputs[0].Name)
		}
	})
	t.Run("no variables (pg.ErrNoRows) yields empty lists", func(t *testing.T) {
		repo := &fakePluginRepoForWfRequest{variablesErr: pg.ErrNoRows}
		inputs, outputs, err := newPluginConditionsTestService(repo).BuildPluginVariableDataForWfRequest(69)
		assert.NoError(t, err)
		assert.Empty(t, inputs)
		assert.Empty(t, outputs)
	})
	t.Run("repository error is propagated", func(t *testing.T) {
		repo := &fakePluginRepoForWfRequest{variablesErr: errors.New("db down")}
		_, _, err := newPluginConditionsTestService(repo).BuildPluginVariableDataForWfRequest(69)
		assert.EqualError(t, err, "db down")
	})
	t.Run("invalid variable format is an error", func(t *testing.T) {
		bad := pluginStepVariable("X", repository2.PLUGIN_VARIABLE_TYPE_INPUT, true, "", "")
		bad.Format = "BOGUS"
		_, _, err := newPluginConditionsTestService(newRepo(bad)).BuildPluginVariableDataForWfRequest(69)
		assert.Error(t, err)
	})
}
