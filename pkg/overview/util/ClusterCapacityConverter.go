/*
 * Copyright (c) 2024. Devtron Inc.
 */

package util

import (
	"fmt"

	clusterBean "github.com/devtron-labs/devtron/pkg/cluster/bean"
	capacityBean "github.com/devtron-labs/devtron/pkg/k8s/capacity/bean"
	overviewBean "github.com/devtron-labs/devtron/pkg/overview/bean"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
)

// ConvertClusterOverviewToCapacityDetails converts ClusterOverviewResponse to ClusterCapacityDetail list
// This is used to serve resource browser API from cluster overview cache
func ConvertClusterOverviewToCapacityDetails(
	logger *zap.SugaredLogger,
	overviewResponse *overviewBean.ClusterOverviewResponse,
	allClusters []clusterBean.ClusterBean,
) []*capacityBean.ClusterCapacityDetail {
	if overviewResponse == nil {
		logger.Warn("overview response is nil, cannot convert to capacity details")
		return nil
	}

	capacityDetails := make([]*capacityBean.ClusterCapacityDetail, 0, len(overviewResponse.ClusterCapacityDistribution))

	// Create a map for quick lookup of cluster beans
	clusterMap := make(map[int]clusterBean.ClusterBean)
	for _, cluster := range allClusters {
		clusterMap[cluster.Id] = cluster
	}

	// Create a map for quick lookup of capacity distribution
	capacityDistMap := make(map[int]overviewBean.ClusterCapacityDistribution)
	for _, capacityDist := range overviewResponse.ClusterCapacityDistribution {
		capacityDistMap[capacityDist.ClusterID] = capacityDist
	}

	// Create a map for node errors by cluster
	nodeErrorsByCluster := make(map[int]map[corev1.NodeConditionType][]string)
	for _, nodeError := range overviewResponse.NodeErrorBreakdown.NodeErrors {
		if _, exists := nodeErrorsByCluster[nodeError.ClusterID]; !exists {
			nodeErrorsByCluster[nodeError.ClusterID] = make(map[corev1.NodeConditionType][]string)
		}
		// Convert error strings to NodeConditionType
		for _, errorStr := range nodeError.Errors {
			conditionType := corev1.NodeConditionType(errorStr)
			nodeErrorsByCluster[nodeError.ClusterID][conditionType] = append(
				nodeErrorsByCluster[nodeError.ClusterID][conditionType],
				nodeError.NodeName,
			)
		}
	}

	// Create a map for node count by cluster
	nodeCountByCluster := make(map[int]int)
	for _, nodeCount := range overviewResponse.NodeDistribution.ByClusters {
		nodeCountByCluster[nodeCount.ClusterID] = nodeCount.NodeCount
	}

	// Build capacity details for each cluster
	for _, cluster := range allClusters {
		capacityDist, hasCapacity := capacityDistMap[cluster.Id]

		var detail *capacityBean.ClusterCapacityDetail
		if hasCapacity {
			// Cluster has capacity data (connected cluster)
			detail = buildCapacityDetailFromOverview(
				cluster,
				capacityDist,
				nodeErrorsByCluster[cluster.Id],
				nodeCountByCluster[cluster.Id],
			)
		} else {
			// Connection failed cluster
			detail = &capacityBean.ClusterCapacityDetail{
				Id:                cluster.Id,
				Name:              cluster.ClusterName,
				ErrorInConnection: cluster.ErrorInConnecting,
				Status:            capacityBean.ClusterStatusConnectionFailed,
				IsVirtualCluster:  cluster.IsVirtualCluster,
				IsProd:            cluster.IsProd,
			}
		}

		capacityDetails = append(capacityDetails, detail)
	}

	logger.Debugw("converted cluster overview to capacity details",
		"totalClusters", len(capacityDetails),
		"connectedClusters", len(overviewResponse.ClusterCapacityDistribution))

	return capacityDetails
}

// buildCapacityDetailFromOverview builds a single ClusterCapacityDetail from overview data
func buildCapacityDetailFromOverview(
	cluster clusterBean.ClusterBean,
	capacityDist overviewBean.ClusterCapacityDistribution,
	nodeErrors map[corev1.NodeConditionType][]string,
	nodeCount int,
) *capacityBean.ClusterCapacityDetail {
	// Determine cluster status based on node errors
	status := capacityBean.ClusterStatusHealthy
	if len(nodeErrors) > 0 {
		status = capacityBean.ClusterStatusUnHealthy
	}

	// Build CPU and Memory resource objects
	cpuResource := &capacityBean.ResourceDetailObject{
		Capacity: fmt.Sprintf("%.2f", capacityDist.CPU.Capacity),
	}

	// Add utilization, requests, and limits percentages if available
	if capacityDist.CPU.UtilizationPercent > 0 {
		cpuResource.UsagePercentage = fmt.Sprintf("%.2f", capacityDist.CPU.UtilizationPercent)
	}
	if capacityDist.CPU.RequestsPercent > 0 {
		cpuResource.RequestPercentage = fmt.Sprintf("%.2f", capacityDist.CPU.RequestsPercent)
	}
	if capacityDist.CPU.LimitsPercent > 0 {
		cpuResource.LimitPercentage = fmt.Sprintf("%.2f", capacityDist.CPU.LimitsPercent)
	}

	memoryResource := &capacityBean.ResourceDetailObject{
		Capacity: fmt.Sprintf("%.2fGi", capacityDist.Memory.Capacity),
	}

	// Add utilization, requests, and limits percentages if available
	if capacityDist.Memory.UtilizationPercent > 0 {
		memoryResource.UsagePercentage = fmt.Sprintf("%.2f", capacityDist.Memory.UtilizationPercent)
	}
	if capacityDist.Memory.RequestsPercent > 0 {
		memoryResource.RequestPercentage = fmt.Sprintf("%.2f", capacityDist.Memory.RequestsPercent)
	}
	if capacityDist.Memory.LimitsPercent > 0 {
		memoryResource.LimitPercentage = fmt.Sprintf("%.2f", capacityDist.Memory.LimitsPercent)
	}

	return &capacityBean.ClusterCapacityDetail{
		Id:               cluster.Id,
		Name:             cluster.ClusterName,
		NodeCount:        nodeCount,
		NodeDetails:      []capacityBean.NodeDetails{}, // Not available in overview cache
		NodeErrors:       nodeErrors,
		NodeK8sVersions:  []string{}, // Not available in overview cache
		ServerVersion:    "",         // Not available in overview cache
		Cpu:              cpuResource,
		Memory:           memoryResource,
		Status:           status,
		IsVirtualCluster: cluster.IsVirtualCluster,
		IsProd:           cluster.IsProd,
	}
}
