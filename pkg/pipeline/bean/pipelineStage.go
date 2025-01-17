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

package bean

import (
	commonBean "github.com/devtron-labs/common-lib/workflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
)

type PipelineStageDto struct {
	Id          int                          `json:"id"`
	Name        string                       `json:"name,omitempty"`
	Description string                       `json:"description,omitempty"`
	Type        repository.PipelineStageType `json:"type,omitempty" validate:"omitempty,oneof=PRE_CI POST_CI PRE_CD POST_CD"`
	Steps       []*PipelineStageStepDto      `json:"steps"`
	TriggerType pipelineConfig.TriggerType   `json:"triggerType,omitempty"`
}

type PipelineStageStepDto struct {
	Id                       int                         `json:"id"`
	Name                     string                      `json:"name" validate:"required"`
	Description              string                      `json:"description"`
	Index                    int                         `json:"index"`
	StepType                 repository.PipelineStepType `json:"stepType" validate:"omitempty,oneof=INLINE REF_PLUGIN"`
	OutputDirectoryPath      []string                    `json:"outputDirectoryPath"`
	InlineStepDetail         *InlineStepDetailDto        `json:"inlineStepDetail" validate:"omitempty,dive"`
	RefPluginStepDetail      *RefPluginStepDetailDto     `json:"pluginRefStepDetail" validate:"omitempty,dive"`
	TriggerIfParentStageFail bool                        `json:"triggerIfParentStageFail"`
}

type InlineStepDetailDto struct {
	ScriptType               repository2.ScriptType                `json:"scriptType" validate:"omitempty,oneof=SHELL DOCKERFILE CONTAINER_IMAGE"`
	Script                   string                                `json:"script"`
	StoreScriptAt            string                                `json:"storeScriptAt"`
	DockerfileExists         bool                                  `json:"dockerfileExists,omitempty"`
	MountPath                string                                `json:"mountPath,omitempty"`
	MountCodeToContainer     bool                                  `json:"mountCodeToContainer,omitempty"`
	MountCodeToContainerPath string                                `json:"mountCodeToContainerPath,omitempty"`
	MountDirectoryFromHost   bool                                  `json:"mountDirectoryFromHost"`
	ContainerImagePath       string                                `json:"containerImagePath,omitempty"`
	ImagePullSecretType      repository2.ScriptImagePullSecretType `json:"imagePullSecretType,omitempty" validate:"omitempty,oneof=CONTAINER_REGISTRY SECRET_PATH"`
	ImagePullSecret          string                                `json:"imagePullSecret,omitempty"`
	MountPathMap             []*MountPathMap                       `json:"mountPathMap,omitempty" validate:"omitempty,dive"`
	CommandArgsMap           []*CommandArgsMap                     `json:"commandArgsMap,omitempty" validate:"omitempty,dive"`
	PortMap                  []*PortMap                            `json:"portMap,omitempty" validate:"omitempty,dive"`
	InputVariables           []*StepVariableDto                    `json:"inputVariables" validate:"dive"`
	OutputVariables          []*StepVariableDto                    `json:"outputVariables" validate:"dive"`
	ConditionDetails         []*ConditionDetailDto                 `json:"conditionDetails" validate:"dive"`
}

type RefPluginStepDetailDto struct {
	PluginId         int                   `json:"pluginId"`
	InputVariables   []*StepVariableDto    `json:"inputVariables"`
	OutputVariables  []*StepVariableDto    `json:"outputVariables"`
	ConditionDetails []*ConditionDetailDto `json:"conditionDetails"`
}

// StepVariableDto is used to define the input/output variables for a step
// TODO: duplicate definition found - bean.PluginVariableDto.
// Have multiple conflicting fields with bean.PluginVariableDto.
type StepVariableDto struct {
	Id              int                                            `json:"id"`
	Name            string                                         `json:"name" validate:"required"`
	Format          repository.PipelineStageStepVariableFormatType `json:"format" validate:"oneof=STRING NUMBER BOOL DATE FILE"`
	Description     string                                         `json:"description"`
	AllowEmptyValue bool                                           `json:"allowEmptyValue,omitempty"`
	DefaultValue    string                                         `json:"defaultValue,omitempty"`
	Value           string                                         `json:"value"`
	// ValueType – Ideally it should have json tag `valueType` instead of `variableType`
	ValueType                 repository.PipelineStageStepVariableValueType `json:"variableType,omitempty" validate:"oneof=NEW FROM_PREVIOUS_STEP GLOBAL"`
	PreviousStepIndex         int                                           `json:"refVariableStepIndex,omitempty"`
	ReferenceVariableName     string                                        `json:"refVariableName,omitempty"`
	VariableStepIndexInPlugin int                                           `json:"variableStepIndexInPlugin,omitempty"`
	ReferenceVariableStage    repository.PipelineStageType                  `json:"refVariableStage,omitempty" validate:"omitempty,oneof=PRE_CI POST_CI PRE_CD POST_CD"`
	StepVariableEntDto
}

