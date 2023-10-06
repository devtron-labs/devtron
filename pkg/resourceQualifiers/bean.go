package resourceQualifiers

type Scope struct {
	AppId     int `json:"appId"`
	EnvId     int `json:"envId"`
	ClusterId int `json:"clusterId"`

	SystemMetadata *SystemMetadata `json:"-"`
}

type SystemMetadata struct {
	EnvironmentName string
	ClusterName     string
	Namespace       string
	ImageTag        string
}

type Qualifier int

const (
	GLOBAL_QUALIFIER Qualifier = 5
)

var CompoundQualifiers []Qualifier

func GetNumOfChildQualifiers(qualifier Qualifier) int {
	return 0
}
