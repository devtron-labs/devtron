package bean

type VariableType string

const (
	VARIABLE_TYPE_VALUE       = "VALUE"
	VARIABLE_TYPE_REF_PRE_CI  = "REF_PRE_CI"
	VARIABLE_TYPE_REF_POST_CI = "REF_POST_CI"
	VARIABLE_TYPE_REF_GLOBAL  = "REF_GLOBAL"
	VARIABLE_TYPE_REF_PLUGIN  = "REF_PLUGIN"
)

type RefPluginObject struct {
	Id    int           `json:"id"`
	Steps []*StepObject `json:"steps"`
}

type StepObject struct {
	Name                     string             `json:"name"`
	Index                    int                `json:"index"`
	StepType                 string             `json:"stepType"`               // REF_PLUGIN or INLINE
	ExecutorType             string             `json:"executorType,omitempty"` //SHELL, DOCKERFILE, CONTAINER_IMAGE
	RefPluginId              int                `json:"refPluginId,omitempty"`
	Script                   string             `json:"script,omitempty"`
	InputVars                []*VariableObject  `json:"inputVars"`
	ExposedPorts             map[int]int        `json:"exposedPorts"` //map of host:container
	OutputVars               []*VariableObject  `json:"outputVars"`
	TriggerSkipConditions    []*ConditionObject `json:"triggerSkipConditions"`
	SuccessFailureConditions []*ConditionObject `json:"successFailureConditions"`
	DockerImage              string             `json:"dockerImage"`
	Command                  string             `json:"command"`
	Args                     []string           `json:"args"`
	CustomScriptMount        *MountPath         `json:"customScriptMount"` // destination path - storeScriptAt
	SourceCodeMount          *MountPath         `json:"sourceCodeMount"`   // destination path - mountCodeToContainerPath
	ExtraVolumeMounts        []*MountPath       `json:"extraVolumeMounts"` // filePathMapping
	ArtifactPaths            []string           `json:"artifactPaths"`
}

type VariableObject struct {
	Name   string `json:"name"`
	Format string `json:"format"` //STRING, NUMBER, BOOL, DATE
	//only for input type
	Value                      string       `json:"value,omitempty"`
	VariableType               VariableType `json:"variableType,omitempty"`
	ReferenceVariableName      string       `json:"referenceVariableName,omitempty"`
	ReferenceVariableStepIndex int          `json:"referenceVariableStepIndex,omitempty"`
	VariableStepIndexInPlugin  int          `json:"variableStepIndexInPlugin,omitempty"`
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
