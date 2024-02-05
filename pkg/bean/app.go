/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package bean

import (
	"encoding/json"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository/imageTagging"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"time"
)

const (
	LayoutISO     = "2006-01-02 15:04:05"
	LayoutUS      = "January 2, 2006 15:04:05"
	LayoutRFC3339 = "2006-01-02T15:04:05Z07:00"
)

type SourceTypeConfig struct {
	Type  pipelineConfig.SourceType `json:"type,omitempty" validate:"oneof=SOURCE_TYPE_BRANCH_FIXED SOURCE_TYPE_BRANCH_REGEX SOURCE_TYPE_TAG_ANY WEBHOOK"`
	Value string                    `json:"value,omitempty" `
	Regex string                    `json:"regex"`
}

type CreateAppDTO struct {
	Id          int                            `json:"id,omitempty" validate:"number"`
	AppName     string                         `json:"appName" validate:"name-component,max=100"`
	Description string                         `json:"description"`
	UserId      int32                          `json:"-"` //not exposed to UI
	Material    []*GitMaterial                 `json:"material" validate:"dive,min=1"`
	TeamId      int                            `json:"teamId,omitempty" validate:"number,required"`
	TemplateId  int                            `json:"templateId"`
	AppLabels   []*Label                       `json:"labels,omitempty" validate:"dive"`
	GenericNote *bean2.GenericNoteResponseBean `json:"genericNote,omitempty"`
	AppType     helper.AppType                 `json:"appType" validate:"gt=-1,lt=3"` //TODO: Change Validation if new AppType is introduced
}

type CreateMaterialDTO struct {
	Id       int            `json:"id,omitempty" validate:"number"`
	AppId    int            `json:"appId" validate:"number"`
	Material []*GitMaterial `json:"material" validate:"dive,min=1"`
	UserId   int32          `json:"-"` //not exposed to UI
}

type UpdateMaterialDTO struct {
	AppId    int          `json:"appId" validate:"number"`
	Material *GitMaterial `json:"material" validate:"dive,min=1"`
	UserId   int32        `json:"-"` //not exposed to UI
}

type GitMaterial struct {
	Name             string   `json:"name,omitempty" ` //not null, //default format pipelineGroup.AppName + "-" + inputMaterial.Name,
	Url              string   `json:"url,omitempty"`   //url of git repo
	Id               int      `json:"id,omitempty" validate:"number"`
	GitProviderId    int      `json:"gitProviderId,omitempty" validate:"gt=0"`
	CheckoutPath     string   `json:"checkoutPath" validate:"checkout-path-component"`
	FetchSubmodules  bool     `json:"fetchSubmodules"`
	IsUsedInCiConfig bool     `json:"isUsedInCiConfig"`
	FilterPattern    []string `json:"filterPattern"`
}

type CiMaterial struct {
	Source          *SourceTypeConfig `json:"source,omitempty" validate:"dive,required"`   //branch for ci
	Path            string            `json:"path,omitempty"`                              // defaults to root of git repo
	CheckoutPath    string            `json:"checkoutPath,omitempty"`                      //path where code will be checked out for single source `./` default for multiSource configured by user
	GitMaterialId   int               `json:"gitMaterialId,omitempty" validate:"required"` //id stored in db GitMaterial( foreign key)
	ScmId           string            `json:"scmId,omitempty"`                             //id of gocd object
	ScmName         string            `json:"scmName,omitempty"`
	ScmVersion      string            `json:"scmVersion,omitempty"`
	Id              int               `json:"id,omitempty"`
	GitMaterialName string            `json:"gitMaterialName"`
	IsRegex         bool              `json:"isRegex"`
}

type CiPipeline struct {
	IsManual                 bool                   `json:"isManual"`
	DockerArgs               map[string]string      `json:"dockerArgs"`
	IsExternal               bool                   `json:"isExternal"`
	ParentCiPipeline         int                    `json:"parentCiPipeline"`
	ParentAppId              int                    `json:"parentAppId"`
	AppId                    int                    `json:"appId"`
	AppName                  string                 `json:"appName,omitempty"`
	AppType                  helper.AppType         `json:"appType,omitempty"`
	ExternalCiConfig         ExternalCiConfig       `json:"externalCiConfig"`
	CiMaterial               []*CiMaterial          `json:"ciMaterial,omitempty" validate:"dive,min=1"`
	Name                     string                 `json:"name,omitempty" validate:"name-component,max=100"` //name suffix of corresponding pipeline. required, unique, validation corresponding to gocd pipelineName will be applicable
	Id                       int                    `json:"id,omitempty" `
	Version                  string                 `json:"version,omitempty"` //matchIf token version in gocd . used for update request
	Active                   bool                   `json:"active,omitempty"`  //pipeline is active or not
	Deleted                  bool                   `json:"deleted,omitempty"`
	BeforeDockerBuild        []*Task                `json:"beforeDockerBuild,omitempty" validate:"dive"`
	AfterDockerBuild         []*Task                `json:"afterDockerBuild,omitempty" validate:"dive"`
	BeforeDockerBuildScripts []*CiScript            `json:"beforeDockerBuildScripts,omitempty" validate:"dive"`
	AfterDockerBuildScripts  []*CiScript            `json:"afterDockerBuildScripts,omitempty" validate:"dive"`
	LinkedCount              int                    `json:"linkedCount"`
	PipelineType             PipelineType           `json:"pipelineType,omitempty"`
	ScanEnabled              bool                   `json:"scanEnabled,notnull"`
	AppWorkflowId            int                    `json:"appWorkflowId,omitempty"`
	PreBuildStage            *bean.PipelineStageDto `json:"preBuildStage,omitempty"`
	PostBuildStage           *bean.PipelineStageDto `json:"postBuildStage,omitempty"`
	TargetPlatform           string                 `json:"targetPlatform,omitempty"`
	IsDockerConfigOverridden bool                   `json:"isDockerConfigOverridden"`
	DockerConfigOverride     DockerConfigOverride   `json:"dockerConfigOverride,omitempty"`
	EnvironmentId            int                    `json:"environmentId,omitempty"`
	LastTriggeredEnvId       int                    `json:"lastTriggeredEnvId"`
	CustomTagObject          *CustomTagData         `json:"customTag,omitempty"`
	DefaultTag               []string               `json:"defaultTag,omitempty"`
	EnableCustomTag          bool                   `json:"enableCustomTag"`
}

