package k8s

import (
	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ClusterResourceListMap struct {
	Headers []string                 `json:"headers"`
	Data    []map[string]interface{} `json:"data"`
}

type EventsResponse struct {
	Events *v1.EventList `json:"events,omitempty"`
}

type ResourceListResponse struct {
	Resources unstructured.UnstructuredList `json:"resources,omitempty"`
}

const K8sClusterResourceNameKey = "name"
const K8sClusterResourcePriorityKey = "priority"
const K8sClusterResourceNamespaceKey = "namespace"
const K8sClusterResourceMetadataKey = "metadata"
const K8sClusterResourceMetadataNameKey = "name"
const K8sClusterResourceOwnerReferenceKey = "ownerReferences"
const K8sClusterResourceCreationTimestampKey = "creationTimestamp"

const K8sClusterResourceRowsKey = "rows"
const K8sClusterResourceCellKey = "cells"
const K8sClusterResourceColumnDefinitionKey = "columnDefinitions"
const K8sClusterResourceObjectKey = "object"

const K8sClusterResourceKindKey = "kind"
const K8sClusterResourceApiVersionKey = "apiVersion"

const K8sClusterResourceRolloutKind = "Rollout"
const K8sClusterResourceRolloutGroup = "argoproj.io"
const K8sClusterResourceReplicationControllerKind = "ReplicationController"
const K8sClusterResourceCronJobKind = "CronJob"
const V1VERSION = "v1"
const BatchGroup = "batch"
const AppsGroup = "apps"
const RestartingNotSupported = "restarting not supported"

var KindVsChildrenGvk = map[string][]schema.GroupVersionKind{
	kube.DeploymentKind:                         append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Group: AppsGroup, Version: V1VERSION, Kind: kube.ReplicaSetKind}, schema.GroupVersionKind{Version: V1VERSION, Kind: kube.PodKind}),
	K8sClusterResourceRolloutKind:               append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Group: AppsGroup, Version: V1VERSION, Kind: kube.ReplicaSetKind}, schema.GroupVersionKind{Version: V1VERSION, Kind: kube.PodKind}),
	K8sClusterResourceCronJobKind:               append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Group: BatchGroup, Version: V1VERSION, Kind: kube.JobKind}, schema.GroupVersionKind{Version: V1VERSION, Kind: kube.PodKind}),
	kube.JobKind:                                append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Version: V1VERSION, Kind: kube.PodKind}),
	kube.ReplicaSetKind:                         append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Version: V1VERSION, Kind: kube.PodKind}),
	kube.DaemonSetKind:                          append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Version: V1VERSION, Kind: kube.PodKind}),
	kube.StatefulSetKind:                        append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Version: V1VERSION, Kind: kube.PodKind}),
	K8sClusterResourceReplicationControllerKind: append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Version: V1VERSION, Kind: kube.PodKind}),
}

const (
	DefaultCluster           = "default_cluster"
	BearerToken              = "bearer_token"
	CertificateAuthorityData = "cert_auth_data"
	CertData                 = "cert_data"
	TlsKey                   = "tls_key"
	LiveZ                    = "/livez"
)

const (
	// EvictionKind represents the kind of evictions object
	EvictionKind = "Eviction"
	// EvictionSubresource represents the kind of evictions object as pod's subresource
	EvictionSubresource = "pods/eviction"
)

type PodLogsRequest struct {
	SinceTime                  *v12.Time `json:"sinceTime,omitempty"`
	TailLines                  int       `json:"tailLines"`
	Follow                     bool      `json:"follow"`
	ContainerName              string    `json:"containerName"`
	IsPrevContainerLogsEnabled bool      `json:"previous"`
}

type ResourceIdentifier struct {
	Name             string                  `json:"name"` //pod name for logs request
	Namespace        string                  `json:"namespace"`
	GroupVersionKind schema.GroupVersionKind `json:"groupVersionKind"`
}

type K8sRequestBean struct {
	ResourceIdentifier ResourceIdentifier `json:"resourceIdentifier"`
	Patch              string             `json:"patch,omitempty"`
	PodLogsRequest     PodLogsRequest     `json:"podLogsRequest,omitempty"`
	ForceDelete        bool               `json:"-"`
}

type GetAllApiResourcesResponse struct {
	ApiResources []*K8sApiResource `json:"apiResources"`
	AllowedAll   bool              `json:"allowedAll"`
}

type K8sApiResource struct {
	Gvk        schema.GroupVersionKind `json:"gvk"`
	Namespaced bool                    `json:"namespaced"`
}

type ApplyResourcesRequest struct {
	Manifest  string `json:"manifest"`
	ClusterId int    `json:"clusterId"`
}

type ApplyResourcesResponse struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Error    string `json:"error"`
	IsUpdate bool   `json:"isUpdate"`
}

type ManifestResponse struct {
	Manifest unstructured.Unstructured `json:"manifest,omitempty"`
}
