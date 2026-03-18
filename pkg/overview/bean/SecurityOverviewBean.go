/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

import "time"

// ============================================================================
// Common Types
// ============================================================================

type EnvType string

const (
	EnvTypeProd    EnvType = "prod"
	EnvTypeNonProd EnvType = "non-prod"
	EnvTypeAll     EnvType = "all"
)

type VulnerabilityCount struct {
	Count       int `json:"count"`       // Total instances (with duplicates)
	UniqueCount int `json:"uniqueCount"` // Unique CVEs
}

type SeverityCount struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Unknown  int `json:"unknown"`
}

type AgeBucketSeverity struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Unknown  int `json:"unknown"`
}

type AgeDistribution struct {
	LessThan30Days    *AgeBucketSeverity `json:"lessThan30Days"`
	Between30To60Days *AgeBucketSeverity `json:"between30To60Days"`
	Between60To90Days *AgeBucketSeverity `json:"between60To90Days"`
	MoreThan90Days    *AgeBucketSeverity `json:"moreThan90Days"`
}

// ============================================================================
// 1. Security Overview API (At a Glance - Organization-wide)
// ============================================================================

type SecurityOverviewRequest struct {
	EnvIds     []int `json:"envIds" schema:"envIds"`
	ClusterIds []int `json:"clusterIds" schema:"clusterIds"`
	AppIds     []int `json:"appIds" schema:"appIds"`
}

type SecurityOverviewResponse struct {
	TotalVulnerabilities   *VulnerabilityCount `json:"totalVulnerabilities"`
	FixableVulnerabilities *VulnerabilityCount `json:"fixableVulnerabilities"`
	ZeroDayVulnerabilities *VulnerabilityCount `json:"zeroDayVulnerabilities"`
}

// ============================================================================
// 2. Severity Insights API (With Prod/Non-Prod Filtering)
// ============================================================================

type SeverityInsightsRequest struct {
	EnvIds     []int   `json:"envIds" schema:"envIds"`
	ClusterIds []int   `json:"clusterIds" schema:"clusterIds"`
	AppIds     []int   `json:"appIds" schema:"appIds"`
	EnvType    EnvType `json:"envType" schema:"envType" validate:"required,oneof=prod non-prod all"`
}

type SeverityInsightsResponse struct {
	SeverityDistribution *SeverityCount   `json:"severityDistribution"`
	AgeDistribution      *AgeDistribution `json:"ageDistribution"`
}

// ============================================================================
// 3. Deployment Security Status API
// ============================================================================

type DeploymentSecurityStatusRequest struct {
	EnvIds     []int `json:"envIds" schema:"envIds"`
	ClusterIds []int `json:"clusterIds" schema:"clusterIds"`
	AppIds     []int `json:"appIds" schema:"appIds"`
}

type DeploymentMetric struct {
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type WorkflowMetric struct {
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type DeploymentSecurityStatusResponse struct {
	ActiveDeploymentsWithVulnerabilities *DeploymentMetric `json:"activeDeploymentsWithVulnerabilities"`
	ActiveDeploymentsWithUnscannedImages *DeploymentMetric `json:"activeDeploymentsWithUnscannedImages"`
	WorkflowsWithScanningEnabled         *WorkflowMetric   `json:"workflowsWithScanningEnabled"`
}

// ============================================================================
// 4. Vulnerability Details API (Paginated List)
// ============================================================================

type VulnerabilitiesRequest struct {
	EnvIds     []int  `json:"envIds" schema:"envIds"`
	ClusterIds []int  `json:"clusterIds" schema:"clusterIds"`
	AppIds     []int  `json:"appIds" schema:"appIds"`
	Severity   string `json:"severity" schema:"severity"` // Optional: critical, high, medium, low, unknown
	Offset     int    `json:"offset" schema:"offset"`
	Size       int    `json:"size" schema:"size" validate:"required,min=1,max=100"`
}

type Vulnerability struct {
	CveName          string    `json:"cveName"`
	Severity         string    `json:"severity"`
	Package          string    `json:"package"`
	CurrentVersion   string    `json:"currentVersion"`
	FixedVersion     string    `json:"fixedVersion"`
	AppCount         int       `json:"appCount"`         // Number of apps affected
	EnvironmentCount int       `json:"environmentCount"` // Number of environments affected
	FirstDetected    time.Time `json:"firstDetected"`
}

type VulnerabilitiesResponse struct {
	Vulnerabilities []*Vulnerability `json:"vulnerabilities"`
	Total           int              `json:"total"`
	Offset          int              `json:"offset"`
	Size            int              `json:"size"`
}

// ============================================================================
// 5. Vulnerability Trend API (Time-series with Prod/Non-Prod Filtering)
// ============================================================================

type VulnerabilityTrendRequest struct {
	TimeWindow string     `json:"timeWindow" schema:"timeWindow" validate:"required,oneof=today thisWeek thisMonth thisQuarter"`
	EnvType    EnvType    `json:"envType" schema:"envType" validate:"required,oneof=prod non-prod all"`
	From       *time.Time `json:"from" schema:"from"`
	To         *time.Time `json:"to" schema:"to"`
}

type VulnerabilityTrendDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Critical  int       `json:"critical"`
	High      int       `json:"high"`
	Medium    int       `json:"medium"`
	Low       int       `json:"low"`
	Unknown   int       `json:"unknown"`
	Total     int       `json:"total"`
}

type VulnerabilityTrendResponse struct {
	Trend []*VulnerabilityTrendDataPoint `json:"trend"`
}

// ============================================================================
// 6. Blocked Deployments Trend API (Organization-wide)
// ============================================================================

type BlockedDeploymentsTrendRequest struct {
	TimeWindow string     `json:"timeWindow" schema:"timeWindow" validate:"required,oneof=today thisWeek thisMonth thisQuarter"`
	From       *time.Time `json:"from" schema:"from"`
	To         *time.Time `json:"to" schema:"to"`
}

type BlockedDeploymentDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

type BlockedDeploymentsTrendResponse struct {
	Trend []*BlockedDeploymentDataPoint `json:"trend"`
}
