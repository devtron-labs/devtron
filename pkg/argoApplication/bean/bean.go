package bean

import (
	client "github.com/devtron-labs/devtron/api/helm-app"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ArgoGroup                    = "argoproj.io"
	ArgoApplicationKind          = "Application"
	VersionV1Alpha1              = "v1alpha1"
	AllNamespaces                = ""
	ArgoLabelForManagedResources = "app.kubernetes.io/instance"
)

var GvkForArgoApplication = schema.GroupVersionKind{
	Group:   ArgoGroup,
	Kind:    ArgoApplicationKind,
	Version: VersionV1Alpha1,
}

type ArgoApplicationListDto struct {
	Name         string `json:"appName"`
	ClusterId    int    `json:"clusterId"`
	ClusterName  string `json:"clusterName"`
	Namespace    string `json:"namespace"`
	HealthStatus string `json:"appStatus"`
	SyncStatus   string `json:"syncStatus"`
}

type ArgoApplicationDetailDto struct {
	*ArgoApplicationListDto
	ResourceTree *client.ResourceTreeResponse `json:"resourceTree"`
}

type ArgoManagedResource struct {
	Group     string
	Kind      string
	Version   string
	Name      string
	Namespace string
}
