package bean

import (
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
)

type PipelineStageDto struct {
	Id          int                          `json:"id"`
	Name        string                       `json:"name,omitempty"`
	Description string                       `json:"description,omitempty"`
	Type        repository.PipelineStageType `json:"type,omitempty" validate:"omitempty,oneof=PRE_CI POST_CI"`
	Steps       []*PipelineStageStepDto      `json:"steps"`
}

type PipelineStageStepDto struct {
	Id                  int                         `json:"id"`
	Name                string                      `json:"name"`
	Description         string                      `json:"description"`
	Index               int                         `json:"index"`
	StepType            repository.PipelineStepType `json:"stepType" validate:"omitempty,oneof=INLINE REF_PLUGIN"`
	OutputDirectoryPath []string                    `json:"outputDirectoryPath"`
	InlineStepDetail    *InlineStepDetailDto        `json:"inlineStepDetail"`
	RefPluginStepDetail *RefPluginStepDetailDto     `json:"pluginRefStepDetail"`
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
	MountPathMap             []*MountPathMap                       `json:"mountPathMap,omitempty"`
	CommandArgsMap           []*CommandArgsMap                     `json:"commandArgsMap,omitempty"`
	PortMap                  []*PortMap                            `json:"portMap,omitempty"`
	InputVariables           []*StepVariableDto                    `json:"inputVariables"`
	OutputVariables          []*StepVariableDto                    `json:"outputVariables"`
	ConditionDetails         []*ConditionDetailDto                 `json:"conditionDetails"`
}

type RefPluginStepDetailDto struct {
	PluginId         int                   `json:"pluginId"`
	InputVariables   []*StepVariableDto    `json:"inputVariables"`
	OutputVariables  []*StepVariableDto    `json:"outputVariables"`
	ConditionDetails []*ConditionDetailDto `json:"conditionDetails"`
}

type StepVariableDto struct {
	Id                        int                                            `json:"id"`
	Name                      string                                         `json:"name"`
	Format                    repository.PipelineStageStepVariableFormatType `json:"format" validate:"oneof=STRING NUMBER BOOL DATE"`
	Description               string                                         `json:"description"`
	IsExposed                 bool                                           `json:"isExposed,omitempty"`
	AllowEmptyValue           bool                                           `json:"allowEmptyValue,omitempty"`
	DefaultValue              string                                         `json:"defaultValue,omitempty"`
	Value                     string                                         `json:"value"`
	ValueType                 repository.PipelineStageStepVariableValueType  `json:"variableType,omitempty" validate:"oneof=NEW FROM_PREVIOUS_STEP GLOBAL"`
	PreviousStepIndex         int                                            `json:"refVariableStepIndex,omitempty"`
	ReferenceVariableName     string                                         `json:"refVariableName,omitempty"`
	VariableStepIndexInPlugin int                                            `json:"variableStepIndexInPlugin,omitempty"`
	ReferenceVariableStage    repository.PipelineStageType                   `json:"refVariableStage"`
}

type ConditionDetailDto struct {
	Id                  int                                       `json:"id"`
	ConditionOnVariable string                                    `json:"conditionOnVariable"` //name of variable on which condition is written
	ConditionType       repository.PipelineStageStepConditionType `json:"conditionType" validate:"oneof=SKIP TRIGGER SUCCESS FAIL"`
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
