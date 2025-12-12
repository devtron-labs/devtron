/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

import (
	"time"

	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/devtron/pkg/overview/constants"
)

type BuildDeploymentActivityRequest struct {
	From *time.Time `json:"from"`
	To   *time.Time `json:"to"`
}

type ActivityKind string

const (
	ActivityKindBuildTrigger      ActivityKind = "buildTrigger"
	ActivityKindDeploymentTrigger ActivityKind = "deploymentTrigger"
	ActivityKindAvgBuildTime      ActivityKind = "avgBuildTime"
)

type BuildDeploymentActivityDetailedRequest struct {
	ActivityKind    ActivityKind              `json:"activityKind" validate:"required,oneof=buildTrigger deploymentTrigger avgBuildTime"`
	AggregationType constants.AggregationType `json:"aggregationType,omitempty"`
	From            *time.Time                `json:"from"`
	To              *time.Time                `json:"to"`
}

type AppMetrics struct {
	Total          int             `json:"total"`
	YourApps       *AppTypeMetrics `json:"yourApps"`
	ThirdPartyApps *AppTypeMetrics `json:"thirdPartyApps"`
}

type PipelineMetrics struct {
	Total         int `json:"total"`
	Production    int `json:"production"`
	NonProduction int `json:"nonProduction"`
}

// Common structure for entity metadata
type EntityMetadata struct {
	Name      string    `json:"name"`
	CreatedOn time.Time `json:"createdOn"`
}

// Time-based aggregated data point
type TimeDataPoint struct {
	Date  string `json:"date"`  // YYYY-MM-DD format for days, YYYY-MM-DD HH:00 format for hours
	Count int    `json:"count"` // Aggregated count for this time period
}

// Enhanced metrics structures with detailed metadata
type ProjectMetrics struct {
	Total   int              `json:"total"`
	Details []EntityMetadata `json:"details"`
}

type AppTypeMetrics struct {
	Total   int              `json:"total"`
	Details []EntityMetadata `json:"details"`
}

type EnvironmentMetrics struct {
	Total   int              `json:"total"`
	Details []EntityMetadata `json:"details"`
}

type BuildPipelineMetrics struct {
	Total               int                    `json:"total"`
	NormalCiPipelines   *CiPipelineTypeMetrics `json:"normalCiPipelines"`
	ExternalCiPipelines *CiPipelineTypeMetrics `json:"externalCiPipelines"`
}

type CiPipelineTypeMetrics struct {
	Total   int              `json:"total"`
	Details []EntityMetadata `json:"details"`
}

type CdPipelineMetrics struct {
	Total         int                         `json:"total"`
	Production    *PipelineEnvironmentMetrics `json:"production"`
	NonProduction *PipelineEnvironmentMetrics `json:"nonProduction"`
}

type PipelineEnvironmentMetrics struct {
	Total   int              `json:"total"`
	Details []EntityMetadata `json:"details"`
}

type DeploymentMetrics struct {
	Total   int              `json:"total"`
	Details []EntityMetadata `json:"details"`
}

// Trend-based metrics structures for aggregated time-series data
type ProjectTrendMetrics struct {
	Total int             `json:"total"`
	Trend []TimeDataPoint `json:"trend"`
}

type AppTrendMetrics struct {
	Total          int                  `json:"total"`
	YourApps       *AppTypeTrendMetrics `json:"yourApps"`
	ThirdPartyApps *AppTypeTrendMetrics `json:"thirdPartyApps"`
}

type AppTypeTrendMetrics struct {
	Total int             `json:"total"`
	Trend []TimeDataPoint `json:"trend"`
}

type EnvironmentTrendMetrics struct {
	Total int             `json:"total"`
	Trend []TimeDataPoint `json:"trend"`
}

type BuildPipelineTrendMetrics struct {
	Total               int                         `json:"total"`
	NormalCiPipelines   *CiPipelineTypeTrendMetrics `json:"normalCiPipelines"`
	ExternalCiPipelines *CiPipelineTypeTrendMetrics `json:"externalCiPipelines"`
}

