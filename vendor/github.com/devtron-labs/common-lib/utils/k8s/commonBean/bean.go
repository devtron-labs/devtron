/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commonBean

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	SecretKind                   = "Secret"
	ServiceKind                  = "Service"
	ServiceAccountKind           = "ServiceAccount"
	EndpointsKind                = "Endpoints"
	EndPointsSlice               = "EndpointSlice"
	DeploymentKind               = "Deployment"
	ReplicaSetKind               = "ReplicaSet"
	StatefulSetKind              = "StatefulSet"
	DaemonSetKind                = "DaemonSet"
	IngressKind                  = "Ingress"
	JobKind                      = "Job"
	PersistentVolumeClaimKind    = "PersistentVolumeClaim"
	CustomResourceDefinitionKind = "CustomResourceDefinition"
	PodKind                      = "Pod"
	APIServiceKind               = "APIService"
	NamespaceKind                = "Namespace"
	HorizontalPodAutoscalerKind  = "HorizontalPodAutoscaler"
	Spec                         = "spec"
	Template                     = "template"
	JobTemplate                  = "jobTemplate"
	Ports                        = "ports"
	Port                         = "port"
	Subsets                      = "subsets"
	Nodes                        = "nodes"
	Containers                   = "containers"
	InitContainers               = "initContainers"
	EphemeralContainers          = "ephemeralContainers"
	Image                        = "image"
)

var defaultContainerPath = []string{Spec, Template, Spec}
var cronJobContainerPath = []string{Spec, JobTemplate, Spec, Template, Spec}
var podContainerPath = []string{Spec}

var kindToPath = map[string][]string{
	PodKind:                       podContainerPath,
	K8sClusterResourceCronJobKind: cronJobContainerPath,
}

func GetContainerSubPathForKind(kind string) []string {
	if path, ok := kindToPath[kind]; ok {
		return path
	}
	return defaultContainerPath
}

const (
	PersistentVolumeClaimsResourceType = "persistentvolumeclaims"
	StatefulSetsResourceType           = "statefulsets"
)

const (
	ContainersType                = "Containers"
	ContainersNamesType           = "ContainerNames"
	InitContainersNamesType       = "InitContainerNames"
	EphemeralContainersInfoType   = "EphemeralContainerInfo"
	EphemeralContainersStatusType = "EphemeralContainerStatuses"
	StatusReason                  = "Status Reason"
	Node                          = "Node"
	RestartCount                  = "Restart Count"
)

const (
	Group   = "group"
	Version = "version"
	Kind    = "kind"
)

type HookType string

const (
	HookTypePreSync  HookType = "PreSync"
	HookTypeSync     HookType = "Sync"
	HookTypePostSync HookType = "PostSync"
	HookTypeSkip     HookType = "Skip"
	HookTypeSyncFail HookType = "SyncFail"
)

type OperationPhase string

const (
	OperationRunning     OperationPhase = "Running"
	OperationTerminating OperationPhase = "Terminating"
	OperationFailed      OperationPhase = "Failed"
	OperationError       OperationPhase = "Error"
	OperationSucceeded   OperationPhase = "Succeeded"
)

const K8sClusterResourceNameKey = "name"
const K8sClusterResourcePriorityKey = "priority"
const K8sClusterResourceNamespaceKey = "namespace"
const K8sClusterResourceMetadataKey = "metadata"
const K8sClusterResourceSpecKey = "spec"
const K8sClusterResourceMetadataNameKey = "name"
const K8sClusterResourceOwnerReferenceKey = "ownerReferences"
const K8sClusterResourceCreationTimestampKey = "creationTimestamp"

const K8sClusterResourceRowsKey = "rows"
const K8sClusterResourceCellKey = "cells"
const K8sClusterResourceColumnDefinitionKey = "columnDefinitions"
const K8sClusterResourceObjectKey = "object"

const K8sClusterResourceGroupKey = "group"
const K8sClusterResourceKindKey = "kind"
const K8sClusterResourceVersionKey = "version"
const K8sClusterResourceApiVersionKey = "apiVersion"

const K8sClusterResourceReplicationControllerKind = "ReplicationController"
const K8sClusterResourceRolloutKind = "Rollout"
const K8sClusterResourceRolloutGroup = "argoproj.io"
const K8sClusterResourceCronJobKind = "CronJob"

