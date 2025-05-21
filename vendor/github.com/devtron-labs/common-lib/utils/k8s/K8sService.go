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
	"context"
	"flag"
	"go.uber.org/zap"
	"io"
	batchV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"net/http"
	"os/user"
	"path/filepath"
)

type K8sService interface {
	GetLogsForAPod(kubeClient *kubernetes.Clientset, namespace string, podName string, container string, follow bool) *rest.Request
	GetMetricsClientSet(restConfig *rest.Config, k8sHttpClient *http.Client) (*metrics.Clientset, error)
	GetNmByName(ctx context.Context, metricsClientSet *metrics.Clientset, name string) (*v1beta1.NodeMetrics, error)
	GetNmList(ctx context.Context, metricsClientSet *metrics.Clientset) (*v1beta1.NodeMetricsList, error)
	GetPodsListForNamespace(ctx context.Context, k8sClientSet *kubernetes.Clientset, namespace string) (*v1.PodList, error)
	GetServerVersionFromDiscoveryClient(k8sClientSet *kubernetes.Clientset) (*version.Info, error)
	GetServerGroups(k8sClientSet *kubernetes.Clientset) (*metav1.APIGroupList, error)
	GetNodeByName(ctx context.Context, k8sClientSet *kubernetes.Clientset, name string) (*v1.Node, error)
	GetNodesList(ctx context.Context, k8sClientSet *kubernetes.Clientset) (*v1.NodeList, error)
	GetCoreV1ClientByRestConfig(restConfig *rest.Config) (*v12.CoreV1Client, error)
	GetCoreV1ClientInCluster(opts ...K8sServiceOpts) (*v12.CoreV1Client, error)
	GetKubeVersion() (*version.Info, error)
	ValidateResource(resourceObj map[string]interface{}, gvk schema.GroupVersionKind, validateCallback func(namespace string, group string, kind string, resourceName string) bool) bool
	BuildK8sObjectListTableData(manifest *unstructured.UnstructuredList, namespaced bool, gvk schema.GroupVersionKind, includeMetadata bool, validateResourceAccess func(namespace string, group string, kind string, resourceName string) bool) (*ClusterResourceListMap, error)
	ValidateForResource(namespace string, resourceRef interface{}, validateCallback func(namespace string, group string, kind string, resourceName string) bool) bool
	GetPodByName(namespace string, name string, client *v12.CoreV1Client) (*v1.Pod, error)
	GetResourceInfoByLabelSelector(ctx context.Context, namespace string, labelSelector string) (*v1.Pod, error)
	GetClientByToken(serverUrl string, token map[string]string) (*v12.CoreV1Client, error)
	ListNamespaces(client *v12.CoreV1Client) (*v1.NamespaceList, error)
	DeleteAndCreateJob(content []byte, namespace string, clusterConfig *ClusterConfig) error
	DeletePodByLabel(namespace string, labels string, clusterConfig *ClusterConfig, opts ...K8sServiceOpts) error
	CreateJob(namespace string, name string, clusterConfig *ClusterConfig, job *batchV1.Job, opts ...K8sServiceOpts) error
	GetLiveZCall(path string, k8sClientSet *kubernetes.Clientset) ([]byte, error)
	DiscoveryClientGetLiveZCall(cluster *ClusterConfig, opts ...K8sServiceOpts) ([]byte, error)
	DeleteJob(namespace string, name string, clusterConfig *ClusterConfig, opts ...K8sServiceOpts) error
	DeleteSecret(namespace string, name string, client *v12.CoreV1Client) error
	UpdateSecret(namespace string, secret *v1.Secret, client *v12.CoreV1Client) (*v1.Secret, error)
	CreateSecretData(namespace string, secret *v1.Secret, v1Client *v12.CoreV1Client) (*v1.Secret, error)
	CreateSecret(namespace string, data map[string][]byte, secretName string, secretType v1.SecretType, client *v12.CoreV1Client, labels map[string]string, stringData map[string]string) (*v1.Secret, error)
	GetSecret(namespace string, name string, client *v12.CoreV1Client) (*v1.Secret, error)
	GetSecretWithCtx(ctx context.Context, namespace string, name string, client *v12.CoreV1Client) (*v1.Secret, error)
	PatchConfigMapJsonType(namespace string, clusterConfig *ClusterConfig, name string, data interface{}, path string) (*v1.ConfigMap, error)
	PatchConfigMap(namespace string, clusterConfig *ClusterConfig, name string, data map[string]interface{}) (*v1.ConfigMap, error)
	UpdateConfigMap(namespace string, cm *v1.ConfigMap, client *v12.CoreV1Client) (*v1.ConfigMap, error)
	CreateConfigMap(namespace string, cm *v1.ConfigMap, client *v12.CoreV1Client) (*v1.ConfigMap, error)
	GetConfigMap(namespace string, name string, client *v12.CoreV1Client) (*v1.ConfigMap, error)
	GetConfigMapWithCtx(ctx context.Context, namespace string, name string, client *v12.CoreV1Client) (*v1.ConfigMap, error)
	GetNsIfExists(namespace string, client *v12.CoreV1Client) (ns *v1.Namespace, exists bool, err error)
	CreateNsIfNotExists(namespace string, clusterConfig *ClusterConfig) (ns *v1.Namespace, nsCreated bool, err error)
	UpdateNSLabels(namespace *v1.Namespace, labels map[string]string, clusterConfig *ClusterConfig) (ns *v1.Namespace, err error)
	GetK8sDiscoveryClientInCluster(opts ...K8sServiceOpts) (*discovery.DiscoveryClient, error)
	GetK8sDiscoveryClient(clusterConfig *ClusterConfig, opts ...K8sServiceOpts) (*discovery.DiscoveryClient, error)
	GetClientForInCluster(opts ...K8sServiceOpts) (*v12.CoreV1Client, error)
	GetCoreV1Client(clusterConfig *ClusterConfig, opts ...K8sServiceOpts) (*v12.CoreV1Client, error)
	GetResource(ctx context.Context, namespace string, name string, gvk schema.GroupVersionKind, restConfig *rest.Config) (*ManifestResponse, error)
	UpdateResource(ctx context.Context, restConfig *rest.Config, gvk schema.GroupVersionKind, namespace string, k8sRequestPatch string) (*ManifestResponse, error)
	DeleteResource(ctx context.Context, restConfig *rest.Config, gvk schema.GroupVersionKind, namespace string, name string, forceDelete bool) (*ManifestResponse, error)
	GetPodListByLabel(namespace, label string, clientSet *kubernetes.Clientset) ([]v1.Pod, error)
	ExtractK8sServerMajorAndMinorVersion(k8sServerVersion *version.Info) (int, int, error)
	GetK8sServerVersion(clientSet *kubernetes.Clientset) (*version.Info, error)
	DecodeGroupKindversion(data string) (*schema.GroupVersionKind, error)
	GetApiResources(restConfig *rest.Config, includeOnlyVerb string) ([]*K8sApiResource, error)
	CreateResources(ctx context.Context, restConfig *rest.Config, manifest string, gvk schema.GroupVersionKind, namespace string) (*ManifestResponse, error)
	PatchResourceRequest(ctx context.Context, restConfig *rest.Config, pt types.PatchType, manifest string, name string, namespace string, gvk schema.GroupVersionKind) (*ManifestResponse, error)
	GetResourceList(ctx context.Context, restConfig *rest.Config, gvk schema.GroupVersionKind, namespace string, asTable bool, listOptions *metav1.ListOptions) (*ResourceListResponse, bool, error)
	GetResourceIfWithAcceptHeader(restConfig *rest.Config, groupVersionKind schema.GroupVersionKind, asTable bool) (resourceIf dynamic.NamespaceableResourceInterface, namespaced bool, err error)
	GetPodLogs(ctx context.Context, restConfig *rest.Config, name string, namespace string, sinceTime *metav1.Time, tailLines int, sinceSeconds int, follow bool, containerName string, isPrevContainerLogsEnabled bool) (io.ReadCloser, error)
	ListEvents(restConfig *rest.Config, namespace string, groupVersionKind schema.GroupVersionKind, ctx context.Context, name string) (*v1.EventList, error)
	GetResourceIf(restConfig *rest.Config, groupVersionKind schema.GroupVersionKind) (resourceIf dynamic.NamespaceableResourceInterface, namespaced bool, err error)
	FetchConnectionStatusForCluster(k8sClientSet *kubernetes.Clientset) error
	CreateK8sClientSet(restConfig *rest.Config) (*kubernetes.Clientset, error)
	CreateOrUpdateSecretByName(client *v12.CoreV1Client, namespace, uniqueSecretName string, secretLabel map[string]string, secretData map[string]string) error

	// below functions are exposed for K8sUtilExtended

	CreateNsWithLabels(namespace string, labels map[string]string, client *v12.CoreV1Client) (ns *v1.Namespace, err error)
	CreateNs(namespace string, client *v12.CoreV1Client) (ns *v1.Namespace, err error)
	GetGVRForCRD(config *rest.Config, CRDName string) (schema.GroupVersionResource, error)
	GetResourceByGVR(ctx context.Context, config *rest.Config, GVR schema.GroupVersionResource, resourceName, namespace string) (*unstructured.Unstructured, error)
	PatchResourceByGVR(ctx context.Context, config *rest.Config, GVR schema.GroupVersionResource, resourceName, namespace string, patchType types.PatchType, patchData []byte) (*unstructured.Unstructured, error)
	DeleteResourceByGVR(ctx context.Context, config *rest.Config, GVR schema.GroupVersionResource, resourceName, namespace string, forceDelete bool) error

	// k8s rest config methods

	GetK8sInClusterRestConfig(opts ...K8sServiceOpts) (*rest.Config, error)
	GetK8sConfigAndClients(clusterConfig *ClusterConfig, opts ...K8sServiceOpts) (*rest.Config, *http.Client, *kubernetes.Clientset, error)
	GetK8sInClusterConfigAndDynamicClients(opts ...K8sServiceOpts) (*rest.Config, *http.Client, dynamic.Interface, error)
	GetK8sInClusterConfigAndClients(opts ...K8sServiceOpts) (*rest.Config, *http.Client, *kubernetes.Clientset, error)
	GetRestConfigByCluster(clusterConfig *ClusterConfig, opts ...K8sServiceOpts) (*rest.Config, error)
	OverrideRestConfigWithCustomTransport(restConfig *rest.Config, opts ...K8sServiceOpts) (*rest.Config, error)
	GetK8sConfigAndClientsByRestConfig(restConfig *rest.Config, opts ...K8sServiceOpts) (*http.Client, *kubernetes.Clientset, error)
}

