package resourceQualifiers

type SystemVariableName string

const (
	DevtronNamespace   SystemVariableName = "DEVTRON_NAMESPACE"
	DevtronClusterName SystemVariableName = "DEVTRON_CLUSTER_NAME"
	DevtronEnvName     SystemVariableName = "DEVTRON_ENV_NAME"
	DevtronImageTag    SystemVariableName = "DEVTRON_IMAGE_TAG"
	DevtronImage       SystemVariableName = "DEVTRON_IMAGE"
	DevtronAppName     SystemVariableName = "DEVTRON_APP_NAME"
)

var SystemVariables = []SystemVariableName{
	DevtronNamespace,
	DevtronClusterName,
	DevtronEnvName,
	DevtronImageTag,
	DevtronAppName,
}
