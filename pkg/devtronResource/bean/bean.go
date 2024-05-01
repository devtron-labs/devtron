package bean

import (
	"time"
)

type DevtronResourceBean struct {
	DisplayName          string                       `json:"displayName,omitempty"`
	Description          string                       `json:"description,omitempty"`
	DevtronResourceId    int                          `json:"devtronResourceId"`
	Kind                 string                       `json:"kind,omitempty"`
	VersionSchemaDetails []*DevtronResourceSchemaBean `json:"versionSchemaDetails,omitempty"`
	LastUpdatedOn        time.Time                    `json:"lastUpdatedOn,omitempty"`
}

type DevtronResourceSchemaBean struct {
	DevtronResourceSchemaId int    `json:"devtronResourceSchemaId"`
	Version                 string `json:"version,omitempty"`
	Schema                  string `json:"schema,omitempty"`
	SampleSchema            string `json:"sampleSchema,omitempty"`
}

type DevtronResourceSchemaRequestBean struct {
	DevtronResourceSchemaId int    `json:"devtronResourceSchemaId"`
	Schema                  string `json:"schema,omitempty"`
	DisplayName             string `json:"displayName,omitempty"`
	Description             string `json:"description,omitempty"`
	UserId                  int    `json:"-"`
}

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

type FilterKeyObject = string

