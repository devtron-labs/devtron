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
)

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

//
//type WorkflowRequest struct {
//	WorkflowNamePrefix         string                            `json:"workflowNamePrefix"`
//	PipelineName               string                            `json:"pipelineName"`
//	PipelineId                 int                               `json:"pipelineId"`
//	DockerImageTag             string                            `json:"dockerImageTag"`
//	DockerRegistryId           string                            `json:"dockerRegistryId"`
//	DockerRegistryType         string                            `json:"dockerRegistryType"`
//	DockerRegistryURL          string                            `json:"dockerRegistryURL"`
//	DockerConnection           string                            `json:"dockerConnection"`
//	DockerCert                 string                            `json:"dockerCert"`
//	DockerRepository           string                            `json:"dockerRepository"`
//	CheckoutPath               string                            `json:"checkoutPath"`
//	DockerUsername             string                            `json:"dockerUsername"`
//	DockerPassword             string                            `json:"dockerPassword"`
//	AwsRegion                  string                            `json:"awsRegion"`
//	AccessKey                  string                            `json:"accessKey"`
//	SecretKey                  string                            `json:"secretKey"`
//	CiCacheLocation            string                            `json:"ciCacheLocation"`
//	CiCacheRegion              string                            `json:"ciCacheRegion"`
//	CiCacheFileName            string                            `json:"ciCacheFileName"`
//	CiProjectDetails           []CiProjectDetails                `json:"ciProjectDetails"`
//	ContainerResources         ContainerResources                `json:"containerResources"`
//	ActiveDeadlineSeconds      int64                             `json:"activeDeadlineSeconds"`
//	CiImage                    string                            `json:"ciImage"`
//	Namespace                  string                            `json:"namespace"`
//	WorkflowId                 int                               `json:"workflowId"`
//	TriggeredBy                int32                             `json:"triggeredBy"`
//	CacheLimit                 int64                             `json:"cacheLimit"`
//	BeforeDockerBuildScripts   []*bean2.CiScript                 `json:"beforeDockerBuildScripts"`
//	AfterDockerBuildScripts    []*bean2.CiScript                 `json:"afterDockerBuildScripts"`
//	CiArtifactLocation         string                            `json:"ciArtifactLocation"`
//	CiArtifactBucket           string                            `json:"ciArtifactBucket"`
//	CiArtifactFileName         string                            `json:"ciArtifactFileName"`
//	CiArtifactRegion           string                            `json:"ciArtifactRegion"`
//	ScanEnabled                bool                              `json:"scanEnabled"`
//	CloudProvider              blob_storage.BlobStorageType      `json:"cloudProvider"`
//	BlobStorageConfigured      bool                              `json:"blobStorageConfigured"`
//	BlobStorageS3Config        *blob_storage.BlobStorageS3Config `json:"blobStorageS3Config"`
//	AzureBlobConfig            *blob_storage.AzureBlobConfig     `json:"azureBlobConfig"`
//	GcpBlobConfig              *blob_storage.GcpBlobConfig       `json:"gcpBlobConfig"`
//	BlobStorageLogsKey         string                            `json:"blobStorageLogsKey"`
//	InAppLoggingEnabled        bool                              `json:"inAppLoggingEnabled"`
//	DefaultAddressPoolBaseCidr string                            `json:"defaultAddressPoolBaseCidr"`
//	DefaultAddressPoolSize     int                               `json:"defaultAddressPoolSize"`
//	PreCiSteps                 []*StepObject                     `json:"preCiSteps"`
//	PostCiSteps                []*StepObject                     `json:"postCiSteps"`
//	RefPlugins                 []*RefPluginObject                `json:"refPlugins"`
//	AppName                    string                            `json:"appName"`
//	TriggerByAuthor            string                            `json:"triggerByAuthor"`
//	CiBuildConfig              *CiBuildConfigBean                `json:"ciBuildConfig"`
//	CiBuildDockerMtuValue      int                               `json:"ciBuildDockerMtuValue"`
//	IgnoreDockerCachePush      bool                              `json:"ignoreDockerCachePush"`
//	IgnoreDockerCachePull      bool                              `json:"ignoreDockerCachePull"`
//	CacheInvalidate            bool                              `json:"cacheInvalidate"`
//	IsPvcMounted               bool                              `json:"IsPvcMounted"`
//	ExtraEnvironmentVariables  map[string]string                 `json:"extraEnvironmentVariables"`
//	EnableBuildContext         bool                              `json:"enableBuildContext"`
//	AppId                      int                               `json:"appId"`
//	EnvironmentId              int                               `json:"environmentId"`
//	OrchestratorHost           string                            `json:"orchestratorHost"`
//	OrchestratorToken          string                            `json:"orchestratorToken"`
//	IsExtRun                   bool                              `json:"isExtRun"`
//	ImageRetryCount            int                               `json:"imageRetryCount"`
//	ImageRetryInterval         int                               `json:"imageRetryInterval"`
//	// Data from CD Workflow service
//	WorkflowRunnerId         int                                 `json:"workflowRunnerId"`
//	CdPipelineId             int                                 `json:"cdPipelineId"`
//	StageYaml                string                              `json:"stageYaml"`
//	ArtifactLocation         string                              `json:"artifactLocation"`
//	CiArtifactDTO            pipeline.CiArtifactDTO              `json:"ciArtifactDTO"`
//	CdImage                  string                              `json:"cdImage"`
//	StageType                string                              `json:"stageType"`
//	CdCacheLocation          string                              `json:"cdCacheLocation"`
//	CdCacheRegion            string                              `json:"cdCacheRegion"`
//	WorkflowPrefixForLog     string                              `json:"workflowPrefixForLog"`
//	DeploymentTriggeredBy    string                              `json:"deploymentTriggeredBy,omitempty"`
//	DeploymentTriggerTime    time.Time                           `json:"deploymentTriggerTime,omitempty"`
//	DeploymentReleaseCounter int                                 `json:"deploymentReleaseCounter,omitempty"`
//	WorkflowExecutor         pipelineConfig.WorkflowExecutorType `json:"workflowExecutor"`
//	PrePostDeploySteps       []*StepObject                       `json:"prePostDeploySteps"`
//	Type                     WorkflowPipelineType
//	Pipeline                 *pipelineConfig.Pipeline
//	Env                      *repository.Environment
//	AppLabels                map[string]string
//}

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
}
type GitOptions struct {
	UserName      string               `json:"userName"`
	Password      string               `json:"password"`
	SshPrivateKey string               `json:"sshPrivateKey"`
	AccessToken   string               `json:"accessToken"`
	AuthMode      repository2.AuthMode `json:"authMode"`
}

type NodeConstraints struct {
	ServiceAccount    string
	TaintKey          string
	TaintValue        string
	NodeLabel         map[string]string
	SkipNodeSelector  bool
	StorageConfigured bool
}

type LimitReqCpuMem struct {
	LimitCpu string
	LimitMem string
	ReqCpu   string
	ReqMem   string
}
