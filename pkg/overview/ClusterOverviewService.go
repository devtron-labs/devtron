/*
 * Copyright (c) 2024. Devtron Inc.
 */

package overview

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/devtron-labs/devtron/pkg/asyncProvider"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	clusterService "github.com/devtron-labs/devtron/pkg/cluster"
	clusterBean "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/k8s"
	capacityService "github.com/devtron-labs/devtron/pkg/k8s/capacity"
	capacityBean "github.com/devtron-labs/devtron/pkg/k8s/capacity/bean"
	"github.com/devtron-labs/devtron/pkg/overview/bean"
	"github.com/devtron-labs/devtron/pkg/overview/cache"
	"github.com/devtron-labs/devtron/pkg/overview/config"
	"github.com/devtron-labs/devtron/pkg/overview/constants"
	overviewUtil "github.com/devtron-labs/devtron/pkg/overview/util"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
)

// ClusterOverviewService provides cluster management overview functionality
type ClusterOverviewService interface {
	GetClusterOverview(ctx context.Context) (*bean.ClusterOverviewResponse, error)
	GetClusterOverviewDetailedNodeInfo(ctx context.Context, request *bean.ClusterOverviewDetailRequest) (*bean.ClusterOverviewNodeDetailedResponse, error)
	RefreshClusterOverviewCache(ctx context.Context) error
}

// ClusterOverviewServiceImpl implements ClusterOverviewService
type ClusterOverviewServiceImpl struct {
	logger              *zap.SugaredLogger
	clusterService      clusterService.ClusterService
	k8sCapacityService  capacityService.K8sCapacityService
	clusterCacheService cache.ClusterCacheService
	k8sCommonService    k8s.K8sCommonService
	enforcer            casbin.Enforcer
	config              *config.ClusterOverviewConfig
}

func NewClusterOverviewServiceImpl(
	logger *zap.SugaredLogger,
	clusterService clusterService.ClusterService,
	k8sCapacityService capacityService.K8sCapacityService,
	clusterCacheService cache.ClusterCacheService,
	k8sCommonService k8s.K8sCommonService,
	enforcer casbin.Enforcer,
	cfg *config.ClusterOverviewConfig,
) *ClusterOverviewServiceImpl {
	service := &ClusterOverviewServiceImpl{
		logger:              logger,
		clusterService:      clusterService,
		k8sCapacityService:  k8sCapacityService,
		clusterCacheService: clusterCacheService,
		k8sCommonService:    k8sCommonService,
		enforcer:            enforcer,
		config:              cfg,
	}

	// Start background refresh worker if enabled
	if cfg.CacheEnabled && cfg.BackgroundRefreshEnabled {
		ctx := context.Background()
		service.StartBackgroundRefresh(ctx)
		logger.Info("Background cache refresh worker started")
	} else {
		logger.Info("Background cache refresh worker disabled")
	}

	return service
}

// StartBackgroundRefresh starts the background cache refresh worker
func (impl *ClusterOverviewServiceImpl) StartBackgroundRefresh(ctx context.Context) {
	if !impl.config.CacheEnabled || !impl.config.BackgroundRefreshEnabled {
		impl.logger.Info("Background refresh disabled")
		return
	}

	impl.logger.Infow("Starting background cache refresh worker",
		"refreshInterval", impl.config.GetRefreshInterval(),
		"maxParallelClusters", impl.config.MaxParallelClusters)

	// Initial cache population
	go func() {
		impl.logger.Info("Performing initial cache population")
		if err := impl.refreshCache(ctx); err != nil {
			impl.logger.Errorw("initial cache population failed", "err", err)
		}
	}()

	// Start periodic refresh
	ticker := time.NewTicker(impl.config.GetRefreshInterval())
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				impl.logger.Info("background refresh worker stopped")
				return
			case <-ticker.C:
				impl.logger.Info("background refresh triggered")
				if err := impl.refreshCache(ctx); err != nil {
					impl.logger.Errorw("background cache refresh failed", "err", err)
				}
			}
		}
	}()
}

// RefreshClusterOverviewCache manually triggers a cache refresh
func (impl *ClusterOverviewServiceImpl) RefreshClusterOverviewCache(ctx context.Context) error {
	return impl.refreshCache(ctx)
}

// refreshCache fetches fresh data and updates cache
func (impl *ClusterOverviewServiceImpl) refreshCache(ctx context.Context) error {
	// Prevent concurrent refreshes
	if impl.clusterCacheService.IsRefreshing() {
		impl.logger.Debug("Cache refresh already in progress, skipping")
		return nil
	}

	impl.clusterCacheService.SetRefreshing(true)
	defer impl.clusterCacheService.SetRefreshing(false)

	startTime := time.Now()
	impl.logger.Debug("Starting cache refresh")

	// Fetch clusters
	clusters, err := impl.clusterService.FindActiveClustersExcludingVirtual()
	if err != nil {
		impl.logger.Errorw("error fetching clusters for cache refresh", "err", err)
		return err
	}

	// Fetch cluster data in parallel
	response, err := impl.fetchClusterDataParallel(ctx, clusters)
	if err != nil {
		impl.logger.Errorw("error fetching cluster data for cache refresh", "err", err)
		return err
	}

	// Update cache
	if err := impl.clusterCacheService.SetClusterOverview(response); err != nil {
		impl.logger.Errorw("error updating cache", "err", err)
		return err
	}

	duration := time.Since(startTime)
	impl.logger.Infow("Cache refresh completed", "duration", duration, "clusterCount", len(clusters), "totalClusters", response.TotalClusters)

	return nil
}