type CiPipelineTypeTrendMetrics struct {
	Total int             `json:"total"`
	Trend []TimeDataPoint `json:"trend"`
}

type CdPipelineTrendMetrics struct {
	Total         int                              `json:"total"`
	Production    *PipelineEnvironmentTrendMetrics `json:"production"`
	NonProduction *PipelineEnvironmentTrendMetrics `json:"nonProduction"`
}

type PipelineEnvironmentTrendMetrics struct {
	Total int             `json:"total"`
	Trend []TimeDataPoint `json:"trend"`
}

type TrendComparison struct {
	Value int    `json:"value"` // The difference value (can be positive or negative)
	Label string `json:"label"` // e.g., "this month", "this week", "this quarter"
}

type AppsOverviewResponse struct {
	Projects         *AtAGlanceMetric `json:"projects"`
	YourApplications *AtAGlanceMetric `json:"yourApplications"`
	HelmApplications *AtAGlanceMetric `json:"helmApplications"`
	Environments     *AtAGlanceMetric `json:"environments"`
}

type WorkflowOverviewResponse struct {
	BuildPipelines                *AtAGlanceMetric `json:"buildPipelines"`
	ExternalImageSource           *AtAGlanceMetric `json:"externalImageSource"`
	AllDeploymentPipelines        *AtAGlanceMetric `json:"allDeploymentPipelines"`
	ScanningEnabledInWorkflows    *AtAGlanceMetric `json:"scanningEnabledInWorkflows"`
	GitOpsComplianceProdPipelines *AtAGlanceMetric `json:"gitOpsComplianceProdPipelines"`
	ProductionPipelines           *AtAGlanceMetric `json:"productionPipelines"`
}

type AtAGlanceMetric struct {
	Total      int     `json:"total"`
	Percentage float64 `json:"percentage,omitempty"` // Optional: percentage value for metrics that represent percentages
}

type BuildDeploymentActivityResponse struct {
	TotalBuildTriggers      int     `json:"totalBuildTriggers"`
	AverageBuildTime        float64 `json:"averageBuildTime"` // in minutes
	TotalDeploymentTriggers int     `json:"totalDeploymentTriggers"`
}

type BuildDeploymentActivityDetailedResponse struct {
	ActivityKind            ActivityKind                `json:"activityKind"`    // Type of activity data returned
	AggregationType         constants.AggregationType   `json:"aggregationType"` // HOURLY, DAILY, or MONTHLY
	BuildTriggersTrend      []BuildStatusDataPoint      `json:"buildTriggersTrend,omitempty"`
	DeploymentTriggersTrend []DeploymentStatusDataPoint `json:"deploymentTriggersTrend,omitempty"`
	AvgBuildTimeTrend       []BuildTimeDataPoint        `json:"avgBuildTimeTrend,omitempty"`
}

type BuildStatusDataPoint struct {
	Timestamp  time.Time `json:"timestamp"`  // Timestamp representing start of aggregation period
	Total      int       `json:"total"`      // Total build triggers
	Successful int       `json:"successful"` // Successful builds
	Failed     int       `json:"failed"`     // Failed builds
}

type DeploymentStatusDataPoint struct {
	Timestamp  time.Time `json:"timestamp"`  // Timestamp representing start of aggregation period
	Total      int       `json:"total"`      // Total deployment triggers
	Successful int       `json:"successful"` // Successful deployments
	Failed     int       `json:"failed"`     // Failed deployments
}

type BuildTimeDataPoint struct {
	Timestamp        time.Time `json:"timestamp"`        // Timestamp representing start of aggregation period
	AverageBuildTime float64   `json:"averageBuildTime"` // in minutes for that time period
}

// Insights Beans
type PipelineType string

