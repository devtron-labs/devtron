package application

import (
	"context"
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
	"io"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	"net/http"
	"strings"
)

type K8sClientService interface {
	GetResource(restConfig *rest.Config, request *K8sRequestBean) (resp *ManifestResponse, err error)
	CreateResource(restConfig *rest.Config, request *K8sRequestBean, manifest string) (resp *ManifestResponse, err error)
	UpdateResource(restConfig *rest.Config, request *K8sRequestBean) (resp *ManifestResponse, err error)
	DeleteResource(restConfig *rest.Config, request *K8sRequestBean) (resp *ManifestResponse, err error)
	ListEvents(restConfig *rest.Config, request *K8sRequestBean) (*EventsResponse, error)
	GetPodLogs(restConfig *rest.Config, request *K8sRequestBean) (io.ReadCloser, error)
	GetApiResources(restConfig *rest.Config, includeOnlyVerb string) ([]*K8sApiResource, error)
	GetResourceList(restConfig *rest.Config, request *K8sRequestBean) (*ResourceListResponse, bool, error)
	ApplyResource(restConfig *rest.Config, request *K8sRequestBean, manifest string) (*ManifestResponse, error)
}

type K8sClientServiceImpl struct {
	logger            *zap.SugaredLogger
	clusterRepository repository.ClusterRepository
}

func NewK8sClientServiceImpl(logger *zap.SugaredLogger, clusterRepository repository.ClusterRepository) *K8sClientServiceImpl {
	return &K8sClientServiceImpl{
		logger:            logger,
		clusterRepository: clusterRepository,
	}
}

type K8sRequestBean struct {
	ResourceIdentifier ResourceIdentifier `json:"resourceIdentifier"`
	Patch              string             `json:"patch,omitempty"`
	PodLogsRequest     PodLogsRequest     `json:"podLogsRequest,omitempty"`
}

type PodLogsRequest struct {
	SinceTime     *metav1.Time `json:"sinceTime,omitempty"`
	TailLines     int          `json:"tailLines"`
	Follow        bool         `json:"follow"`
	ContainerName string       `json:"containerName"`
}

type ResourceIdentifier struct {
	Name             string                  `json:"name"` //pod name for logs request
	Namespace        string                  `json:"namespace"`
	GroupVersionKind schema.GroupVersionKind `json:"groupVersionKind"`
}

type ManifestResponse struct {
	Manifest unstructured.Unstructured `json:"manifest,omitempty"`
}

type EventsResponse struct {
	Events *apiv1.EventList `json:"events,omitempty"`
}

type ResourceListResponse struct {
	Resources unstructured.UnstructuredList `json:"resources,omitempty"`
}

func (impl K8sClientServiceImpl) GetResource(restConfig *rest.Config, request *K8sRequestBean) (*ManifestResponse, error) {
	resourceIf, namespaced, err := impl.GetResourceIf(restConfig, request)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, err
	}
	resourceIdentifier := request.ResourceIdentifier
	var resp *unstructured.Unstructured
	if len(resourceIdentifier.Namespace) > 0 && namespaced {
		resp, err = resourceIf.Namespace(resourceIdentifier.Namespace).Get(context.Background(), resourceIdentifier.Name, metav1.GetOptions{})
	} else {
		resp, err = resourceIf.Get(context.Background(), resourceIdentifier.Name, metav1.GetOptions{})
	}
	if err != nil {
		impl.logger.Errorw("error in getting resource", "err", err, "resource", resourceIdentifier.Name)
		return nil, err
	}
	return &ManifestResponse{*resp}, nil
}