type StepVariableEntDto struct {
}

func (s *StepVariableDto) GetValue() string {
	if s == nil {
		return ""
	} else if len(s.Value) != 0 {
		return s.Value
	} else {
		return s.DefaultValue
	}
}

func (s *StepVariableDto) IsEmptyValue() bool {
	if s == nil {
		return true
	}
	// If the variable is global, then the value is empty, but referenceVariableName should not be empty
	if s.ValueType.IsGlobalDefinedValue() {
		return len(s.ReferenceVariableName) == 0
	} else if s.ValueType.IsPreviousOutputDefinedValue() {
		return len(s.ReferenceVariableName) == 0 || s.PreviousStepIndex == 0
	}
	return len(s.GetValue()) == 0
}

func (s *StepVariableDto) IsEmptyValueAllowed(isTriggerStage bool) bool {
	if s == nil {
		return false
	}
	// If the variable is not exposed as runtime arg OR if it is a trigger stage,
	// then empty value refers to StepVariableDto.AllowEmptyValue
	return s.AllowEmptyValue
}

type ConditionDetailDto struct {
	Id                  int                                       `json:"id"`
	ConditionOnVariable string                                    `json:"conditionOnVariable"` //name of variable on which condition is written
	ConditionType       repository.PipelineStageStepConditionType `json:"conditionType" validate:"oneof=SKIP TRIGGER FAIL PASS"`
	ConditionalOperator string                                    `json:"conditionOperator"`
	ConditionalValue    string                                    `json:"conditionalValue"`
}

type MountPathMap struct {
	FilePathOnDisk      string `json:"filePathOnDisk"`
	FilePathOnContainer string `json:"filePathOnContainer"`
}

type CommandArgsMap struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type PortMap struct {
	PortOnLocal     int `json:"portOnLocal" validate:"number,gt=0"`
	PortOnContainer int `json:"portOnContainer" validate:"number,gt=0"`
}

const (
	VULNERABILITY_SCANNING_PLUGIN string = "Vulnerability Scanning"

	NotTriggered       string = "Not Triggered"
	NotDeployed               = "Not Deployed"
	WorkflowTypeDeploy        = "DEPLOY"
	WorkflowTypePre           = "PRE"
	WorkflowTypePost          = "POST"
)

// BuildPrePostStepDataRequest is a request object for func BuildPrePostAndRefPluginStepsDataForWfRequest
type BuildPrePostStepDataRequest struct {
	PipelineId int
	StageType  string
	Scope      resourceQualifiers.Scope
}

// NewBuildPrePostStepDataReq creates a new BuildPrePostStepDataRequest object
func NewBuildPrePostStepDataReq(pipelineId int, stageType string, scope resourceQualifiers.Scope) *BuildPrePostStepDataRequest {
	return &BuildPrePostStepDataRequest{
		PipelineId: pipelineId,
		StageType:  stageType,
		Scope:      scope,
	}
}

type VariableAndConditionDataForStep struct {
	inputVariables           []*commonBean.VariableObject
	outputVariables          []*commonBean.VariableObject
	triggerSkipConditions    []*ConditionObject
	successFailureConditions []*ConditionObject
}

func NewVariableAndConditionDataForStep() *VariableAndConditionDataForStep {
	return &VariableAndConditionDataForStep{}
}

func (v *VariableAndConditionDataForStep) AddInputVariable(variable *commonBean.VariableObject) *VariableAndConditionDataForStep {
	v.inputVariables = append(v.inputVariables, variable)
	return v
}

func (v *VariableAndConditionDataForStep) AddOutputVariable(variable *commonBean.VariableObject) *VariableAndConditionDataForStep {
	v.outputVariables = append(v.outputVariables, variable)
	return v
}

func (v *VariableAndConditionDataForStep) AddTriggerSkipCondition(condition *ConditionObject) *VariableAndConditionDataForStep {
	v.triggerSkipConditions = append(v.triggerSkipConditions, condition)
	return v
}

func (v *VariableAndConditionDataForStep) AddSuccessFailureCondition(condition *ConditionObject) *VariableAndConditionDataForStep {
	v.successFailureConditions = append(v.successFailureConditions, condition)
	return v
}

func (v *VariableAndConditionDataForStep) GetInputVariables() []*commonBean.VariableObject {
	return v.inputVariables
}

func (v *VariableAndConditionDataForStep) GetOutputVariables() []*commonBean.VariableObject {
	return v.outputVariables
}

func (v *VariableAndConditionDataForStep) GetTriggerSkipConditions() []*ConditionObject {
	return v.triggerSkipConditions
}

func (v *VariableAndConditionDataForStep) GetSuccessFailureConditions() []*ConditionObject {
	return v.successFailureConditions
}