type DockerConfigOverride struct {
	DockerRegistry   string                  `json:"dockerRegistry,omitempty"`
	DockerRepository string                  `json:"dockerRepository,omitempty"`
	CiBuildConfig    *bean.CiBuildConfigBean `json:"ciBuildConfig,omitEmpty"`
	//DockerBuildConfig *DockerBuildConfig  `json:"dockerBuildConfig,omitempty"`
}

type CiPipelineMin struct {
	Name             string       `json:"name,omitempty" validate:"name-component,max=100"` //name suffix of corresponding pipeline. required, unique, validation corresponding to gocd pipelineName will be applicable
	Id               int          `json:"id,omitempty" `
	Version          string       `json:"version,omitempty"` //matchIf token version in gocd . used for update request
	IsExternal       bool         `json:"isExternal,omitempty"`
	ParentCiPipeline int          `json:"parentCiPipeline"`
	ParentAppId      int          `json:"parentAppId"`
	PipelineType     PipelineType `json:"pipelineType,omitempty"`
	ScanEnabled      bool         `json:"scanEnabled,notnull"`
}

type CiScript struct {
	Id             int    `json:"id"`
	Index          int    `json:"index"`
	Name           string `json:"name" validate:"required"`
	Script         string `json:"script"`
	OutputLocation string `json:"outputLocation"`
}

type ExternalCiConfig struct {
	Id            int                    `json:"id"`
	WebhookUrl    string                 `json:"webhookUrl"`
	Payload       string                 `json:"payload"`
	AccessKey     string                 `json:"accessKey"`
	PayloadOption []PayloadOptionObject  `json:"payloadOption"`
	Schema        map[string]interface{} `json:"schema"`
	Responses     []ResponseSchemaObject `json:"responses"`
	ExternalCiConfigRole
}

type ExternalCiConfigRole struct {
	ProjectId             int    `json:"projectId"`
	ProjectName           string `json:"projectName"`
	EnvironmentId         string `json:"environmentId"`
	EnvironmentName       string `json:"environmentName"`
	EnvironmentIdentifier string `json:"environmentIdentifier"`
	AppId                 int    `json:"appId"`
	AppName               string `json:"appName"`
	Role                  string `json:"role"`
}

// -------------------
type PatchAction int
type PipelineType string

const (
	CREATE        PatchAction = iota
	UPDATE_SOURCE             //update value of SourceTypeConfig
	DELETE                    //delete this pipeline
	//DEACTIVATE     //pause/deactivate this pipeline
)

const (
	NORMAL    PipelineType = "NORMAL"
	LINKED    PipelineType = "LINKED"
	EXTERNAL  PipelineType = "EXTERNAL"
	CI_JOB    PipelineType = "CI_JOB"
	LINKED_CD PipelineType = "LINKED_CD"
)

const (
	CASCADE_DELETE int = iota
	NON_CASCADE_DELETE
	FORCE_DELETE
)
const (
	WEBHOOK_SELECTOR_UNIQUE_ID_NAME          string = "unique id"
	WEBHOOK_SELECTOR_REPOSITORY_URL_NAME     string = "repository url"
	WEBHOOK_SELECTOR_HEADER_NAME             string = "header"
	WEBHOOK_SELECTOR_GIT_URL_NAME            string = "git url"
	WEBHOOK_SELECTOR_AUTHOR_NAME             string = "author"
	WEBHOOK_SELECTOR_DATE_NAME               string = "date"
	WEBHOOK_SELECTOR_TARGET_CHECKOUT_NAME    string = "target checkout"
	WEBHOOK_SELECTOR_SOURCE_CHECKOUT_NAME    string = "source checkout"
	WEBHOOK_SELECTOR_TARGET_BRANCH_NAME_NAME string = "target branch name"
	WEBHOOK_SELECTOR_SOURCE_BRANCH_NAME_NAME string = "source branch name"

	WEBHOOK_EVENT_MERGED_ACTION_TYPE     string = "merged"
	WEBHOOK_EVENT_NON_MERGED_ACTION_TYPE string = "non-merged"
)

