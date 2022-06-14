package k8s

import (
	metav1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ClusterCapacityDetail struct {
	Id                int                                   `json:"id,omitempty"`
	Name              string                                `json:"name,omitempty"`
	ErrorInConnection string                                `json:"errorInNodeListing,omitempty"`
	NodeCount         int                                   `json:"nodeCount,omitempty"`
	NodeErrors        map[metav1.NodeConditionType][]string `json:"nodeErrors"`
	NodeK8sVersions   []string                              `json:"nodeK8sVersions"`
	ServerVersion     string                                `json:"serverVersion,omitempty"`
	Cpu               *ResourceDetailObject                 `json:"cpu"`
	Memory            *ResourceDetailObject                 `json:"memory"`
}

type NodeCapacityDetail struct {
	Name          string                              `json:"name"`
	Version       string                              `json:"version,omitempty"`
	Kind          string                              `json:"kind,omitempty"`
	Roles         []string                            `json:"roles"`
	K8sVersion    string                              `json:"k8sVersion"`
	Cpu           *ResourceDetailObject               `json:"cpu,omitempty"`
	Memory        *ResourceDetailObject               `json:"memory,omitempty"`
	Age           string                              `json:"age,omitempty"`
	Status        string                              `json:"status,omitempty"`
	PodCount      int                                 `json:"podCount,omitempty"`
	TaintCount    int                                 `json:"taintCount,omitempty"`
	Errors        map[metav1.NodeConditionType]string `json:"errors"`
	InternalIp    string                              `json:"internalIp"`
	ExternalIp    string                              `json:"externalIp"`
	Unschedulable bool                                `json:"unschedulable"`
	CreatedAt     string                              `json:"createdAt"`
	Labels        []*LabelAnnotationTaintObject       `json:"labels,omitempty"`
	Annotations   []*LabelAnnotationTaintObject       `json:"annotations,omitempty"`
	Taints        []*LabelAnnotationTaintObject       `json:"taints,omitempty"`
	Conditions    []*NodeConditionObject              `json:"conditions,omitempty"`
	Resources     []*ResourceDetailObject             `json:"resources,omitempty"`
	Pods          []*PodCapacityDetail                `json:"pods,omitempty"`
	Manifest      unstructured.Unstructured           `json:"manifest,omitempty"`
	ClusterName   string                              `json:"clusterName,omitempty"`
}

type PodCapacityDetail struct {
	Name      string                `json:"name"`
	Namespace string                `json:"namespace"`
	Cpu       *ResourceDetailObject `json:"cpu"`
	Memory    *ResourceDetailObject `json:"memory"`
	Age       string                `json:"age"`
	CreatedAt string                `json:"createdAt"`
}

type ResourceDetailObject struct {
	ResourceName      string `json:"name,omitempty"`
	Capacity          string `json:"capacity,omitempty"`
	Allocatable       string `json:"allocatable,omitempty"`
	Usage             string `json:"usage,omitempty"`
	Request           string `json:"request,omitempty"`
	Limit             string `json:"limit,omitempty"`
	UsagePercentage   string `json:"usagePercentage,omitempty"`
	RequestPercentage string `json:"requestPercentage,omitempty"`
	LimitPercentage   string `json:"limitPercentage,omitempty"`
	//below fields to be used at FE for sorting
	CapacityInBytes    int64 `json:"capacityInBytes,omitempty"`
	AllocatableInBytes int64 `json:"allocatableInBytes,omitempty"`
	UsageInBytes       int64 `json:"usageInBytes,omitempty"`
	RequestInBytes     int64 `json:"requestInBytes,omitempty"`
	LimitInBytes       int64 `json:"limitInBytes,omitempty"`
}

type LabelAnnotationTaintObject struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect,omitempty"`
}

type NodeConditionObject struct {
	Type      string `json:"type"`
	HaveIssue bool   `json:"haveIssue"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
}

type NodeManifestUpdateDto struct {
	ClusterId     int    `json:"clusterId"`
	Name          string `json:"name"`
	ManifestPatch string `json:"manifestPatch"`
	Version       string `json:"version"`
	Kind          string `json:"kind"`
}