func (impl K8sClientServiceImpl) CreateResource(restConfig *rest.Config, request *K8sRequestBean, manifest string) (*ManifestResponse, error) {
	resourceIf, namespaced, err := impl.GetResourceIf(restConfig, request)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, err
	}
	var createObj map[string]interface{}
	err = json.Unmarshal([]byte(manifest), &createObj)
	if err != nil {
		impl.logger.Errorw("error in json un-marshaling patch(manifest) string for creating resource", "err", err, "manifest", request.Patch)
		return nil, err
	}
	resourceIdentifier := request.ResourceIdentifier
	var resp *unstructured.Unstructured
	if len(resourceIdentifier.Namespace) > 0 && namespaced {
		resp, err = resourceIf.Namespace(resourceIdentifier.Namespace).Create(context.Background(), &unstructured.Unstructured{Object: createObj}, metav1.CreateOptions{})
	} else {
		resp, err = resourceIf.Create(context.Background(), &unstructured.Unstructured{Object: createObj}, metav1.CreateOptions{})
	}
	if err != nil {
		impl.logger.Errorw("error in creating resource", "err", err)
		return nil, err
	}
	return &ManifestResponse{*resp}, nil
}

func (impl K8sClientServiceImpl) UpdateResource(restConfig *rest.Config, request *K8sRequestBean) (*ManifestResponse, error) {
	resourceIf, namespaced, err := impl.GetResourceIf(restConfig, request)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, err
	}
	var updateObj map[string]interface{}
	err = json.Unmarshal([]byte(request.Patch), &updateObj)
	if err != nil {
		impl.logger.Errorw("error in json un-marshaling patch string for updating resource ", "err", err)
		return nil, err
	}
	resourceIdentifier := request.ResourceIdentifier
	var resp *unstructured.Unstructured
	if len(resourceIdentifier.Namespace) > 0 && namespaced {
		resp, err = resourceIf.Namespace(resourceIdentifier.Namespace).Update(context.Background(), &unstructured.Unstructured{Object: updateObj}, metav1.UpdateOptions{})
	} else {
		resp, err = resourceIf.Update(context.Background(), &unstructured.Unstructured{Object: updateObj}, metav1.UpdateOptions{})
	}
	if err != nil {
		impl.logger.Errorw("error in updating resource", "err", err, "resource", resourceIdentifier.Name)
		return nil, err
	}
	return &ManifestResponse{*resp}, nil
}
func (impl K8sClientServiceImpl) DeleteResource(restConfig *rest.Config, request *K8sRequestBean) (*ManifestResponse, error) {
	resourceIf, namespaced, err := impl.GetResourceIf(restConfig, request)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, err
	}
	resourceIdentifier := request.ResourceIdentifier
	var obj *unstructured.Unstructured
	if len(resourceIdentifier.Namespace) > 0 && namespaced {
		obj, err = resourceIf.Namespace(resourceIdentifier.Namespace).Get(context.Background(), request.ResourceIdentifier.Name, metav1.GetOptions{})
		if err != nil {
			impl.logger.Errorw("error in getting resource", "err", err, "resource", resourceIdentifier.Name)
			return nil, err
		}
		err = resourceIf.Namespace(resourceIdentifier.Namespace).Delete(context.Background(), request.ResourceIdentifier.Name, metav1.DeleteOptions{})
	} else {
		obj, err = resourceIf.Get(context.Background(), request.ResourceIdentifier.Name, metav1.GetOptions{})
		if err != nil {
			impl.logger.Errorw("error in getting resource", "err", err, "resource", resourceIdentifier.Name)
			return nil, err
		}
		err = resourceIf.Delete(context.Background(), request.ResourceIdentifier.Name, metav1.DeleteOptions{})
	}
	if err != nil {
		impl.logger.Errorw("error in deleting resource", "err", err, "resource", resourceIdentifier.Name)
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}

