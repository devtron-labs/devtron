package application

import "k8s.io/apimachinery/pkg/runtime/schema"

type K8sApiResource struct {
	Gvk        schema.GroupVersionKind `json:"gvk"`
	Namespaced bool                    `json:"namespaced"`
}
