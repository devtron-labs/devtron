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
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
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
}

type CreateAppDTO struct {
	Id         int            `json:"id,omitempty" validate:"number"`
	AppName    string         `json:"appName" validate:"name-component,max=100"`
	UserId     int32          `json:"-"` //not exposed to UI
	Material   []*GitMaterial `json:"material" validate:"dive,min=1"`
	TeamId     int            `json:"teamId,omitempty" validate:"number,required"`
	TemplateId int            `json:"templateId"`
	AppLabels  []*Label       `json:"labels,omitempty" validate:"dive"`
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
	Name            string `json:"name,omitempty" ` //not null, //default format pipelineGroup.AppName + "-" + inputMaterial.Name,
	Url             string `json:"url,omitempty"`   //url of git repo
	Id              int    `json:"id,omitempty" validate:"number"`
	GitProviderId   int    `json:"gitProviderId,omitempty" validate:"gt=0"`
	CheckoutPath    string `json:"checkoutPath" validate:"checkout-path-component"`
	FetchSubmodules bool   `json:"fetchSubmodules"`
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
}

type CiPipeline struct {
	IsManual                 bool              `json:"isManual"`
	DockerArgs               map[string]string `json:"dockerArgs"`
	IsExternal               bool              `json:"isExternal"`
	ParentCiPipeline         int               `json:"parentCiPipeline"`
	ParentAppId              int               `json:"parentAppId"`
	ExternalCiConfig         ExternalCiConfig  `json:"externalCiConfig"`
	CiMaterial               []*CiMaterial     `json:"ciMaterial,omitempty" validate:"dive,min=1"`
	Name                     string            `json:"name,omitempty" validate:"name-component,max=100"` //name suffix of corresponding pipeline. required, unique, validation corresponding to gocd pipelineName will be applicable
	Id                       int               `json:"id,omitempty" `
	Version                  string            `json:"version,omitempty"` //matchIf token version in gocd . used for update request
	Active                   bool              `json:"active,omitempty"`  //pipeline is active or not
	Deleted                  bool              `json:"deleted,omitempty"`
	BeforeDockerBuild        []*Task           `json:"beforeDockerBuild,omitempty" validate:"dive"`
	AfterDockerBuild         []*Task           `json:"afterDockerBuild,omitempty" validate:"dive"`
	BeforeDockerBuildScripts []*CiScript       `json:"beforeDockerBuildScripts,omitempty" validate:"dive"`
	AfterDockerBuildScripts  []*CiScript       `json:"afterDockerBuildScripts,omitempty" validate:"dive"`
	LinkedCount              int               `json:"linkedCount"`
	PipelineType             PipelineType      `json:"pipelineType,omitempty"`
	ScanEnabled              bool              `json:"scanEnabled,notnull"`
	AppWorkflowId            int               `json:"appWorkflowId,omitempty"`
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
	Id         int    `json:"id"`
	WebhookUrl string `json:"webhookUrl"`
	Payload    string `json:"payload"`
	AccessKey  string `json:"accessKey"`
}

//-------------------
type PatchAction int
type PipelineType string

const (
	CREATE        PatchAction = iota
	UPDATE_SOURCE             //update value of SourceTypeConfig
	DELETE                    //delete this pipeline
	//DEACTIVATE     //pause/deactivate this pipeline
)

