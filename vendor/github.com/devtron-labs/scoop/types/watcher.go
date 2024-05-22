package types

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"
)

type EnvConfig struct {
	Token           string `env:"TOKEN"`
	ClusterId       int    `env:"CLUSTER_ID" envDefault:"1"`
	OrchestratorUrl string `env:"ORCHESTRATOR_URL" envDefault:"http://localhost:8080"`
}

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
	Id   int    `json:"id"`
	Name string `json:"name"`
	// Namespaces            map[string]bool           `json:"namespaces"`
	SelectedActions       []EventType               `json:"selectedActions"`
	GVKs                  []schema.GroupVersionKind `json:"groupVersionKinds"`
	EventFilterExpression string                    `json:"eventFilterExpression"`
	ClusterId             int                       `json:"clusterId"`
	JobConfigured         bool                      `json:"jobConfigured"`
	Selectors             NamespaceSelector         `json:"namespaceSelector"`
}

func (watcher *Watcher) IntrestedInAction(action EventType) bool {

	for _, selectedAction := range watcher.SelectedActions {
		if selectedAction == action {
			return true
		}
	}

	return false
}

func (watcher *Watcher) IntrestedInGVK(gvk schema.GroupVersionKind) bool {

	for _, interestedGVK := range watcher.GVKs {
		if interestedGVK.Group == gvk.Group && interestedGVK.Kind == gvk.Kind && interestedGVK.Version == gvk.Version {
			return true
		}
	}

	return false
}

func (watcher *Watcher) IntrestedInNamespace(ns string, isProd bool) bool {

	if watcher.Selectors.InterestGroup == All {
		return true
	}

	if watcher.Selectors.InterestGroup == AllProd || watcher.Selectors.InterestGroup == AllNonProd {
		return isProd
	}

	existsInList := false
	for _, namespace := range watcher.Selectors.Namespaces {
		if namespace == ns {
			existsInList = true
			break
		}
	}

	// should present in included list
	if watcher.Selectors.InterestGroup == Included {
		return existsInList
	}

	// should not present in excluded group
	return !existsInList
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
	InterestGroup InterestCriteria `json:"subGroup"`
	Namespaces    []string         `json:"names"`
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
