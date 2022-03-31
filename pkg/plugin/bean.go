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
	Deleted     bool     `json:"deleted,omitempty"`
	Tags        []string `json:"tags"`
}

type PluginVariableDto struct {
	Id                    int    `json:"id"`
	Name                  string `json:"name"`
	Format                string `json:"format"`
	Description           string `json:"description"`
	IsExposed             bool   `json:"is_exposed"`
	AllowEmptyValue       bool   `json:"allow_empty_value"`
	DefaultValue          string `json:"default_value"`
	Value                 string `json:"value,omitempty"`
	VariableType          string `json:"variable_type,omitempty"` //INPUT or OUTPUT
	ValueType             string `json:"value_type"`              //NEW, FROM_PREVIOUS_STEP or GLOBAL
	PreviousStepIndex     int    `json:"previous_step_index"`
	ReferenceVariableName string `json:"reference_variable_name"`
	Deleted               bool   `json:"deleted,omitempty"`
}
