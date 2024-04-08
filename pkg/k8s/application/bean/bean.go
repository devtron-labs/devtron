package bean

import (
	"github.com/devtron-labs/common-lib/utils/k8s"
)

const (
	DEFAULT_NAMESPACE = "default"
	EVENT_K8S_KIND    = "Event"
	LIST_VERB         = "list"
	Delete            = "delete"
)

const (
	// App Type Identifiers
	DevtronAppType = 0 // Identifier for Devtron Apps
	HelmAppType    = 1 // Identifier for Helm Apps
	ArgoAppType    = 2
	// Deployment Type Identifiers
	HelmInstalledType = 0 // Identifier for Helm deployment
	ArgoInstalledType = 1 // Identifier for ArgoCD deployment
)

const (
	LastEventID                         = "Last-Event-ID"
	TimestampOffsetToAvoidDuplicateLogs = 1
	IntegerBase                         = 10
	IntegerBitSize                      = 64
)

const (
	LocalTimezoneInGMT = "GMT+0530"
	LocalTimeOffset    = 5*60*60 + 30*60
)

type ResourceInfo struct {
	PodName string `json:"podName"`
}

type DevtronAppIdentifier struct {
	ClusterId int `json:"clusterId"`
	AppId     int `json:"appId"`
	EnvId     int `json:"envId"`
}

type Response struct {
	Kind     string   `json:"kind"`
	Name     string   `json:"name"`
	PointsTo string   `json:"pointsTo"`
	Urls     []string `json:"urls"`
}

type RotatePodResourceResponse struct {
	k8s.ResourceIdentifier
	ErrorResponse string `json:"errorResponse"`
}

//const Job = "Job"
//const Deployment = "Deployment"
//const StatefulSet = "StatefulSet"
//const DaemonSet = "DaemonSet"
//const ReplicaSet = "ReplicaSet"
//const Rollout = "Rollout"
//const ReplicationController = "ReplicationController"
//const Pod = "Pod"
//const CronJob = "CronJob"
//const Containers = "containers"
//const InitContainers = "initContainers"
//const EphemeralContainers = "ephemeralContainers"
//const IMAGE = "image"

//var commonContainerPath = []string{"spec", "template", "spec"}
//var cronJobContainerPath = []string{"spec", "jobTemplate", "spec", "spec"}
//var podContainerPath = []string{"spec"}
//
//var KindToPath = map[string][]string{
//	Deployment:            commonContainerPath,
//	Job:                   commonContainerPath,
//	StatefulSet:           commonContainerPath,
//	DaemonSet:             commonContainerPath,
//	ReplicaSet:            commonContainerPath,
//	Rollout:               commonContainerPath,
//	ReplicationController: commonContainerPath,
//	Pod:                   podContainerPath,
//	CronJob:               cronJobContainerPath,
//}