type CiPatchStatus string

const (
	CI_PATCH_SUCCESS        CiPatchStatus = "Succeeded"
	CI_PATCH_FAILED         CiPatchStatus = "Failed"
	CI_PATCH_NOT_AUTHORIZED CiPatchStatus = "Not authorised"
	CI_PATCH_SKIP           CiPatchStatus = "Skipped"
)

type CiPatchMessage string

const (
	CI_PATCH_NOT_AUTHORIZED_MESSAGE CiPatchMessage = "You don't have permission to change branch"
	CI_PATCH_MULTI_GIT_ERROR        CiPatchMessage = "Build pipeline is connected to multiple git repositories"
	CI_PATCH_REGEX_ERROR            CiPatchMessage = "Provided branch does not match regex "
	CI_BRANCH_TYPE_ERROR            CiPatchMessage = "Branch cannot be changed for pipeline as source type is “Pull request or Tag”"
	CI_PATCH_SKIP_MESSAGE           CiPatchMessage = "Skipped for pipeline as source type is "
)

func (a PatchAction) String() string {
	return [...]string{"CREATE", "UPDATE_SOURCE", "DELETE", "DEACTIVATE"}[a]

}

// ----------------

type CiMaterialPatchRequest struct {
	AppId         int               `json:"appId" validate:"required"`
	EnvironmentId int               `json:"environmentId" validate:"required"`
	Source        *SourceTypeConfig `json:"source" validate:"required"`
}

type CustomTagData struct {
	TagPattern string `json:"tagPattern"`
	CounterX   int    `json:"counterX"`
	Enabled    bool   `json:"enabled"`
}

type CiMaterialValuePatchRequest struct {
	AppId         int `json:"appId" validate:"required"`
	EnvironmentId int `json:"environmentId" validate:"required"`
}

type CiMaterialBulkPatchRequest struct {
	AppIds        []int  `json:"appIds" validate:"required"`
	EnvironmentId int    `json:"environmentId" validate:"required"`
	Value         string `json:"value,omitempty" validate:"required"`
}

type CiMaterialBulkPatchResponse struct {
	Apps []CiMaterialPatchResponse `json:"apps"`
}

type CiMaterialPatchResponse struct {
	AppId   int           `json:"appId"`
	Status  CiPatchStatus `json:"status"`
	Message string        `json:"message"`
}

type CiPatchRequest struct {
	CiPipeline    *CiPipeline `json:"ciPipeline"`
	AppId         int         `json:"appId,omitempty"`
	Action        PatchAction `json:"action"`
	AppWorkflowId int         `json:"appWorkflowId,omitempty"`
	UserId        int32       `json:"-"`
	IsJob         bool        `json:"-"`
	IsCloneJob    bool        `json:"isCloneJob,omitempty"`

	ParentCDPipeline               int          `json:"parentCDPipeline"`
	DeployEnvId                    int          `json:"deployEnvId"`
	SwitchFromCiPipelineId         int          `json:"switchFromCiPipelineId"`
	SwitchFromExternalCiPipelineId int          `json:"switchFromExternalCiPipelineId"`
	SwitchFromCiPipelineType       PipelineType `json:"-"`
	SwitchToCiPipelineType         PipelineType `json:"-"`
}

func (ciPatchRequest CiPatchRequest) IsSwitchCiPipelineRequest() bool {
	return (ciPatchRequest.SwitchFromCiPipelineId != 0 || ciPatchRequest.SwitchFromExternalCiPipelineId != 0)
}

type CiRegexPatchRequest struct {
	CiPipelineMaterial []*CiPipelineMaterial `json:"ciPipelineMaterial,omitempty"`
	Id                 int                   `json:"id,omitempty" `
	AppId              int                   `json:"appId,omitempty"`
	UserId             int32                 `json:"-"`
}

type GitCiTriggerRequest struct {
	CiPipelineMaterial        CiPipelineMaterial `json:"ciPipelineMaterial" validate:"required"`
	TriggeredBy               int32              `json:"triggeredBy"`
	ExtraEnvironmentVariables map[string]string  `json:"extraEnvironmentVariables"` // extra env variables which will be used for CI
}

type SourceType string

type CiPipelineMaterial struct {
	Id            int                      `json:"Id"`
	GitMaterialId int                      `json:"GitMaterialId"`
	Type          string                   `json:"Type"`
	Value         string                   `json:"Value"`
	Active        bool                     `json:"Active"`
	GitCommit     pipelineConfig.GitCommit `json:"GitCommit"`
	GitTag        string                   `json:"GitTag"`
}

