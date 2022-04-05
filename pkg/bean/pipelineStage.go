package bean

type PipelineStageDto struct {
	Id          int                     `json:"id"`
	Name        string                  `json:"name,omitempty"`
	Description string                  `json:"description,omitempty"`
	Type        string                  `json:"type,omitempty"`
	Steps       []*PipelineStageStepDto `json:"steps"`
}

type PipelineStageStepDto struct {
	Id                  int                     `json:"id"`
	Name                string                  `json:"name"`
	Description         string                  `json:"description"`
	Index               int                     `json:"index"`
	StepType            string                  `json:"stepType"`
	ReportDirectoryPath string                  `json:"reportDirectoryPath"`
	InlineStepDetail    *InlineStepDetailDto    `json:"inlineStepDetail"`
	RefPluginStepDetail *RefPluginStepDetailDto `json:"pluginRefStepDetail"`
}

type InlineStepDetailDto struct {
	ScriptType           string                `json:"scriptType"`
	Script               string                `json:"script"`
	DockerfileExists     bool                  `json:"dockerfileExists,omitempty"`
	StoreScriptAt        string                `json:"storeScriptAt,omitempty"`
	MountPath            string                `json:"mountPath,omitempty"`
	MountCodeToContainer bool                  `json:"mountCodeToContainer,omitempty"`
	ConfigureMountPath   bool                  `json:"configureMountPath,omitempty"`
	ContainerImagePath   string                `json:"containerImagePath,omitempty"`
	ImagePullSecretType  string                `json:"imagePullSecretType,omitempty"`
	ImagePullSecret      string                `json:"imagePullSecret,omitempty"`
	MountPathMap         []*MountPathMap       `json:"mountPathMap,omitempty"`
	InputVariables       []*StepVariableDto    `json:"inputVariables"`
	OutputVariables      []*StepVariableDto    `json:"outputVariables"`
	ConditionDetails     []*ConditionDetailDto `json:"conditionDetails"`
}

type RefPluginStepDetailDto struct {
	PluginId         int                   `json:"pluginId,omitempty"`
	InputVariables   []*StepVariableDto    `json:"inputVariables"`
	OutputVariables  []*StepVariableDto    `json:"outputVariables"`
	ConditionDetails []*ConditionDetailDto `json:"conditionDetails"`
}

type StepVariableDto struct {
	Id                    int    `json:"id"`
	Name                  string `json:"name"`
	Format                string `json:"format"`
	Description           string `json:"description"`
	IsExposed             bool   `json:"isExposed,omitempty"`
	AllowEmptyValue       bool   `json:"allowEmptyValue,omitempty"`
	DefaultValue          string `json:"defaultValue,omitempty"`
	Value                 string `json:"value"`
	ValueType             string `json:"valueType,omitempty"`
	PreviousStepIndex     int    `json:"previousStepIndex,omitempty"`
	ReferenceVariableName string `json:"referenceVariableName,omitempty"`
}

type ConditionDetailDto struct {
	Id                  int    `json:"id"`
	ConditionOnVariable string `json:"conditionOnVariable"` //name of variable on which condition is written
	ConditionType       string `json:"conditionType"`       //SKIP, TRIGGER, SUCCESS, FAIL
	ConditionalOperator string `json:"conditionalOperator"`
	ConditionalValue    string `json:"conditionalValue"`
}

type MountPathMap struct {
	FilePathOnDisk      string `json:"filePathOnDisk"`
	FilePathOnContainer string `json:"filePathOnContainer"`
}