type K8sServiceImpl struct {
	logger              *zap.SugaredLogger
	runTimeConfig       *RuntimeConfig
	httpTransportConfig *HttpTransportConfig
	kubeconfig          *string
	opts                Options
}

func (impl *K8sServiceImpl) SetCustomHttpClientConfig(customHttpClientConfig HttpTransportInterface) *K8sServiceImpl {
	impl.httpTransportConfig.customHttpClientConfig = customHttpClientConfig
	return impl
}

func (impl *K8sServiceImpl) GetCustomHttpClientConfig() HttpTransportInterface {
	return impl.httpTransportConfig.customHttpClientConfig
}

func (impl *K8sServiceImpl) SetDefaultHttpClientConfig(defaultHttpClientConfig HttpTransportInterface) *K8sServiceImpl {
	impl.httpTransportConfig.defaultHttpClientConfig = defaultHttpClientConfig
	return impl
}

func (impl *K8sServiceImpl) GetDefaultHttpClientConfig() HttpTransportInterface {
	return impl.httpTransportConfig.defaultHttpClientConfig
}

type Options struct {
	transportType TransportType
}

func (opt *Options) SetTransportType(transportType TransportType) {
	opt.transportType = transportType
}

func (opt *Options) GetTransportType() TransportType {
	return opt.transportType
}

