package application

import (
	"go.uber.org/zap"
	v1b1 "k8s.io/api/events/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/typed/events/v1beta1"
	"k8s.io/client-go/rest"
)

type K8sApplicationService interface {
	GetResource(token string, request *GetRequest) (resp *ManifestResponse, err error)
	UpdateResource(token string, request *UpdateRequest) (resp *ManifestResponse, err error)
	DeleteResource(token string, request *DeleteRequest) (resp *ManifestResponse, err error)
	ListEvents(token string, request *GetRequest) (*EventsResponse, error)
}

type K8sApplicationServiceImpl struct {
	logger        *zap.SugaredLogger
}

func NewK8sApplicationServiceImpl(logger *zap.SugaredLogger) *K8sApplicationServiceImpl {
	return &K8sApplicationServiceImpl{
		logger:        logger,
	}
}

type GetRequest struct {
	//TODO : update validations
	Name             string                  `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Namespace        string                  `protobuf:"bytes,2,req,name=namespace" json:"namespace,omitempty"`
	GroupVersionKind schema.GroupVersionKind `protobuf:"bytes,3,req,name=groupVersionKind" json:"groupVersionKind,omitempty"`
}

type UpdateRequest struct {
	//TODO : update validations
	Name             string                  `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Namespace        string                  `protobuf:"bytes,2,req,name=namespace" json:"namespace,omitempty"`
	GroupVersionKind schema.GroupVersionKind `protobuf:"bytes,3,req,name=groupVersionKind" json:"groupVersionKind,omitempty"`
	Patch            string                  `protobuf:"bytes,4,req,name=patch" json:"patch,omitempty"`
	PatchType        string                  `protobuf:"bytes,5,req,name=patchType" json:"patchType,omitempty"`
}

type DeleteRequest struct {
	//TODO : update validations
	Name             string                  `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Namespace        string                  `protobuf:"bytes,2,req,name=namespace" json:"namespace,omitempty"`
	GroupVersionKind schema.GroupVersionKind `protobuf:"bytes,3,req,name=groupVersionKind" json:"groupVersionKind,omitempty"`
	Force            *bool                   `protobuf:"bytes,4,req,name=force" json:"force,omitempty"`
}

type ManifestResponse struct {
	//TODO : update validations
	Manifest unstructured.Unstructured `protobuf:"bytes,1,req,name=manifest" json:"manifest,omitempty"`
}

type EventsResponse struct {
	//TODO : update validations
	Events *v1b1.EventList
}

func (impl K8sApplicationServiceImpl) GetResource(token string, request *GetRequest) (*ManifestResponse, error) {
	restConfig := &rest.Config{
		BearerToken: token,
	}
	dynamicIf, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	apiResource, err := ServerResourceForGroupVersionKind(discoveryClient, request.GroupVersionKind)
	if err != nil {
		return nil, err
	}
	resource := request.GroupVersionKind.GroupVersion().WithResource(apiResource.Name)
	//TODO : confirm for client-go version, updated version have changed arguments
	obj, err := dynamicIf.Resource(resource).Namespace(request.Namespace).Get(request.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}

func (impl K8sApplicationServiceImpl) UpdateResource(token string, request *UpdateRequest) (*ManifestResponse, error) {
	restConfig := &rest.Config{
		BearerToken: token,
	}
	dynamicIf, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	apiResource, err := ServerResourceForGroupVersionKind(discoveryClient, request.GroupVersionKind)
	if err != nil {
		return nil, err
	}
	resource := request.GroupVersionKind.GroupVersion().WithResource(apiResource.Name)
	obj, err := dynamicIf.Resource(resource).Namespace(request.Namespace).Patch(request.Name, types.PatchType(request.PatchType), []byte(request.Patch), metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}
func (impl K8sApplicationServiceImpl) DeleteResource(token string, request *DeleteRequest) (*ManifestResponse, error) {
	restConfig := &rest.Config{
		BearerToken: token,
	}
	dynamicIf, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	apiResource, err := ServerResourceForGroupVersionKind(discoveryClient, request.GroupVersionKind)
	if err != nil {
		return nil, err
	}
	resource := request.GroupVersionKind.GroupVersion().WithResource(apiResource.Name)
	obj, err := dynamicIf.Resource(resource).Namespace(request.Namespace).Get(request.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = dynamicIf.Resource(resource).Namespace(request.Namespace).Delete(request.Name, &metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}

func (impl K8sApplicationServiceImpl) ListEvents(token string, request *GetRequest) (*EventsResponse, error) {
	restConfig := &rest.Config{
		BearerToken: token,
	}
	var resourceMap map[string]string
	resourceMap["involvedObject.apiVersion"] = request.GroupVersionKind.GroupVersion().String()
	resourceMap["involvedObject.kind"] = request.GroupVersionKind.Kind
	resourceMap["involvedObject.name"] = request.Name
	resourceMap["involvedObject.namespace"] = request.Namespace
	listOptions := metav1.ListOptions{
		FieldSelector: fields.Set(resourceMap).AsSelector().String(),
	}
	eventsClient, err := v1beta1.NewForConfig(restConfig)
	events, err := eventsClient.Events(request.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}
	return &EventsResponse{events}, nil
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
