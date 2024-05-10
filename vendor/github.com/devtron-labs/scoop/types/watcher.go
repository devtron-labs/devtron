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
	SelectedActions       []EventType               `json:"selectedActions"`
	GVKs                  []schema.GroupVersionKind `json:"groupVersionKinds"`
	EventFilterExpression string                    `json:"eventFilterExpression"`
	ClusterId             int                       `json:"clusterId"`
	JobConfigured         bool                      `json:"jobConfigured"`
	Selectors             NamespaceSelector         `json:"namespaceSelector"`
}

type InterestCriteria string

const (
	// Included is to only include selected namespace
	Included InterestCriteria = "INCLUDED"

	// Excluded is to exclude selected namespace
	Excluded InterestCriteria = "EXCLUDED"

	// AllProd is to only show interest in all prod namespaces
	AllProd InterestCriteria = "ALL_PROD"

	// AllNonProd is to only show interest in all non-prod namespaces
	AllNonProd InterestCriteria = "ALL_NON_PROD"

	// All is to show interest in any namespace
	All InterestCriteria = "ALL"
)

type NamespaceSelector struct {
	InerestGroup InterestCriteria `json:"subGroup"`
	Namespaces   []string         `json:"names"`
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
const NAMESPACE_CUD_URL = "/k8s/namespace"

type EventType string

const (
	DELETED EventType = "DELETED"
	CREATED EventType = "CREATED"
	UPDATED EventType = "UPDATED"
)

const (
	NamespaceKey   = "namespace"
	IsProdKey      = "isProd"
	ActionKey      = "action"
	NewResourceKey = "newResource"
	OldResourceKey = "oldResource"
)