const (
	NORMAL   PipelineType = "NORMAL"
	LINKED   PipelineType = "LINKED"
	EXTERNAL PipelineType = "EXTERNAL"
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

func (a PatchAction) String() string {
	return [...]string{"CREATE", "UPDATE_SOURCE", "DELETE", "DEACTIVATE"}[a]

}

//----------------
type CiPatchRequest struct {
	CiPipeline    *CiPipeline `json:"ciPipeline"`
	AppId         int         `json:"appId,omitempty"`
	Action        PatchAction `json:"action"`
	AppWorkflowId int         `json:"appWorkflowId,omitempty"`
	UserId        int32       `json:"-"`
}

type GitCiTriggerRequest struct {
	CiPipelineMaterial CiPipelineMaterial `json:"ciPipelineMaterial" validate:"required"`
	TriggeredBy        int32              `json:"triggeredBy"`
}

type GitCommit struct {
	Commit                 string //git hash
	Author                 string
	Date                   time.Time
	Message                string
	Changes                []string
	WebhookData            *WebhookData
	GitRepoUrl             string
	GitRepoName            string
	CiConfigureSourceType  pipelineConfig.SourceType
	CiConfigureSourceValue string
}

type WebhookData struct {
	Id              int               `json:"id"`
	EventActionType string            `json:"eventActionType"`
	Data            map[string]string `json:"data"`
}

type SourceType string

type CiPipelineMaterial struct {
	Id            int       `json:"Id"`
	GitMaterialId int       `json:"GitMaterialId"`
	Type          string    `json:"Type"`
	Value         string    `json:"Value"`
	Active        bool      `json:"Active"`
	GitCommit     GitCommit `json:"GitCommit"`
	GitTag        string    `json:"GitTag"`
}

type CiTriggerRequest struct {
	PipelineId         int                  `json:"pipelineId"`
	CiPipelineMaterial []CiPipelineMaterial `json:"ciPipelineMaterials" validate:"required"`
	TriggeredBy        int32                `json:"triggeredBy"`
	InvalidateCache    bool                 `json:"invalidateCache"`
}

type CiTrigger struct {
	CiMaterialId int    `json:"ciMaterialId"`
	CommitHash   string `json:"commitHash"`
}

type Material struct {
	GitMaterialId int    `json:"gitMaterialId"`
	MaterialName  string `json:"materialName"`
}

type CiConfigRequest struct {
	Id                int                `json:"id,omitempty" validate:"number"` //ciTemplateId
	AppId             int                `json:"appId,omitempty" validate:"required,number"`
	DockerRegistry    string             `json:"dockerRegistry,omitempty" `  //repo id example ecr mapped one-one with gocd registry entry
	DockerRepository  string             `json:"dockerRepository,omitempty"` // example test-app-1 which is inside ecr
	DockerBuildConfig *DockerBuildConfig `json:"dockerBuildConfig,omitempty" validate:"required,dive"`
	CiPipelines       []*CiPipeline      `json:"ciPipelines,omitempty" validate:"dive"` //a pipeline will be built for each ciMaterial
	AppName           string             `json:"appName,omitempty"`
	Version           string             `json:"version,omitempty"` //gocd etag used for edit purpose
	DockerRegistryUrl string             `json:"-"`
	CiTemplateName    string             `json:"-"`
	UserId            int32              `json:"-"`
	Materials         []Material         `json:"materials"`
	AppWorkflowId     int                `json:"appWorkflowId,omitempty"`
	BeforeDockerBuild []*Task            `json:"beforeDockerBuild,omitempty" validate:"dive"`
	AfterDockerBuild  []*Task            `json:"afterDockerBuild,omitempty" validate:"dive"`
	ScanEnabled       bool               `json:"scanEnabled,notnull"`
}

type TestExecutorImageProperties struct {
	ImageName string `json:"imageName,omitempty"`
	Arg       string `json:"arg,omitempty"`
	ReportDir string `json:"reportDir,omitempty"`
}

type DockerBuildConfig struct {
	GitMaterialId  int               `json:"gitMaterialId,omitempty" validate:"required"`
	DockerfilePath string            `json:"dockerfileRelativePath,omitempty" validate:"required"`
	Args           map[string]string `json:"args,omitempty"`
	//Name Tag DockerfilePath RepoUrl
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

//used for automated unit and integration test
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

//set of unique attributes which corresponds to a cluster
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

//--------- cd related struct ---------
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
	Id                            int                               `json:"id,omitempty"  validate:"number" `
	EnvironmentId                 int                               `json:"environmentId,omitempty"  validate:"number,required" `
	EnvironmentName               string                            `json:"environmentName,omitempty" `
	CiPipelineId                  int                               `json:"ciPipelineId,omitempty" validate:"number,required"`
	TriggerType                   pipelineConfig.TriggerType        `json:"triggerType,omitempty" validate:"oneof=AUTOMATIC MANUAL"`
	Name                          string                            `json:"name,omitempty" validate:"name-component,max=50"` //pipelineName
	Strategies                    []Strategy                        `json:"strategies,omitempty"`
	Namespace                     string                            `json:"namespace,omitempty" validate:"name-component,max=50"` //namespace
	AppWorkflowId                 int                               `json:"appWorkflowId,omitempty" `
	DeploymentTemplate            pipelineConfig.DeploymentTemplate `json:"deploymentTemplate,omitempty" validate:"oneof=BLUE-GREEN ROLLING CANARY RECREATE"` //
	PreStage                      CdStage                           `json:"preStage"`
	PostStage                     CdStage                           `json:"postStage"`
	PreStageConfigMapSecretNames  PreStageConfigMapSecretNames      `json:"preStageConfigMapSecretNames"`
	PostStageConfigMapSecretNames PostStageConfigMapSecretNames     `json:"postStageConfigMapSecretNames"`
	RunPreStageInEnv              bool                              `json:"runPreStageInEnv"`
	RunPostStageInEnv             bool                              `json:"runPostStageInEnv"`
	CdArgoSetup                   bool                              `json:"isClusterCdActive"`
	//Downstream         []int                             `json:"downstream"` //PipelineCounter of downstream	(for future reference only)
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
	DeploymentTemplate pipelineConfig.DeploymentTemplate `json:"deploymentTemplate,omitempty" validate:"oneof=BLUE-GREEN ROLLING CANARY RECREATE"` //
	Config             json.RawMessage                   `json:"config,omitempty" validate:"string"`
	Default            bool                              `json:"default"`
}

type CdPipelines struct {
	Pipelines []*CDPipelineConfigObject `json:"pipelines,omitempty" validate:"dive"`
	AppId     int                       `json:"appId,omitempty"  validate:"number,required" `
	UserId    int32                     `json:"-"`
}

type CDPatchRequest struct {
	Pipeline    *CDPipelineConfigObject `json:"pipeline,omitempty"`
	AppId       int                     `json:"appId,omitempty"`
	Action      CdPatchAction           `json:"action,omitempty"`
	UserId      int32                   `json:"-"`
	ForceDelete bool                    `json:"-"`
}

type CdPatchAction int

const (
	CD_CREATE CdPatchAction = iota
	CD_DELETE               //delete this pipeline
	CD_UPDATE
)

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
	Id           int             `json:"id"`
	Image        string          `json:"image,notnull"`
	ImageDigest  string          `json:"image_digest,notnull"`
	MaterialInfo json.RawMessage `json:"material_info"` //git material metadata json array string
	DataSource   string          `json:"data_source,notnull"`
	DeployedTime string          `json:"deployed_time"`
	Deployed     bool            `json:"deployed,notnull"`
	Latest       bool            `json:"latest,notnull"`
	IsVulnerable bool            `json:"vulnerable,notnull"`
	ScanEnabled  bool            `json:"scanEnabled,notnull"`
	Scanned      bool            `json:"scanned,notnull"`
}

type CiArtifactResponse struct {
	//AppId           int      `json:"app_id"`
	CdPipelineId int              `json:"cd_pipeline_id,notnull"`
	CiArtifacts  []CiArtifactBean `json:"ci_artifacts,notnull"`
}

type AppLabelsDto struct {
	Labels []*Label `json:"labels" validate:"dive"`
	AppId  int      `json:"appId"`
	UserId int32    `json:"-"`
}

type AppLabelDto struct {
	Key    string `json:"key,notnull"`
	Value  string `json:"value,notnull"`
	AppId  int    `json:"appId,omitempty"`
	UserId int32  `json:"-"`
}

type Label struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type AppMetaInfoDto struct {
	AppId       int       `json:"appId"`
	AppName     string    `json:"appName"`
	ProjectId   int       `json:"projectId"`
	ProjectName string    `json:"projectName"`
	CreatedBy   string    `json:"createdBy"`
	CreatedOn   time.Time `json:"createdOn"`
	Active      bool      `json:"active,notnull"`
	Labels      []*Label  `json:"labels"`
	UserId      int32     `json:"-"`
}