type CiTriggerRequest struct {
	PipelineId          int                  `json:"pipelineId"`
	CiPipelineMaterial  []CiPipelineMaterial `json:"ciPipelineMaterials" validate:"required"`
	TriggeredBy         int32                `json:"triggeredBy"`
	InvalidateCache     bool                 `json:"invalidateCache"`
	EnvironmentId       int                  `json:"environmentId"`
	PipelineType        string               `json:"pipelineType"`
	CiArtifactLastFetch time.Time            `json:"ciArtifactLastFetch"`
}

type CiTrigger struct {
	CiMaterialId int    `json:"ciMaterialId"`
	CommitHash   string `json:"commitHash"`
}

type Material struct {
	GitMaterialId int    `json:"gitMaterialId"`
	MaterialName  string `json:"materialName"`
}

type TriggerViewCiConfig struct {
	CiGitMaterialId int           `json:"ciGitConfiguredId"`
	CiPipelines     []*CiPipeline `json:"ciPipelines,omitempty" validate:"dive"` //a pipeline will be built for each ciMaterial
	Materials       []Material    `json:"materials"`
}

type CiConfigRequest struct {
	Id                 int                             `json:"id,omitempty" validate:"number"` //ciTemplateId
	AppId              int                             `json:"appId,omitempty" validate:"required,number"`
	DockerRegistry     string                          `json:"dockerRegistry,omitempty" `  //repo id example ecr mapped one-one with gocd registry entry
	DockerRepository   string                          `json:"dockerRepository,omitempty"` // example test-app-1 which is inside ecr
	CiBuildConfig      *bean.CiBuildConfigBean         `json:"ciBuildConfig"`
	CiPipelines        []*CiPipeline                   `json:"ciPipelines,omitempty" validate:"dive"` //a pipeline will be built for each ciMaterial
	AppName            string                          `json:"appName,omitempty"`
	Version            string                          `json:"version,omitempty"` //gocd etag used for edit purpose
	DockerRegistryUrl  string                          `json:"-"`
	CiTemplateName     string                          `json:"-"`
	UserId             int32                           `json:"-"`
	Materials          []Material                      `json:"materials"`
	AppWorkflowId      int                             `json:"appWorkflowId,omitempty"`
	BeforeDockerBuild  []*Task                         `json:"beforeDockerBuild,omitempty" validate:"dive"`
	AfterDockerBuild   []*Task                         `json:"afterDockerBuild,omitempty" validate:"dive"`
	ScanEnabled        bool                            `json:"scanEnabled,notnull"`
	CreatedOn          time.Time                       `sql:"created_on,type:timestamptz"`
	CreatedBy          int32                           `sql:"created_by,type:integer"`
	UpdatedOn          time.Time                       `sql:"updated_on,type:timestamptz"`
	UpdatedBy          int32                           `sql:"updated_by,type:integer"`
	IsJob              bool                            `json:"-"`
	CiGitMaterialId    int                             `json:"ciGitConfiguredId"`
	IsCloneJob         bool                            `json:"isCloneJob,omitempty"`
	AppWorkflowMapping *appWorkflow.AppWorkflowMapping `json:"-"`
	Artifact           *repository3.CiArtifact         `json:"-"`
}

type CiPipelineMinResponse struct {
	Id               int    `json:"id,omitempty" validate:"number"` //ciTemplateId
	AppId            int    `json:"appId,omitempty" validate:"required,number"`
	AppName          string `json:"appName,omitempty"`
	ParentCiPipeline int    `json:"parentCiPipeline"`
	ParentAppId      int    `json:"parentAppId"`
	PipelineType     string `json:"pipelineType"`
}

type TestExecutorImageProperties struct {
	ImageName string `json:"imageName,omitempty"`
	Arg       string `json:"arg,omitempty"`
	ReportDir string `json:"reportDir,omitempty"`
}

type PipelineCreateResponse struct {
	AppName string `json:"appName,omitempty"`
	AppId   int    `json:"appId,omitempty"`
}

/*
user should be able to compose multiple sequential and parallel steps for building binary.
*/
type BuildBinaryConfig struct {
	Name   string  `json:"name"`
	Stages []Stage `json:"stages"` //stages will be executed sequentially
}

type Stage struct {
	Name string `json:"name"`
	Jobs []Job  `json:"jobs"` //job will run in parallel
}

type Job struct {
	Name  string `json:"name"`
	Tasks []Task `json:"tasks"` //task will run sequentially
}

type Task struct {
	Name string   `json:"name"`
	Type string   `json:"type"` //for now ignore this input
	Cmd  string   `json:"cmd"`
	Args []string `json:"args"`
}

/*
tag git
build binary
push binary to artifact store
build docker image
push docker image
docker args
*/
type PackagingConfig struct {
}

/*
contains reference to chart and values.yaml changes for next deploy
*/
type HelmConfig struct {
}

// used for automated unit and integration test
type Test struct {
	Name    string
	Command string
}

//pipeline

type Pipeline struct {
	Environment Environment

	//Test ->
}

/*
if Environments has multiple entries then application of them will be deployed simultaneously
*/
type EnvironmentGroup struct {
	Name         string
	Environments []Environment
}

// set of unique attributes which corresponds to a cluster
// different environment of gocd and k8s cluster.
type Environment struct {
	Values string
}

