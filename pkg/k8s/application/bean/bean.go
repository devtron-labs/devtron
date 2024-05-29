/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

import (
	"github.com/devtron-labs/common-lib/utils/k8s"
)

const (
	Authorization     = "Authorization"
	BaseForK8sProxy   = "/orchestrator/k8s/proxy"
	Cluster           = "cluster"
	Empty             = ""
	Env               = "env"
	ClusterIdentifier = "clusterIdentifier"
	EnvIdentifier     = "envIdentifier"
	RoleView          = "View"
	RoleAdmin         = "Admin"
	API               = "api"
	APIs              = "apis"
	K8sEmpty          = "k8sempty"
	V1                = "v1"
	ALL               = "*"
	NAMESPACES        = "namespaces"
	NODES             = "nodes"
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

type PortForwardRequest struct {
	ClusterId   int
	Namespace   string
	ServiceName string
	TargetPort  string // []string{"5432:5432"}
}

type K8sProxyRequest struct {
	ClusterId   int
	ClusterName string
	EnvId       int
	EnvName     string
}

type InterClusterCommunicationConfig struct {
	ProxyUpTime int64 `env:"PROXY_UP_TIME" envDefault:"60"`
}