func (impl K8sClientServiceImpl) ListEvents(restConfig *rest.Config, request *K8sRequestBean) (*EventsResponse, error) {
	_, namespaced, err := impl.GetResourceIf(restConfig, request)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, err
	}

	resourceIdentifier := request.ResourceIdentifier
	resourceIdentifier.GroupVersionKind.Kind = "List"
	if !namespaced {
		resourceIdentifier.Namespace = "default"
	}
	eventsClient, err := v1.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting client for resource", "err", err)
		return nil, err
	}
	eventsIf := eventsClient.Events(resourceIdentifier.Namespace)
	eventsExp := eventsIf.(v1.EventExpansion)
	fieldSelector := eventsExp.GetFieldSelector(pointer.StringPtr(resourceIdentifier.Name), pointer.StringPtr(resourceIdentifier.Namespace), nil, nil)
	listOptions := metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       resourceIdentifier.GroupVersionKind.Kind,
			APIVersion: resourceIdentifier.GroupVersionKind.GroupVersion().String(),
		},
		FieldSelector: fieldSelector.String(),
	}
	list, err := eventsIf.List(context.Background(), listOptions)
	if err != nil {
		impl.logger.Errorw("error in getting events list", "err", err)
		return nil, err
	}
	return &EventsResponse{list}, nil
}

func (impl K8sClientServiceImpl) GetPodLogs(restConfig *rest.Config, request *K8sRequestBean) (io.ReadCloser, error) {
	resourceIdentifier := request.ResourceIdentifier
	podLogsRequest := request.PodLogsRequest
	podClient, err := v1.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting client for resource", "err", err)
		return nil, err
	}
	tailLines := int64(podLogsRequest.TailLines)
	podLogOptions := &apiv1.PodLogOptions{
		Follow:     podLogsRequest.Follow,
		TailLines:  &tailLines,
		Container:  podLogsRequest.ContainerName,
		Timestamps: true,
	}
	if podLogsRequest.SinceTime != nil {
		podLogOptions.SinceTime = podLogsRequest.SinceTime
	}
	podIf := podClient.Pods(resourceIdentifier.Namespace)
	logsRequest := podIf.GetLogs(resourceIdentifier.Name, podLogOptions)
	stream, err := logsRequest.Stream(context.Background())
	if err != nil {
		impl.logger.Errorw("error in streaming pod logs", "err", err)
		return nil, err
	}
	return stream, nil
}

func (impl K8sClientServiceImpl) GetResourceIf(restConfig *rest.Config, request *K8sRequestBean) (resourceIf dynamic.NamespaceableResourceInterface, namespaced bool, err error) {
	resourceIdentifier := request.ResourceIdentifier
	dynamicIf, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, false, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting k8s client", "err", err)
		return nil, false, err
	}
	apiResource, err := ServerResourceForGroupVersionKind(discoveryClient, resourceIdentifier.GroupVersionKind)
	if err != nil {
		impl.logger.Errorw("error in getting server resource", "err", err)
		return nil, false, err
	}
	resource := resourceIdentifier.GroupVersionKind.GroupVersion().WithResource(apiResource.Name)
	return dynamicIf.Resource(resource), apiResource.Namespaced, nil
}

func (impl K8sClientServiceImpl) GetResourceIfWithAcceptHeader(restConfig *rest.Config, request *K8sRequestBean) (resourceIf dynamic.NamespaceableResourceInterface, namespaced bool, err error) {
	resourceIdentifier := request.ResourceIdentifier
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting k8s client", "err", err)
		return nil, false, err
	}
	apiResource, err := ServerResourceForGroupVersionKind(discoveryClient, resourceIdentifier.GroupVersionKind)
	if err != nil {
		impl.logger.Errorw("error in getting server resource", "err", err)
		return nil, false, err
	}
	resource := resourceIdentifier.GroupVersionKind.GroupVersion().WithResource(apiResource.Name)
	wt := restConfig.WrapTransport // Reference: https://github.com/kubernetes/client-go/issues/407
	restConfig.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		if wt != nil {
			rt = wt(rt)
		}
		return &HeaderAdder{
			rt: rt,
		}
	}
	dynamicIf, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, false, err
	}
	return dynamicIf.Resource(resource), apiResource.Namespaced, nil
}

