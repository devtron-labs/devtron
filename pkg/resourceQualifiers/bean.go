package resourceQualifiers

type Scope struct {
	AppId     int `json:"appId"`
	EnvId     int `json:"envId"`
	ClusterId int `json:"clusterId"`
	ProjectId int `json:"projectId"`
}

type Qualifier int

const (
	APP_AND_ENV_QUALIFIER Qualifier = 1
	APP_QUALIFIER         Qualifier = 2
	ENV_QUALIFIER         Qualifier = 3
	CLUSTER_QUALIFIER     Qualifier = 4
	GLOBAL_QUALIFIER      Qualifier = 5
)

var CompoundQualifiers = []Qualifier{APP_AND_ENV_QUALIFIER}

func GetNumOfChildQualifiers(qualifier Qualifier) int {
	switch qualifier {
	case APP_AND_ENV_QUALIFIER:
		return 1
	}
	return 0
}
