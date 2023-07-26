package util

import (
	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ClusterResourceListMap struct {
	Headers       []string                 `json:"headers"`
	Data          []map[string]interface{} `json:"data"`
	ServerVersion string                   `json:"serverVersion"`
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