// fetchClusterDataParallel fetches cluster data using worker pool for parallel execution
func (impl *ClusterOverviewServiceImpl) fetchClusterDataParallel(ctx context.Context, clusters []clusterBean.ClusterBean) (*bean.ClusterOverviewResponse, error) {
	if len(clusters) == 0 {
		return NewEmptyClusterOverviewResponse(), nil
	}

	// Separate clusters into valid and error clusters
	validClusters := make([]clusterBean.ClusterBean, 0, len(clusters))
	errorClusters := make([]clusterBean.ClusterBean, 0)

	for _, cluster := range clusters {
		if len(cluster.ErrorInConnecting) > 0 {
			impl.logger.Debugw("Skipping cluster with connection error", "clusterId", cluster.Id, "clusterName", cluster.ClusterName, "error", cluster.ErrorInConnecting)
			errorClusters = append(errorClusters, cluster)
			continue
		}
		validClusters = append(validClusters, cluster)
	}

	if len(errorClusters) > 0 {
		impl.logger.Infow("Skipped clusters with connection errors", "skippedCount", len(errorClusters), "validCount", len(validClusters), "totalCount", len(clusters))
	}

	// Create placeholder capacity details for clusters with errors
	errorClusterDetails := make([]*capacityBean.ClusterCapacityDetail, 0, len(errorClusters))
	for _, cluster := range errorClusters {
		errorClusterDetails = append(errorClusterDetails, &capacityBean.ClusterCapacityDetail{
			Id:                cluster.Id,
			Name:              cluster.ClusterName,
			ErrorInConnection: cluster.ErrorInConnecting,
			Status:            capacityBean.ClusterStatusConnectionFailed,
			IsVirtualCluster:  cluster.IsVirtualCluster,
			IsProd:            cluster.IsProd,
		})
	}

	// If all clusters have connection errors, return response with error clusters only
	if len(validClusters) == 0 {
		impl.logger.Warn("All clusters have connection errors, returning response with error clusters only")
		allClusterPointers := make([]*clusterBean.ClusterBean, len(clusters))
		for i := range clusters {
			allClusterPointers[i] = &clusters[i]
		}
		return impl.aggregateClusterCapacityDetails(ctx, errorClusterDetails, allClusterPointers), nil
	}

	// Create worker pool with configured parallelism
	wp := asyncProvider.NewBatchWorker[*capacityBean.ClusterCapacityDetail](
		impl.config.MaxParallelClusters,
		impl.logger,
	)
	wp.InitializeResponse()

	// Convert to pointer slice (only valid clusters)
	clusterPointers := make([]*clusterBean.ClusterBean, len(validClusters))
	for i := range validClusters {
		clusterPointers[i] = &validClusters[i]
	}

	// Submit cluster fetch tasks to worker pool
	for _, cluster := range clusterPointers {
		clusterCopy := cluster // Capture for closure
		wp.Submit(func() (*capacityBean.ClusterCapacityDetail, error) {
			impl.logger.Debugw("Fetching cluster capacity", "clusterId", clusterCopy.Id, "clusterName", clusterCopy.ClusterName)

			// Fetch cluster capacity detail
			detail, err := impl.k8sCapacityService.GetClusterCapacityDetail(ctx, clusterCopy, false)
			if err != nil {
				impl.logger.Warnw("error fetching cluster capacity, skipping", "clusterId", clusterCopy.Id, "clusterName", clusterCopy.ClusterName, "err", err)
				// Populate error for this cluster
				detail = &capacityBean.ClusterCapacityDetail{
					ErrorInConnection: err.Error(),
					Status:            capacityBean.ClusterStatusConnectionFailed,
				}
				// Continue to next cluster, returning error will stop the worker pool from further processing
			}

			// Set cluster metadata
			detail.Id = clusterCopy.Id
			detail.Name = clusterCopy.ClusterName
			detail.IsVirtualCluster = clusterCopy.IsVirtualCluster
			detail.IsProd = clusterCopy.IsProd

			return detail, nil
		})
	}

	// Wait for all tasks to complete
	if err := wp.StopWait(); err != nil {
		impl.logger.Errorw("error waiting for worker pool tasks", "err", err)
		// Continue anyway to return partial results
	}

	// Get results from worker pool
	results := wp.GetResponse()

	// Combine successful results with error cluster placeholders
	allClusterDetails := make([]*capacityBean.ClusterCapacityDetail, 0, len(results)+len(errorClusterDetails))
	allClusterDetails = append(allClusterDetails, results...)
	allClusterDetails = append(allClusterDetails, errorClusterDetails...)

	// Create combined cluster bean list (all clusters)
	allClusterPointers := make([]*clusterBean.ClusterBean, len(clusters))
	for i := range clusters {
		allClusterPointers[i] = &clusters[i]
	}

	// Log summary
	successCount := len(results)
	failedCount := len(validClusters) - successCount
	if failedCount > 0 || len(errorClusters) > 0 {
		impl.logger.Infow("Cluster fetch summary", "successCount", successCount, "failedCount", failedCount, "skippedCount", len(errorClusters), "totalClusters", len(clusters))
	}

	// Aggregate all results (including error clusters) into response
	return impl.aggregateClusterCapacityDetails(ctx, allClusterDetails, allClusterPointers), nil
}

// aggregateClusterCapacityDetails aggregates cluster capacity details into overview response
func (impl *ClusterOverviewServiceImpl) aggregateClusterCapacityDetails(ctx context.Context, details []*capacityBean.ClusterCapacityDetail, clusterBeans []*clusterBean.ClusterBean) *bean.ClusterOverviewResponse {
	return impl.buildClusterOverviewResponse(ctx, details, clusterBeans)
}

// GetClusterOverview retrieves comprehensive cluster management overview
// Returns from cache if enabled and available, otherwise fetches directly
func (impl *ClusterOverviewServiceImpl) GetClusterOverview(ctx context.Context) (*bean.ClusterOverviewResponse, error) {
	// If cache is disabled, fetch directly
	if !impl.config.CacheEnabled {
		impl.logger.Debug("Cache disabled, fetching cluster overview directly")
		return impl.fetchClusterOverviewDirect(ctx)
	}

	// Try to get from cache
	if cachedData, found := impl.clusterCacheService.GetClusterOverview(); found {
		return impl.handleCacheHit(cachedData)
	}

	// Cache miss - fallback to direct fetch
	return impl.handleCacheMiss(ctx)
}

// handleCacheHit processes a cache hit and returns the cached data
func (impl *ClusterOverviewServiceImpl) handleCacheHit(cachedData *bean.ClusterOverviewResponse) (*bean.ClusterOverviewResponse, error) {
	cacheAge := impl.clusterCacheService.GetCacheAge()

	// Warn if cache is stale but return it anyway
	if cacheAge > impl.config.GetMaxStaleDataDuration() {
		impl.logger.Warnw("cache is stale but returning anyway",
			"cacheAge", cacheAge,
			"maxStaleAge", impl.config.GetMaxStaleDataDuration())
	}

	impl.logger.Infow("returning cluster overview from cache", "cacheAge", cacheAge)
	return cachedData, nil
}

