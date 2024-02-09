package bean

type EnvironmentBean struct {
	Id                     int      `json:"id,omitempty" validate:"number"`
	Environment            string   `json:"environment_name,omitempty" validate:"required,max=50"`
	ClusterId              int      `json:"cluster_id,omitempty" validate:"number,required"`
	ClusterName            string   `json:"cluster_name,omitempty"`
	Active                 bool     `json:"active"`
	Default                bool     `json:"default"`
	PrometheusEndpoint     string   `json:"prometheus_endpoint,omitempty"`
	Namespace              string   `json:"namespace,omitempty" validate:"name-space-component,max=50"`
	CdArgoSetup            bool     `json:"isClusterCdActive"`
	EnvironmentIdentifier  string   `json:"environmentIdentifier"`
	Description            string   `json:"description" validate:"max=40"`
	AppCount               int      `json:"appCount"`
	IsVirtualEnvironment   bool     `json:"isVirtualEnvironment"`
	AllowedDeploymentTypes []string `json:"allowedDeploymentTypes"`
	ClusterServerUrl       string   `json:"-"`
	ErrorInConnecting      string   `json:"-"`
}

type EnvDto struct {
	EnvironmentId         int    `json:"environmentId" validate:"number"`
	EnvironmentName       string `json:"environmentName,omitempty" validate:"max=50"`
	Namespace             string `json:"namespace,omitempty" validate:"name-space-component,max=50"`
	EnvironmentIdentifier string `json:"environmentIdentifier,omitempty"`
	Description           string `json:"description" validate:"max=40"`
	IsVirtualEnvironment  bool   `json:"isVirtualEnvironment"`
}

type ClusterEnvDto struct {
	ClusterId        int       `json:"clusterId"`
	ClusterName      string    `json:"clusterName,omitempty"`
	Environments     []*EnvDto `json:"environments,omitempty"`
	IsVirtualCluster bool      `json:"isVirtualCluster"`
}

type ResourceGroupingResponse struct {
	EnvList  []EnvironmentBean `json:"envList"`
	EnvCount int               `json:"envCount"`
}

const (
	PIPELINE_DEPLOYMENT_TYPE_HELM = "helm"
	PIPELINE_DEPLOYMENT_TYPE_ACD  = "argo_cd"
)
