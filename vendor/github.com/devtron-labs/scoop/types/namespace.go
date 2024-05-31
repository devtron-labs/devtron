package types

type ClusterEnvDto struct {
	ClusterId        int       `json:"clusterId"`
	ClusterName      string    `json:"clusterName,omitempty"`
	Environments     []*EnvDto `json:"environments,omitempty"`
	IsVirtualCluster bool      `json:"isVirtualCluster"`
}

type EnvDto struct {
	EnvironmentId         int    `json:"environmentId" validate:"number"`
	EnvironmentName       string `json:"environmentName,omitempty" validate:"max=50"`
	Namespace             string `json:"namespace,omitempty" validate:"name-space-component,max=50"`
	EnvironmentIdentifier string `json:"environmentIdentifier,omitempty"`
	Description           string `json:"description" validate:"max=40"`
	IsVirtualEnvironment  bool   `json:"isVirtualEnvironment"`
	Default               bool   `json:"default"`
}
