package k8s

import (
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8sObjectsUtil"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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

type ClusterConfig struct {
	ClusterName                     string
	Host                            string
	BearerToken                     string
	InsecureSkipTLSVerify           bool
	KeyData                         string
	CertData                        string
	CAData                          string
	ClusterId                       int
	ProxyUrl                        string
	ToConnectForClusterVerification bool
	ToConnectWithSSHTunnel          bool
	SSHTunnelUser                   string
	SSHTunnelPassword               string
	SSHTunnelAuthKey                string
	SSHTunnelServerAddress          string
}

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
	// EphemeralContainers are set for Pod kind manifest response only.
	// will only contain ephemeral containers which are in running state
	// +optional
	EphemeralContainers []*k8sObjectsUtil.EphemeralContainerData `json:"ephemeralContainers,omitempty"`
}

// SetRunningEphemeralContainers will extract out all the running ephemeral containers of the given pod manifest and sets in manifestResponse.EphemeralContainers
// if given manifest is not of pod kind
func (manifestResponse *ManifestResponse) SetRunningEphemeralContainers() error {
	if manifestResponse != nil {
		if podManifest := manifestResponse.Manifest; k8sObjectsUtil.IsPod(podManifest.GetKind(), podManifest.GroupVersionKind().Group) {
			pod := v1.Pod{}
			// Convert the unstructured object to a Pod object
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(podManifest.Object, &pod)
			if err != nil {
				return err
			}
			runningEphemeralContainers := k8sObjectsUtil.ExtractEphemeralContainers([]v1.Pod{pod})
			manifestResponse.EphemeralContainers = runningEphemeralContainers[pod.Name]
		}
	}
	return nil
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