const (
	BuildPipelines      PipelineType = "buildPipelines"
	DeploymentPipelines PipelineType = "deploymentPipelines"
)

type SortOrder string

const (
	ASC  SortOrder = "ASC"
	DESC SortOrder = "DESC"
)

type InsightsRequest struct {
	TimeRangeRequest *utils.TimeRangeRequest `json:"timeRangeRequest"`
	PipelineType     PipelineType            `json:"pipelineType"`
	SortOrder        SortOrder               `json:"sortOrder"`
	Limit            int                     `json:"limit"`
	Offset           int                     `json:"offset"`
}

type InsightsResponse struct {
	Pipelines  []PipelineUsageItem `json:"pipelines"`
	TotalCount int                 `json:"totalCount"`
}

type PipelineUsageItem struct {
	AppID        int    `json:"appId"`           // Required for both CI and CD pipelines
	EnvID        int    `json:"envId,omitempty"` // Only for deployment pipelines
	PipelineID   int    `json:"pipelineId"`
	PipelineName string `json:"pipelineName"`
	AppName      string `json:"appName"`
	EnvName      string `json:"envName,omitempty"` // Only for deployment pipelines
	TriggerCount int    `json:"triggerCount"`
}

type ApprovalPolicyOverviewResponse struct {
	TotalProdPipelineCount              int `json:"totalProdPipelineCount"`
	PipelineCountWithConfigApproval     int `json:"pipelineCountWithConfigApproval"`
	PipelineCountWithDeploymentApproval int `json:"pipelineCountWithDeploymentApproval"`
}

// Cluster Management Overview Beans

// ClusterOverviewRequest represents the request for cluster management overview
type ClusterOverviewRequest struct {
	// No specific filters needed for now - returns all cluster data
}

// ClusterOverviewResponse represents the comprehensive cluster management overview
type ClusterOverviewResponse struct {
	TotalClusters               int                           `json:"totalClusters"`
	TotalCpuCapacity            *ResourceCapacity             `json:"totalCpuCapacity"`
	TotalMemoryCapacity         *ResourceCapacity             `json:"totalMemoryCapacity"`
	ClusterStatusBreakdown      *ClusterStatusBreakdown       `json:"clusterStatusBreakdown"`
	NodeSchedulingBreakdown     *NodeSchedulingBreakdown      `json:"nodeSchedulingBreakdown"`
	NodeErrorBreakdown          *NodeErrorBreakdown           `json:"nodeErrorBreakdown"`
	ClusterDistribution         *ClusterDistribution          `json:"clusterDistribution"`
	ClusterCapacityDistribution []ClusterCapacityDistribution `json:"clusterCapacityDistribution"`
	NodeDistribution            *NodeDistribution             `json:"nodeDistribution"`
}

