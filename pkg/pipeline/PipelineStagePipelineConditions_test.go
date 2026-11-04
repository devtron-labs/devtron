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
	"testing"

	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type fakePipelineStageRepoForWfRequest struct {
	repository.PipelineStageRepository
	variables  []*repository.PipelineStageStepVariable
	conditions []*repository.PipelineStageStepCondition
}

func (f *fakePipelineStageRepoForWfRequest) GetVariablesByStepId(stepId int) ([]*repository.PipelineStageStepVariable, error) {
	return f.variables, nil
}

func (f *fakePipelineStageRepoForWfRequest) GetConditionsByStepId(stepId int) ([]*repository.PipelineStageStepCondition, error) {
	return f.conditions, nil
}

// Regression guard for the plugin-condition change: the conditions saved on the pipeline step
// (pipeline_stage_step_condition) are the ones that must keep reaching ci-runner.
func TestBuildVariableAndConditionDataForWfRequest_PipelineLevelConditionsAreSent(t *testing.T) {
	repo := &fakePipelineStageRepoForWfRequest{
		variables: []*repository.PipelineStageStepVariable{
			{Id: 198, PipelineStageStepId: 225, Name: "FILE_PATH", Format: repository.PIPELINE_STAGE_STEP_VARIABLE_FORMAT_TYPE_STRING, VariableType: repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_INPUT, ValueType: repository.PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_NEW, Value: "/tmp/other", AllowEmptyValue: true, IsExposed: true, VariableStepIndexInPlugin: 1},
			{Id: 199, PipelineStageStepId: 225, Name: "CONTENT", Format: repository.PIPELINE_STAGE_STEP_VARIABLE_FORMAT_TYPE_STRING, VariableType: repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_OUTPUT, ValueType: repository.PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_NEW, IsExposed: true, VariableStepIndexInPlugin: 1},
		},
		conditions: []*repository.PipelineStageStepCondition{
			{Id: 15, PipelineStageStepId: 225, ConditionVariableId: 198, ConditionType: repository.PIPELINE_STAGE_STEP_CONDITION_TYPE_TRIGGER, ConditionalOperator: "==", ConditionalValue: "/tmp/other"},
			{Id: 16, PipelineStageStepId: 225, ConditionVariableId: 199, ConditionType: repository.PIPELINE_STAGE_STEP_CONDITION_TYPE_SUCCESS, ConditionalOperator: "==", ConditionalValue: "ok"},
			{Id: 17, PipelineStageStepId: 225, ConditionVariableId: 0, ConditionType: repository.PIPELINE_STAGE_STEP_CONDITION_TYPE_SKIP, ConditionalOperator: "!=", ConditionalValue: "x"},
		},
	}
	impl := &PipelineStageServiceImpl{logger: zap.NewNop().Sugar(), pipelineStageRepository: repo}
	data, err := impl.buildVariableAndConditionDataForWfRequest(225)
	assert.NoError(t, err)
	if assert.Len(t, data.GetInputVariables(), 1) {
		assert.Equal(t, "/tmp/other", data.GetInputVariables()[0].Value)
		assert.Equal(t, 1, data.GetInputVariables()[0].VariableStepIndexInPlugin)
	}
	assert.Len(t, data.GetOutputVariables(), 1)
	triggerSkip := data.GetTriggerSkipConditions()
	if assert.Len(t, triggerSkip, 2) {
		assert.Equal(t, "FILE_PATH", triggerSkip[0].ConditionOnVariable)
		assert.Equal(t, "TRIGGER", triggerSkip[0].ConditionType)
		assert.Equal(t, "==", triggerSkip[0].ConditionalOperator)
		assert.Equal(t, "/tmp/other", triggerSkip[0].ConditionalValue)
		assert.Equal(t, "", triggerSkip[1].ConditionOnVariable, "condition on an unknown variable id is sent without a variable name (pre-existing behaviour)")
	}
	if assert.Len(t, data.GetSuccessFailureConditions(), 1) {
		assert.Equal(t, "CONTENT", data.GetSuccessFailureConditions()[0].ConditionOnVariable)
		assert.Equal(t, "PASS", data.GetSuccessFailureConditions()[0].ConditionType)
	}
}
