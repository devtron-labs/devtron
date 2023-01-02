package application

import (
	"context"
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type dynamicClient struct {
	client *rest.RESTClient
}

type dynamicResourceClient struct {
	client    *dynamicClient
	namespace string
	resource  schema.GroupVersionResource
}

var parameterScheme = runtime.NewScheme()
var dynamicParameterCodec = runtime.NewParameterCodec(parameterScheme)
var versionV1 = schema.GroupVersion{Version: "v1"}
var deleteScheme = runtime.NewScheme()
var deleteOptionsCodec = serializer.NewCodecFactory(deleteScheme)

func init() {
	metav1.AddToGroupVersion(parameterScheme, versionV1)
	metav1.AddToGroupVersion(deleteScheme, versionV1)
}

type Interface interface {
	Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface
}

func (c *dynamicClient) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return &dynamicResourceClient{client: c, resource: resource}
}

func (c *dynamicResourceClient) Create(ctx context.Context, obj *unstructured.Unstructured, opts metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, errors.New("not implemented")
}

func (c *dynamicResourceClient) Update(ctx context.Context, obj *unstructured.Unstructured, opts metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, errors.New("not implemented")
}

func (c *dynamicResourceClient) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, opts metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, errors.New("not implemented")
}

func (c *dynamicResourceClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions, subresources ...string) error {
	return errors.New("not implemented")
}

func (c *dynamicResourceClient) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return errors.New("not implemented")
}

func (c *dynamicResourceClient) Get(ctx context.Context, name string, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, errors.New("not implemented")
}

func (c *dynamicResourceClient) Namespace(ns string) dynamic.ResourceInterface {
	ret := *c
	ret.namespace = ns
	return &ret
}

func (c *dynamicResourceClient) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	result := c.client.client.Get().AbsPath(c.makeURLSegments("")...).SpecificallyVersionedParams(&opts, dynamicParameterCodec, versionV1).Do(ctx)
	if err := result.Error(); err != nil {
		return nil, err
	}
	retBytes, err := result.Raw()
	if err != nil {
		return nil, err
	}
	uncastObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, retBytes)
	if err != nil {
		return nil, err
	}
	if list, ok := uncastObj.(*unstructured.UnstructuredList); ok {
		return list, nil
	}

	list, err := uncastObj.(*unstructured.Unstructured).ToList()
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (c *dynamicResourceClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, errors.New("not implemented")
}

func (c *dynamicResourceClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, errors.New("not implemented")
}

func (c *dynamicResourceClient) makeURLSegments(name string) []string {
	url := []string{}
	if len(c.resource.Group) == 0 {
		url = append(url, "api")
	} else {
		url = append(url, "apis", c.resource.Group)
	}
	url = append(url, c.resource.Version)

	if len(c.namespace) > 0 {
		url = append(url, "namespaces", c.namespace)
	}
	url = append(url, c.resource.Resource)

	if len(name) > 0 {
		url = append(url, name)
	}

	return url
}