// ResourceCapacity represents capacity with value and unit
type ResourceCapacity struct {
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

// ClusterStatusBreakdown represents cluster health status breakdown
type ClusterStatusBreakdown struct {
	Healthy          int `json:"healthy"`
	Unhealthy        int `json:"unhealthy"`
	ConnectionFailed int `json:"connectionFailed"`
}

// NodeErrorBreakdown represents breakdown of node errors with detailed node information
type NodeErrorBreakdown struct {
	ErrorCounts map[string]int    `json:"errorCounts"` // Map of error types to their counts
	Total       int               `json:"total"`       // Total number of node errors
	NodeErrors  []NodeErrorDetail `json:"nodeErrors"`  // Detailed list of nodes with errors
}

// NodeErrorDetail represents detailed error information for a single node
type NodeErrorDetail struct {
	NodeName    string   `json:"nodeName"`    // Name of the node with errors
	ClusterName string   `json:"clusterName"` // Name of the cluster the node belongs to
	ClusterID   int      `json:"clusterId"`   // ID of the cluster
	Errors      []string `json:"errors"`      // List of error types
	NodeStatus  string   `json:"nodeStatus"`  // Current status of the node (Ready/Not Ready)
}

// NodeSchedulingBreakdown represents breakdown of node scheduling status with detailed node information
type NodeSchedulingBreakdown struct {
	Schedulable        int                    `json:"schedulable"`        // Count of schedulable nodes
	Unschedulable      int                    `json:"unschedulable"`      // Count of unschedulable nodes
	Total              int                    `json:"total"`              // Total number of nodes
	SchedulableNodes   []NodeSchedulingDetail `json:"schedulableNodes"`   // Detailed list of schedulable nodes
	UnschedulableNodes []NodeSchedulingDetail `json:"unschedulableNodes"` // Detailed list of unschedulable nodes
}

// NodeSchedulingDetail represents detailed information about a node's scheduling status
type NodeSchedulingDetail struct {
	NodeName    string `json:"nodeName"`    // Name of the node
	ClusterName string `json:"clusterName"` // Name of the cluster the node belongs to
	ClusterID   int    `json:"clusterId"`   // ID of the cluster
	Schedulable bool   `json:"schedulable"` // Whether the node is schedulable
}

// ClusterDistribution represents cluster distribution by provider and cluster version
type ClusterDistribution struct {
	ByProvider []ProviderDistribution `json:"byProvider"`
	ByVersion  []VersionDistribution  `json:"byVersion"`
}

// ProviderDistribution represents cluster count by cloud provider
type ProviderDistribution struct {
	Provider string `json:"provider"` // AWS, GCP, Azure, On-Premise, etc.
	Count    int    `json:"count"`
}

// VersionDistribution represents cluster count by Kubernetes version (major.minor only)
type VersionDistribution struct {
	Version string `json:"version"` // e.g., "1.28", "1.29", "1.30" (major.minor only, patch ignored)
	Count   int    `json:"count"`
}

// ClusterCapacityDistribution represents capacity distribution for individual clusters
type ClusterCapacityDistribution struct {
	ClusterID     int                    `json:"clusterId"`
	ClusterName   string                 `json:"clusterName"`
	ServerVersion string                 `json:"serverVersion"` // Kubernetes server version (e.g., "v1.28.3")
	CPU           *ClusterResourceMetric `json:"cpu"`
	Memory        *ClusterResourceMetric `json:"memory"`
}

// ClusterResourceMetric represents resource metrics for a cluster
type ClusterResourceMetric struct {
	Capacity           float64 `json:"capacity"`           // Capacity in cores for CPU, Gi for memory (with decimal precision)
	UtilizationPercent float64 `json:"utilizationPercent"` // Utilization percentage
	RequestsPercent    float64 `json:"requestsPercent"`    // Requests percentage
	LimitsPercent      float64 `json:"limitsPercent"`      // Limits percentage
}

// NodeDistribution represents node distribution by clusters and autoscaler
type NodeDistribution struct {
	ByClusters   []ClusterNodeCount    `json:"byClusters"`   // Node count grouped by cluster
	ByAutoscaler []AutoscalerNodeCount `json:"byAutoscaler"` // Node count grouped by autoscaler type
}

// Removed old structs - ClusterSummary, ResourceSummary, NodeCountSummary not needed in new API spec

// ClusterNodeCount represents node count for a specific cluster
type ClusterNodeCount struct {
	ClusterID   int    `json:"clusterId"`   // ID of the cluster
	ClusterName string `json:"clusterName"` // Name of the cluster
	NodeCount   int    `json:"nodeCount"`   // Total number of nodes in this cluster
}

// AutoscalerNodeCount represents node count for a specific autoscaler type with detailed node information
type AutoscalerNodeCount struct {
	AutoscalerType string                 `json:"autoscalerType"` // Type of autoscaler (EKS, Karpenter, Cast AI, GKE, CAS)
	NodeCount      int                    `json:"nodeCount"`      // Total number of nodes managed by this autoscaler
	NodeDetails    []AutoscalerNodeDetail `json:"nodeDetails"`    // Detailed list of nodes managed by this autoscaler
}

// AutoscalerNodeDetail represents detailed information for a single node managed by autoscaler
type AutoscalerNodeDetail struct {
	NodeName    string `json:"nodeName"`    // Name of the node
	ClusterName string `json:"clusterName"` // Name of the cluster the node belongs to
	ClusterID   int    `json:"clusterId"`   // ID of the cluster
	ManagedBy   string `json:"managedBy"`   // Display name of the autoscaler managing this node
}

// Cluster Upgrade Overview Beans

// ClusterUpgradeOverviewResponse represents the response for cluster upgrade overview
type ClusterUpgradeOverviewResponse struct {
	CanCurrentUserUpgrade bool                    `json:"canCurrentUserUpgrade"`
	LatestVersion         string                  `json:"latestVersion"`
	ClusterList           []ClusterUpgradeDetails `json:"clusterList"`
}

// ClusterUpgradeDetails represents upgrade details for a single cluster
type ClusterUpgradeDetails struct {
	ClusterId      int      `json:"clusterId"`
	ClusterName    string   `json:"clusterName"`
	CurrentVersion string   `json:"currentVersion"`
	UpgradePath    []string `json:"upgradePath"`
}

// NodeViewGroupType represents the type of node view grouping
type NodeViewGroupType string

const (
	NodeViewGroupTypeNodeErrors     NodeViewGroupType = "nodeErrors"
	NodeViewGroupTypeNodeScheduling NodeViewGroupType = "nodeScheduling"
	NodeViewGroupTypeAutoscaler     NodeViewGroupType = "autoscalerManaged"
)

// ClusterOverviewDetailRequest represents request parameters for detailed drill-down API
type ClusterOverviewDetailRequest struct {
	GroupBy   NodeViewGroupType `schema:"groupBy" validate:"required,oneof=nodeErrors nodeScheduling autoscalerManaged"`
	Offset    int               `schema:"offset"`
	Limit     int               `schema:"limit"`
	SortBy    string            `schema:"sortBy"`
	SortOrder string            `schema:"sortOrder"` // asc or desc
	SearchKey string            `schema:"searchKey"`

	// Filter parameters (optional, used based on GroupBy)
	AutoscalerType  string `schema:"autoscalerType"`  // Filter by autoscaler type (only for autoscalerManaged groupBy)
	ErrorType       string `schema:"errorType"`       // Filter by error type (only for nodeErrors groupBy)
	SchedulableType string `schema:"schedulableType"` // Filter by schedulable type: "schedulable" or "unschedulable" (only for nodeScheduling groupBy)
}

// ClusterOverviewNodeDetailedResponse represents the unified response for all node view group types
// Fields are conditionally included based on the groupBy parameter
type ClusterOverviewNodeDetailedResponse struct {
	TotalCount int                               `json:"totalCount"`
	NodeList   []ClusterOverviewNodeDetailedItem `json:"nodeList"`
}

// ClusterOverviewNodeDetailedItem represents a single node item in the detailed response
// Different fields are populated based on the NodeViewGroupType
type ClusterOverviewNodeDetailedItem struct {
	// Common fields (always present)
	NodeName    string `json:"nodeName"`
	ClusterName string `json:"clusterName"`
	ClusterID   int    `json:"clusterId,omitempty"`

	// NodeErrors specific fields
	NodeErrors []string `json:"nodeErrors,omitempty"` // List of error types (only for nodeErrors type)
	NodeStatus string   `json:"nodeStatus,omitempty"` // Node status: Ready/Not Ready (only for nodeErrors type)

	// NodeScheduling specific fields
	Schedulable bool `json:"schedulable,omitempty"` // Whether node is schedulable (only for nodeScheduling type)

	// Autoscaler specific fields
	AutoscalerType string `json:"autoscalerType,omitempty"` // Type of autoscaler managing the node (only for autoscaler type)
}