type MaterialMetadata struct {
	ProgrammingLang      string
	LanguageRuntime      string
	BuildTool            string
	Executables          []string
	Profiles             map[string]string // pipeline-stage, profile
	LogDirs              map[string]string //file, log pattern
	EnvironmentVariables map[string]string
	PropertiesConfig     []PropertiesConfig
	ExposeConfig         []ServiceExposeConfig //a mocroservice can be exposed in multiple ways
	MonitoringConfig     MonitoringConfig
}

type PropertiesConfig struct {
	Name          string
	Location      string
	MountLocation string //MountLocation and Location might be same

	//figure out way to templatize the properties file
	//Vars map[string]string
}

type MonitoringConfig struct {
	port                   string
	ReadinessProbeEndpoint string
	InitialDelaySeconds    int32
	PeriodSeconds          int32
	TimeoutSeconds         int32
	SuccessThreshold       int32
	FailureThreshold       int32
	HttpHeaders            map[string]string
	TpMonitoringConf       []ThirdPartyMonitoringConfig
	//alertReceiver -> user who would receive alert with threshold
	// alert threshold
}

type ThirdPartyMonitoringConfig struct {
}

type ExposeType string
type Scheme string

const (
	EXPOSE_INTERNAL ExposeType = "clusterIp"
	EXPOSE_EXTERNAL ExposeType = "elb"
	SCHEME_HTTP     Scheme     = "http"
	SCHEME_HTTPS    Scheme     = "https"
	SCHEME_TCP      Scheme     = "tcp"
)

type ServiceExposeConfig struct {
	ExposeType  ExposeType
	Scheme      Scheme
	Port        string
	Path        string
	BackendPath string
	Host        string
}

type MaterialOperations interface {
	MaterialExists(material *GitMaterial) (bool, error)
	SaveMaterial(material *GitMaterial) error
	GenerateMaterialMetaData(material *GitMaterial) (*MaterialMetadata, error)
	ValidateMaterialMetaData(material *GitMaterial, metadata *MaterialMetadata) (bool, error)
	SaveMaterialMetaData(metadata *MaterialMetadata) error
}

// --------- cd related struct ---------
type CDMaterialMetadata struct {
	Url    string `json:"url,omitempty"`
	Branch string `json:"branch,omitempty"`
	Tag    string `json:"tag,omitempty"`
}

type CDSourceObject struct {
	Id          int                `json:"id"`
	DisplayName string             `json:"displayName"`
	Metadata    CDMaterialMetadata `json:"metadata"`
}

type CDPipelineConfigObject struct {
	Id                            int                                    `json:"id,omitempty"  validate:"number" `
	EnvironmentId                 int                                    `json:"environmentId,omitempty"  validate:"number,required" `
	EnvironmentName               string                                 `json:"environmentName,omitempty" `
	Description                   string                                 `json:"description" validate:"max=40"`
	CiPipelineId                  int                                    `json:"ciPipelineId,omitempty" validate:"number"`
	TriggerType                   pipelineConfig.TriggerType             `json:"triggerType,omitempty" validate:"oneof=AUTOMATIC MANUAL"`
	Name                          string                                 `json:"name,omitempty" validate:"name-component,max=50"` //pipelineName
	Strategies                    []Strategy                             `json:"strategies,omitempty"`
	Namespace                     string                                 `json:"namespace,omitempty"` //namespace
	AppWorkflowId                 int                                    `json:"appWorkflowId,omitempty" `
	DeploymentTemplate            chartRepoRepository.DeploymentStrategy `json:"deploymentTemplate,omitempty"` //
	PreStage                      CdStage                                `json:"preStage,omitempty"`
	PostStage                     CdStage                                `json:"postStage,omitempty"`
	PreStageConfigMapSecretNames  PreStageConfigMapSecretNames           `json:"preStageConfigMapSecretNames,omitempty"`
	PostStageConfigMapSecretNames PostStageConfigMapSecretNames          `json:"postStageConfigMapSecretNames,omitempty"`
	RunPreStageInEnv              bool                                   `json:"runPreStageInEnv,omitempty"`
	RunPostStageInEnv             bool                                   `json:"runPostStageInEnv,omitempty"`
	CdArgoSetup                   bool                                   `json:"isClusterCdActive"`
	ParentPipelineId              int                                    `json:"parentPipelineId"`
	ParentPipelineType            string                                 `json:"parentPipelineType"`
	DeploymentAppType             string                                 `json:"deploymentAppType"`
	AppName                       string                                 `json:"appName"`
	DeploymentAppDeleteRequest    bool                                   `json:"deploymentAppDeleteRequest"`
	DeploymentAppCreated          bool                                   `json:"deploymentAppCreated"`
	AppId                         int                                    `json:"appId"`
	TeamId                        int                                    `json:"-"`
	EnvironmentIdentifier         string                                 `json:"-" `
	IsVirtualEnvironment          bool                                   `json:"isVirtualEnvironment"`
	HelmPackageName               string                                 `json:"helmPackageName"`
	ChartName                     string                                 `json:"chartName"`
	ChartBaseVersion              string                                 `json:"chartBaseVersion"`
	ContainerRegistryId           int                                    `json:"containerRegistryId"`
	RepoUrl                       string                                 `json:"repoUrl"`
	ManifestStorageType           string                                 `json:"manifestStorageType"`
	PreDeployStage                *bean.PipelineStageDto                 `json:"preDeployStage,omitempty"`
	PostDeployStage               *bean.PipelineStageDto                 `json:"postDeployStage,omitempty"`
	SourceToNewPipelineId         map[int]int                            `json:"sourceToNewPipelineId,omitempty"`
	RefPipelineId                 int                                    `json:"refPipelineId,omitempty"`
	ExternalCiPipelineId          int                                    `json:"externalCiPipelineId,omitempty"`
	CustomTagObject               *CustomTagData                         `json:"customTag"`
	CustomTagStage                *repository.PipelineStageType          `json:"customTagStage"`
	EnableCustomTag               bool                                   `json:"enableCustomTag"`
	SwitchFromCiPipelineId        int                                    `json:"switchFromCiPipelineId"`
	CDPipelineAddType             CDPipelineAddType                      `json:"addType"`
	ChildPipelineId               int                                    `json:"childPipelineId"`
	IsDigestEnforcedForPipeline   bool                                   `json:"isDigestEnforcedForPipeline"`
	IsDigestEnforcedForEnv        bool                                   `json:"isDigestEnforcedForEnv"`
}