// handleCacheMiss handles cache miss by attempting to refresh cache or fetching directly
func (impl *ClusterOverviewServiceImpl) handleCacheMiss(ctx context.Context) (*bean.ClusterOverviewResponse, error) {
	impl.logger.Warn("cache miss - background refresh may not be running, attempting fallback")

	// Try to refresh cache synchronously
	if err := impl.refreshCache(ctx); err != nil {
		impl.logger.Errorw("error refreshing cache synchronously, falling back to direct fetch", "err", err)
		// Fallback to direct fetch without caching
		return impl.fetchClusterOverviewDirect(ctx)
	}

	// Try to get from cache after refresh
	if cachedData, found := impl.clusterCacheService.GetClusterOverview(); found {
		impl.logger.Info("successfully populated cache, returning data")
		return cachedData, nil
	}

	// Cache refresh succeeded but data not in cache (shouldn't happen)
	impl.logger.Warn("cache refresh succeeded but data not found in cache, falling back to direct fetch")
	return impl.fetchClusterOverviewDirect(ctx)
}

// fetchClusterOverviewDirect fetches cluster overview without using cache
func (impl *ClusterOverviewServiceImpl) fetchClusterOverviewDirect(ctx context.Context) (*bean.ClusterOverviewResponse, error) {
	impl.logger.Debug("Fetching cluster overview directly (bypassing cache)")

	// Fetch active clusters
	clusters, err := impl.clusterService.FindActiveClustersExcludingVirtual()
	if err != nil {
		impl.logger.Errorw("error fetching clusters", "err", err)
		return nil, fmt.Errorf("failed to fetch clusters: %w", err)
	}

	// Fetch cluster data in parallel
	response, err := impl.fetchClusterDataParallel(ctx, clusters)
	if err != nil {
		impl.logger.Errorw("error fetching cluster data", "err", err)
		return nil, fmt.Errorf("failed to fetch cluster data: %w", err)
	}

	impl.logger.Infow("successfully fetched cluster overview directly", "totalClusters", response.TotalClusters)
	return response, nil
}

func (impl *ClusterOverviewServiceImpl) buildClusterOverviewResponse(ctx context.Context, clusterCapacityDetails []*capacityBean.ClusterCapacityDetail, clusterBeans []*clusterBean.ClusterBean) *bean.ClusterOverviewResponse {
	// Initialize response using adapter
	response := NewClusterOverviewResponse(len(clusterCapacityDetails))

	// Tracking variables for aggregation
	var totalCpuCapacityCores, totalMemoryCapacityGi float64
	providerCounts := make(map[string]int)
	versionCounts := make(map[string]int)
	autoscalerCounts := make(map[string]int)
	autoscalerNodeDetailsMap := make(map[string][]bean.AutoscalerNodeDetail)

	// Create a map of cluster ID to cluster bean for quick lookup
	clusterBeanMap := make(map[int]*clusterBean.ClusterBean)
	for _, cb := range clusterBeans {
		clusterBeanMap[cb.Id] = cb
	}

	// Process each cluster to extract and aggregate data
	for _, cluster := range clusterCapacityDetails {
		impl.processClusterStatus(cluster, response)

		if len(cluster.ErrorInConnection) == 0 {
			metrics := impl.processClusterCapacity(cluster, &totalCpuCapacityCores, &totalMemoryCapacityGi)
			impl.addClusterCapacityDistribution(cluster, response, metrics)

			// Get the corresponding cluster bean for autoscaler detection
			if clusterBeanForAutoscaler, exists := clusterBeanMap[cluster.Id]; exists {
				impl.processNodeDistributionAndAutoscaler(ctx, cluster, clusterBeanForAutoscaler, response, autoscalerCounts, autoscalerNodeDetailsMap)
			} else {
				impl.logger.Warnw("cluster bean not found for autoscaler detection",
					"clusterId", cluster.Id,
					"clusterName", cluster.Name)
			}
			impl.aggregateClusterMetadata(cluster, providerCounts, versionCounts)
		}
		impl.processNodeDetails(cluster, response)
		impl.aggregateNodeErrorCounts(cluster, response)

	}

	impl.finalizeResponse(response, totalCpuCapacityCores, totalMemoryCapacityGi, providerCounts, versionCounts, autoscalerNodeDetailsMap)

	return response
}

// processClusterStatus updates cluster status breakdown based on cluster health
func (impl *ClusterOverviewServiceImpl) processClusterStatus(cluster *capacityBean.ClusterCapacityDetail, response *bean.ClusterOverviewResponse) {
	if cluster.Status == capacityBean.ClusterStatusHealthy {
		response.ClusterStatusBreakdown.Healthy++
	} else if cluster.Status == capacityBean.ClusterStatusConnectionFailed {
		response.ClusterStatusBreakdown.ConnectionFailed++
	} else {
		response.ClusterStatusBreakdown.Unhealthy++
	}
}

// clusterCapacityMetrics holds parsed capacity metrics for a cluster
type clusterCapacityMetrics struct {
	cpuCapacity    float64
	cpuUtil        float64
	cpuRequest     float64
	cpuLimit       float64
	memoryCapacity float64
	memoryUtil     float64
	memoryRequest  float64
	memoryLimit    float64
}

// processClusterCapacity extracts and aggregates CPU and memory metrics from cluster
func (impl *ClusterOverviewServiceImpl) processClusterCapacity(cluster *capacityBean.ClusterCapacityDetail, totalCpu, totalMemory *float64) clusterCapacityMetrics {
	metrics := clusterCapacityMetrics{}

	// Process CPU metrics
	if cluster.Cpu != nil {
		cpuCapacityFloat, err := strconv.ParseFloat(cluster.Cpu.Capacity, 64)
		if err != nil {
			impl.logger.Errorw("error in parsing cpu capacity", "err", err, "capacity", cluster.Cpu.Capacity)
			cpuCapacityFloat = 0
		}
		metrics.cpuCapacity = cpuCapacityFloat
		*totalCpu += cpuCapacityFloat

		metrics.cpuUtil, _ = strconv.ParseFloat(strings.TrimSuffix(cluster.Cpu.UsagePercentage, "%"), 64)
		metrics.cpuRequest, _ = strconv.ParseFloat(strings.TrimSuffix(cluster.Cpu.RequestPercentage, "%"), 64)
		metrics.cpuLimit, _ = strconv.ParseFloat(strings.TrimSuffix(cluster.Cpu.LimitPercentage, "%"), 64)
	}

	// Process Memory metrics
	if cluster.Memory != nil {
		memoryCapacityStr := strings.TrimSuffix(cluster.Memory.Capacity, "Gi")
		memoryCapacityFloat, err := strconv.ParseFloat(memoryCapacityStr, 64)
		if err != nil {
			impl.logger.Errorw("error in parsing memory capacity", "err", err, "capacity", cluster.Memory.Capacity)
			memoryCapacityFloat = 0
		}
		metrics.memoryCapacity = memoryCapacityFloat
		*totalMemory += memoryCapacityFloat

		metrics.memoryUtil, _ = strconv.ParseFloat(strings.TrimSuffix(cluster.Memory.UsagePercentage, "%"), 64)
		metrics.memoryRequest, _ = strconv.ParseFloat(strings.TrimSuffix(cluster.Memory.RequestPercentage, "%"), 64)
		metrics.memoryLimit, _ = strconv.ParseFloat(strings.TrimSuffix(cluster.Memory.LimitPercentage, "%"), 64)
	}

	return metrics
}

