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

type K8sApplicationService interface {
	GetResource(restConfig *rest.Config, request *K8sRequestBean) (resp *ManifestResponse, err error)
	UpdateResource(restConfig *rest.Config, request *K8sRequestBean) (resp *ManifestResponse, err error)
	DeleteResource(restConfig *rest.Config, request *K8sRequestBean) (resp *ManifestResponse, err error)
	ListEvents(restConfig *rest.Config, request *K8sRequestBean) (*EventsResponse, error)
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

func (impl K8sApplicationServiceImpl) GetResource(restConfig *rest.Config, request *K8sRequestBean) (*ManifestResponse, error) {
	resourceIf, err := impl.GetResourceIf(restConfig, request)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, err
	}
	resourceIdentifier := request.ResourceIdentifier
	var resp *unstructured.Unstructured
	if len(resourceIdentifier.Namespace) > 0 {
		resp, err = resourceIf.Namespace(resourceIdentifier.Namespace).Get(resourceIdentifier.Name, metav1.GetOptions{})
	} else {
		resp, err = resourceIf.Get(resourceIdentifier.Name, metav1.GetOptions{})
	}
	if err != nil {
		impl.logger.Errorw("error in getting resource", "err", err, "resource", resourceIdentifier.Name)
		return nil, err
	}
	return &ManifestResponse{*resp}, nil
}

func (impl K8sApplicationServiceImpl) UpdateResource(restConfig *rest.Config, request *K8sRequestBean) (*ManifestResponse, error) {
	resourceIf, err := impl.GetResourceIf(restConfig, request)
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
	if len(resourceIdentifier.Namespace) > 0 {
		resp, err = resourceIf.Namespace(resourceIdentifier.Namespace).Update(&unstructured.Unstructured{Object: updateObj}, metav1.UpdateOptions{})
	} else {
		resp, err = resourceIf.Update(&unstructured.Unstructured{Object: updateObj}, metav1.UpdateOptions{})
	}
	if err != nil {
		impl.logger.Errorw("error in updating resource", "err", err, "resource", resourceIdentifier.Name)
		return nil, err
	}
	return &ManifestResponse{*resp}, nil
}
func (impl K8sApplicationServiceImpl) DeleteResource(restConfig *rest.Config, request *K8sRequestBean) (*ManifestResponse, error) {
	resourceIf, err := impl.GetResourceIf(restConfig, request)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, err
	}
	resourceIdentifier := request.ResourceIdentifier
	var obj *unstructured.Unstructured
	if len(resourceIdentifier.Namespace) > 0 {
		obj, err = resourceIf.Namespace(resourceIdentifier.Namespace).Get(request.ResourceIdentifier.Name, metav1.GetOptions{})
		if err != nil {
			impl.logger.Errorw("error in getting resource", "err", err, "resource", resourceIdentifier.Name)
			return nil, err
		}
		err = resourceIf.Namespace(resourceIdentifier.Namespace).Delete(request.ResourceIdentifier.Name, &metav1.DeleteOptions{})
	} else {
		obj, err = resourceIf.Get(request.ResourceIdentifier.Name, metav1.GetOptions{})
		if err != nil {
			impl.logger.Errorw("error in getting resource", "err", err, "resource", resourceIdentifier.Name)
			return nil, err
		}
		err = resourceIf.Delete(request.ResourceIdentifier.Name, &metav1.DeleteOptions{})
	}
	if err != nil {
		impl.logger.Errorw("error in deleting resource", "err", err, "resource", resourceIdentifier.Name)
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}

func (impl K8sApplicationServiceImpl) ListEvents(restConfig *rest.Config, request *K8sRequestBean) (*EventsResponse, error) {
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

func (impl K8sApplicationServiceImpl) GetResourceIf(restConfig *rest.Config, request *K8sRequestBean) (resourceIf dynamic.NamespaceableResourceInterface, err error) {
	resourceIdentifier := request.ResourceIdentifier
	dynamicIf, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting k8s client", "err", err)
		return nil, err
	}
	apiResource, err := ServerResourceForGroupVersionKind(discoveryClient, resourceIdentifier.GroupVersionKind)
	if err != nil {
		impl.logger.Errorw("error in getting server resource", "err", err)
		return nil, err
	}
	resource := resourceIdentifier.GroupVersionKind.GroupVersion().WithResource(apiResource.Name)
	return dynamicIf.Resource(resource), nil
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
