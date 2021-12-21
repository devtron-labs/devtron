package application

import (
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type k8sApplication interface {
	GetResource(request *GetRequest)(resp *ManifestResponse, err error)
	UpdateResource(request *UpdateRequest)(resp *ManifestResponse, err error)
	DeleteResource(request *DeleteRequest)(resp *ManifestResponse, err error)
}

type K8sApplicationServiceImpl struct {
	restConfig *rest.Config
	logger     *zap.SugaredLogger
}

func NewK8sApplicationServiceImpl(restConfig *rest.Config, logger *zap.SugaredLogger) *K8sApplicationServiceImpl {
	return &K8sApplicationServiceImpl{
		restConfig: restConfig,
		logger:     logger,
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
	//TODO : update validations, confirm for force delete flag
	Name             string                  `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Namespace        string                  `protobuf:"bytes,2,req,name=namespace" json:"namespace,omitempty"`
	GroupVersionKind schema.GroupVersionKind `protobuf:"bytes,3,req,name=groupVersionKind" json:"groupVersionKind,omitempty"`
	Force            *bool                   `protobuf:"bytes,4,req,name=force" json:"force,omitempty"`
}

type ManifestResponse struct {
	Manifest unstructured.Unstructured `protobuf:"bytes,1,req,name=manifest" json:"manifest,omitempty"`
}

func (impl K8sApplicationServiceImpl) GetResource(request *GetRequest)(*ManifestResponse, error) {
	dynamicIf, err := dynamic.NewForConfig(impl.restConfig)
	if err != nil {
		return nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(impl.restConfig)
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

func (impl K8sApplicationServiceImpl) UpdateResource(request *UpdateRequest) (*ManifestResponse, error) {
	dynamicIf, err := dynamic.NewForConfig(impl.restConfig)
	if err != nil {
		return nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(impl.restConfig)
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
func (impl K8sApplicationServiceImpl) DeleteResource(request *DeleteRequest) (*ManifestResponse, error) {
	dynamicIf, err := dynamic.NewForConfig(impl.restConfig)
	if err != nil {
		return nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(impl.restConfig)
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

func ServerResourceForGroupVersionKind(discoveryClient discovery.DiscoveryInterface, gvk schema.GroupVersionKind) (*metav1.APIResource, error) {
	resources, err := discoveryClient.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		return nil, err
	}
	for _, r := range resources.APIResources {
		if r.Kind == gvk.Kind {
			//log.Debugf("Chose API '%s' for %s", r.Name, gvk)
			return &r, nil
		}
	}
	return nil, errors.NewNotFound(schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind}, "")
}