// addClusterCapacityDistribution adds cluster capacity distribution entry to response
func (impl *ClusterOverviewServiceImpl) addClusterCapacityDistribution(cluster *capacityBean.ClusterCapacityDetail, response *bean.ClusterOverviewResponse, metrics clusterCapacityMetrics) {
	response.ClusterCapacityDistribution = append(response.ClusterCapacityDistribution,
		NewClusterCapacityDistribution(
			cluster.Id,
			cluster.Name,
			cluster.ServerVersion,
			metrics.cpuCapacity,
			metrics.cpuUtil,
			metrics.cpuRequest,
			metrics.cpuLimit,
			metrics.memoryCapacity,
			metrics.memoryUtil,
			metrics.memoryRequest,
			metrics.memoryLimit,
		))
}

// processNodeDistributionAndAutoscaler adds cluster node count to distribution and aggregates autoscaler counts across all clusters
func (impl *ClusterOverviewServiceImpl) processNodeDistributionAndAutoscaler(ctx context.Context, cluster *capacityBean.ClusterCapacityDetail, clusterBean *clusterBean.ClusterBean, response *bean.ClusterOverviewResponse, autoscalerCounts map[string]int, autoscalerNodeDetailsMap map[string][]bean.AutoscalerNodeDetail) {
	// Add cluster node count to distribution
	response.NodeDistribution.ByClusters = append(response.NodeDistribution.ByClusters,
		NewClusterNodeCount(cluster.Id, cluster.Name, cluster.NodeCount))

	// Fetch node details with labels to determine autoscaler types
	nodeCapacityDetails, err := impl.k8sCapacityService.GetNodeCapacityDetailsListByCluster(ctx, clusterBean)
	if err != nil {
		impl.logger.Errorw("error fetching node capacity details for autoscaler detection, skipping autoscaler aggregation",
			"clusterId", cluster.Id,
			"clusterName", cluster.Name,
			"err", err)
		return
	}

	// Process each node to determine autoscaler type and aggregate globally
	for _, nodeDetail := range nodeCapacityDetails {
		autoscalerType := overviewUtil.DetermineAutoscalerTypeFromLabelArray(nodeDetail.Labels)

		// Add to global autoscaler counts
		autoscalerCounts[autoscalerType]++

		// Collect node details for this autoscaler type globally across all clusters
		autoscalerNodeDetailsMap[autoscalerType] = append(autoscalerNodeDetailsMap[autoscalerType], bean.AutoscalerNodeDetail{
			NodeName:    nodeDetail.Name,
			ClusterName: cluster.Name,
			ClusterID:   cluster.Id,
			ManagedBy:   autoscalerType,
		})
	}
}

// processNodeDetails processes node details to populate scheduling and error information
func (impl *ClusterOverviewServiceImpl) processNodeDetails(cluster *capacityBean.ClusterCapacityDetail, response *bean.ClusterOverviewResponse) {
	if cluster.NodeDetails == nil {
		return
	}

	// Build node errors map for quick lookup
	nodeErrorsMap := impl.buildNodeErrorsMap(cluster.NodeErrors)

	// Process each node
	for _, nodeDetail := range cluster.NodeDetails {
		if errorTypes, hasErrors := nodeErrorsMap[nodeDetail.NodeName]; hasErrors {
			impl.addNodeWithErrors(nodeDetail, cluster, errorTypes, response)
		} else {
			impl.addSchedulableNode(nodeDetail, cluster, response)
		}
	}
}

// buildNodeErrorsMap creates a map of node names to their error types
func (impl *ClusterOverviewServiceImpl) buildNodeErrorsMap(nodeErrors map[corev1.NodeConditionType][]string) map[string][]string {
	nodeErrorsMap := make(map[string][]string)
	for conditionType, nodeNames := range nodeErrors {
		for _, nodeName := range nodeNames {
			errorType := impl.getHumanReadableErrorType(conditionType)
			nodeErrorsMap[nodeName] = append(nodeErrorsMap[nodeName], errorType)
		}
	}
	return nodeErrorsMap
}

// addNodeWithErrors adds a node with errors to error breakdown and unschedulable nodes
func (impl *ClusterOverviewServiceImpl) addNodeWithErrors(nodeDetail capacityBean.NodeDetails, cluster *capacityBean.ClusterCapacityDetail, errorTypes []string, response *bean.ClusterOverviewResponse) {
	nodeStatus := "Not Ready"
	if len(errorTypes) == 0 {
		nodeStatus = "Ready"
	}

	// Store errors as array directly, no need to convert to comma-separated string
	response.NodeErrorBreakdown.NodeErrors = append(response.NodeErrorBreakdown.NodeErrors,
		NewNodeErrorDetail(nodeDetail.NodeName, cluster.Name, cluster.Id, errorTypes, nodeStatus))

	response.NodeSchedulingBreakdown.UnschedulableNodes = append(response.NodeSchedulingBreakdown.UnschedulableNodes,
		NewNodeSchedulingDetail(nodeDetail.NodeName, cluster.Name, cluster.Id, false))
	response.NodeSchedulingBreakdown.Unschedulable++
}

// addSchedulableNode adds a schedulable node to the scheduling breakdown
func (impl *ClusterOverviewServiceImpl) addSchedulableNode(nodeDetail capacityBean.NodeDetails, cluster *capacityBean.ClusterCapacityDetail, response *bean.ClusterOverviewResponse) {
	response.NodeSchedulingBreakdown.SchedulableNodes = append(response.NodeSchedulingBreakdown.SchedulableNodes,
		NewNodeSchedulingDetail(nodeDetail.NodeName, cluster.Name, cluster.Id, true))
	response.NodeSchedulingBreakdown.Schedulable++
}

