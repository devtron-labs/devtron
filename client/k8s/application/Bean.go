package application

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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

type ClusterResourceListResponse struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	Status     string   `json:"status"`
	Age        string   `json:"age"`
	Ready      string   `json:"ready,omitempty"`
	Restarts   string   `json:"restarts,omitempty"`
	Url        string   `json:"url,omitempty"`
	Containers []string `json:"containers,omitempty"`
}

type ClusterResourceListMap struct {
	Headers []string                 `json:"headers"`
	Data    []map[string]interface{} `json:"data"`
}

const K8sClusterResourceNameKey = "name"
const K8sClusterResourcePriorityKey = "priority"
const K8sClusterResourceNamespaceKey = "namespace"
const K8sClusterResourceMetadataKey = "metadata"
const K8sClusterResourceCreationTimestampKey = "creationTimestamp"

const K8sClusterResourceObjectKey = "object"
const K8sClusterResourceRowsKey = "rows"
const K8sClusterResourceCellKey = "cells"
const K8sClusterResourceColumnDefinitionKey = "columnDefinitions"
