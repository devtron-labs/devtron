package plugin

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
	Id                    int    `json:"id,omitempty"`
	Name                  string `json:"name"`
	Format                string `json:"format"`
	Description           string `json:"description"`
	IsExposed             bool   `json:"isExposed,omitempty"`
	AllowEmptyValue       bool   `json:"allowEmptyValue,omitempty"`
	DefaultValue          string `json:"defaultValue,omitempty"`
	Value                 string `json:"value,omitempty"`
	ValueType             string `json:"valueType"` //NEW, FROM_PREVIOUS_STEP or GLOBAL
	PreviousStepIndex     int    `json:"previousStepIndex,omitempty"`
	ReferenceVariableName string `json:"referenceVariableName,omitempty"`
}
