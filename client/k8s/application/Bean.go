package application

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type K8sApiResource struct {
	Gvk        schema.GroupVersionKind `json:"gvk"`
	Namespaced bool                    `json:"namespaced"`
}

type CreateResourcesRequest struct {
	Manifest  string `json:"manifest"`
	ClusterId int    `json:"clusterId"`
}

type CreateResourcesResponse struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Error    string `json:"error"`
	isUpdate bool   `json:"isUpdate"`
}