// aggregateNodeErrorCounts aggregates node error counts by error type
func (impl *ClusterOverviewServiceImpl) aggregateNodeErrorCounts(cluster *capacityBean.ClusterCapacityDetail, response *bean.ClusterOverviewResponse) {
	for conditionType, nodeNames := range cluster.NodeErrors {
		errorCount := len(nodeNames)

		switch conditionType {
		case constants.NodeConditionNetworkUnavailable:
			response.NodeErrorBreakdown.ErrorCounts[constants.NodeErrorNetworkUnavailable] += errorCount
		case constants.NodeConditionMemoryPressure:
			response.NodeErrorBreakdown.ErrorCounts[constants.NodeErrorMemoryPressure] += errorCount
		case constants.NodeConditionDiskPressure:
			response.NodeErrorBreakdown.ErrorCounts[constants.NodeErrorDiskPressure] += errorCount
		case constants.NodeConditionPIDPressure:
			response.NodeErrorBreakdown.ErrorCounts[constants.NodeErrorPIDPressure] += errorCount
		case constants.NodeConditionReady:
			response.NodeErrorBreakdown.ErrorCounts[constants.NodeErrorKubeletNotReady] += errorCount
		default:
			response.NodeErrorBreakdown.ErrorCounts[constants.NodeErrorOthers] += errorCount
		}
	}
}

// aggregateClusterMetadata aggregates cluster metadata (provider and version)
func (impl *ClusterOverviewServiceImpl) aggregateClusterMetadata(cluster *capacityBean.ClusterCapacityDetail, providerCounts, versionCounts map[string]int) {
	provider := impl.determineProviderFromCluster(cluster)
	providerCounts[provider]++

	version := impl.extractMajorMinorVersion(cluster.ServerVersion)
	versionCounts[version]++
}

// finalizeResponse sets total values and builds distribution arrays
func (impl *ClusterOverviewServiceImpl) finalizeResponse(response *bean.ClusterOverviewResponse, totalCpu, totalMemory float64, providerCounts, versionCounts map[string]int, autoscalerNodeDetailsMap map[string][]bean.AutoscalerNodeDetail) {
	// Set total capacity values with 2 decimal precision
	response.TotalCpuCapacity.Value = fmt.Sprintf("%.2f", overviewUtil.RoundToTwoDecimals(totalCpu))
	response.TotalMemoryCapacity.Value = fmt.Sprintf("%.2f", overviewUtil.RoundToTwoDecimals(totalMemory))

	// Build provider distribution
	for provider, count := range providerCounts {
		response.ClusterDistribution.ByProvider = append(response.ClusterDistribution.ByProvider,
			NewProviderDistribution(provider, count))
	}

	// Build version distribution
	for version, count := range versionCounts {
		response.ClusterDistribution.ByVersion = append(response.ClusterDistribution.ByVersion,
			NewVersionDistribution(version, count))
	}

	// Build autoscaler distribution - aggregated across all clusters
	for autoscalerType, nodeDetails := range autoscalerNodeDetailsMap {
		response.NodeDistribution.ByAutoscaler = append(response.NodeDistribution.ByAutoscaler, bean.AutoscalerNodeCount{
			AutoscalerType: autoscalerType,
			NodeCount:      len(nodeDetails),
			NodeDetails:    nodeDetails,
		})
	}

	// Set total counts for breakdowns
	response.NodeSchedulingBreakdown.Total = response.NodeSchedulingBreakdown.Schedulable + response.NodeSchedulingBreakdown.Unschedulable
	response.NodeErrorBreakdown.Total = len(response.NodeErrorBreakdown.NodeErrors)
}

// getHumanReadableErrorType converts Kubernetes node condition types to human-readable error types
func (impl *ClusterOverviewServiceImpl) getHumanReadableErrorType(conditionType corev1.NodeConditionType) string {
	switch conditionType {
	case corev1.NodeNetworkUnavailable:
		return constants.NodeErrorNetworkUnavailable
	case corev1.NodeMemoryPressure:
		return constants.NodeErrorMemoryPressure
	case corev1.NodeDiskPressure:
		return constants.NodeErrorDiskPressure
	case corev1.NodePIDPressure:
		return constants.NodeErrorPIDPressure
	case corev1.NodeReady:
		return constants.NodeErrorKubeletNotReady
	default:
		return constants.NodeErrorOthers
	}
}

func (impl *ClusterOverviewServiceImpl) determineProviderFromCluster(cluster *capacityBean.ClusterCapacityDetail) string {
	if cluster.NodeDetails != nil && len(cluster.NodeDetails) > 0 {
		for _, nodeDetail := range cluster.NodeDetails {
			provider := impl.determineProviderFromNodeName(nodeDetail.NodeName)
			if provider != constants.ProviderUnknown {
				return provider
			}
		}
	}

	return constants.ProviderUnknown
}

