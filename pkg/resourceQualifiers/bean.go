package resourceQualifiers

type Scope struct {
	AppId          int             `json:"appId"`
	EnvId          int             `json:"envId"`
	ClusterId      int             `json:"clusterId"`
	PipelineId     int             `json:"pipelineId"`
	SystemMetadata *SystemMetadata `json:"-"`
}

type SystemMetadata struct {
	EnvironmentName string
	ClusterName     string
	Namespace       string
	ImageTag        string
	Image           string
	AppName         string
}

func (metadata *SystemMetadata) GetDataFromSystemVariable(variable SystemVariableName) string {
	switch variable {
	case DevtronNamespace:
		return metadata.Namespace
	case DevtronClusterName:
		return metadata.ClusterName
	case DevtronEnvName:
		return metadata.EnvironmentName
	case DevtronImageTag:
		return metadata.ImageTag
	case DevtronAppName:
		return metadata.AppName
	case DevtronImage:
		return metadata.Image
	}
	return ""
}

type Qualifier int

const (
	GLOBAL_QUALIFIER   Qualifier = 5
	PIPELINE_QUALIFIER Qualifier = 6
)

var CompoundQualifiers []Qualifier

func GetNumOfChildQualifiers(qualifier Qualifier) int {
	return 0
}