type DevtronResourceObjectBean struct {
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

type DevtronResourceObjectGetAPIBean struct {
	*DevtronResourceObjectDescriptorBean
	*DevtronResourceObjectBasicDataBean
	ChildObjects []*DevtronResourceObjectGetAPIBean `json:"childObjects,omitempty"`
}

type DevtronResourceDependencyPatchAPIBean struct {
	*DevtronResourceObjectDescriptorBean
	DependencyPatch []*DependencyPatchBean `json:"dependencyPatch,omitempty"`
}

type DependencyPatchBean struct {
	PatchQuery     []PatchQuery    `json:"query,omitempty"`
	DependencyInfo *DependencyInfo `json:"dependencyInfo,omitempty"`
}

type DevtronResourceObjectBasicDataBean struct {
	Schema       string              `json:"schema,omitempty"`
	CatalogData  string              `json:"objectData,omitempty"` //json key not changed for backward compatibility
	Overview     *ResourceOverview   `json:"overview,omitempty"`
	ConfigStatus *ConfigStatus       `json:"configStatus,omitempty"`
	ParentConfig *ResourceIdentifier `json:"parentConfig,omitempty"`
}

type ResourceOverview struct {
	Description    string            `json:"description,omitempty"`
	Note           *NoteBean         `json:"note,omitempty"`
	ReleaseVersion string            `json:"releaseVersion,omitempty"`
	CreatedBy      *UserSchema       `json:"createdBy,omitempty"`
	CreatedOn      time.Time         `json:"createdOn,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
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

type Environment struct {
	Id   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type ArtifactConfig struct {
	ArtifactId   int             `json:"artifactId"`
	Image        string          `json:"image"`
	RegistryType string          `json:"registryType"`
	RegistryName string          `json:"registryName"`
	CommitSource []GitCommitData `json:"commitSource,omitempty"`
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
	Status   Status `json:"status"`
	Comment  string `json:"comment,omitempty"`
	IsLocked bool   `json:"isLocked"`
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

type FilterCriteriaDecoder struct {
	Resource DevtronResourceKind
	Type     FilterCriteriaIdentifier
	Value    string
}

type SearchCriteriaDecoder struct {
	SearchBy SearchPropertyBy
	Value    string
}

type FilterCriteriaIdentifier string

const (
	Identifier FilterCriteriaIdentifier = "identifier"
	Id         FilterCriteriaIdentifier = "id"
)

type Status string

type IdType string

func (s Status) ToString() string {
	return string(s)
}

const (
	DraftStatus           Status = "draft"
	ReadyForReleaseStatus Status = "readyForRelease"
	HoldStatus            Status = "hold"
)

type ReleaseStatus string //status of release, i.e. rollout status of the release. Not to be confused with config status

const (
	NotDeployedReleaseStatus        ReleaseStatus = "notDeployed"
	PartiallyDeployedReleaseStatus  ReleaseStatus = "partiallyDeployed"
	CompletelyDeployedReleaseStatus ReleaseStatus = "completelyDeployed"
)

type DependencyArtifactStatus string

const (
	NotSelectedDependencyArtifactStatus     DependencyArtifactStatus = "noImageSelected"
	PartialSelectedDependencyArtifactStatus DependencyArtifactStatus = "partialImagesSelected"
	AllSelectedDependencyArtifactStatus     DependencyArtifactStatus = "allImagesSelected"
)

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
	Success bool `json:"success"`
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

type SearchPropertyBy string

const (
	ArtifactTag SearchPropertyBy = "artifactTag"
	ImageTag    SearchPropertyBy = "imageTag"
)

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

type DevtronResourceUIComponent string

func (d DevtronResourceUIComponent) ToString() string {
	return string(d)
}

const (
	UIComponentAll          DevtronResourceUIComponent = "*"
	UIComponentCatalog      DevtronResourceUIComponent = "catalog"
	UIComponentOverview     DevtronResourceUIComponent = "overview"
	UIComponentConfigStatus DevtronResourceUIComponent = "configStatus"
	UIComponentNote         DevtronResourceUIComponent = "note"
)

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

	ResourceObjectDependenciesPath = "dependencies"

	DependencyConfigImageKey              = "artifactConfig.image"
	DependencyConfigArtifactIdKey         = "artifactConfig.artifactId"
	DependencyConfigRegistryNameKey       = "artifactConfig.registryName"
	DependencyConfigRegistryTypeKey       = "artifactConfig.registryType"
	DependencyConfigCiWorkflowKey         = "ciWorkflowId"
	DependencyConfigCommitSourceKey       = "commitSource"
	DependencyConfigReleaseInstructionKey = "releaseInstruction"
	DependencyChildInheritanceKey         = "childInheritance"

	ResourceSchemaMetadataPath       = "properties.overview.properties.metadata"
	ResourceObjectMetadataPath       = "overview.metadata"
	ResourceObjectOverviewPath       = "overview"
	ResourceObjectIdPath             = "overview.id"
	ResourceObjectNamePath           = "overview.name"
	ResourceObjectDescriptionPath    = "overview.description"
	ResourceObjectCreatedOnPath      = "overview.createdOn"
	ResourceObjectCreatedByPath      = "overview.createdBy"
	ResourceObjectReleaseNotePath    = "overview.releaseNote"
	ResourceObjectReleaseVersionPath = "overview.releaseVersion"
	ResourceObjectTagsPath           = "overview.tags"
	ResourceObjectIdTypePath         = "overview.idType"

	ResourceObjectCreatedByIdPath   = "overview.createdBy.id"
	ResourceObjectCreatedByNamePath = "overview.createdBy.name"
	ResourceObjectCreatedByIconPath = "overview.createdBy.icon"

	ResourceConfigStatusPath         = "status.config"
	ResourceConfigStatusStatusPath   = "status.config.status"
	ResourceConfigStatusCommentPath  = "status.config.comment"
	ResourceConfigStatusIsLockedPath = "status.config.lock"

	SchemaValidationFailedErrorUserMessage = "Something went wrong. Please check internal message in console for more details."
	BadRequestDependenciesErrorMessage     = "Invalid request. Please check internal message in console for more details."

	EmptyJsonObject = "{}"
)

type PatchQueryPath string
type PatchQueryOperation string

const (
	Replace PatchQueryOperation = "replace"
	Add     PatchQueryOperation = "add"
	Remove  PatchQueryOperation = "remove"
)
const (
	ReleaseInstructionQueryPath PatchQueryPath = "releaseInstruction"
	CommitQueryPath             PatchQueryPath = "commit"
	ImageQueryPath              PatchQueryPath = "image"
	DescriptionQueryPath        PatchQueryPath = "description"
	NoteQueryPath               PatchQueryPath = "note"
	ReadMeQueryPath             PatchQueryPath = "readme"
	NameQueryPath               PatchQueryPath = "name"
	StatusQueryPath             PatchQueryPath = "status"
	LockQueryPath               PatchQueryPath = "lock"
	TagsQueryPath               PatchQueryPath = "tags"
	ApplicationQueryPath        PatchQueryPath = "application"
)

const (
	ResourceAlreadyExistsMessage         = "Resource already exists!"
	ResourceDoesNotExistMessage          = "Resource does not exists!"
	InvalidResourceSchemaId              = "Invalid resource schema id! No resource schema found."
	InvalidResourceKindOrVersion         = "Invalid resource kind or version! No resource schema found."
	ReleaseVersionNotFound               = "Invalid overview data! overview.releaseVersion is required."
	ResourceNameNotFound                 = "Invalid payload data! name is required."
	ResourceParentConfigNotFound         = "parentConfig is required! parent dependency not defined."
	ResourceParentConfigDataNotFound     = "parentConfig.data is required! parent dependency data not found."
	InvalidResourceParentConfigData      = "Invalid parentConfig data! either id or identifier is required."
	InvalidResourceRequestDescriptorData = "Invalid request! either id or identifier is required."
	InvalidResourceParentConfigId        = "Invalid parentConfig! incorrect parent dependency."
	InvalidResourceParentConfigKind      = "Invalid parentConfig kind! incorrect parent dependency."
	InvalidResourceKindOrComponent       = "Invalid resource kind or component! Implementation not available."
	InvalidResourceKind                  = "Invalid resource kind! Implementation not supported."
	InvalidQueryDependencyInfo           = "Invalid query param: dependencyInfo!"
	InvalidQueryConfigOption             = "Invalid query param: configOption!"
	InvalidResourceVersion               = "Invalid resource version! Implementation not supported."
	PatchPathNotSupportedError           = "patch path not supported"
	IdTypeNotSupportedError              = "resource object id type not supported"
	InvalidNoDependencyRequest           = "Invalid dependency request. No dependencies present. "
	InvalidFilterCriteria                = "invalid format filter criteria!"
	InvalidSearchKey                     = "invalid format search key!"
	InvalidPatchOperation                = "invalid patch operation or not supported or dependency info not found"
	ApplicationDependencyFoundError      = "application cannot be patched as other dependencies are dependent on this application"
	ApplicationDependencyNotFoundError   = "no application found "
)

type ChildObjectType string

const DefaultCdPipelineSelector string = "*"
const (
	EnvironmentChildObjectType ChildObjectType = "environments"
)
