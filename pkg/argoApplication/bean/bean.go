package bean

import (
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ArgoGroup                    = "argoproj.io"
	ArgoApplicationKind          = "Application"
	VersionV1Alpha1              = "v1alpha1"
	AllNamespaces                = ""
	DevtronCDNamespae            = "devtroncd"
	ArgoLabelForManagedResources = "app.kubernetes.io/instance"
)

const (
	Server      = "server"
	Destination = "destination"
	Config      = "config"
)

var GvkForArgoApplication = schema.GroupVersionKind{
	Group:   ArgoGroup,
	Kind:    ArgoApplicationKind,
	Version: VersionV1Alpha1,
}

var GvkForSecret = schema.GroupVersionKind{
	Kind:    k8sCommonBean.SecretKind,
	Version: k8sCommonBean.V1VERSION,
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
	Manifest     map[string]interface{}       `json:"manifest"`
}

type ArgoManagedResource struct {
	Group     string
	Kind      string
	Version   string
	Name      string
	Namespace string
}

type ArgoClusterConfigObj struct {
	BearerToken     string `json:"bearerToken"`
	TlsClientConfig struct {
		Insecure bool `json:"insecure"`
	} `json:"tlsClientConfig"`
}
