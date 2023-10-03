package bean

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
)

func (n DevtronResourceSearchableKeyName) ToString() string {
	return string(n)
}

type DevtronResourceName string

const (
	DEVTRON_RESOURCE_PROJECT     DevtronResourceName = "PROJECT"
	DEVTRON_RESOURCE_APP         DevtronResourceName = "APP"
	DEVTRON_RESOURCE_CLUSTER     DevtronResourceName = "CLUSTER"
	DEVTRON_RESOURCE_ENVIRONMENT DevtronResourceName = "ENVIRONMENT"
	DEVTRON_RESOURCE_CI_PIPELINE DevtronResourceName = "CI_PIPELINE"
	DEVTRON_RESOURCE_CD_PIPELINE DevtronResourceName = "CD_PIPELINE"
)

func (n DevtronResourceName) ToString() string {
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
