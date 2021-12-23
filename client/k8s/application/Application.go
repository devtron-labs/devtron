package application

import (
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
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
	PatchType          string             `json:"patchType"`
}

type ResourceIdentifier struct {
	//TODO : update validations
	Name             string                  `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Namespace        string                  `protobuf:"bytes,2,req,name=namespace" json:"namespace,omitempty"`
	GroupVersionKind schema.GroupVersionKind `protobuf:"bytes,3,req,name=groupVersionKind" json:"groupVersionKind,omitempty"`
}

type ManifestResponse struct {
	//TODO : update validations
	Manifest unstructured.Unstructured `protobuf:"bytes,1,req,name=manifest" json:"manifest,omitempty"`
}

type EventsResponse struct {
	//TODO : update validations
	Events *v1b1.EventList
}

func (impl K8sApplicationServiceImpl) GetResource(request *K8sRequestBean) (*ManifestResponse, error) {
	dynamicIf, resource, err := impl.GetResourceAndDynamicIf(request)
	if err != nil {
		return nil, err
	}
	obj, err := dynamicIf.Resource(resource).Namespace(request.ResourceIdentifier.Namespace).Get(request.ResourceIdentifier.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}

func (impl K8sApplicationServiceImpl) UpdateResource(request *K8sRequestBean) (*ManifestResponse, error) {
	dynamicIf, resource, err := impl.GetResourceAndDynamicIf(request)
	if err != nil {
		return nil, err
	}
	obj, err := dynamicIf.Resource(resource).Namespace(request.ResourceIdentifier.Namespace).Patch(request.ResourceIdentifier.Name, types.PatchType(request.PatchType), []byte(request.Patch), metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}
func (impl K8sApplicationServiceImpl) DeleteResource(request *K8sRequestBean) (*ManifestResponse, error) {
	dynamicIf, resource, err := impl.GetResourceAndDynamicIf(request)
	if err != nil {
		return nil, err
	}
	obj, err := dynamicIf.Resource(resource).Namespace(request.ResourceIdentifier.Namespace).Get(request.ResourceIdentifier.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = dynamicIf.Resource(resource).Namespace(request.ResourceIdentifier.Namespace).Delete(request.ResourceIdentifier.Name, &metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}
	return &ManifestResponse{*obj}, nil
}

func (impl K8sApplicationServiceImpl) ListEvents(request *K8sRequestBean) (*EventsResponse, error) {
	restConfig, err := impl.GetClusterConfig(request.ClusterId)
	if err != nil {
		return nil, err
	}
	resourceIdentifier := request.ResourceIdentifier
	var resourceMap map[string]string
	resourceMap["involvedObject.apiVersion"] = resourceIdentifier.GroupVersionKind.GroupVersion().String()
	resourceMap["involvedObject.kind"] = resourceIdentifier.GroupVersionKind.Kind
	resourceMap["involvedObject.name"] = resourceIdentifier.Name
	resourceMap["involvedObject.namespace"] = resourceIdentifier.Namespace
	listOptions := metav1.ListOptions{
		FieldSelector: fields.Set(resourceMap).AsSelector().String(),
	}
	eventsClient, err := v1beta1.NewForConfig(restConfig)
	events, err := eventsClient.Events(resourceIdentifier.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}
	return &EventsResponse{events}, nil
}

func (impl K8sApplicationServiceImpl) GetResourceAndDynamicIf(request *K8sRequestBean) (dynamicIf dynamic.Interface, resource schema.GroupVersionResource, err error) {
	restConfig, err := impl.GetClusterConfig(request.ClusterId)
	//TODO : remove hardcoding for token and url
	restConfig.Host = "https://api.demo.devtron.info"
	restConfig.BearerToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6Ink4bnhvVGFaS011V1c3clRGR2pCTlhuY0hDcXFJT0NtaWUtYTR6bXJXSVkifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZXZ0cm9uY2QiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlY3JldC5uYW1lIjoiY2QtdXNlci10b2tlbi13azRmZyIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJjZC11c2VyIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiODY4ZWM0MTMtOWYxMS00ZmEyLWEzNDctMDU2NzE3ODVhNjNiIiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OmRldnRyb25jZDpjZC11c2VyIn0.sEa6N_kGUwt8HztMzPPIGexxUStefLdw3v0JkGDUNJXLnJXtnAO9NiY068ZRkSLo8yCRgpkmfoqSbQHIRj8qdGuVltVdqNaZHVIMzy9wNoDOcuW_VDBwuMGdm7duotEtDdrGkSXOVe2ezqQGfa1ZTWCFAS6CSozsjUmyyQoAofsSTYB7h6yYsqqDaz4AVXuhuJ1wwKBcI_pLu_Rhv8COkjKcz-Hwk-xGB4x8b_cBiZlnXnhhCsX6ClIvV0Qv2F-Q90k9h9RwNG26XwsAGqXafKEC6LlZ-FT51Q7Ta_ZjrG-GGKixPpd-oSAqEi19bjW1syFMfY4cCR2uC072ZQZirA"
	resourceIdentifier := request.ResourceIdentifier
	dynamicIf, err = dynamic.NewForConfig(restConfig)
	if err != nil {
		return dynamicIf, resource, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return dynamicIf, resource, err
	}
	apiResource, err := ServerResourceForGroupVersionKind(discoveryClient, resourceIdentifier.GroupVersionKind)
	if err != nil {
		return dynamicIf, resource, err
	}
	resource = resourceIdentifier.GroupVersionKind.GroupVersion().WithResource(apiResource.Name)
	return dynamicIf, resource, nil
}

func (impl K8sApplicationServiceImpl) GetClusterConfig(clusterId int) (*rest.Config, error) {
	cluster, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		return nil, err
	}
	configMap := cluster.Config
	bearerToken := configMap["bearer_token"]
	var restConfig *rest.Config
	if cluster.ClusterName == DEFAULT_CLUSTER && len(bearerToken) == 0 {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
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
