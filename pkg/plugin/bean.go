package plugin

import (
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
)

const (
	CREATEPLUGIN      = 0
	UPDATEPLUGIN      = 1
	DELETEPLUGIN      = 2
	CI_TYPE_PLUGIN    = "CI"
	CD_TYPE_PLUGIN    = "CD"
	CI_CD_TYPE_PLUGIN = "CI_CD"
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
	Type        string            `json:"type" validate:"oneof=SHARED PRESET"` // SHARED, PRESET etc
	Icon        string            `json:"icon"`
	Tags        []string          `json:"tags"`
	Action      int               `json:"action"`
	PluginStage string            `json:"pluginStage,omitempty"`
	PluginSteps []*PluginStepsDto `json:"pluginSteps,omitempty"`
}

func (r *PluginMetadataDto) getPluginMetadataSqlObj(userId int32) *repository.PluginMetadata {
	return &repository.PluginMetadata{
		Name:        r.Name,
		Description: r.Description,
		Type:        repository.PluginType(r.Type),
		Icon:        r.Icon,
		AuditLog:    sql.NewDefaultAuditLog(userId),
	}
}

type PluginStepsDto struct {
	Id                   int                       `json:"id,pk"`
	Name                 string                    `json:"name"`
	Description          string                    `json:"description"`
	Index                int                       `json:"index"`
	StepType             repository.PluginStepType `json:"stepType"`
	RefPluginId          int                       `json:"refPluginId"` //id of plugin used as reference
	OutputDirectoryPath  []string                  `json:"outputDirectoryPath"`
	DependentOnStep      string                    `json:"dependentOnStep"`
	PluginStepVariable   []*PluginVariableDto      `json:"pluginStepVariable,omitempty"`
	PluginPipelineScript *PluginPipelineScript     `json:"pluginPipelineScript,omitempty"`
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
	VariableType              repository.PluginStepVariableType       `json:"variableType"`
	ValueType                 repository.PluginStepVariableValueType  `json:"valueType,omitempty"`
	PreviousStepIndex         int                                     `json:"previousStepIndex,omitempty"`
	VariableStepIndex         int                                     `json:"variableStepIndex"`
	VariableStepIndexInPlugin int                                     `json:"variableStepIndexInPlugin"`
	ReferenceVariableName     string                                  `json:"referenceVariableName,omitempty"`
	PluginStepCondition       []*PluginStepCondition                  `json:"pluginStepCondition,omitempty"`
}

type PluginPipelineScript struct {
	Id                       int                                  `json:"id"`
	Script                   string                               `json:"script"`
	StoreScriptAt            string                               `json:"storeScriptAt"`
	Type                     repository.ScriptType                `json:"type"`
	DockerfileExists         bool                                 `json:"dockerfileExists"`
	MountPath                string                               `json:"mountPath"`
	MountCodeToContainer     bool                                 `json:"mountCodeToContainer"`
	MountCodeToContainerPath string                               `json:"mountCodeToContainerPath"`
	MountDirectoryFromHost   bool                                 `json:"mountDirectoryFromHost"`
	ContainerImagePath       string                               `json:"containerImagePath"`
	ImagePullSecretType      repository.ScriptImagePullSecretType `json:"imagePullSecretType"`
	ImagePullSecret          string                               `json:"imagePullSecret"`
	Deleted                  bool                                 `json:"deleted"`
	PathArgPortMapping       []*ScriptPathArgPortMapping          `json:"pathArgPortMapping"`
}

type PluginStepCondition struct {
	Id                  int                                `json:"id"`
	PluginStepId        int                                `json:"pluginStepId"`
	ConditionVariableId int                                `json:"conditionVariableId"` //id of variable on which condition is written
	ConditionType       repository.PluginStepConditionType `json:"conditionType"`
	ConditionalOperator string                             `json:"conditionalOperator"`
	ConditionalValue    string                             `json:"conditionalValue"`
	Deleted             bool                               `json:"deleted"`
}

type ScriptPathArgPortMapping struct {
	Id                  int                          `json:"id"`
	TypeOfMapping       repository.ScriptMappingType `json:"typeOfMapping"`
	FilePathOnDisk      string                       `json:"filePathOnDisk"`
	FilePathOnContainer string                       `json:"filePathOnContainer"`
	Command             string                       `json:"command"`
	Args                []string                     `json:"args"`
	PortOnLocal         int                          `json:"portOnLocal"`
	PortOnContainer     int                          `json:"portOnContainer"`
	ScriptId            int                          `json:"scriptId"`
}
