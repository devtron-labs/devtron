package types

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"time"
)

type InterceptedEvent struct {
	Action         watch.EventType         `json:"action"`
	InvolvedObject map[string]interface{}  `json:"involvedObject"`
	GVK            schema.GroupVersionKind `json:"gvk"`

	InterceptedAt time.Time  `json:"interceptedAt"`
	Watchers      []*Watcher `json:"watchers"`

	ClusterId int    `json:"clusterId"`
	Namespace string `json:"namespace"`
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
