/*
 * Copyright (c) 2024. Devtron Inc.
 */

package adaptor

import "github.com/devtron-labs/devtron/pkg/overview/bean"

// SecurityOverviewAdapter provides factory methods for initializing security overview bean structs

// NewSeverityCount returns a new initialized SeverityCount with all fields set to zero
func NewSeverityCount() *bean.SeverityCount {
	return &bean.SeverityCount{
		Critical: 0,
		High:     0,
		Medium:   0,
		Low:      0,
		Unknown:  0,
	}
}

// NewAgeBucketSeverity returns a new initialized AgeBucketSeverity with all fields set to zero
func NewAgeBucketSeverity() *bean.AgeBucketSeverity {
	return &bean.AgeBucketSeverity{
		Critical: 0,
		High:     0,
		Medium:   0,
		Low:      0,
		Unknown:  0,
	}
}

// NewAgeDistribution returns a new initialized AgeDistribution with all nested structs initialized
func NewAgeDistribution() *bean.AgeDistribution {
	return &bean.AgeDistribution{
		LessThan30Days:    NewAgeBucketSeverity(),
		Between30To60Days: NewAgeBucketSeverity(),
		Between60To90Days: NewAgeBucketSeverity(),
		MoreThan90Days:    NewAgeBucketSeverity(),
	}
}

// NewVulnerabilityCount returns a new initialized VulnerabilityCount with all fields set to zero
func NewVulnerabilityCount() *bean.VulnerabilityCount {
	return &bean.VulnerabilityCount{
		Count:       0,
		UniqueCount: 0,
	}
}

// NewSecurityOverviewResponse returns a new initialized SecurityOverviewResponse with all nested structs initialized
func NewSecurityOverviewResponse() *bean.SecurityOverviewResponse {
	return &bean.SecurityOverviewResponse{
		TotalVulnerabilities:   NewVulnerabilityCount(),
		FixableVulnerabilities: NewVulnerabilityCount(),
		ZeroDayVulnerabilities: NewVulnerabilityCount(),
	}
}

// NewSeverityInsightsResponse returns a new initialized SeverityInsightsResponse with all nested structs initialized
func NewSeverityInsightsResponse() *bean.SeverityInsightsResponse {
	return &bean.SeverityInsightsResponse{
		SeverityDistribution: NewSeverityCount(),
		AgeDistribution:      NewAgeDistribution(),
	}
}

// NewDeploymentMetric returns a new initialized DeploymentMetric with all fields set to zero
func NewDeploymentMetric() *bean.DeploymentMetric {
	return &bean.DeploymentMetric{
		Count:      0,
		Percentage: 0.0,
	}
}

// NewWorkflowMetric returns a new initialized WorkflowMetric with all fields set to zero
func NewWorkflowMetric() *bean.WorkflowMetric {
	return &bean.WorkflowMetric{
		Count:      0,
		Percentage: 0.0,
	}
}

// NewDeploymentSecurityStatusResponse returns a new initialized DeploymentSecurityStatusResponse with all nested structs initialized
func NewDeploymentSecurityStatusResponse() *bean.DeploymentSecurityStatusResponse {
	return &bean.DeploymentSecurityStatusResponse{
		ActiveDeploymentsWithVulnerabilities: NewDeploymentMetric(),
		ActiveDeploymentsWithUnscannedImages: NewDeploymentMetric(),
		WorkflowsWithScanningEnabled:         NewWorkflowMetric(),
	}
}

// NewVulnerabilitiesResponse returns a new initialized VulnerabilitiesResponse with empty slice and pagination info
func NewVulnerabilitiesResponse(offset, size int) *bean.VulnerabilitiesResponse {
	return &bean.VulnerabilitiesResponse{
		Vulnerabilities: []*bean.Vulnerability{},
		Total:           0,
		Offset:          offset,
		Size:            size,
	}
}

// NewVulnerabilityTrendResponse returns a new initialized VulnerabilityTrendResponse with empty trend slice
func NewVulnerabilityTrendResponse() *bean.VulnerabilityTrendResponse {
	return &bean.VulnerabilityTrendResponse{
		Trend: []*bean.VulnerabilityTrendDataPoint{},
	}
}

// NewBlockedDeploymentsTrendResponse returns a new initialized BlockedDeploymentsTrendResponse with empty trend slice
func NewBlockedDeploymentsTrendResponse() *bean.BlockedDeploymentsTrendResponse {
	return &bean.BlockedDeploymentsTrendResponse{
		Trend: []*bean.BlockedDeploymentDataPoint{},
	}
}
