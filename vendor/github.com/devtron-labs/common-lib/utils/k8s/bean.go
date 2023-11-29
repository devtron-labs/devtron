package k8s

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	DEFAULT_CLUSTER          = "default_cluster"
	DEVTRON_SERVICE_NAME     = "devtron-service"
	DefaultClusterUrl        = "https://kubernetes.default.svc"
	BearerToken              = "bearer_token"
	CertificateAuthorityData = "cert_auth_data"
	CertData                 = "cert_data"
	TlsKey                   = "tls_key"
	LiveZ                    = "/livez"
	Running                  = "Running"
	RestartingNotSupported   = "restarting not supported"
	DEVTRON_APP_LABEL_KEY    = "app"
	DEVTRON_APP_LABEL_VALUE1 = "devtron"
	DEVTRON_APP_LABEL_VALUE2 = "orchestrator"
)

type ClusterResourceListMap struct {
	Headers       []string                 `json:"headers"`
	Data          []map[string]interface{} `json:"data"`
	ServerVersion string                   `json:"serverVersion"`
}

type EventsResponse struct {
	Events *v1.EventList `json:"events,omitempty"`
}

type ResourceListResponse struct {
	Resources unstructured.UnstructuredList `json:"resources,omitempty"`
}

type PodLogsRequest struct {
	SinceTime                  *v12.Time `json:"sinceTime,omitempty"`
	TailLines                  int       `json:"tailLines"`
	Follow                     bool      `json:"follow"`
	ContainerName              string    `json:"containerName"`
	IsPrevContainerLogsEnabled bool      `json:"previous"`
}

type ResourceIdentifier struct {
	Name             string                  `json:"name"` //pod name for logs request
	Namespace        string                  `json:"namespace"`
	GroupVersionKind schema.GroupVersionKind `json:"groupVersionKind"`
}

type K8sRequestBean struct {
	ResourceIdentifier ResourceIdentifier `json:"resourceIdentifier"`
	Patch              string             `json:"patch,omitempty"`
	PodLogsRequest     PodLogsRequest     `json:"podLogsRequest,omitempty"`
	ForceDelete        bool               `json:"-"`
}

type GetAllApiResourcesResponse struct {
	ApiResources []*K8sApiResource `json:"apiResources"`
	AllowedAll   bool              `json:"allowedAll"`
}

type K8sApiResource struct {
	Gvk        schema.GroupVersionKind     `json:"gvk"`
	Gvr        schema.GroupVersionResource `json:"gvr"`
	Namespaced bool                        `json:"namespaced"`
}

type ApplyResourcesRequest struct {
	Manifest  string `json:"manifest"`
	ClusterId int    `json:"clusterId"`
}

type ApplyResourcesResponse struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Error    string `json:"error"`
	IsUpdate bool   `json:"isUpdate"`
}

type ManifestResponse struct {
	Manifest unstructured.Unstructured `json:"manifest,omitempty"`
}

type ResourceKey struct {
	Group     string
	Kind      string
	Namespace string
	Name      string
}

func (k *ResourceKey) String() string {
	return fmt.Sprintf("%s/%s/%s/%s", k.Group, k.Kind, k.Namespace, k.Name)
}

func (k ResourceKey) GroupKind() schema.GroupKind {
	return schema.GroupKind{Group: k.Group, Kind: k.Kind}
}

func NewResourceKey(group string, kind string, namespace string, name string) ResourceKey {
	return ResourceKey{Group: group, Kind: kind, Namespace: namespace, Name: name}
}

func GetResourceKey(obj *unstructured.Unstructured) ResourceKey {
	gvk := obj.GroupVersionKind()
	return NewResourceKey(gvk.Group, gvk.Kind, obj.GetNamespace(), obj.GetName())
}

type DeletionPropagationOptions string

const (
	DeletePropagationOrphan     DeletionPropagationOptions = "Orphan"
	DeletePropagationBackground DeletionPropagationOptions = "Background"
	DeletePropagationForeground DeletionPropagationOptions = "Foreground"
)

type DeleteOptions struct {
	Kind               string
	APIVersion         string
	GracePeriodSeconds *int64
	Preconditions      string
	OrphanDependents   *bool
	PropagationPolicy  DeletionPropagationOptions
	DryRun             []string
}

type DeleteAndCreateJobRequest struct {
	Content       []byte
	Namespace     string
	ClusterConfig *ClusterConfig
	DeleteOptions *DeleteOptions
}

func (impl DeleteAndCreateJobRequest) GetK8sDeleteOptions() v12.DeleteOptions {
	if impl.DeleteOptions == nil {
		return v12.DeleteOptions{}
	}
	deleteOptions := v12.DeleteOptions{
		TypeMeta: v12.TypeMeta{
			Kind:       impl.DeleteOptions.Kind,
			APIVersion: impl.DeleteOptions.APIVersion,
		},
		GracePeriodSeconds: impl.DeleteOptions.GracePeriodSeconds,
		DryRun:             impl.DeleteOptions.DryRun,
	}
	deletionPropagation := v12.DeletionPropagation(impl.DeleteOptions.PropagationPolicy)
	if len(deletionPropagation) > 0 {
		deleteOptions.PropagationPolicy = &deletionPropagation
	}
	return deleteOptions
}
