package bean

type RefPluginObject struct {
	Id    int           `json:"id"`
	Steps []*StepObject `json:"steps"`
}

type StepObject struct {
	Index                    int                `json:"index"`
	StepType                 string             `json:"stepType"`     // REF_PLUGIN or INLINE
	ExecutorType             string             `json:"executorType"` //SHELL, DOCKERFILE, CONTAINER_IMAGE
	RefPluginId              int                `json:"refPluginId"`
	Script                   string             `json:"script"`
	InputVars                []*VariableObject  `json:"inputVars"`
	ExposedPorts             map[int]int        `json:"exposedPorts"` //map of host:container
	OutputVars               []*VariableObject  `json:"outputVars"`
	TriggerSkipConditions    []*ConditionObject `json:"triggerSkipConditions"`
	SuccessFailureConditions []*ConditionObject `json:"successFailureConditions"`
	DockerImage              string             `json:"dockerImage"`
	Command                  string             `json:"command"`
	Args                     []string           `json:"args"`
	CustomScriptMount        *MountPath         `json:"customScriptMountDestinationPath"` // destination path - storeScriptAt
	SourceCodeMount          *MountPath         `json:"sourceCodeMountDestinationPath"`   // destination path - mountCodeToContainerPath
	ExtraVolumeMounts        []*MountPath       `json:"extraVolumeMounts"`                // filePathMapping
	ArtifactPaths            []string           `json:"artifactPaths"`
}

type VariableObject struct {
	Name   string `json:"name"`
	Format string `json:"format"` //STRING, NUMBER, BOOL, DATE
	//only for input type
	Value                      string `json:"value,omitempty"`
	ValueType                  string `json:"valueType,omitempty"`
	ReferenceVariableName      string `json:"referenceVariableName,omitempty"`
	ReferenceVariableStepIndex int    `json:"referenceVariableStepIndex,omitempty"`
	ReferenceVariableStage     string `json:"referenceVariableStage,omitempty"`
}

type ConditionObject struct {
	ConditionType       string `json:"conditionType"`       //TRIGGER, SKIP, SUCCESS, FAIL
	ConditionOnVariable string `json:"conditionOnVariable"` //name of variable
	ConditionalOperator string `json:"conditionalOperator"`
	ConditionalValue    string `json:"conditionalValue"`
}

type MountPath struct {
	SourcePath      string `json:"sourcePath"`
	DestinationPath string `json:"destinationPath"`
}
