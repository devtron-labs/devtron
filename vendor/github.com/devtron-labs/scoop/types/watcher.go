package types

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"
)

type InterceptedEvent struct {
	Action          EventType              `json:"action"`
	InvolvedObjects map[string]interface{} `json:"involvedObject"`
	ObjectMeta      InvolvedObjectMetadata `json:"metadata"`

	InterceptedAt time.Time  `json:"interceptedAt"`
	Watchers      []*Watcher `json:"watchers"`

	ClusterId int    `json:"clusterId"`
	Namespace string `json:"namespace"`
}

type InvolvedObjectMetadata struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func GetObjectMetadata(resource *unstructured.Unstructured) InvolvedObjectMetadata {
	gvk := resource.GroupVersionKind()
	name := resource.GetName()
	ns := resource.GetNamespace()
	return InvolvedObjectMetadata{
		Group:     gvk.Group,
		Version:   gvk.Version,
		Kind:      gvk.Kind,
		Name:      name,
		Namespace: ns,
	}
}

type Watcher struct {
	Id                    int                       `json:"id"`
	Name                  string                    `json:"name"`
	Namespaces            map[string]bool           `json:"namespaces"`
	GVKs                  []schema.GroupVersionKind `json:"groupVersionKinds"`
	EventFilterExpression string                    `json:"eventFilterExpression"`
	ClusterId             int                       `json:"clusterId"`
}

type Action string

const (
	DELETE Action = "DELETE"
	ADD    Action = "ADD"
	UPDATE Action = "UPDATE"
)

type Payload struct {
	Action  Action
	Watcher *Watcher
}

const WATCHER_CUD_URL = "/k8s/watcher"
const API_RESOURCES_URL = "/k8s/api-resources"
const RESOURCE_LIST_URL = "/k8s/resources"
const K8S_CACHE_CONFIG_URL = "/k8s/cache/config"

type EventType string

const (
	DELETED EventType = "DELETED"
	CREATED EventType = "CREATED"
	UPDATED EventType = "UPDATED"
)