func NewK8sUtil(
	logger *zap.SugaredLogger,
	runTimeConfig *RuntimeConfig,
) (*K8sServiceImpl, error) {
	var kubeconfig *string
	if runTimeConfig.LocalDevMode {
		usr, err := user.Current()
		if err != nil {
			logger.Errorw("error in NewK8sUtil, failed to get current user", "err", err)
			return nil, err
		}
		kubeconfig = flag.String("kubeconfig-authenticator-xyz", filepath.Join(usr.HomeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	}
	flag.Parse()
	return &K8sServiceImpl{
		logger:              logger,
		runTimeConfig:       runTimeConfig,
		kubeconfig:          kubeconfig,
		httpTransportConfig: NewHttpTransportConfig(logger),
	}, nil
}

func (impl *K8sServiceImpl) NewKubeConfigImpl(
	httpTransportConfig HttpTransportInterface,
	kubeConfigBuilder KubeConfigBuilderInterface,
) *KubeConfigImpl {
	return NewKubeConfigImpl(
		impl.logger,
		impl.runTimeConfig,
		impl.kubeconfig,
		httpTransportConfig,
		kubeConfigBuilder,
	)
}

// WithHttpTransport toggles between default and overridden transport.
// This is used to create a new KubeConfigImpl with the specified transport type.
// It is used in K8sUtilExtended to override the NewKubeConfigBuilder.
// NOTE: Any modifications here is subject to change in K8sUtilExtended as well.
func (impl *K8sServiceImpl) WithHttpTransport(opt Options) KubeConfigInterface {
	switch opt.GetTransportType() {
	case TransportTypeDefault:
		return impl.NewKubeConfigImpl(impl.GetDefaultHttpClientConfig(), NewKubeConfigBuilder())
	default:
		// default fallback is custom transport
		return impl.NewKubeConfigImpl(impl.GetCustomHttpClientConfig(), NewKubeConfigBuilder())
	}
}

type K8sServiceOpts func(Options) Options

// WithDefaultHttpTransport ensures the transport type is default and rest config is not modified
// This is necessary when we use clients such as SPDY, that has its own transport
func WithDefaultHttpTransport() K8sServiceOpts {
	return func(opts Options) Options {
		opts.transportType = TransportTypeDefault
		return opts
	}
}

// WithOverriddenHttpTransport ensures the transport is overridden and rest config is modified
func WithOverriddenHttpTransport() K8sServiceOpts {
	return func(opts Options) Options {
		opts.transportType = TransportTypeOverridden
		return opts
	}
}