const V1VERSION = "v1"
const BatchGroup = "batch"
const AppsGroup = "apps"

const HelmHookAnnotation = "helm.sh/hook"
const HibernateReplicaAnnotation = "hibernator.devtron.ai/replicas"

const (
	K8sResourceColumnDefinitionName         = "Name"
	K8sResourceColumnDefinitionSyncStatus   = "Sync Status"
	K8sResourceColumnDefinitionHealthStatus = "Health Status"
	K8sClusterResourceStatusKey             = "status"
	K8sClusterResourceHealthKey             = "health"
	K8sClusterResourceResourcesKey          = "resources"
	K8sClusterResourceSyncKey               = "sync"
)

var KindVsChildrenGvk = map[string][]schema.GroupVersionKind{
	DeploymentKind:                append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Group: AppsGroup, Version: V1VERSION, Kind: ReplicaSetKind}, schema.GroupVersionKind{Version: V1VERSION, Kind: PodKind}),
	K8sClusterResourceRolloutKind: append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Group: AppsGroup, Version: V1VERSION, Kind: ReplicaSetKind}, schema.GroupVersionKind{Version: V1VERSION, Kind: PodKind}),
	K8sClusterResourceCronJobKind: append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Group: BatchGroup, Version: V1VERSION, Kind: JobKind}, schema.GroupVersionKind{Version: V1VERSION, Kind: PodKind}),
	JobKind:                       append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Version: V1VERSION, Kind: PodKind}),
	ReplicaSetKind:                append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Version: V1VERSION, Kind: PodKind}),
	DaemonSetKind:                 append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Version: V1VERSION, Kind: PodKind}),
	StatefulSetKind:               append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Version: V1VERSION, Kind: PodKind}),
	K8sClusterResourceReplicationControllerKind: append(make([]schema.GroupVersionKind, 0), schema.GroupVersionKind{Version: V1VERSION, Kind: PodKind}),
}

const (
	// EvictionKind represents the kind of evictions object
	EvictionKind = "Eviction"
	// EvictionSubresource represents the kind of evictions object as pod's subresource
	EvictionSubresource = "pods/eviction"
)

// constants starts
var podsGvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}, Scope: meta.RESTScopeNameNamespace}
var replicaSetGvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicasets"}, Scope: meta.RESTScopeNameNamespace}
var jobGvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}, Scope: meta.RESTScopeNameNamespace}
var endpointsGvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "endpoints"}, Scope: meta.RESTScopeNameNamespace}
var endpointSliceV1Beta1GvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "discovery.k8s.io", Version: "v1beta1", Resource: "endpointslices"}, Scope: meta.RESTScopeNameNamespace}
var endpointSliceV1GvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "discovery.k8s.io", Version: "v1", Resource: "endpointslices"}, Scope: meta.RESTScopeNameNamespace}
var pvGvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumes"}, Scope: meta.RESTScopeNameRoot}
var pvcGvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumeclaims"}, Scope: meta.RESTScopeNameNamespace}
var stsGvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}, Scope: meta.RESTScopeNameNamespace}
var configGvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}, Scope: meta.RESTScopeNameNamespace}
var hpaGvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers"}, Scope: meta.RESTScopeNameNamespace}
var deployGvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, Scope: meta.RESTScopeNameNamespace}
var serviceGvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}, Scope: meta.RESTScopeNameNamespace}
var daemonGvrAndScope = &GvrAndScope{Gvr: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}, Scope: meta.RESTScopeNameNamespace}

