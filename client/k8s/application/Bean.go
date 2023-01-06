package application

import (
	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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

type ClusterResourceListMap struct {
	Headers []string                 `json:"headers"`
	Data    []map[string]interface{} `json:"data"`
}

const K8sClusterResourceNameKey = "name"
const K8sClusterResourcePriorityKey = "priority"
const K8sClusterResourceNamespaceKey = "namespace"
const K8sClusterResourceOwnerReferenceKey = "ownerReferences"
const K8sClusterResourceMetadataKey = "metadata"
const K8sClusterResourceMetadataNameKey = "name"
const K8sClusterResourceCreationTimestampKey = "creationTimestamp"

const K8sClusterResourceObjectKey = "object"
const K8sClusterResourceRowsKey = "rows"
const K8sClusterResourceCellKey = "cells"
const K8sClusterResourceColumnDefinitionKey = "columnDefinitions"

const K8sClusterResourceKindKey = "kind"
const K8sClusterResourceApiVersionKey = "apiVersion"

const K8sClusterResourceRolloutKind = "Rollout"
const K8sClusterResourceRolloutGroup = "argoproj.io"
const K8sClusterResourceReplicationControllerKind = "ReplicationController"
const K8sClusterResourceCronJobKind = "CronJob"
const V1VERSION = "v1"
const BatchGroup = "batch"
const AppsGroup = "apps"

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
