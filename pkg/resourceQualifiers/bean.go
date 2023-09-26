package resourceQualifiers

type Scope struct {
	AppId     int `json:"appId"`
	EnvId     int `json:"envId"`
	ClusterId int `json:"clusterId"`
}

type Qualifier int

const (
	GLOBAL_QUALIFIER Qualifier = 5
)