type CDPipelineAddType string

const (
	SEQUENTIAL CDPipelineAddType = "SEQUENTIAL"
	PARALLEL   CDPipelineAddType = "PARALLEL"
)

func (cdpipelineConfig *CDPipelineConfigObject) IsSwitchCiPipelineRequest() bool {
	return cdpipelineConfig.SwitchFromCiPipelineId > 0 && cdpipelineConfig.AppWorkflowId > 0
}

type PreStageConfigMapSecretNames struct {
	ConfigMaps []string `json:"configMaps"`
	Secrets    []string `json:"secrets"`
}

type PostStageConfigMapSecretNames struct {
	ConfigMaps []string `json:"configMaps"`
	Secrets    []string `json:"secrets"`
}

type CdStage struct {
	TriggerType pipelineConfig.TriggerType `json:"triggerType,omitempty"`
	Name        string                     `json:"name,omitempty"`
	Status      string                     `json:"status,omitempty"`
	Config      string                     `json:"config,omitempty"`
	//CdWorkflowId       int                        `json:"cdWorkflowId,omitempty" validate:"number"`
	//CdWorkflowRunnerId int                        `json:"cdWorkflowRunnerId,omitempty" validate:"number"`
}

type Strategy struct {
	DeploymentTemplate chartRepoRepository.DeploymentStrategy `json:"deploymentTemplate,omitempty"` //
	Config             json.RawMessage                        `json:"config,omitempty" validate:"string"`
	Default            bool                                   `json:"default"`
}

type CdPipelines struct {
	Pipelines         []*CDPipelineConfigObject `json:"pipelines,omitempty" validate:"dive"`
	AppId             int                       `json:"appId,omitempty"  validate:"number,required" `
	UserId            int32                     `json:"-"`
	AppDeleteResponse *AppDeleteResponseDTO     `json:"deleteResponse,omitempty"`
}

type AppDeleteResponseDTO struct {
	DeleteInitiated  bool   `json:"deleteInitiated"`
	ClusterReachable bool   `json:"clusterReachable"`
	ClusterName      string `json:"clusterName"`
}

type CDPatchRequest struct {
	Pipeline         *CDPipelineConfigObject `json:"pipeline,omitempty"`
	AppId            int                     `json:"appId,omitempty"`
	Action           CdPatchAction           `json:"action,omitempty"`
	UserId           int32                   `json:"-"`
	ForceDelete      bool                    `json:"-"`
	NonCascadeDelete bool                    `json:"-"`
}

type CdPatchAction int

const (
	CD_CREATE CdPatchAction = iota
	CD_DELETE               //delete this pipeline
	CD_UPDATE
	CD_DELETE_PARTIAL // Partially delete means it will only delete ACD app
)

type DeploymentAppTypeChangeRequest struct {
	EnvId                 int            `json:"envId,omitempty" validate:"required"`
	DesiredDeploymentType DeploymentType `json:"desiredDeploymentType,omitempty" validate:"required"`
	ExcludeApps           []int          `json:"excludeApps"`
	IncludeApps           []int          `json:"includeApps"`
	AutoTriggerDeployment bool           `json:"autoTriggerDeployment"`
	UserId                int32          `json:"-"`
}

type DeploymentChangeStatus struct {
	Id      int    `json:"id,omitempty"`
	AppId   int    `json:"appId,omitempty"`
	AppName string `json:"appName,omitempty"`
	EnvId   int    `json:"envId,omitempty"`
	EnvName string `json:"envName,omitempty"`
	Error   string `json:"error,omitempty"`
	Status  Status `json:"status,omitempty"`
}