// determineProviderFromNodeName determines cloud provider from node name patterns
func (impl *ClusterOverviewServiceImpl) determineProviderFromNodeName(nodeName string) string {
	nodeNameLower := strings.ToLower(nodeName)

	// Google Cloud Platform (GKE) patterns
	// Examples: gke-shared-cluster-ci-nodes-818049c0-6knz, gke-cluster-default-pool-12345678-abcd
	if strings.HasPrefix(nodeNameLower, constants.NodePrefixGKE) {
		return constants.ProviderGCP
	}

	// Azure (AKS) patterns
	// Examples: aks-newpool-37469834-vmss000000, aks-nodepool1-12345678-vmss000001
	if strings.HasPrefix(nodeNameLower, constants.NodePrefixAKS) {
		return constants.ProviderAzure
	}

	// AWS (EKS) patterns
	// Examples: ip-192-168-1-100.us-west-2.compute.internal, ip-10-0-1-50.ec2.internal
	if strings.Contains(nodeNameLower, constants.NodePatternAWSComputeInternal) || strings.Contains(nodeNameLower, constants.NodePatternAWSEC2Internal) {
		return constants.ProviderAWS
	}
	// EKS managed node groups: eks-nodegroup-12345678-abcd
	if strings.HasPrefix(nodeNameLower, constants.NodePrefixEKS) {
		return constants.ProviderAWS
	}

	// Additional AWS patterns: nodes with AWS region patterns
	for _, pattern := range constants.AWSRegionPatterns {
		if strings.Contains(nodeNameLower, pattern) {
			return constants.ProviderAWS
		}
	}

	// Oracle Cloud (OKE) patterns
	// Examples: oke-cywiqripuyg-nsgagklgnst-st2qczvnmba-0, oke-c1a2b3c4d5e-n6f7g8h9i0j-s1k2l3m4n5o-1
	if strings.HasPrefix(nodeNameLower, constants.NodePrefixOKE) {
		return constants.ProviderOracle
	}

	// DigitalOcean (DOKS) patterns
	// Examples: pool-<pool-id>-<random>, nodes often contain "digitalocean" in metadata
	if strings.Contains(nodeNameLower, constants.NodePatternDigitalOcean) {
		return constants.ProviderDigitalOcean
	}

	// IBM Cloud (IKS) patterns
	// Examples: kube-<cluster-id>-<worker-id>, 10.x.x.x.kube-<cluster-id>
	if strings.Contains(nodeNameLower, constants.NodePatternIBMKube) {
		return constants.ProviderIBM
	}

	// Alibaba Cloud (ACK) patterns
	// Examples: aliyun.com-59176-test, cn-hangzhou.i-bp12h6biv9bg24lmdc2o
	// Nodes often contain "aliyun" in their names or "cn-" prefix for Chinese regions
	if strings.Contains(nodeNameLower, constants.NodePatternAliyun) {
		return constants.ProviderAlibaba
	}
	// Alibaba Cloud region patterns (cn-hangzhou, cn-beijing, etc.)
	if strings.HasPrefix(nodeNameLower, constants.NodePatternAlibabaRegion) {
		return constants.ProviderAlibaba
	}

	// Additional Azure patterns: nodes with Azure region indicators
	if strings.Contains(nodeNameLower, constants.NodePatternAzureVMSS) || strings.Contains(nodeNameLower, constants.NodePatternAzureScaleSets) {
		return constants.ProviderAzure
	}

	// Additional GCP patterns
	if strings.Contains(nodeNameLower, constants.NodePatternGCP) || strings.Contains(nodeNameLower, constants.NodePatternGoogle) {
		return constants.ProviderGCP
	}

	return constants.ProviderUnknown
}

// extractMajorMinorVersion extracts major.minor version from Kubernetes version string using semver
// Examples: "v1.28.3" -> "1.28", "1.29.0-gke.1234" -> "1.29", "v1.30" -> "1.30"
func (impl *ClusterOverviewServiceImpl) extractMajorMinorVersion(version string) string {
	if version == "" {
		return constants.VersionUnknown
	}

	cleanVersion := version
	if strings.HasPrefix(version, "v") {
		cleanVersion = version[1:]
	}

	// Parse using semver library (same as ClusterUpgradeService)
	semverVersion, err := semver.Parse(cleanVersion)
	if err != nil {
		impl.logger.Warnw("failed to parse version using semver, falling back to string parsing", "version", version, "err", err)
		// Fallback to simple string parsing if semver fails
		parts := strings.Split(cleanVersion, ".")
		if len(parts) >= 2 {
			return fmt.Sprintf("%s.%s", parts[0], parts[1])
		}
		return constants.VersionUnknown
	}

	// Use the same approach as ClusterUpgradeService: TrimToMajorAndMinorVersion
	return fmt.Sprintf("%d.%d", semverVersion.Major, semverVersion.Minor)
}

// GetClusterOverviewDetailedNodeInfo retrieves paginated and filtered node details from cache based on node view group type
func (impl *ClusterOverviewServiceImpl) GetClusterOverviewDetailedNodeInfo(ctx context.Context, request *bean.ClusterOverviewDetailRequest) (*bean.ClusterOverviewNodeDetailedResponse, error) {
	clusterOverview, found := impl.clusterCacheService.GetClusterOverview()
	if !found {
		impl.logger.Warnw("cluster overview cache not found, returning empty response")
		return NewEmptyClusterOverviewNodeDetailedResponse(), nil
	}

	switch request.GroupBy {
	case bean.NodeViewGroupTypeNodeErrors:
		return impl.getNodeErrorsDetail(clusterOverview, request)
	case bean.NodeViewGroupTypeNodeScheduling:
		return impl.getNodeSchedulingDetail(clusterOverview, request)
	case bean.NodeViewGroupTypeAutoscaler:
		return impl.getAutoscalerDetail(clusterOverview, request)
	default:
		return nil, fmt.Errorf("invalid node view group type: %s", request.GroupBy)
	}
}

// getNodeErrorsDetail retrieves paginated and filtered node error details
func (impl *ClusterOverviewServiceImpl) getNodeErrorsDetail(clusterOverview *bean.ClusterOverviewResponse, request *bean.ClusterOverviewDetailRequest) (*bean.ClusterOverviewNodeDetailedResponse, error) {
	// Get all node errors from cache
	allNodes := clusterOverview.NodeErrorBreakdown.NodeErrors

	// Apply error type filter if specified
	if request.ErrorType != "" {
		allNodes = impl.filterNodeErrorsByType(allNodes, request.ErrorType)
	}

	// Apply search filter
	filteredNodes := impl.filterNodeErrors(allNodes, request.SearchKey)

	// Apply sorting
	sortedNodes := impl.sortNodeErrors(filteredNodes, request.SortBy, request.SortOrder)

	// Apply pagination
	totalCount := len(sortedNodes)
	paginatedNodes := impl.paginateNodeErrors(sortedNodes, request.Offset, request.Limit)

	// Convert to unified response format
	unifiedNodes := make([]bean.ClusterOverviewNodeDetailedItem, len(paginatedNodes))
	for i, node := range paginatedNodes {
		unifiedNodes[i] = bean.ClusterOverviewNodeDetailedItem{
			NodeName:    node.NodeName,
			ClusterName: node.ClusterName,
			ClusterID:   node.ClusterID,
			NodeErrors:  node.Errors,
			NodeStatus:  node.NodeStatus,
		}
	}

	return NewClusterOverviewNodeDetailedResponse(totalCount, unifiedNodes), nil
}

