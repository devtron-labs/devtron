package generateManifest

import (
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	Values   RequestDataMode = 1
	Manifest RequestDataMode = 2
)

type Kind string

const (
	Deployment  Kind = "Deployment"
	StatefulSet Kind = "StatefulSet"
	DemonSet    Kind = "DemonSet"
	Rollout     Kind = "Rollout"
)

const LabelReleaseKey = "release"

type DeploymentTemplateRequest struct {
	AppId                       int                               `json:"appId"`
	EnvId                       int                               `json:"envId,omitempty"`
	ChartRefId                  int                               `json:"chartRefId"`
	RequestDataMode             RequestDataMode                   `json:"valuesAndManifestFlag"`
	Values                      string                            `json:"values"`
	Type                        repository.DeploymentTemplateType `json:"type"`
	DeploymentTemplateHistoryId int                               `json:"deploymentTemplateHistoryId,omitempty"`
	ResourceName                string                            `json:"resourceName"`
	PipelineId                  int                               `json:"pipelineId"`
}

type RequestDataMode int

var ChartRepository = &gRPC.ChartRepository{
	Name:     "repo",
	Url:      "http://localhost:8080/",
	Username: "admin",
	Password: "password",
}

var ReleaseIdentifier = &gRPC.ReleaseIdentifier{
	ReleaseNamespace: "devtron-demo",
	ReleaseName:      "release-name",
}

type DeploymentTemplateResponse struct {
	Data             string            `json:"data"`
	ResolvedData     string            `json:"resolvedData"`
	VariableSnapshot map[string]string `json:"variableSnapshot"`
}

type RestartPodResponse struct {
	EnvironmentId int                                 `json:"environmentId" `
	Namespace     string                              `json:"namespace"`
	RestartPodMap map[int]*ResourceIdentifierResponse `json:"restartPodMap"`
}

type ResourceIdentifierResponse struct {
	ResourceMetaData []*ResourceMetadata `json:"resourceMetaData,omitempty"`
	AppName          string              `json:"appName"`
}
type ResourceMetadata struct {
	Name             string                  `json:"name"`
	GroupVersionKind schema.GroupVersionKind `json:"groupVersionKind"`
}
type RestartWorkloadConfig struct {
	WorkerPoolSize   int `env:"FEATURE_RESTART_WORKLOAD_WORKER_POOL_SIZE" envDefault:"5"`
	RequestBatchSize int `env:"FEATURE_RESTART_WORKLOAD_BATCH_SIZE" envDefault:"1"`
}
