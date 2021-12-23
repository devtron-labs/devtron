package application

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
)

const DEFAULT_CLUSTER = "default_cluster"

type K8sApplicationService interface {
	GetResource(request *K8sRequestBean) (resp *ManifestResponse, err error)
	UpdateResource(request *K8sRequestBean) (resp *ManifestResponse, err error)
	DeleteResource(request *K8sRequestBean) (resp *ManifestResponse, err error)
	ListEvents(request *K8sRequestBean) (*EventsResponse, error)
}

type K8sApplicationServiceImpl struct {
	logger            *zap.SugaredLogger
	clusterRepository repository.ClusterRepository
}

func NewK8sApplicationServiceImpl(logger *zap.SugaredLogger, clusterRepository repository.ClusterRepository) *K8sApplicationServiceImpl {
	return &K8sApplicationServiceImpl{
		logger:            logger,
		clusterRepository: clusterRepository,
	}
}

type K8sRequestBean struct {
	//TODO : update validations
	AppId              int                `json:"appId"`
	ClusterId          int                `json:"clusterId"`
	ResourceIdentifier ResourceIdentifier `json:"resourceIdentifier"`
	Patch              string             `json:"patch"`
}

type ResourceIdentifier struct {
	//TODO : update validations
	Name             string                  `json:"name,omitempty"`
	Namespace        string                  `json:"namespace,omitempty"`
	GroupVersionKind schema.GroupVersionKind `json:"groupVersionKind,omitempty"`
}

type ManifestResponse struct {
	//TODO : update validations
	Manifest unstructured.Unstructured `json:"manifest,omitempty"`
}

type EventsResponse struct {
	//TODO : update validations
	Events *apiv1.EventList
}

func (impl K8sApplicationServiceImpl) GetResource(request *K8sRequestBean) (*ManifestResponse, error) {
	dynamicIf, resource, err := impl.GetResourceAndDynamicIf(request)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, err
	}
	obj, err := dynamicIf.Resource(resource).Namespace(request.ResourceIdentifier.Namespace).Get(request.ResourceIdentifier.Name, metav1.GetOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting resource", "err", err, "resource", request.ResourceIdentifier.Name)
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}

func (impl K8sApplicationServiceImpl) UpdateResource(request *K8sRequestBean) (*ManifestResponse, error) {
	dynamicIf, resource, err := impl.GetResourceAndDynamicIf(request)
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
	obj, err := dynamicIf.Resource(resource).Namespace(request.ResourceIdentifier.Namespace).Update(&unstructured.Unstructured{Object: updateObj}, metav1.UpdateOptions{})
	if err != nil {
		impl.logger.Errorw("error in updating resource", "err", err, "resource", request.ResourceIdentifier.Name)
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}
func (impl K8sApplicationServiceImpl) DeleteResource(request *K8sRequestBean) (*ManifestResponse, error) {
	dynamicIf, resource, err := impl.GetResourceAndDynamicIf(request)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, err
	}
	obj, err := dynamicIf.Resource(resource).Namespace(request.ResourceIdentifier.Namespace).Get(request.ResourceIdentifier.Name, metav1.GetOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting resource", "err", err, "resource", request.ResourceIdentifier.Name)
		return nil, err
	}
	err = dynamicIf.Resource(resource).Namespace(request.ResourceIdentifier.Namespace).Delete(request.ResourceIdentifier.Name, &metav1.DeleteOptions{})
	if err != nil {
		impl.logger.Errorw("error in deleting resource", "err", err, "resource", request.ResourceIdentifier.Name)
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}

func (impl K8sApplicationServiceImpl) ListEvents(request *K8sRequestBean) (*EventsResponse, error) {
	restConfig, err := impl.GetClusterConfig(request.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster config by ID", "err", err, "clusterid", request.ClusterId)
		return nil, err
	}
	resourceIdentifier := request.ResourceIdentifier
	resourceIdentifier.GroupVersionKind.Kind = "List"
	eventsClient, err := v1.NewForConfig(restConfig)
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
	list, err := eventsIf.List(listOptions)
	if err != nil {
		impl.logger.Errorw("error in getting events list", "err", err)
		return nil, err
	}
	return &EventsResponse{list}, nil
}

func (impl K8sApplicationServiceImpl) GetResourceAndDynamicIf(request *K8sRequestBean) (dynamicIf dynamic.Interface, resource schema.GroupVersionResource, err error) {
	restConfig, err := impl.GetClusterConfig(request.ClusterId)
	resourceIdentifier := request.ResourceIdentifier
	dynamicIf, err = dynamic.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return dynamicIf, resource, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting k8s client", "err", err)
		return dynamicIf, resource, err
	}
	apiResource, err := ServerResourceForGroupVersionKind(discoveryClient, resourceIdentifier.GroupVersionKind)
	if err != nil {
		impl.logger.Errorw("error in getting server resource", "err", err)
		return dynamicIf, resource, err
	}
	resource = resourceIdentifier.GroupVersionKind.GroupVersion().WithResource(apiResource.Name)
	return dynamicIf, resource, nil
}

func (impl K8sApplicationServiceImpl) GetClusterConfig(clusterId int) (*rest.Config, error) {
	cluster, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster by ID", "err", err, "clusterId")
		return nil, err
	}
	configMap := cluster.Config
	bearerToken := configMap["bearer_token"]
	var restConfig *rest.Config
	if cluster.ClusterName == DEFAULT_CLUSTER && len(bearerToken) == 0 {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			impl.logger.Errorw("error in getting cluster config for default cluster", "err", err)
			return nil, err
		}
	} else {
		restConfig = &rest.Config{Host: cluster.ServerUrl, BearerToken: bearerToken, TLSClientConfig: rest.TLSClientConfig{Insecure: true}}
	}
	return restConfig, nil
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