// getNodeSchedulingDetail retrieves paginated and filtered node scheduling details
func (impl *ClusterOverviewServiceImpl) getNodeSchedulingDetail(clusterOverview *bean.ClusterOverviewResponse, request *bean.ClusterOverviewDetailRequest) (*bean.ClusterOverviewNodeDetailedResponse, error) {
	// Filter by schedulable type if specified
	var allNodes []bean.NodeSchedulingDetail
	if request.SchedulableType != "" {
		switch request.SchedulableType {
		case constants.SchedulableTypeSchedulable:
			allNodes = clusterOverview.NodeSchedulingBreakdown.SchedulableNodes
		case constants.SchedulableTypeUnschedulable:
			allNodes = clusterOverview.NodeSchedulingBreakdown.UnschedulableNodes
		default:
			// Invalid schedulableType, return all nodes
			allNodes = append(clusterOverview.NodeSchedulingBreakdown.SchedulableNodes, clusterOverview.NodeSchedulingBreakdown.UnschedulableNodes...)
		}
	} else {
		// Combine schedulable and unschedulable nodes if no filter specified
		allNodes = append(clusterOverview.NodeSchedulingBreakdown.SchedulableNodes, clusterOverview.NodeSchedulingBreakdown.UnschedulableNodes...)
	}

	// Apply search filter
	filteredNodes := impl.filterNodeScheduling(allNodes, request.SearchKey)

	// Apply sorting
	sortedNodes := impl.sortNodeScheduling(filteredNodes, request.SortBy, request.SortOrder)

	// Apply pagination
	totalCount := len(sortedNodes)
	paginatedNodes := impl.paginateNodeScheduling(sortedNodes, request.Offset, request.Limit)

	unifiedNodes := make([]bean.ClusterOverviewNodeDetailedItem, len(paginatedNodes))
	for i, node := range paginatedNodes {
		unifiedNodes[i] = bean.ClusterOverviewNodeDetailedItem{
			NodeName:    node.NodeName,
			ClusterName: node.ClusterName,
			ClusterID:   node.ClusterID,
			Schedulable: node.Schedulable,
		}
	}

	return NewClusterOverviewNodeDetailedResponse(totalCount, unifiedNodes), nil
}

// getAutoscalerDetail retrieves paginated and filtered autoscaler node details
func (impl *ClusterOverviewServiceImpl) getAutoscalerDetail(clusterOverview *bean.ClusterOverviewResponse, request *bean.ClusterOverviewDetailRequest) (*bean.ClusterOverviewNodeDetailedResponse, error) {
	// Filter by autoscaler type if specified
	var allNodes []bean.AutoscalerNodeDetail
	if request.AutoscalerType != "" {
		// Get nodes only for the specified autoscaler type
		for _, autoscalerGroup := range clusterOverview.NodeDistribution.ByAutoscaler {
			if autoscalerGroup.AutoscalerType == request.AutoscalerType {
				allNodes = append(allNodes, autoscalerGroup.NodeDetails...)
				break
			}
		}
	} else {
		// Combine all autoscaler nodes from all autoscaler types
		for _, autoscalerGroup := range clusterOverview.NodeDistribution.ByAutoscaler {
			allNodes = append(allNodes, autoscalerGroup.NodeDetails...)
		}
	}

	// Apply search filter
	filteredNodes := impl.filterAutoscalerNodes(allNodes, request.SearchKey)

	// Apply sorting
	sortedNodes := impl.sortAutoscalerNodes(filteredNodes, request.SortBy, request.SortOrder)

	// Apply pagination
	totalCount := len(sortedNodes)
	paginatedNodes := impl.paginateAutoscalerNodes(sortedNodes, request.Offset, request.Limit)

	unifiedNodes := make([]bean.ClusterOverviewNodeDetailedItem, len(paginatedNodes))
	for i, node := range paginatedNodes {
		unifiedNodes[i] = bean.ClusterOverviewNodeDetailedItem{
			NodeName:       node.NodeName,
			ClusterName:    node.ClusterName,
			ClusterID:      node.ClusterID,
			AutoscalerType: node.ManagedBy,
		}
	}

	return NewClusterOverviewNodeDetailedResponse(totalCount, unifiedNodes), nil
}

// Helper methods for node error filtering, sorting, and pagination

// filterNodeErrorsByType filters nodes by specific error type
func (impl *ClusterOverviewServiceImpl) filterNodeErrorsByType(nodes []bean.NodeErrorDetail, errorType string) []bean.NodeErrorDetail {
	if errorType == "" {
		return nodes
	}

	var filtered []bean.NodeErrorDetail
	errorTypeLower := strings.ToLower(errorType)

	for _, node := range nodes {
		// Check if the node's error array contains the specified error type
		for _, err := range node.Errors {
			if strings.ToLower(err) == errorTypeLower {
				filtered = append(filtered, node)
				break // Found the error, no need to check other errors for this node
			}
		}
	}

	return filtered
}

func (impl *ClusterOverviewServiceImpl) filterNodeErrors(nodes []bean.NodeErrorDetail, searchKey string) []bean.NodeErrorDetail {
	if searchKey == "" {
		return nodes
	}

	searchKey = strings.ToLower(searchKey)
	var filtered []bean.NodeErrorDetail

	for _, node := range nodes {
		// Check if search key matches node name, cluster name, or node status
		if strings.Contains(strings.ToLower(node.NodeName), searchKey) ||
			strings.Contains(strings.ToLower(node.ClusterName), searchKey) ||
			strings.Contains(strings.ToLower(node.NodeStatus), searchKey) {
			filtered = append(filtered, node)
			continue
		}

		// Check if search key matches any error in the errors array
		for _, err := range node.Errors {
			if strings.Contains(strings.ToLower(err), searchKey) {
				filtered = append(filtered, node)
				break // Found match, no need to check other errors
			}
		}
	}

	return filtered
}

func (impl *ClusterOverviewServiceImpl) sortNodeErrors(nodes []bean.NodeErrorDetail, sortBy, sortOrder string) []bean.NodeErrorDetail {
	if sortBy == "" {
		sortBy = constants.SortFieldNodeName // default sort
	}
	if sortOrder == "" {
		sortOrder = constants.SortOrderAsc // default order
	}

	// Create a copy to avoid modifying the original
	sorted := make([]bean.NodeErrorDetail, len(nodes))
	copy(sorted, nodes)

	sort.Slice(sorted, func(i, j int) bool {
		var compareResult int

		switch sortBy {
		case constants.SortFieldNodeName:
			compareResult = strings.Compare(sorted[i].NodeName, sorted[j].NodeName)
		case constants.SortFieldClusterName:
			compareResult = strings.Compare(sorted[i].ClusterName, sorted[j].ClusterName)
		case constants.SortFieldNodeErrors:
			// Sort by joining the error array for comparison
			errorsI := strings.Join(sorted[i].Errors, ", ")
			errorsJ := strings.Join(sorted[j].Errors, ", ")
			compareResult = strings.Compare(errorsI, errorsJ)
		case constants.SortFieldNodeStatus:
			compareResult = strings.Compare(sorted[i].NodeStatus, sorted[j].NodeStatus)
		default:
			compareResult = strings.Compare(sorted[i].NodeName, sorted[j].NodeName)
		}

		if sortOrder == constants.SortOrderDesc {
			return compareResult > 0
		}
		return compareResult < 0
	})

	return sorted
}

