package application

import (
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
