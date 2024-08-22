/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package k8s

import (
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib/utils/k8sObjectsUtil"
	"github.com/devtron-labs/common-lib/utils/remoteConnection/bean"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/rest"
	"log"
	"net"
	"net/http"
	"time"
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
	ToConnectForClusterVerification bool
	RemoteConnectionConfig          *bean.RemoteConnectionConfigBean
}

func (clusterConfig *ClusterConfig) PopulateTlsConfigurationsInto(restConfig *rest.Config) {
	restConfig.TLSClientConfig = rest.TLSClientConfig{Insecure: clusterConfig.InsecureSkipTLSVerify}
	if clusterConfig.InsecureSkipTLSVerify == false {
		restConfig.TLSClientConfig.ServerName = restConfig.ServerName
		restConfig.TLSClientConfig.KeyData = []byte(clusterConfig.KeyData)
		restConfig.TLSClientConfig.CertData = []byte(clusterConfig.CertData)
		restConfig.TLSClientConfig.CAData = []byte(clusterConfig.CAData)
	}
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
	SinceSeconds               int       `json:"sinceSeconds,omitempty"`
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
	ForceDelete        bool               `json:"forceDelete,omitempty"`
}

type GetAllApiResourcesResponse struct {
	ApiResources []*K8sApiResource `json:"apiResources"`
	AllowedAll   bool              `json:"allowedAll"`
}

type K8sApiResource struct {
	Gvk        schema.GroupVersionKind     `json:"gvk"`
	Gvr        schema.GroupVersionResource `json:"gvr"`
	Namespaced bool                        `json:"namespaced"`
	ShortNames []string                    `json:"shortNames"`
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

type CustomK8sHttpTransportConfig struct {
	UseCustomTransport  bool `env:"USE_CUSTOM_HTTP_TRANSPORT" envDefault:"false"`
	TimeOut             int  `env:"K8s_TCP_TIMEOUT" envDefault:"30"`
	KeepAlive           int  `env:"K8s_TCP_KEEPALIVE" envDefault:"30"`
	TLSHandshakeTimeout int  `env:"K8s_TLS_HANDSHAKE_TIMEOUT" envDefault:"10"`
	MaxIdleConnsPerHost int  `env:"K8s_CLIENT_MAX_IDLE_CONNS_PER_HOST" envDefault:"25"`
	IdleConnTimeout     int  `env:"K8s_TCP_IDLE_CONN_TIMEOUT" envDefault:"300"`
}

type LocalDevMode bool

type RuntimeConfig struct {
	LocalDevMode LocalDevMode `env:"RUNTIME_CONFIG_LOCAL_DEV" envDefault:"false"`
}

func GetRuntimeConfig() (*RuntimeConfig, error) {
	cfg := &RuntimeConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func NewCustomK8sHttpTransportConfig() *CustomK8sHttpTransportConfig {
	customK8sHttpTransportConfig := &CustomK8sHttpTransportConfig{}
	err := env.Parse(customK8sHttpTransportConfig)
	if err != nil {
		log.Println("error in parsing custom k8s http configurations from env : ", "err : ", err)
	}
	return customK8sHttpTransportConfig
}

// OverrideConfigWithCustomTransport
// overrides the given rest config with custom transport if UseCustomTransport is enabled.
// if the config already has a defined transport, we don't override it.
func (impl *CustomK8sHttpTransportConfig) OverrideConfigWithCustomTransport(config *rest.Config) (*rest.Config, error) {
	if !impl.UseCustomTransport || config.Transport != nil {
		return config, nil
	}

	dial := (&net.Dialer{
		Timeout:   time.Duration(impl.TimeOut) * time.Second,
		KeepAlive: time.Duration(impl.KeepAlive) * time.Second,
	}).DialContext

	// Get the TLS options for this client config
	tlsConfig, err := rest.TLSConfigFor(config)
	if err != nil {
		return nil, err
	}

	transport := utilnet.SetTransportDefaults(&http.Transport{
		Proxy:               config.Proxy,
		TLSHandshakeTimeout: time.Duration(impl.TLSHandshakeTimeout) * time.Second,
		TLSClientConfig:     tlsConfig,
		MaxIdleConns:        impl.MaxIdleConnsPerHost,
		MaxConnsPerHost:     impl.MaxIdleConnsPerHost,
		MaxIdleConnsPerHost: impl.MaxIdleConnsPerHost,
		DialContext:         dial,
		DisableCompression:  config.DisableCompression,
		IdleConnTimeout:     time.Duration(impl.IdleConnTimeout) * time.Second,
	})

	rt, err := rest.HTTPWrappersForConfig(config, transport)
	if err != nil {
		return nil, err
	}

	config.Transport = rt
	config.Timeout = time.Duration(impl.TimeOut) * time.Second

	// set default tls config and remove auth/exec provides since we use it in a custom transport.
	// we already set tls config in the transport
	config.TLSClientConfig = rest.TLSClientConfig{}
	config.AuthProvider = nil
	config.ExecProvider = nil

	return config, nil
}

var NotFoundError = errors.New("not found")

func IsNotFoundError(err error) bool {
	return errors.Is(err, NotFoundError)
}
