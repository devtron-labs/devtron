package bean

import (
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
)

type VariableType string

const (
	VARIABLE_TYPE_VALUE       = "VALUE"
	VARIABLE_TYPE_REF_PRE_CI  = "REF_PRE_CI"
	VARIABLE_TYPE_REF_POST_CI = "REF_POST_CI"
	VARIABLE_TYPE_REF_GLOBAL  = "REF_GLOBAL"
	VARIABLE_TYPE_REF_PLUGIN  = "REF_PLUGIN"
	IMAGE_SCANNER_ENDPOINT    = "IMAGE_SCANNER_ENDPOINT"
)

const CI_JOB string = "CI_JOB"

type WorkflowPipelineType string

const (
	CI_WORKFLOW_PIPELINE_TYPE  WorkflowPipelineType = "CI"
	CD_WORKFLOW_PIPELINE_TYPE  WorkflowPipelineType = "CD"
	JOB_WORKFLOW_PIPELINE_TYPE WorkflowPipelineType = "JOB"
)

type RefPluginObject struct {
	Id    int           `json:"id"`
	Steps []*StepObject `json:"steps"`
}

type PrePostAndRefPluginStepsResponse struct {
	PreStageSteps    []*StepObject
	PostStageSteps   []*StepObject
	RefPluginData    []*RefPluginObject
	VariableSnapshot map[string]string
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
	TriggerIfParentStageFail bool               `json:"triggerIfParentStageFail"`
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

type ContainerResources struct {
	MinCpu        string `json:"minCpu"`
	MaxCpu        string `json:"maxCpu"`
	MinStorage    string `json:"minStorage"`
	MaxStorage    string `json:"maxStorage"`
	MinEphStorage string `json:"minEphStorage"`
	MaxEphStorage string `json:"maxEphStorage"`
	MinMem        string `json:"minMem"`
	MaxMem        string `json:"maxMem"`
}
type CiProjectDetails struct {
	GitRepository   string `json:"gitRepository"`
	MaterialName    string `json:"materialName"`
	CheckoutPath    string `json:"checkoutPath"`
	FetchSubmodules bool   `json:"fetchSubmodules"`
	CommitHash      string `json:"commitHash"`
	GitTag          string `json:"gitTag"`
	CommitTime      string `json:"commitTime"`
	//Branch        string          `json:"branch"`
	Type        string                    `json:"type"`
	Message     string                    `json:"message"`
	Author      string                    `json:"author"`
	GitOptions  GitOptions                `json:"gitOptions"`
	SourceType  pipelineConfig.SourceType `json:"sourceType"`
	SourceValue string                    `json:"sourceValue"`
	WebhookData pipelineConfig.WebhookData
	CloningMode string `json:"cloningMode"`
}
type GitOptions struct {
	UserName      string               `json:"userName"`
	Password      string               `json:"password"`
	SshPrivateKey string               `json:"sshPrivateKey"`
	AccessToken   string               `json:"accessToken"`
	AuthMode      repository2.AuthMode `json:"authMode"`
}

type NodeConstraints struct {
	ServiceAccount   string
	TaintKey         string
	TaintValue       string
	NodeLabel        map[string]string
	SkipNodeSelector bool
}

type LimitReqCpuMem struct {
	LimitCpu string
	LimitMem string
	ReqCpu   string
	ReqMem   string
}