var gvkVsChildGvrAndScope = map[schema.GroupVersionKind][]*GvrAndScope{
	schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}:                     append(make([]*GvrAndScope, 0), replicaSetGvrAndScope),
	schema.GroupVersionKind{Group: "argoproj.io", Version: "v1alpha1", Kind: "Rollout"}:           append(make([]*GvrAndScope, 0), replicaSetGvrAndScope),
	schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"}:                     append(make([]*GvrAndScope, 0), podsGvrAndScope),
	schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "CronJob"}:                       append(make([]*GvrAndScope, 0), jobGvrAndScope),
	schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"}:                           append(make([]*GvrAndScope, 0), podsGvrAndScope),
	schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}:                    append(make([]*GvrAndScope, 0), podsGvrAndScope, pvcGvrAndScope, stsGvrAndScope),
	schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"}:                      append(make([]*GvrAndScope, 0), podsGvrAndScope),
	schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}:                            append(make([]*GvrAndScope, 0), endpointsGvrAndScope, endpointSliceV1Beta1GvrAndScope, endpointSliceV1GvrAndScope),
	schema.GroupVersionKind{Group: "monitoring.coreos.com", Version: "v1", Kind: "Prometheus"}:    append(make([]*GvrAndScope, 0), stsGvrAndScope, configGvrAndScope),
	schema.GroupVersionKind{Group: "monitoring.coreos.com", Version: "v1", Kind: "Alertmanager"}:  append(make([]*GvrAndScope, 0), stsGvrAndScope, configGvrAndScope),
	schema.GroupVersionKind{Group: "keda.sh", Version: "v1alpha1", Kind: "ScaledObject"}:          append(make([]*GvrAndScope, 0), hpaGvrAndScope),
	schema.GroupVersionKind{Group: "autoscaling", Version: "v2", Kind: "HorizontalPodAutoscaler"}: append(make([]*GvrAndScope, 0), podsGvrAndScope),
	schema.GroupVersionKind{Group: "flagger.app", Version: "v1beta1", Kind: "Canary"}:             append(make([]*GvrAndScope, 0), deployGvrAndScope, serviceGvrAndScope, daemonGvrAndScope),
}

// constants end

type GvrAndScope struct {
	Gvr   schema.GroupVersionResource
	Scope meta.RESTScopeName
}

//var K8sNativeGroups = []string{"", "admissionregistration.k8s.io", "apiextensions.k8s.io", "apiregistration.k8s.io", "apps", "authentication.k8s.io", "authorization.k8s.io",
//	"autoscaling", "batch", "certificates.k8s.io", "coordination.k8s.io", "core", "discovery.k8s.io", "events.k8s.io", "flowcontrol.apiserver.k8s.io", "argoproj.io",
//	"internal.apiserver.k8s.io", "networking.k8s.io", "node.k8s.io", "policy", "rbac.authorization.k8s.io", "resource.k8s.io", "scheduling.k8s.io", "storage.k8s.io"}

func GetGvkVsChildGvrAndScope() map[schema.GroupVersionKind][]*GvrAndScope {
	return gvkVsChildGvrAndScope
}

// ResourceNode contains information about live resource and its children
type ResourceNode struct {
	*ResourceRef    `json:",inline" protobuf:"bytes,1,opt,name=resourceRef"`
	ParentRefs      []*ResourceRef          `json:"parentRefs,omitempty" protobuf:"bytes,2,opt,name=parentRefs"`
	NetworkingInfo  *ResourceNetworkingInfo `json:"networkingInfo,omitempty" protobuf:"bytes,4,opt,name=networkingInfo"`
	ResourceVersion string                  `json:"resourceVersion,omitempty" protobuf:"bytes,5,opt,name=resourceVersion"`
	Health          *HealthStatus           `json:"health,omitempty" protobuf:"bytes,7,opt,name=health"`
	IsHibernated    bool                    `json:"isHibernated"`
	CanBeHibernated bool                    `json:"canBeHibernated"`
	Info            []InfoItem              `json:"info,omitempty"`
	Port            []int64                 `json:"port,omitempty"`
	CreatedAt       string                  `json:"createdAt,omitempty"`
	IsHook          bool                    `json:"isHook,omitempty"`
	HookType        string                  `json:"hookType,omitempty"`
	// UpdateRevision is used when a pod's owner is a StatefulSet for identifying if the pod is new or old
	UpdateRevision string `json:"updateRevision,omitempty"`
	// DeploymentPodHash is the podHash in deployment manifest and is used to compare replicaSet's podHash for identifying new vs old pod
	DeploymentPodHash        string `json:"deploymentPodHash,omitempty"`
	DeploymentCollisionCount *int32 `json:"deploymentCollisionCount,omitempty"`
	// RolloutCurrentPodHash is the podHash in rollout manifest and is used to compare replicaSet's podHash for identifying new vs old pod
	RolloutCurrentPodHash string `json:"rolloutCurrentPodHash,omitempty"`
}