func ServerResourceForGroupVersionKind(discoveryClient discovery.DiscoveryInterface, gvk schema.GroupVersionKind) (*metav1.APIResource, error) {
	resources, err := discoveryClient.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		return nil, err
	}
	for _, r := range resources.APIResources {
		if r.Kind == gvk.Kind {
			return &r, nil
		}
	}
	return nil, errors.NewNotFound(schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind}, "")
}

// if verb is supplied empty, that means - return all
func (impl K8sClientServiceImpl) GetApiResources(restConfig *rest.Config, includeOnlyVerb string) ([]*K8sApiResource, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic k8s client", "err", err)
		return nil, err
	}

	apiResourcesListFromK8s, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		impl.logger.Errorw("error in getting api-resources from k8s", "err", err)
		return nil, err
	}

	apiResources := make([]*K8sApiResource, 0)
	for _, apiResourceListFromK8s := range apiResourcesListFromK8s {
		if apiResourceListFromK8s != nil {
			for _, apiResourceFromK8s := range apiResourceListFromK8s.APIResources {
				var includeResource bool
				if len(includeOnlyVerb) > 0 {
					for _, verb := range apiResourceFromK8s.Verbs {
						if verb == includeOnlyVerb {
							includeResource = true
							break
						}
					}
				} else {
					includeResource = true
				}
				if !includeResource {
					continue
				}
				var group string
				var version string
				gv := apiResourceListFromK8s.GroupVersion
				if len(gv) > 0 {
					splitGv := strings.Split(gv, "/")
					if len(splitGv) == 1 {
						version = splitGv[0]
					} else {
						group = splitGv[0]
						version = splitGv[1]
					}
				}
				apiResources = append(apiResources, &K8sApiResource{
					Gvk: schema.GroupVersionKind{
						Group:   group,
						Version: version,
						Kind:    apiResourceFromK8s.Kind,
					},
					Namespaced: apiResourceFromK8s.Namespaced,
				})
			}
		}
	}
	return apiResources, nil
}

func (impl K8sClientServiceImpl) GetResourceList(restConfig *rest.Config, request *K8sRequestBean) (*ResourceListResponse, bool, error) {
	resourceIf, namespaced, err := impl.GetResourceIfWithAcceptHeader(restConfig, request)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, namespaced, err
	}
	resourceIdentifier := request.ResourceIdentifier
	var resp *unstructured.UnstructuredList
	listOptions := metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       resourceIdentifier.GroupVersionKind.Kind,
			APIVersion: resourceIdentifier.GroupVersionKind.GroupVersion().String(),
		},
	}
	if len(resourceIdentifier.Namespace) > 0 && namespaced {
		resp, err = resourceIf.Namespace(resourceIdentifier.Namespace).List(context.Background(), listOptions)
	} else {
		resp, err = resourceIf.List(context.Background(), listOptions)
	}
	if err != nil {
		impl.logger.Errorw("error in getting resource", "err", err, "resource", resourceIdentifier)
		return nil, namespaced, err
	}
	return &ResourceListResponse{*resp}, namespaced, nil
}

func (impl K8sClientServiceImpl) ApplyResource(restConfig *rest.Config, request *K8sRequestBean, manifest string) (*ManifestResponse, error) {
	resourceIf, namespaced, err := impl.GetResourceIf(restConfig, request)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, err
	}
	resourceIdentifier := request.ResourceIdentifier
	var resp *unstructured.Unstructured
	if len(resourceIdentifier.Namespace) > 0 && namespaced {
		resp, err = resourceIf.Namespace(resourceIdentifier.Namespace).Patch(context.Background(), resourceIdentifier.Name, types.StrategicMergePatchType, []byte(manifest), metav1.PatchOptions{FieldManager: "patch"})
	} else {
		resp, err = resourceIf.Patch(context.Background(), resourceIdentifier.Name, types.StrategicMergePatchType, []byte(manifest), metav1.PatchOptions{FieldManager: "patch"})
	}
	if err != nil {
		impl.logger.Errorw("error in applying resource", "err", err)
		return nil, err
	}
	return &ManifestResponse{*resp}, nil
}