type DeploymentAppTypeChangeResponse struct {
	EnvId                 int                       `json:"envId,omitempty"`
	DesiredDeploymentType DeploymentType            `json:"desiredDeploymentType,omitempty"`
	SuccessfulPipelines   []*DeploymentChangeStatus `json:"successfulPipelines"`
	FailedPipelines       []*DeploymentChangeStatus `json:"failedPipelines"`
	TriggeredPipelines    []*CdPipelineTrigger      `json:"-"` // Disabling auto-trigger until bulk trigger API is fixed
}

type CdPipelineTrigger struct {
	CiArtifactId int `json:"ciArtifactId"`
	PipelineId   int `json:"pipelineId"`
}

type DeploymentType = string

const (
	Helm                    DeploymentType = "helm"
	ArgoCd                  DeploymentType = "argo_cd"
	ManifestDownload        DeploymentType = "manifest_download"
	GitOpsWithoutDeployment DeploymentType = "git_ops_without_deployment"
)

func IsAcdApp(deploymentType string) bool {
	return deploymentType == ArgoCd
}

func IsHelmApp(deploymentType string) bool {
	return deploymentType == Helm
}

type Status string

const (
	Success         Status = "Success"
	Failed          Status = "Failed"
	INITIATED       Status = "Migration initiated"
	NOT_YET_DELETED Status = "Not yet deleted"
)

const RELEASE_NOT_EXIST = "release not exist"
const NOT_FOUND = "not found"

func (a CdPatchAction) String() string {
	return [...]string{"CREATE", "DELETE", "CD_UPDATE"}[a]
}

type CDPipelineViewObject struct {
	Id                 int                         `json:"id"`
	PipelineCounter    int                         `json:"pipelineCounter"`
	Environment        string                      `json:"environment"`
	Downstream         []int                       `json:"downstream"` //PipelineCounter of downstream
	Status             string                      `json:"status"`
	Message            string                      `json:"message"`
	ProgressText       string                      `json:"progress_text"`
	PipelineType       pipelineConfig.PipelineType `json:"pipelineType"`
	GitDiffUrl         string                      `json:"git_diff_url"`
	PipelineHistoryUrl string                      `json:"pipeline_history_url"` //remove
	Rollback           Rollback                    `json:"rollback"`
	Name               string                      `json:"-"`
	CDSourceObject
}

//Trigger materials in different API

type Rollback struct {
	url     string `json:"url"` //remove
	enabled bool   `json:"enabled"`
}

type CiArtifactBean struct {
	Id                            int                       `json:"id"`
	Image                         string                    `json:"image,notnull"`
	ImageDigest                   string                    `json:"image_digest,notnull"`
	MaterialInfo                  json.RawMessage           `json:"material_info"` //git material metadata json array string
	DataSource                    string                    `json:"data_source,notnull"`
	DeployedTime                  string                    `json:"deployed_time"`
	Deployed                      bool                      `json:"deployed,notnull"`
	Latest                        bool                      `json:"latest,notnull"`
	LastSuccessfulTriggerOnParent bool                      `json:"lastSuccessfulTriggerOnParent,notnull"`
	RunningOnParentCd             bool                      `json:"runningOnParentCd,omitempty"`
	IsVulnerable                  bool                      `json:"vulnerable,notnull"`
	ScanEnabled                   bool                      `json:"scanEnabled,notnull"`
	Scanned                       bool                      `json:"scanned,notnull"`
	WfrId                         int                       `json:"wfrId"`
	DeployedBy                    string                    `json:"deployedBy"`
	CiConfigureSourceType         pipelineConfig.SourceType `json:"ciConfigureSourceType"`
	CiConfigureSourceValue        string                    `json:"ciConfigureSourceValue"`
	ImageReleaseTags              []*repository2.ImageTag   `json:"imageReleaseTags"`
	ImageComment                  *repository2.ImageComment `json:"imageComment"`
	CreatedTime                   string                    `json:"createdTime"`
	ExternalCiPipelineId          int                       `json:"-"`
	ParentCiArtifact              int                       `json:"-"`
	CiWorkflowId                  int                       `json:"-"`
	RegistryType                  string                    `json:"registryType"`
	RegistryName                  string                    `json:"registryName"`
	CiPipelineId                  int                       `json:"-"`
	CredentialsSourceType         string                    `json:"-"`
	CredentialsSourceValue        string                    `json:"-"`
}

type CiArtifactResponse struct {
	//AppId           int      `json:"app_id"`
	CdPipelineId               int              `json:"cd_pipeline_id,notnull"`
	LatestWfArtifactId         int              `json:"latest_wf_artifact_id"`
	LatestWfArtifactStatus     string           `json:"latest_wf_artifact_status"`
	CiArtifacts                []CiArtifactBean `json:"ci_artifacts,notnull"`
	TagsEditable               bool             `json:"tagsEditable"`
	AppReleaseTagNames         []string         `json:"appReleaseTagNames"` //unique list of tags exists in the app
	HideImageTaggingHardDelete bool             `json:"hideImageTaggingHardDelete"`
	TotalCount                 int              `json:"totalCount"`
}