func (impl *ClusterOverviewServiceImpl) paginateNodeErrors(nodes []bean.NodeErrorDetail, offset, limit int) []bean.NodeErrorDetail {
	if limit <= 0 {
		limit = 10 // default limit
	}

	start := offset
	if start < 0 {
		start = 0
	}
	if start >= len(nodes) {
		return []bean.NodeErrorDetail{}
	}

	end := start + limit
	if end > len(nodes) {
		end = len(nodes)
	}

	return nodes[start:end]
}

// Helper methods for node scheduling filtering, sorting, and pagination

func (impl *ClusterOverviewServiceImpl) filterNodeScheduling(nodes []bean.NodeSchedulingDetail, searchKey string) []bean.NodeSchedulingDetail {
	if searchKey == "" {
		return nodes
	}

	searchKey = strings.ToLower(searchKey)
	var filtered []bean.NodeSchedulingDetail

	for _, node := range nodes {
		schedulableStr := "schedulable"
		if !node.Schedulable {
			schedulableStr = "unschedulable"
		}

		if strings.Contains(strings.ToLower(node.NodeName), searchKey) ||
			strings.Contains(strings.ToLower(node.ClusterName), searchKey) ||
			strings.Contains(schedulableStr, searchKey) {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

func (impl *ClusterOverviewServiceImpl) sortNodeScheduling(nodes []bean.NodeSchedulingDetail, sortBy, sortOrder string) []bean.NodeSchedulingDetail {
	if sortBy == "" {
		sortBy = constants.SortFieldNodeName // default sort
	}
	if sortOrder == "" {
		sortOrder = constants.SortOrderAsc // default order
	}

	// Create a copy to avoid modifying the original
	sorted := make([]bean.NodeSchedulingDetail, len(nodes))
	copy(sorted, nodes)

	sort.Slice(sorted, func(i, j int) bool {
		var compareResult int

		switch sortBy {
		case constants.SortFieldNodeName:
			compareResult = strings.Compare(sorted[i].NodeName, sorted[j].NodeName)
		case constants.SortFieldClusterName:
			compareResult = strings.Compare(sorted[i].ClusterName, sorted[j].ClusterName)
		case constants.SortFieldSchedulable:
			// For boolean comparison: false < true
			if sorted[i].Schedulable == sorted[j].Schedulable {
				compareResult = 0
			} else if sorted[i].Schedulable {
				compareResult = 1
			} else {
				compareResult = -1
			}
		default:
			compareResult = strings.Compare(sorted[i].NodeName, sorted[j].NodeName)
		}

		if sortOrder == constants.SortOrderDesc {
			return compareResult > 0
		}
		return compareResult < 0
	})

	return sorted
}

func (impl *ClusterOverviewServiceImpl) paginateNodeScheduling(nodes []bean.NodeSchedulingDetail, offset, limit int) []bean.NodeSchedulingDetail {
	if limit <= 0 {
		limit = 10 // default limit
	}

	start := offset
	if start < 0 {
		start = 0
	}
	if start >= len(nodes) {
		return []bean.NodeSchedulingDetail{}
	}

	end := start + limit
	if end > len(nodes) {
		end = len(nodes)
	}

	return nodes[start:end]
}

// Helper methods for autoscaler node filtering, sorting, and pagination

func (impl *ClusterOverviewServiceImpl) filterAutoscalerNodes(nodes []bean.AutoscalerNodeDetail, searchKey string) []bean.AutoscalerNodeDetail {
	if searchKey == "" {
		return nodes
	}

	searchKey = strings.ToLower(searchKey)
	var filtered []bean.AutoscalerNodeDetail

	for _, node := range nodes {
		if strings.Contains(strings.ToLower(node.NodeName), searchKey) ||
			strings.Contains(strings.ToLower(node.ClusterName), searchKey) ||
			strings.Contains(strings.ToLower(node.ManagedBy), searchKey) {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

func (impl *ClusterOverviewServiceImpl) sortAutoscalerNodes(nodes []bean.AutoscalerNodeDetail, sortBy, sortOrder string) []bean.AutoscalerNodeDetail {
	if sortBy == "" {
		sortBy = constants.SortFieldNodeName // default sort
	}
	if sortOrder == "" {
		sortOrder = constants.SortOrderAsc // default order
	}

	// Create a copy to avoid modifying the original
	sorted := make([]bean.AutoscalerNodeDetail, len(nodes))
	copy(sorted, nodes)

	sort.Slice(sorted, func(i, j int) bool {
		var compareResult int

		switch sortBy {
		case constants.SortFieldNodeName:
			compareResult = strings.Compare(sorted[i].NodeName, sorted[j].NodeName)
		case constants.SortFieldClusterName:
			compareResult = strings.Compare(sorted[i].ClusterName, sorted[j].ClusterName)
		case constants.SortFieldAutoscalerType:
			compareResult = strings.Compare(sorted[i].ManagedBy, sorted[j].ManagedBy)
		default:
			compareResult = strings.Compare(sorted[i].NodeName, sorted[j].NodeName)
		}

		if sortOrder == constants.SortOrderDesc {
			return compareResult > 0
		}
		return compareResult < 0
	})

	return sorted
}

func (impl *ClusterOverviewServiceImpl) paginateAutoscalerNodes(nodes []bean.AutoscalerNodeDetail, offset, limit int) []bean.AutoscalerNodeDetail {
	if limit <= 0 {
		limit = 10 // default limit
	}

	start := offset
	if start < 0 {
		start = 0
	}
	if start >= len(nodes) {
		return []bean.AutoscalerNodeDetail{}
	}

	end := start + limit
	if end > len(nodes) {
		end = len(nodes)
	}

	return nodes[start:end]
}
