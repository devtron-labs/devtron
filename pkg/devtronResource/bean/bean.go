package bean

import (
	"github.com/devtron-labs/devtron/api/bean"
	"time"
)

// TODO: rename this and apiBean, InternalBean files as resource object bean files

type DevtronResourceObjectDescriptorBean struct {
	Kind         string                       `json:"kind,omitempty"`
	SubKind      string                       `json:"subKind,omitempty"`
	Version      string                       `json:"version,omitempty"`
	OldObjectId  int                          `json:"id,omitempty"`   // here at FE we are still calling this id since id is used everywhere w.r.t resource's own tables
	Name         string                       `json:"name,omitempty"` // Name will only be used as a metadata for resource object json. Name can not be privileged for getting repository.DevtronResourceObject
	SchemaId     int                          `json:"schemaId"`
	Identifier   string                       `json:"identifier,omitempty"` // Identifier should not be used in code anywhere only just a user-friendly way to get repository.DevtronResourceObject
	UIComponents []DevtronResourceUIComponent `json:"-"`
	Id           int                          `json:"-"` //this is the field which holds the resourceObjectId i.e. id in devtron_resource_object table. Have not exposed this and taking this value from oldObjectId to maintain backward compatibility
	IdType       IdType                       `json:"-"` // internal , for release and release-track IdType will be repository.DevtronResourceObject.(Id)
	UserId       int32                        `json:"-"`
}

// GetResourceIdByIdType will return the resource id based on id type
func (reqBean *DevtronResourceObjectDescriptorBean) GetResourceIdByIdType() int {
	if reqBean.IdType == OldObjectId {
		return reqBean.OldObjectId
	} else if reqBean.IdType == ResourceObjectIdType {
		return reqBean.Id
	}
	return 0
}

// not used anymore, TODO: remove if testing passed and not required anymore
type devtronResourceObjectBean struct {
	*DevtronResourceObjectDescriptorBean
	Schema            string                           `json:"schema,omitempty"`
	ObjectData        string                           `json:"objectData"`
	Dependencies      []*DevtronResourceDependencyBean `json:"dependencies"`
	ChildDependencies []*DevtronResourceDependencyBean `json:"childDependencies"`
	Overview          *ResourceOverview                `json:"overview,omitempty"`
	ConfigStatus      *ConfigStatus                    `json:"configStatus,omitempty"`
	ParentConfig      *ResourceIdentifier              `json:"parentConfig,omitempty"`
	PatchQuery        []PatchQuery                     `json:"query,omitempty"`
	DependencyInfo    *DependencyInfo                  `json:"DependencyInfo,omitempty"`
}

type DependencyConfigOptions[T any] struct {
	Id              int                    `json:"id"`
	Identifier      string                 `json:"identifier"`
	ResourceKind    DevtronResourceKind    `json:"resourceKind"`
	ResourceVersion DevtronResourceVersion `json:"resourceVersion"`
	Data            T                      `json:"data,omitempty"`
}

type DevtronResourceDependencyPatchAPIBean struct {
	*DevtronResourceObjectDescriptorBean
	DependencyPatch []*DependencyPatchBean `json:"dependencyPatch,omitempty"`
}

type DependencyPatchBean struct {
	PatchQuery     []PatchQuery    `json:"query,omitempty"`
	DependencyInfo *DependencyInfo `json:"dependencyInfo,omitempty"`
}

