/*
 * Copyright (c) 2024. Devtron Inc.
 */

package overview

import (
	"github.com/devtron-labs/devtron/pkg/overview/bean"
	"github.com/devtron-labs/devtron/pkg/overview/constants"
	"github.com/devtron-labs/devtron/pkg/overview/util"
)

// NewClusterOverviewResponse creates and initializes a new ClusterOverviewResponse with default values
func NewClusterOverviewResponse(totalClusters int) *bean.ClusterOverviewResponse {
	return &bean.ClusterOverviewResponse{
		TotalClusters:               totalClusters,
		TotalCpuCapacity:            NewResourceCapacity("0", "cores"),
		TotalMemoryCapacity:         NewResourceCapacity("0", "Gi"),
		ClusterStatusBreakdown:      NewClusterStatusBreakdown(),
		NodeSchedulingBreakdown:     NewNodeSchedulingBreakdown(),
		NodeErrorBreakdown:          NewNodeErrorBreakdown(),
		ClusterDistribution:         NewClusterDistribution(),
		ClusterCapacityDistribution: []bean.ClusterCapacityDistribution{},
		NodeDistribution:            NewNodeDistribution(),
	}
}

// NewEmptyClusterOverviewResponse creates an empty ClusterOverviewResponse with zero values
func NewEmptyClusterOverviewResponse() *bean.ClusterOverviewResponse {
	return NewClusterOverviewResponse(0)
}

// NewResourceCapacity creates a new ResourceCapacity with the given value and unit
func NewResourceCapacity(value, unit string) *bean.ResourceCapacity {
	return &bean.ResourceCapacity{
		Value: value,
		Unit:  unit,
	}
}

// NewClusterStatusBreakdown creates a new ClusterStatusBreakdown with zero values
func NewClusterStatusBreakdown() *bean.ClusterStatusBreakdown {
	return &bean.ClusterStatusBreakdown{
		Healthy:          0,
		Unhealthy:        0,
		ConnectionFailed: 0,
	}
}

// NewNodeSchedulingBreakdown creates a new NodeSchedulingBreakdown with initialized slices
func NewNodeSchedulingBreakdown() *bean.NodeSchedulingBreakdown {
	return &bean.NodeSchedulingBreakdown{
		Schedulable:        0,
		Unschedulable:      0,
		Total:              0,
		SchedulableNodes:   []bean.NodeSchedulingDetail{},
		UnschedulableNodes: []bean.NodeSchedulingDetail{},
	}
}

// NewNodeErrorBreakdown creates a new NodeErrorBreakdown with initialized error counts map
func NewNodeErrorBreakdown() *bean.NodeErrorBreakdown {
	errorCounts := make(map[string]int)
	// Initialize all error types with zero counts
	errorCounts[constants.NodeErrorNetworkUnavailable] = 0
	errorCounts[constants.NodeErrorMemoryPressure] = 0
	errorCounts[constants.NodeErrorDiskPressure] = 0
	errorCounts[constants.NodeErrorPIDPressure] = 0
	errorCounts[constants.NodeErrorKubeletNotReady] = 0
	errorCounts[constants.NodeErrorOthers] = 0

	return &bean.NodeErrorBreakdown{
		ErrorCounts: errorCounts,
		Total:       0,
		NodeErrors:  []bean.NodeErrorDetail{},
	}
}

// NewClusterDistribution creates a new ClusterDistribution with empty slices
func NewClusterDistribution() *bean.ClusterDistribution {
	return &bean.ClusterDistribution{
		ByProvider: []bean.ProviderDistribution{},
		ByVersion:  []bean.VersionDistribution{},
	}
}

// NewNodeDistribution creates a new NodeDistribution with empty slices
func NewNodeDistribution() *bean.NodeDistribution {
	return &bean.NodeDistribution{
		ByClusters:   []bean.ClusterNodeCount{},
		ByAutoscaler: []bean.AutoscalerNodeCount{},
	}
}