type AppLabelsDto struct {
	Labels []*Label `json:"labels" validate:"dive"`
	AppId  int      `json:"appId"`
	UserId int32    `json:"-"`
}

type AppLabelDto struct {
	Key       string `json:"key,notnull"`
	Value     string `json:"value,notnull"`
	Propagate bool   `json:"propagate,notnull"`
	AppId     int    `json:"appId,omitempty"`
	UserId    int32  `json:"-"`
}

type Label struct {
	Key       string `json:"key" validate:"required"`
	Value     string `json:"value"` // intentionally not added required tag as tag can be added without value
	Propagate bool   `json:"propagate"`
}

type AppMetaInfoDto struct {
	AppId       int                            `json:"appId"`
	AppName     string                         `json:"appName"`
	Description string                         `json:"description"`
	ProjectId   int                            `json:"projectId"`
	ProjectName string                         `json:"projectName"`
	CreatedBy   string                         `json:"createdBy"`
	CreatedOn   time.Time                      `json:"createdOn"`
	Active      bool                           `json:"active,notnull"`
	Labels      []*Label                       `json:"labels"`
	Note        *bean2.GenericNoteResponseBean `json:"note"`
	UserId      int32                          `json:"-"`
	//below field is only valid for helm apps
	ChartUsed    *ChartUsedDto         `json:"chartUsed,omitempty"`
	GitMaterials []*GitMaterialMetaDto `json:"gitMaterials,omitempty"`
}

type GitMaterialMetaDto struct {
	DisplayName    string `json:"displayName"`
	RedirectionUrl string `json:"redirectionUrl"` // here we are converting ssh urls to https for redirection at FE
	OriginalUrl    string `json:"originalUrl"`
}

type ChartUsedDto struct {
	AppStoreChartName  string `json:"appStoreChartName,omitempty"`
	AppStoreChartId    int    `json:"appStoreChartId,omitempty"`
	AppStoreAppName    string `json:"appStoreAppName,omitempty"`
	AppStoreAppVersion string `json:"appStoreAppVersion,omitempty"`
	ChartAvatar        string `json:"chartAvatar,omitempty"`
}

type AppLabelsJsonForDeployment struct {
	Labels map[string]string `json:"appLabels"`
}

type UpdateProjectBulkAppsRequest struct {
	AppIds []int `json:"appIds"`
	TeamId int   `json:"teamId"`
	UserId int32 `json:"-"`
}

type CdBulkAction int

const (
	CD_BULK_DELETE CdBulkAction = iota
)

type CdBulkActionRequestDto struct {
	Action        CdBulkAction `json:"action"`
	EnvIds        []int        `json:"envIds"`
	AppIds        []int        `json:"appIds"`
	ProjectIds    []int        `json:"projectIds"`
	ForceDelete   bool         `json:"forceDelete"`
	CascadeDelete bool         `json:"cascadeDelete"`
	UserId        int32        `json:"-"`
}

type CdBulkActionResponseDto struct {
	PipelineName    string `json:"pipelineName"`
	AppName         string `json:"appName"`
	EnvironmentName string `json:"environmentName"`
	DeletionResult  string `json:"deletionResult,omitempty"`
}

type SchemaObject struct {
	Description string      `json:"description"`
	DataType    string      `json:"dataType"`
	Example     string      `json:"example"`
	Optional    bool        `json:"optional"`
	Child       interface{} `json:"child"`
}

type PayloadOptionObject struct {
	Key        string   `json:"key"`
	PayloadKey []string `json:"payloadKey"`
	Label      string   `json:"label"`
	Mandatory  bool     `json:"mandatory"`
}

type ResponseSchemaObject struct {
	Description ResponseDescriptionSchemaObject `json:"description"`
	Code        string                          `json:"code"`
}

type ResponseDescriptionSchemaObject struct {
	Description  string                 `json:"description,omitempty"`
	ExampleValue ExampleValueDto        `json:"exampleValue,omitempty"`
	Schema       map[string]interface{} `json:"schema,omitempty"`
}

type ErrorDto struct {
	Code        int    `json:"code"`
	UserMessage string `json:"userMessage"`
}

type ExampleValueDto struct {
	Code   int        `json:"code,omitempty"`
	Errors []ErrorDto `json:"errors,omitempty"`
	Result string     `json:"result,omitempty"`
	Status string     `json:"status,omitempty"`
}

type ManifestStorage = string

const (
	ManifestStorageGit ManifestStorage = "git"
)

func IsGitStorage(storageType string) bool {
	return storageType == ManifestStorageGit
}

const CustomAutoScalingEnabledPathKey = "CUSTOM_AUTOSCALING_ENABLED_PATH"
const CustomAutoscalingReplicaCountPathKey = "CUSTOM_AUTOSCALING_REPLICA_COUNT_PATH"
const CustomAutoscalingMinPathKey = "CUSTOM_AUTOSCALING_MIN_PATH"
const CustomAutoscalingMaxPathKey = "CUSTOM_AUTOSCALING_MAX_PATH"
