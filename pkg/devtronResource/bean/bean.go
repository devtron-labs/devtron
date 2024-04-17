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
	OldObjectId  int                          `json:"id,omitempty"` //here at FE we are still calling this id since id is used everywhere w.r.t resource's own tables
	Name         string                       `json:"name,omitempty"`
	SchemaId     int                          `json:"schemaId"`
	Identifier   string                       `json:"identifier,omitempty"`
	UIComponents []DevtronResourceUIComponent `json:"-"`
	Id           int                          `json:"-"` //this is the field which holds the resourceObjectId i.e. id in devtron_resource_object table. Have not exposed this and taking this value from oldObjectId to maintain backward compatibility
	IdType       IdType                       `json:"-"` // internal , for release and release-track idType will be resourceObjectId
}

type DevtronResourceObjectBean struct {
	*DevtronResourceObjectDescriptorBean
	Schema            string                           `json:"schema,omitempty"`
	ObjectData        string                           `json:"objectData"`
	Dependencies      []*DevtronResourceDependencyBean `json:"dependencies"`
	ChildDependencies []*DevtronResourceDependencyBean `json:"childDependencies"`
	Overview          *ResourceOverview                `json:"overview,omitempty"`
	ConfigStatus      *ConfigStatus                    `json:"configStatus,omitempty"`
	ParentConfig      *ResourceParentConfig            `json:"parentConfig,omitempty"`
	PatchQuery        []PatchQuery                     `json:"query,omitempty"`
	UserId            int32                            `json:"-"`
}

type DevtronResourceObjectGetAPIBean struct {
	*DevtronResourceObjectDescriptorBean
	*DevtronResourceObjectBasicDataBean
	ChildObjects []*DevtronResourceObjectGetAPIBean `json:"childObjects,omitempty"`
}

type DevtronResourceObjectBasicDataBean struct {
	Schema       string                `json:"schema,omitempty"`
	CatalogData  string                `json:"objectData,omitempty"` //json key not changed for backward compatibility
	Overview     *ResourceOverview     `json:"overview,omitempty"`
	ConfigStatus *ConfigStatus         `json:"configStatus,omitempty"`
	ParentConfig *ResourceParentConfig `json:"parentConfig,omitempty"`
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

type ResourceParentConfig struct {
	Type DevtronResourceKind `json:"type"`
	Data *ResourceParentData `json:"data,omitempty"`
}

type PatchQuery struct {
	Operation string         `json:"op"` // default is replace
	Path      PatchQueryPath `json:"path" validate:"required"`
	Value     interface{}    `json:"value"`
}

type DependencyInfo struct {
	DependencyName string              `json:"dependencyName,omitempty"`
	DependencyType DevtronResourceKind `json:"dependencyResourceKind,omitempty"`
}

type ResourceParentData struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type DevtronResourceDependencyBean struct {
	Name                    string                           `json:"name"`
	OldObjectId             int                              `json:"id"` //have both oldObjectId and resourceObjectId
	DevtronResourceId       int                              `json:"devtronResourceId"`
	DevtronResourceSchemaId int                              `json:"devtronResourceSchemaId"`
	DependentOnIndex        int                              `json:"dependentOnIndex,omitempty"`
	DependentOnParentIndex  int                              `json:"dependentOnParentIndex,omitempty"`
	TypeOfDependency        DevtronResourceDependencyType    `json:"typeOfDependency"`
	Index                   int                              `json:"index"`
	Dependencies            []*DevtronResourceDependencyBean `json:"dependencies,omitempty"`
	Metadata                interface{}                      `json:"metadata,omitempty"`
	IdType                  IdType                           `json:"idType,omitempty"`
}

type UpdateSchemaResponseBean struct {
	Message       string   `json:"message"`
	PathsToRemove []string `json:"pathsToRemove"`
}

type ResourceObjectRequirementRequest struct {
	ReqBean                  *DevtronResourceObjectBean
	ObjectDataPath           string
	SkipJsonSchemaValidation bool
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
	DependentOnParentIndexKey  = "dependentOnParentIndex"
	IconKey                    = "icon"
	EnumKey                    = "enum"
	EnumNamesKey               = "enumNames"
	IndexKey                   = "index"

	OldObjectIdDbColumnKey = "old_object_id"
	NameDbColumnKey        = "name"

	ResourceObjectDependenciesPath = "dependencies"

	ResourceSchemaMetadataPath           = "properties.overview.properties.metadata"
	ResourceObjectMetadataPath           = "overview.metadata"
	ResourceObjectOverviewPath           = "overview"
	ResourceObjectIdPath                 = "overview.id"
	ResourceObjectNamePath               = "overview.name"
	ResourceObjectDescriptionPath        = "overview.description"
	ResourceObjectCreatedOnPath          = "overview.createdOn"
	ResourceObjectCreatedByPath          = "overview.createdBy"
	ResourceObjectReleaseNotePath        = "overview.releaseNote"
	ResourceObjectReleaseVersionPath     = "overview.releaseVersion"
	ResourceObjectReleaseInstructionPath = "overview.releaseInstruction"
	ResourceObjectTagsPath               = "overview.tags"
	ResourceObjectIdTypePath             = "overview.idType"

	ResourceObjectCreatedByIdPath   = "overview.createdBy.id"
	ResourceObjectCreatedByNamePath = "overview.createdBy.name"
	ResourceObjectCreatedByIconPath = "overview.createdBy.icon"

	ResourceConfigStatusPath         = "status.config"
	ResourceConfigStatusStatusPath   = "status.config.status"
	ResourceConfigStatusCommentPath  = "status.config.comment"
	ResourceConfigStatusIsLockedPath = "status.config.isLocked"

	SchemaValidationFailedErrorUserMessage = "Something went wrong. Please check internal message in console for more details."
	BadRequestDependenciesErrorMessage     = "Invalid request. Please check internal message in console for more details."

	EmptyJsonObject = "{}"
)

type PatchQueryPath string

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
)

const (
	ResourceAlreadyExistsMessage     = "Resource already exists!"
	ResourceDoesNotExistMessage      = "Resource does not exists!"
	InvalidResourceKindOrVersion     = "Invalid resource kind or version! No resource schema found."
	ReleaseVersionNotFound           = "Invalid overview data! overview.releaseVersion is required."
	ResourceNameNotFound             = "Invalid payload data! name is required."
	ResourceParentConfigNotFound     = "parentConfig is required! parent dependency not defined."
	ResourceParentConfigDataNotFound = "parentConfig.data is required! parent dependency data not found."
	InvalidResourceParentConfigData  = "Invalid parentConfig.data! either id or name is required."
	InvalidResourceParentConfigId    = "Invalid parentConfig id! incorrect parent dependency."
	InvalidResourceParentConfigType  = "Invalid parentConfig type! incorrect parent dependency."
	InvalidResourceKindOrComponent   = "Invalid resource kind or component! Implementation not available."
	InvalidResourceKind              = "Invalid resource kind! Implementation not supported."
	PatchPathNotSupportedError       = "patch path not supported"
)
