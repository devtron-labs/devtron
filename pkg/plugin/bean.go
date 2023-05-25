package plugin

import "github.com/devtron-labs/devtron/pkg/plugin/repository"

type PluginDetailDto struct {
	Metadata        *PluginMetadataDto   `json:"metadata"`
	InputVariables  []*PluginVariableDto `json:"inputVariables"`
	OutputVariables []*PluginVariableDto `json:"outputVariables"`
}

type PluginMetadataDto struct {
	Id          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"` // SHARED, PRESET etc
	Icon        string   `json:"icon"`
	Tags        []string `json:"tags"`
}

type PluginVariableDto struct {
	Id                    int                                     `json:"id,omitempty"`
	Name                  string                                  `json:"name"`
	Format                repository.PluginStepVariableFormatType `json:"format"`
	Description           string                                  `json:"description"`
	IsExposed             bool                                    `json:"isExposed"`
	AllowEmptyValue       bool                                    `json:"allowEmptyValue"`
	DefaultValue          string                                  `json:"defaultValue"`
	Value                 string                                  `json:"value,omitempty"`
	ValueType             repository.PluginStepVariableValueType  `json:"variableType,omitempty"`
	PreviousStepIndex     int                                     `json:"previousStepIndex,omitempty"`
	VariableStepIndex     int                                     `json:"variableStepIndex"`
	ReferenceVariableName string                                  `json:"referenceVariableName,omitempty"`
}