// ResourceRef includes fields which unique identify resource
type ResourceRef struct {
	Group     string                    `json:"group,omitempty" protobuf:"bytes,1,opt,name=group"`
	Version   string                    `json:"version,omitempty" protobuf:"bytes,2,opt,name=version"`
	Kind      string                    `json:"kind,omitempty" protobuf:"bytes,3,opt,name=kind"`
	Namespace string                    `json:"namespace,omitempty" protobuf:"bytes,4,opt,name=namespace"`
	Name      string                    `json:"name,omitempty" protobuf:"bytes,5,opt,name=name"`
	UID       string                    `json:"uid,omitempty" protobuf:"bytes,6,opt,name=uid"`
	Manifest  unstructured.Unstructured `json:"-"`
}

func (r *ResourceRef) GetGvk() schema.GroupVersionKind {
	if r == nil {
		return schema.GroupVersionKind{}
	}
	return schema.GroupVersionKind{
		Group:   r.Group,
		Version: r.Version,
		Kind:    r.Kind,
	}
}

// ResourceNetworkingInfo holds networking resource related information
type ResourceNetworkingInfo struct {
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,3,opt,name=labels"`
}

type HealthStatus struct {
	Status  HealthStatusCode `json:"status,omitempty" protobuf:"bytes,1,opt,name=status"`
	Message string           `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
}

type HealthStatusCode = string

const (
	HealthStatusUnknown             HealthStatusCode = "Unknown"
	HealthStatusProgressing         HealthStatusCode = "Progressing"
	HealthStatusHealthy             HealthStatusCode = "Healthy"
	HealthStatusSuspended           HealthStatusCode = "Suspended"
	HealthStatusDegraded            HealthStatusCode = "Degraded"
	HealthStatusMissing             HealthStatusCode = "Missing"
	HealthStatusHibernated          HealthStatusCode = "Hibernated"
	HealthStatusPartiallyHibernated HealthStatusCode = "Partially Hibernated"
)

type ResourceTreeResponse struct {
	Nodes       []*ResourceNode `json:"nodes,omitempty"`
	PodMetadata []*PodMetadata  `json:"podMetadata,omitempty"`
}

type PodMetadata struct {
	Name                string                    `json:"name"`
	UID                 string                    `json:"uid"`
	Containers          []string                  `json:"containers"`
	InitContainers      []string                  `json:"initContainers"`
	IsNew               bool                      `json:"isNew"`
	EphemeralContainers []*EphemeralContainerData `json:"ephemeralContainers"`
}

// use value field as generic type
// InfoItem contains arbitrary, human readable information about an application
type InfoItem struct {
	// Name is a human readable title for this piece of information.
	Name string `json:"name,omitempty"`
	// Value is human readable content.
	Value interface{} `json:"value,omitempty"`
}

type EphemeralContainerData struct {
	Name       string `json:"name"`
	IsExternal bool   `json:"isExternal"`
}

type EphemeralContainerStatusesInfo struct {
	Name  string
	State v1.ContainerState
}

type EphemeralContainerInfo struct {
	Name    string
	Command []string
}
type ExtraNodeInfo struct {
	// UpdateRevision is only used for StatefulSets, if not empty, indicates the version of the StatefulSet used to generate Pods in the sequence
	UpdateRevision         string
	ResourceNetworkingInfo *ResourceNetworkingInfo
	RolloutCurrentPodHash  string
}

const (
	DEFAULT_CLUSTER          = "default_cluster"
	DEVTRON_SERVICE_NAME     = "devtron-service"
	DefaultClusterUrl        = "https://kubernetes.default.svc"
	BearerToken              = "bearer_token"
	CertificateAuthorityData = "cert_auth_data"
	CertData                 = "cert_data"
	TlsKey                   = "tls_key"
	LiveZ                    = "/livez"
	Running                  = "Running"
	RestartingNotSupported   = "restarting not supported"
	DEVTRON_APP_LABEL_KEY    = "app"
	DEVTRON_APP_LABEL_VALUE1 = "devtron"
	DEVTRON_APP_LABEL_VALUE2 = "orchestrator"
)
