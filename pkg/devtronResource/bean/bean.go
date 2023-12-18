package bean

import "time"

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
	Kind        string `json:"kind,omitempty"`
	SubKind     string `json:"subKind,omitempty"`
	Version     string `json:"version,omitempty"`
	OldObjectId int    `json:"id,omitempty"` //here at FE we are still calling this id since id is used everywhere w.r.t resource's own tables
	Name        string `json:"name,omitempty"`
	SchemaId    int    `json:"schemaId"`
}

type DevtronResourceObjectBean struct {
	*DevtronResourceObjectDescriptorBean
	Schema            string                           `json:"schema,omitempty"`
	ObjectData        string                           `json:"objectData"`
	Dependencies      []*DevtronResourceDependencyBean `json:"dependencies"`
	ChildDependencies []*DevtronResourceDependencyBean `json:"childDependencies"`
	UserId            int32                            `json:"-"`
}

type DevtronResourceDependencyBean struct {
	Name                    string                           `json:"name"`
	OldObjectId             int                              `json:"id"`
	DevtronResourceId       int                              `json:"devtronResourceId"`
	DevtronResourceSchemaId int                              `json:"devtronResourceSchemaId"`
	DependentOnIndex        int                              `json:"dependentOnIndex"`
	DependentOnParentIndex  int                              `json:"dependentOnParentIndex"`
	TypeOfDependency        DevtronResourceDependencyType    `json:"typeOfDependency"`
	Index                   int                              `json:"index"`
	Dependencies            []*DevtronResourceDependencyBean `json:"dependencies,omitempty"`
	Metadata                interface{}                      `json:"metadata,omitempty"`
}

type UpdateSchemaResponseBean struct {
	Message       string   `json:"message"`
	PathsToRemove []string `json:"pathsToRemove"`
}

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
)

func (n DevtronResourceKind) ToString() string {
	return string(n)
}

type DevtronResourceVersion string

const (
	DevtronResourceVersion1 DevtronResourceVersion = "v1"
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

	ResourceSchemaMetadataPath     = "properties.overview.properties.metadata"
	ResourceObjectMetadataPath     = "overview.metadata"
	ResourceObjectDependenciesPath = "dependencies"
	ResourceObjectIdPath           = "overview.id"

	SchemaValidationFailedErrorUserMessage = "Something went wrong. Please check internal message in console for more details."
	BadRequestDependenciesErrorMessage     = "Invalid request. Please check internal message in console for more details."

	EmptyJsonObject = "{}"
)