// NewClusterCapacityDistribution creates a new ClusterCapacityDistribution entry
func NewClusterCapacityDistribution(clusterID int, clusterName string, serverVersion string, cpuCapacity float64, cpuUtil, cpuRequest, cpuLimit float64, memCapacity float64, memUtil, memRequest, memLimit float64) bean.ClusterCapacityDistribution {
	return bean.ClusterCapacityDistribution{
		ClusterID:     clusterID,
		ClusterName:   clusterName,
		ServerVersion: serverVersion,
		CPU:           NewClusterResourceMetric(cpuCapacity, cpuUtil, cpuRequest, cpuLimit),
		Memory:        NewClusterResourceMetric(memCapacity, memUtil, memRequest, memLimit),
	}
}

// NewClusterResourceMetric creates a new ClusterResourceMetric with capacity and percentages rounded to 2 decimal places
func NewClusterResourceMetric(capacity float64, utilPercent, requestPercent, limitPercent float64) *bean.ClusterResourceMetric {
	return &bean.ClusterResourceMetric{
		Capacity:           util.RoundToTwoDecimals(capacity),
		UtilizationPercent: util.RoundToTwoDecimals(utilPercent),
		RequestsPercent:    util.RoundToTwoDecimals(requestPercent),
		LimitsPercent:      util.RoundToTwoDecimals(limitPercent),
	}
}

// NewClusterNodeCount creates a new ClusterNodeCount entry
func NewClusterNodeCount(clusterID int, clusterName string, nodeCount int) bean.ClusterNodeCount {
	return bean.ClusterNodeCount{
		ClusterID:   clusterID,
		ClusterName: clusterName,
		NodeCount:   nodeCount,
	}
}

// NewNodeErrorDetail creates a new NodeErrorDetail entry
func NewNodeErrorDetail(nodeName, clusterName string, clusterID int, errors []string, nodeStatus string) bean.NodeErrorDetail {
	return bean.NodeErrorDetail{
		NodeName:    nodeName,
		ClusterName: clusterName,
		ClusterID:   clusterID,
		Errors:      errors,
		NodeStatus:  nodeStatus,
	}
}

// NewNodeSchedulingDetail creates a new NodeSchedulingDetail entry
func NewNodeSchedulingDetail(nodeName, clusterName string, clusterID int, schedulable bool) bean.NodeSchedulingDetail {
	return bean.NodeSchedulingDetail{
		NodeName:    nodeName,
		ClusterName: clusterName,
		ClusterID:   clusterID,
		Schedulable: schedulable,
	}
}

// NewProviderDistribution creates a new ProviderDistribution entry
func NewProviderDistribution(provider string, count int) bean.ProviderDistribution {
	return bean.ProviderDistribution{
		Provider: provider,
		Count:    count,
	}
}

// NewVersionDistribution creates a new VersionDistribution entry
func NewVersionDistribution(version string, count int) bean.VersionDistribution {
	return bean.VersionDistribution{
		Version: version,
		Count:   count,
	}
}

// NewClusterOverviewNodeDetailedResponse creates a new ClusterOverviewNodeDetailedResponse
func NewClusterOverviewNodeDetailedResponse(totalCount int, nodeList []bean.ClusterOverviewNodeDetailedItem) *bean.ClusterOverviewNodeDetailedResponse {
	return &bean.ClusterOverviewNodeDetailedResponse{
		TotalCount: totalCount,
		NodeList:   nodeList,
	}
}

// NewEmptyClusterOverviewNodeDetailedResponse creates an empty response for when cache is not found
func NewEmptyClusterOverviewNodeDetailedResponse() *bean.ClusterOverviewNodeDetailedResponse {
	return &bean.ClusterOverviewNodeDetailedResponse{
		TotalCount: 0,
		NodeList:   []bean.ClusterOverviewNodeDetailedItem{},
	}
}

// NewClusterUpgradeOverviewResponse creates a new ClusterUpgradeOverviewResponse
func NewClusterUpgradeOverviewResponse(canUpgrade bool, latestVersion string, clusterList []bean.ClusterUpgradeDetails) *bean.ClusterUpgradeOverviewResponse {
	return &bean.ClusterUpgradeOverviewResponse{
		CanCurrentUserUpgrade: canUpgrade,
		LatestVersion:         latestVersion,
		ClusterList:           clusterList,
	}
}