type ResourceOverview struct {
	Description     string            `json:"description,omitempty"`
	Note            *NoteBean         `json:"note,omitempty"`
	ReleaseVersion  string            `json:"releaseVersion,omitempty"`
	CreatedBy       *UserSchema       `json:"createdBy,omitempty"`
	CreatedOn       time.Time         `json:"createdOn,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
	FirstReleasedOn time.Time         `json:"firstReleasedOn,omitempty"`
	RunSource       *RunSource        `json:"runSource,omitempty"`
}

type DevtronResourceTaskRunBean struct {
	*DevtronResourceObjectDescriptorBean
	Overview *ResourceOverview `json:"overview,omitempty"`
	Action   []*TaskRunAction  `json:"action,omitempty"`
}

type RunSource struct {
	Id                      int               `json:"id,omitempty"`
	IdType                  IdType            `json:"idType,omitempty"`
	DevtronResourceId       int               `json:"devtronResourceId,omitempty"`
	DevtronResourceSchemaId int               `json:"devtronResourceSchemaId,omitempty"`
	DependencyDetail        *DependencyDetail `json:"dependencyDetail,omitempty"`
}

type DependencyDetail struct {
	Id                      int    `json:"id,omitempty"`
	IdType                  IdType `json:"idType,omitempty"`
	DevtronResourceId       int    `json:"devtronResourceId,omitempty"`
	DevtronResourceSchemaId int    `json:"devtronResourceSchemaId,omitempty"`
}

type TaskRunAction struct {
	TaskType           TaskType `json:"taskType,omitempty"`
	CdWorkflowRunnerId int      `json:"cdWfrId,omitempty"`
}

type NoteBean struct {
	Value     string      `json:"value"`
	UpdatedOn time.Time   `json:"updatedOn"`
	UpdatedBy *UserSchema `json:"updatedBy"`
}

type ResourceIdentifier struct {
	Id         int    `json:"id"`
	Identifier string `json:"identifier,omitempty"` // Identifier should not be used in code anywhere only just a user-friendly way to get repository.DevtronResourceObject
	DevtronResourceTypeReq
}

type DevtronResourceTypeReq struct {
	ResourceKind    DevtronResourceKind    `json:"resourceKind"`
	ResourceSubKind DevtronResourceKind    `json:"-"` // ResourceSubKind will be derived internally from the given ResourceKind
	ResourceVersion DevtronResourceVersion `json:"resourceVersion"`
	SchemaId        int                    `json:"-"`
}

type PatchQuery struct {
	Operation PatchQueryOperation `json:"op"` // default is replace
	Path      PatchQueryPath      `json:"path" validate:"required"`
	Value     interface{}         `json:"value"`
}

type DependencyInfo struct {
	Id              int                    `json:"id"`
	Identifier      string                 `json:"identifier"`
	ResourceKind    DevtronResourceKind    `json:"resourceKind,omitempty"`
	ResourceVersion DevtronResourceVersion `json:"resourceVersion,omitempty"`
}

const (
	AllIdentifierQueryString = "*"
	IdentifierQueryString    = "identifier"
	IdQueryString            = "id"
)

type ResourceParentData struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type DevtronResourceDependencyBean struct {
	*DevtronResourceTypeReq
	OldObjectId             int                              `json:"id"` //have both oldObjectId and resourceObjectId
	DevtronResourceId       int                              `json:"devtronResourceId"`
	DevtronResourceSchemaId int                              `json:"devtronResourceSchemaId"`
	DependentOnIndex        int                              `json:"dependentOnIndex,omitempty"`
	DependentOnIndexes      []int                            `json:"dependentOnIndexes,omitempty"`
	DependentOnParentIndex  int                              `json:"dependentOnParentIndex,omitempty"`
	TypeOfDependency        DevtronResourceDependencyType    `json:"typeOfDependency"`
	Index                   int                              `json:"index"`
	Dependencies            []*DevtronResourceDependencyBean `json:"dependencies,omitempty"`
	Metadata                interface{}                      `json:"metadata,omitempty"`
	IdType                  IdType                           `json:"idType,omitempty"`
	Identifier              string                           `json:"identifier,omitempty"`
	Config                  *DependencyConfigBean            `json:"config,omitempty"`
	ChildObjects            []*ChildObject                   `json:"childObjects,omitempty"`
	ChildInheritance        []*ChildInheritance              `json:"childInheritance,omitempty"` // right now being used internally for release, cd pipeline is being inherited.
}

func NewDevtronResourceDependencyBean() *DevtronResourceDependencyBean {
	return &DevtronResourceDependencyBean{}
}

func (d *DevtronResourceDependencyBean) WithOldObjectId(id int) *DevtronResourceDependencyBean {
	d.OldObjectId = id
	return d
}

func (d *DevtronResourceDependencyBean) WithDevtronResourceId(dtResourceId int) *DevtronResourceDependencyBean {
	d.DevtronResourceId = dtResourceId
	return d
}

func (d *DevtronResourceDependencyBean) WithDevtronResourceSchemaId(dtResourceSchemaId int) *DevtronResourceDependencyBean {
	d.DevtronResourceSchemaId = dtResourceSchemaId
	return d
}

func (d *DevtronResourceDependencyBean) WithDependentOnIndex(dependentOnIndex int) *DevtronResourceDependencyBean {
	d.DependentOnIndex = dependentOnIndex
	return d
}

func (d *DevtronResourceDependencyBean) WithDependentOnIndexes(dependentOnIndexes ...int) *DevtronResourceDependencyBean {
	d.DependentOnIndexes = append(d.DependentOnIndexes, dependentOnIndexes...)
	return d
}

func (d *DevtronResourceDependencyBean) WithTypeOfDependency(typeOfDependency DevtronResourceDependencyType) *DevtronResourceDependencyBean {
	d.TypeOfDependency = typeOfDependency
	return d
}

func (d *DevtronResourceDependencyBean) WithDependentOnParentIndex(dependentOnParentIndex int) *DevtronResourceDependencyBean {
	d.DependentOnParentIndex = dependentOnParentIndex
	return d
}

func (d *DevtronResourceDependencyBean) WithIndex(index int) *DevtronResourceDependencyBean {
	d.Index = index
	return d
}

func (d *DevtronResourceDependencyBean) WithIdType(idType IdType) *DevtronResourceDependencyBean {
	d.IdType = idType
	return d
}

func (d *DevtronResourceDependencyBean) WithChildInheritance(childInheritance ...*ChildInheritance) *DevtronResourceDependencyBean {
	d.ChildInheritance = append(d.ChildInheritance, childInheritance...)
	return d
}

type DependencyFilterCondition struct {
	filterByTypes            []DevtronResourceDependencyType
	filterByIndexes          []int
	filterByDependentOnIndex int
	filterByIdAndSchemaId    []IdAndSchemaIdFilter
	fetchChildInheritance    bool
}

func (c *DependencyFilterCondition) GetFilterByTypes() []DevtronResourceDependencyType {
	if c == nil {
		return []DevtronResourceDependencyType{}
	}
	return c.filterByTypes
}

func (c *DependencyFilterCondition) GetFilterByIndexes() []int {
	if c == nil {
		return []int{}
	}
	return c.filterByIndexes
}

func (c *DependencyFilterCondition) GetFilterByDependentOnIndex() int {
	if c == nil {
		return 0
	}
	return c.filterByDependentOnIndex
}

func (c *DependencyFilterCondition) GetFilterByFilterByIdAndSchemaId() []IdAndSchemaIdFilter {
	if c == nil {
		return nil
	}
	return c.filterByIdAndSchemaId
}

func (c *DependencyFilterCondition) GetChildInheritance() bool {
	if c == nil {
		return false
	}
	return c.fetchChildInheritance
}

func NewDependencyFilterCondition() *DependencyFilterCondition {
	return &DependencyFilterCondition{}
}

func (c *DependencyFilterCondition) WithFilterByTypes(types ...DevtronResourceDependencyType) *DependencyFilterCondition {
	c.filterByTypes = append(c.filterByTypes, types...)
	return c
}

func (c *DependencyFilterCondition) WithFilterByIndexes(indexes ...int) *DependencyFilterCondition {
	c.filterByIndexes = append(c.filterByIndexes, indexes...)
	return c
}

func (c *DependencyFilterCondition) WithFilterByDependentOnIndex(dependentOnIndex int) *DependencyFilterCondition {
	c.filterByDependentOnIndex = dependentOnIndex
	return c
}

func (c *DependencyFilterCondition) WithChildInheritance() *DependencyFilterCondition {
	c.fetchChildInheritance = true
	return c
}

func (c *DependencyFilterCondition) WithFilterByIdAndSchemaId(ids []int, schemaId int) *DependencyFilterCondition {
	idAndSchemaIdFilters := make([]IdAndSchemaIdFilter, 0, len(ids))
	for _, id := range ids {
		idAndSchemaIdFilters = append(idAndSchemaIdFilters, IdAndSchemaIdFilter{Id: id, DevtronResourceSchemaId: schemaId})
	}
	c.filterByIdAndSchemaId = idAndSchemaIdFilters
	return c
}

type DependencyConfigBean struct {
	*DevtronAppDependencyConfig
}

type ChildObject struct {
	Data interface{}     `json:"data,omitempty"`
	Type ChildObjectType `json:"type,omitempty"`
}

type ChildInheritance struct {
	ResourceId int      `json:"resourceId"` // signifies devtron resource kind id , currently for application dependency - resourceid of Cdpipline resourceKind
	Selector   []string `json:"selector"`   // ["*"] means all
}

type CdPipelineEnvironment struct {
	Id                int    `json:"id,omitempty"`
	Name              string `json:"name,omitempty"`
	PipelineId        int    `json:"pipelineId,omitempty"`
	DeploymentAppType string `json:"deploymentAppType,omitempty"`
}

type ArtifactConfig struct {
	ArtifactId          int                                     `json:"artifactId"`
	Image               string                                  `json:"image"`
	RegistryType        string                                  `json:"registryType"`
	RegistryName        string                                  `json:"registryName"`
	CommitSource        []GitCommitData                         `json:"commitSource,omitempty"`
	SourceAppWorkflowId int                                     `json:"artifactSourceAppWorkflowId,omitempty"`
	SourceReleaseConfig *DtResourceObjectInternalDescriptorBean `json:"sourceReleaseConfiguration,omitempty"`
}

type GitCommitData struct {
	Author       string               `json:"author"`
	Branch       string               `json:"branch"`
	Message      string               `json:"message"`
	ModifiedTime string               `json:"modifiedTime"`
	Revision     string               `json:"revision"`
	Tag          string               `json:"tag"`
	Url          string               `json:"url"`
	WebhookData  *WebHookMaterialInfo `json:"webhookData,omitempty"`
}

type WebHookMaterialInfo struct {
	Id              int         `json:"id"`
	EventActionType string      `json:"eventActionType"`
	Data            interface{} `json:"data"`
}

type DevtronAppDependencyConfig struct {
	ArtifactConfig *ArtifactConfig `json:"artifactConfig"`
	CiWorkflowId   int             `json:"ciWorkflowId"`

	ReleaseInstruction string `json:"releaseInstruction,omitempty"`
}

// DependencyMetaDataBean is used internally to set the value of DevtronResourceDependencyBean.Metadata
type DependencyMetaDataBean struct {
	// set as DevtronResourceDependencyBean.Metadata if DevtronResourceDependencyBean.TypeOfDependency --> DevtronResourceApplication
	MapOfAppsMetadata map[int]interface{}
	// set as DevtronResourceDependencyBean.Metadata if DevtronResourceDependencyBean.TypeOfDependency --> DevtronResourceApplication
	MapOfCdPipelinesMetadata map[int]interface{}
}

type UpdateSchemaResponseBean struct {
	Message       string   `json:"message"`
	PathsToRemove []string `json:"pathsToRemove"`
}

type ConfigStatus struct {
	Status   ReleaseConfigStatus `json:"status"`
	Comment  string              `json:"comment,omitempty"`
	IsLocked bool                `json:"isLocked"`
}

type ReleaseConfigSchema struct {
	Lock    bool   `json:"lock"`
	Comment string `json:"comment"`
	Status  string `json:"status"`
}

type UserSchema struct {
	Id   int32  `json:"id"`
	Icon bool   `json:"icon"`
	Name string `json:"name"`
}

type DevtronResourceTaskExecutionBean struct {
	*DevtronResourceObjectDescriptorBean
	DryRun        bool      `json:"dryRun"`
	Tasks         []*Task   `json:"tasks" validate:"required"`
	TriggeredTime time.Time `json:"-"` // for internal use
}

type Task struct {
	AppId                int               `json:"appId" validate:"required"`
	CdWorkflowType       bean.WorkflowType `json:"cdWorkflowType" validate:"required"`
	DeploymentWithConfig string            `json:"deploymentWithConfig" validate:"required"`
	PipelineId           int               `json:"pipelineId" validate:"required"`
	LevelIndex           int               `json:"levelIndex" validate:"required"`
}

type TaskExecutionResponseBean struct {
	AppId         int    `json:"appId"`
	EnvId         int    `json:"envId"`
	IsVirtualEnv  bool   `json:"isVirtualEnv"`
	AppName       string `json:"appName"`
	EnvName       string `json:"envName"`
	Feasibility   error  `json:"feasibility"`
	TriggerStatus error  `json:"triggerStatus"`
}

// IdType is used for identifying nature of id stored in object json or to implement logics. As we are using devtron_resource_object for storing all resource types across
// devtron we also faced a problem where id of resource object will be unique across all resource types, but old resources are stored in different tables and their id value
// can be same leading to conflict. To avoid this we are using idType where oldObjectId denotes ids of old resources(of their diff tables) and resourceObjectId is the unique
// id got from the primary key of devtron_resource_object table.
type IdType string

const (
	ResourceObjectIdType IdType = "resourceObjectId"
	OldObjectId          IdType = "oldObjectId"
)

const (
	SchemaUpdateSuccessMessage = "Schema updated successfully."
	DryRunSuccessfullMessage   = "Dry run successful"
)

const (
	Enum                 = "enum"
	Required             = "required"
	Properties           = "properties"
	Items                = "items"
	AdditionalProperties = "additionalProperties"
)

type SuccessResponse struct {
	Success       bool   `json:"success"`
	UserMessage   string `json:"userMessage"`
	DetailMessage string `json:"detailMessage"`
}

type DevtronResourceDependencyType string

const (
	DevtronResourceDependencyTypeParent     DevtronResourceDependencyType = "parent"
	DevtronResourceDependencyTypeChild      DevtronResourceDependencyType = "child"
	DevtronResourceDependencyTypeUpstream   DevtronResourceDependencyType = "upstream"
	DevtronResourceDependencyTypeDownStream DevtronResourceDependencyType = "downstream"
	DevtronResourceDependencyTypeLevel      DevtronResourceDependencyType = "level"
)

func (n DevtronResourceDependencyType) ToString() string {
	return string(n)
}

type DevtronResourceSearchableKeyName string

const (
	DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME           DevtronResourceSearchableKeyName = "PROJECT_APP_NAME"
	DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME           DevtronResourceSearchableKeyName = "CLUSTER_ENV_NAME"
	DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV      DevtronResourceSearchableKeyName = "IS_ALL_PRODUCTION_ENV"
	DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH         DevtronResourceSearchableKeyName = "CI_PIPELINE_BRANCH"
	DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_TRIGGER_ACTION DevtronResourceSearchableKeyName = "CI_PIPELINE_TRIGGER_ACTION"
	DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID                     DevtronResourceSearchableKeyName = "APP_ID"
	DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID                     DevtronResourceSearchableKeyName = "ENV_ID"
	DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID                 DevtronResourceSearchableKeyName = "CLUSTER_ID"
	DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_ID                 DevtronResourceSearchableKeyName = "PROJECT_ID"
	DEVTRON_RESOURCE_SEARCHABLE_KEY_PIPELINE_ID                DevtronResourceSearchableKeyName = "PIPELINE_ID"
)

func (n DevtronResourceSearchableKeyName) ToString() string {
	return string(n)
}

type DevtronResourceKind string

const (
	DevtronResourceApplication        DevtronResourceKind = "application"
	DevtronResourceDevtronApplication DevtronResourceKind = "devtron-application"
	DevtronResourceHelmApplication    DevtronResourceKind = "helm-application"
	DevtronResourceCluster            DevtronResourceKind = "cluster"
	DevtronResourceJob                DevtronResourceKind = "job"
	DevtronResourceUser               DevtronResourceKind = "users"
	DevtronResourceCdPipeline         DevtronResourceKind = "cd-pipeline"
	DevtronResourceReleaseTrack       DevtronResourceKind = "release-track"
	DevtronResourceRelease            DevtronResourceKind = "release"
	DevtronResourceTaskRun            DevtronResourceKind = "task-run"

	DevtronResourceEnvironment DevtronResourceKind = "environment" // DevtronResourceEnvironment is an internal only resource kind used for filtering
	DevtronResourceAppWorkflow DevtronResourceKind = "appWorkflow" // DevtronResourceAppWorkflow is an internal only resource kind used for filtering
)

func (n DevtronResourceKind) ToString() string {
	return string(n)
}

type DevtronResourceVersion string

const (
	DevtronResourceVersion1      DevtronResourceVersion = "v1"
	DevtronResourceVersionAlpha1 DevtronResourceVersion = "alpha1"
)

func (n DevtronResourceVersion) ToString() string {
	return string(n)
}

type DevtronResourceAttributeName string

const (
	DEVTRON_RESOURCE_ATTRIBUTE_APP_NAME                  DevtronResourceAttributeName = "APP_NAME"
	DEVTRON_RESOURCE_ATTRIBUTE_PROJECT_NAME              DevtronResourceAttributeName = "PROJECT_NAME"
	DEVTRON_RESOURCE_ATTRIBUTE_CLUSTER_NAME              DevtronResourceAttributeName = "CLUSTER_NAME"
	DEVTRON_RESOURCE_ATTRIBUTE_ENVIRONMENT_NAME          DevtronResourceAttributeName = "ENVIRONMENT_NAME"
	DEVTRON_RESOURCE_ATTRIBUTE_ENVIRONMENT_IS_PRODUCTION DevtronResourceAttributeName = "IS_PRODUCTION_ENVIRONMENT"
	DEVTRON_RESOURCE_ATTRIBUTE_CI_PIPELINE_BRANCH_VALUE  DevtronResourceAttributeName = "CI_PIPELINE_BRANCH_VALUE"
	DEVTRON_RESOURCE_ATTRIBUTE_CI_PIPELINE_STAGE         DevtronResourceAttributeName = "CI_PIPELINE_STAGE"
)

func (n DevtronResourceAttributeName) ToString() string {
	return string(n)
}

type DevtronResourceAttributeType string

const (
	DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN DevtronResourceAttributeType = "PLUGIN"
)

func (n DevtronResourceAttributeType) ToString() string {
	return string(n)
}

type ValueType string

const (
	VALUE_TYPE_REGEX ValueType = "REGEX"
	VALUE_TYPE_FIXED ValueType = "FIXED"
)

func (v ValueType) ToString() string {
	return string(v)
}

const (
	KindKey                    = "kind"
	VersionKey                 = "version"
	RefValues                  = "values"
	TypeKey                    = "type"
	RefKey                     = "$ref"
	RefTypeKey                 = "refType"
	ReferencesPrefix           = "#/references"
	RefTypePath                = "#/references/users"
	ReferencesKey              = "references"
	IdKey                      = "id"
	NameKey                    = "name"
	IdTypeKey                  = "idType"
	TypeOfDependencyKey        = "typeOfDependency"
	DevtronResourceIdKey       = "devtronResourceId"
	DevtronResourceSchemaIdKey = "devtronResourceSchemaId"
	DependentOnIndexKey        = "dependentOnIndex"
	DependentOnIndexesKey      = "dependentOnIndexes"
	DependentOnParentIndexKey  = "dependentOnParentIndex"
	ConfigKey                  = "config"
	IconKey                    = "icon"
	EnumKey                    = "enum"
	EnumNamesKey               = "enumNames"
	IndexKey                   = "index"

	IdDbColumnKey          = "id"
	OldObjectIdDbColumnKey = "old_object_id"
	NameDbColumnKey        = "name"
	IdentifierDbColumnKey  = "identifier"

	SchemaValidationFailedErrorUserMessage = "Something went wrong. Please check internal message in console for more details."
	BadRequestDependenciesErrorMessage     = "Invalid request. Please check internal message in console for more details."

	EmptyJsonObject = "{}"
)

const (
	ResourceAlreadyExistsMessage                = "Resource already exists!"
	ResourceDoesNotExistMessage                 = "Resource does not exists!"
	InvalidResourceSchemaId                     = "Invalid resource schema id! No resource schema found."
	InvalidResourceKindOrVersion                = "Invalid resource kind or version! No resource schema found."
	ReleaseVersionNotValid                      = "Invalid releaseVersion data! Version absent or not following semantic versioning."
	ResourceNameNotFound                        = "Invalid payload data! name is required."
	ResourceParentConfigNotFound                = "parentConfig is required! parent dependency not defined."
	ResourceParentConfigDataNotFound            = "parentConfig.data is required! parent dependency data not found."
	InvalidResourceParentConfigData             = "Invalid parentConfig data! either id or identifier is required."
	InvalidResourceRequestDescriptorData        = "Invalid request! either id or identifier is required."
	InvalidResourceParentConfigId               = "Invalid parentConfig! incorrect parent dependency."
	InvalidResourceParentConfigKind             = "Invalid parentConfig kind! incorrect parent dependency."
	InvalidResourceKindOrComponent              = "Invalid resource kind or component! Implementation not available."
	InvalidResourceKind                         = "Invalid resource kind! Implementation not supported."
	InvalidQueryDependencyInfo                  = "Invalid query param: dependencyInfo!"
	InvalidQueryConfigOption                    = "Invalid query param: configOption!"
	InvalidResourceVersion                      = "Invalid resource version! Implementation not supported."
	UnimplementedResourceKindOrVersion          = "Invalid resource kind or version! Implementation not supported."
	PatchPathNotSupportedError                  = "patch path not supported"
	PatchValueNotSupportedError                 = "patch value not supported"
	IdTypeNotSupportedError                     = "resource object id type not supported"
	InvalidNoDependencyRequest                  = "Invalid dependency request. No dependencies present. "
	InvalidFilterCriteria                       = "invalid format filter criteria!"
	InvalidSearchKey                            = "invalid format search key!"
	InvalidPatchOperation                       = "invalid patch operation or not supported or dependency info not found"
	InvalidTaskRunOperation                     = "invalid taskRun trigger"
	ApplicationDependencyFoundError             = "application cannot be patched as other dependencies are dependent on this application"
	ApplicationDependencyNotFoundError          = "no application found "
	InvalidDeleteRequest                        = "invalid delete request, action not allowed"
	NoTaskFoundMessage                          = "no task found to execute"
	CanTriggerMessage                           = "Can trigger"
	DeploymentByPassingMessage                  = "You are authorised to deploy outside maintenance window"
	InvalidParentConfigIdOrIdentifier           = "invalid parent id or identifier"
	ActionPolicyInValidDueToStatusErrMessage    = "Operation not allowed with the current status."
	InvalidLevelIndexOrLevelIndexChangedMessage = "invalid level(stages) index or level(stage) index has been changed"
	StageTaskExecutionNotAllowedMessage         = "cannot execute task as all applications in above stages are not deployed successfully on any env."
	CloneSourceDoesNotExistsErrMessage          = "Clone source does not exists."
)

type ChildObjectType string

const DefaultCdPipelineSelector string = "*"
const (
	EnvironmentChildObjectType ChildObjectType = "environments"
)

type CdPipelineReleaseInfo struct {
	AppId                      int                     `json:"appId"`
	AppName                    string                  `json:"appName"`
	EnvId                      int                     `json:"envId"`
	EnvName                    string                  `json:"envName"`
	PipelineId                 int                     `json:"pipelineId"`
	DeploymentAppDeleteRequest bool                    `json:"deploymentAppDeleteRequest"`
	ExistingStages             *ExistingStage          `json:"existingStages"`
	DeployStatus               string                  `json:"deployStatus"`
	PreStatus                  string                  `json:"preStatus"`
	PostStatus                 string                  `json:"postStatus"`
	PreCdWorkflowRunnerId      int                     `json:"preCdWorkflowRunnerId,omitempty"`
	CdWorkflowRunnerId         int                     `json:"cdWorkflowRunnerId,omitempty"`
	PostCdWorkflowRunnerId     int                     `json:"postCdWorkflowRunnerId,omitempty"`
	ReleaseDeploymentStatus    ReleaseDeploymentStatus `json:"releaseDeploymentRolloutStatus,omitempty"`
}

type ReleaseDeploymentStatus string

const (
	YetToTrigger ReleaseDeploymentStatus = "yetToTrigger"
	Ongoing      ReleaseDeploymentStatus = "onGoing"
	Failed       ReleaseDeploymentStatus = "failed"
	Completed    ReleaseDeploymentStatus = "completed"
)

func (r ReleaseDeploymentStatus) ToString() string {
	return string(r)
}

type ExistingStage struct {
	Pre    bool `json:"pre"`
	Deploy bool `json:"deploy"`
	Post   bool `json:"post"`
}

type TaskInfoPostApiBean struct {
	*DevtronResourceObjectDescriptorBean
	FilterCriteria []string `json:"filterCriteria"`
}
