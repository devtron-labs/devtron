/*
 * Copyright (c) 2024. Devtron Inc.
 */

package util

import (
	capacityBean "github.com/devtron-labs/devtron/pkg/k8s/capacity/bean"
	"github.com/devtron-labs/devtron/pkg/overview/constants"
)

// DetermineAutoscalerTypeFromLabels determines the autoscaler type based on node labels (map format)
// This is used by the informer layer which works with native Kubernetes node labels
// Priority order: EKS Auto Mode > Karpenter > Cast AI > GKE > Not Detected
// Note: Cluster Autoscaler (CAS) cannot be reliably detected from node labels alone as it works
// with existing node groups and doesn't add its own labels. Nodes managed by CAS will show as "Not Detected".
func DetermineAutoscalerTypeFromLabels(labels map[string]string) string {
	// Check for EKS Auto Mode: eks.amazonaws.com/compute-type=auto
	if computeType, exists := labels[constants.LabelEKSComputeType]; exists && computeType == constants.LabelEKSComputeAuto {
		return constants.AutoscalerEKS
	}

	// Check for Karpenter: karpenter.sh/initialized=true
	if initialized, exists := labels[constants.LabelKarpenterInitialized]; exists && initialized == constants.LabelKarpenterTrue {
		return constants.AutoscalerKarpenter
	}

	// Check for Cast AI: provisioner.cast.ai/managed-by=cast.ai
	if managedBy, exists := labels[constants.LabelCastAIManagedBy]; exists && managedBy == constants.LabelCastAIValue {
		return constants.AutoscalerCastAI
	}

	// Check for GKE: cloud.google.com/gke-provisioning=standard
	if provisioning, exists := labels[constants.LabelGKEProvisioning]; exists && provisioning == constants.LabelGKEAutoPilot {
		return constants.AutoscalerGKE
	}

	// If none of the known autoscaler labels are found, return Not Detected
	// This includes nodes managed by Cluster Autoscaler (CAS) as CAS doesn't add unique labels
	return constants.AutoscalerNotDetected
}

// DetermineAutoscalerTypeFromLabelArray determines the autoscaler type based on node labels (array format)
// This is used by the service layer which works with capacity service label objects
// It converts the label array to a map and delegates to DetermineAutoscalerTypeFromLabels
func DetermineAutoscalerTypeFromLabelArray(labels []*capacityBean.LabelAnnotationTaintObject) string {
	// Convert label array to map for easier lookup
	labelMap := make(map[string]string)
	for _, label := range labels {
		if label != nil {
			labelMap[label.Key] = label.Value
		}
	}

	// Delegate to the main detection function
	return DetermineAutoscalerTypeFromLabels(labelMap)
}
