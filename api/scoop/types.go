package scoop

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"
)

type InterceptedEvent struct {
	Message        string                  `json:"message"`
	MessageType    string                  `json:"type"`
	Event          map[string]interface{}  `json:"event"` // raw k8s event
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
