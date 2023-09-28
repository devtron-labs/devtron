package plugin

import (
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
)

type PluginDetailDto struct {
	Metadata        *PluginMetadataDto   `json:"metadata"`
	InputVariables  []*PluginVariableDto `json:"inputVariables"`
	OutputVariables []*PluginVariableDto `json:"outputVariables"`
}

type PluginListComponentDto struct { //created new struct for backward compatibility (needed to add input and output Vars along with metadata fields)
	*PluginMetadataDto
	InputVariables  []*PluginVariableDto `json:"inputVariables"`
	OutputVariables []*PluginVariableDto `json:"outputVariables"`
}

type PluginMetadataDto struct {
	Id          int               `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"` // SHARED, PRESET etc
	Icon        string            `json:"icon"`
	Tags        []string          `json:"tags"`
	Action      int               `json:"action"`
	PluginStage string            `json:"pluginStage,omitempty"`
	PluginSteps []*PluginStepsDto `json:"pluginSteps,omitempty"`
}

type PluginStepsDto struct {
	Id                   int                              `json:"id,pk"`
	Name                 string                           `json:"name"`
	Description          string                           `json:"description"`
	Index                int                              `json:"index"`
	StepType             repository.PluginStepType        `json:"stepType"`
	RefPluginId          int                              `json:"refPluginId"` //id of plugin used as reference
	OutputDirectoryPath  []string                         `json:"outputDirectoryPath"`
	DependentOnStep      string                           `json:"dependentOnStep"`
	PluginStepVariable   []*PluginVariableDto             `json:"pluginStepVariable,omitempty"`
	PluginPipelineScript *repository.PluginPipelineScript `json:"pluginPipelineScript,omitempty"`
}

type PluginVariableDto struct {
	Id                        int                                     `json:"id,omitempty"`
	Name                      string                                  `json:"name"`
	Format                    repository.PluginStepVariableFormatType `json:"format"`
	Description               string                                  `json:"description"`
	IsExposed                 bool                                    `json:"isExposed"`
	AllowEmptyValue           bool                                    `json:"allowEmptyValue"`
	DefaultValue              string                                  `json:"defaultValue"`
	Value                     string                                  `json:"value,omitempty"`
	VariableType              string                                  `json:"variableType"`
	ValueType                 repository.PluginStepVariableValueType  `json:"valueType,omitempty"`
	PreviousStepIndex         int                                     `json:"previousStepIndex,omitempty"`
	VariableStepIndex         int                                     `json:"variableStepIndex"`
	VariableStepIndexInPlugin int                                     `json:"variableStepIndexInPlugin"`
	ReferenceVariableName     string                                  `json:"referenceVariableName,omitempty"`
	PluginStepCondition       []*repository.PluginStepCondition       `json:"pluginStepCondition,omitempty"`
}
