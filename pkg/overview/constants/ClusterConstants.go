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

package constants

// Cloud Provider Constants
const (
	ProviderAWS          = "AWS"
	ProviderGCP          = "GCP"
	ProviderAzure        = "Azure"
	ProviderOracle       = "Oracle"
	ProviderDigitalOcean = "DigitalOcean"
	ProviderIBM          = "IBM"
	ProviderAlibaba      = "Alibaba"
	ProviderUnknown      = "Unknown"
)

// Node Condition Type Constants
// These map to Kubernetes node condition types
const (
	NodeConditionNetworkUnavailable = "NetworkUnavailable"
	NodeConditionMemoryPressure     = "MemoryPressure"
	NodeConditionDiskPressure       = "DiskPressure"
	NodeConditionPIDPressure        = "PIDPressure"
	NodeConditionReady              = "Ready"
	NodeConditionOthers             = "Others"
)

// Node Error Breakdown Keys
// These are used as keys in the NodeErrorBreakdown map
const (
	NodeErrorNetworkUnavailable = "NetworkUnavailable"
	NodeErrorMemoryPressure     = "MemoryPressure"
	NodeErrorDiskPressure       = "DiskPressure"
	NodeErrorPIDPressure        = "PIDPressure"
	NodeErrorKubeletNotReady    = "KubeletNotReady"
	NodeErrorOthers             = "Others"
)

// Version Constants
const (
	VersionUnknown = "Unknown"
)

// Autoscaler Type Constants
const (
	AutoscalerKarpenter         = "karpenter"
	AutoscalerGKE               = "gke"
	AutoscalerEKS               = "eks"
	AutoscalerAKS               = "aks"
	AutoscalerCastAI            = "castAi"
	AutoscalerClusterAutoscaler = "clusterAutoscaler"
	AutoscalerNotDetected       = "notDetected"
)

// Autoscaler Label Constants
// These labels are used to identify which autoscaler manages a node
const (
	// EKS Auto Mode label
	LabelEKSComputeType = "eks.amazonaws.com/compute-type"
	LabelEKSComputeAuto = "auto"

	// Karpenter label
	LabelKarpenterInitialized = "karpenter.sh/initialized"
	LabelKarpenterTrue        = "true"

	// Cast AI label
	LabelCastAIManagedBy = "provisioner.cast.ai/managed-by"
	LabelCastAIValue     = "cast.ai"

	// GKE label
	LabelGKEProvisioning = "cloud.google.com/gke-provisioning"
	LabelGKEAutoPilot    = "spot"
)

// Node Name Prefix Constants
const (
	NodePrefixGKE = "gke-"
	NodePrefixAKS = "aks-"
	NodePrefixEKS = "eks-"
	NodePrefixOKE = "oke-"
)

// Node Name Pattern Constants
const (
	NodePatternAWSComputeInternal = ".compute.internal"
	NodePatternAWSEC2Internal     = ".ec2.internal"
	NodePatternAzureVMSS          = "vmss"
	NodePatternAzureScaleSets     = "scalesets"
	NodePatternGCP                = "gcp"
	NodePatternGoogle             = "google"
	NodePatternDigitalOcean       = "digitalocean"
	NodePatternIBMKube            = "kube"
	NodePatternAliyun             = "aliyun"
	NodePatternAlibabaRegion      = "cn-"
)

// AWS Region Pattern Constants
var AWSRegionPatterns = []string{
	"us-east-", "us-west-", "eu-west-", "eu-central-", "ap-south-",
	"ap-southeast-", "ap-northeast-", "ca-central-", "sa-east-",
}

// Sort Field Constants for Cluster Overview Detail API
const (
	SortFieldNodeName       = "nodeName"
	SortFieldClusterName    = "clusterName"
	SortFieldNodeErrors     = "nodeErrors"
	SortFieldNodeStatus     = "nodeStatus"
	SortFieldSchedulable    = "schedulable"
	SortFieldAutoscalerType = "autoscalerType"
)

// Sort Order Constants
const (
	SortOrderAsc  = "ASC"
	SortOrderDesc = "DESC"
)

// Schedulable Type Constants for filtering
const (
	SchedulableTypeSchedulable   = "schedulable"
	SchedulableTypeUnschedulable = "unschedulable"
)
