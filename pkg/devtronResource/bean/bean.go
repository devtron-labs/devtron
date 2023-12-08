package bean

type DevtronResourceObjectDescriptorBean struct {
	Kind        string `json:"kind,omitempty"`
	SubKind     string `json:"subKind,omitempty"`
	Version     string `json:"version,omitempty"`
	OldObjectId int    `json:"id,omitempty"` //here at FE we are still calling this id since id is used everywhere w.r.t resource's own tables
	Name        string `json:"name,omitempty"`
}

type DevtronResourceObjectBean struct {
	*DevtronResourceObjectDescriptorBean
	Schema     string `json:"schema,omitempty"`
	ObjectData string `json:"objectData"`
	UserId     int32  `json:"-"`
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
	DEVTRON_RESOURCE_APPLICATION         DevtronResourceKind = "application"
	DEVTRON_RESOURCE_DEVTRON_APPLICATION DevtronResourceKind = "devtron-application"
	DEVTRON_RESOURCE_HELM_APPLICATION    DevtronResourceKind = "helm-application"
	DEVTRON_RESOURCE_CLUSTER             DevtronResourceKind = "cluster"
	DEVTRON_RESOURCE_JOB                 DevtronResourceKind = "job"
	DEVTRON_RESOURCE_USER                DevtronResourceKind = "users"
)

func (n DevtronResourceKind) ToString() string {
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
	KindKey          = "kind"
	VersionKey       = "version"
	RefValues        = "values"
	TypeKey          = "type"
	RefKey           = "$ref"
	RefTypeKey       = "refType"
	ReferencesPrefix = "#/references"
	ReferencesKey    = "references"
	IdKey            = "id"
	NameKey          = "name"
	IconKey          = "icon"
	EnumKey          = "enum"
	EnumNamesKey     = "enumNames"

	ResourceSchemaMetadataPath = "properties.overview.properties.metadata"
	ResourceObjectMetadataPath = "overview.metadata"
	ResourceObjectIdPath       = "overview.id"

	SchemaValidationFailedErrorUserMessage = "Something went wrong. Please check internalMessage in console for more details."
)
