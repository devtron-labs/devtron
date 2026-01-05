/*
 * Copyright (c) 2024. Devtron Inc.
 */

package overview

import (
	"context"
	"fmt"
	"time"

	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/overview/adaptor"
	"github.com/devtron-labs/devtron/pkg/overview/bean"
	"github.com/devtron-labs/devtron/pkg/overview/constants"
	"github.com/devtron-labs/devtron/pkg/overview/util"
	imageScanRepo "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository"
	scanBean "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository/bean"
	"go.uber.org/zap"
)

type SecurityOverviewService interface {
	// 1. Security Overview API - "At a Glance" metrics (organization-wide)
	GetSecurityOverview(ctx context.Context, request *bean.SecurityOverviewRequest) (*bean.SecurityOverviewResponse, error)

	// 2. Severity Insights API - With prod/non-prod filtering
	GetSeverityInsights(ctx context.Context, request *bean.SeverityInsightsRequest) (*bean.SeverityInsightsResponse, error)

	// 3. Deployment Security Status API
	GetDeploymentSecurityStatus(ctx context.Context, request *bean.DeploymentSecurityStatusRequest) (*bean.DeploymentSecurityStatusResponse, error)

	// 5. Vulnerability Trend API - Time-series with prod/non-prod filtering
	GetVulnerabilityTrend(ctx context.Context, currentTimeRange *utils.TimeRangeRequest, envType bean.EnvType, aggregationType constants.AggregationType) (*bean.VulnerabilityTrendResponse, error)

	// 6. Blocked Deployments Trend API - Organization-wide
	GetBlockedDeploymentsTrend(ctx context.Context, currentTimeRange *utils.TimeRangeRequest, aggregationType constants.AggregationType) (*bean.BlockedDeploymentsTrendResponse, error)
}

type SecurityOverviewServiceImpl struct {
	logger                        *zap.SugaredLogger
	imageScanResultRepository     imageScanRepo.ImageScanResultRepository
	imageScanDeployInfoRepository imageScanRepo.ImageScanDeployInfoRepository
	cveStoreRepository            imageScanRepo.CveStoreRepository
	ciPipelineRepository          pipelineConfig.CiPipelineRepository
	cdWorkflowRepository          pipelineConfig.CdWorkflowRepository
}

func NewSecurityOverviewServiceImpl(
	logger *zap.SugaredLogger,
	imageScanResultRepository imageScanRepo.ImageScanResultRepository,
	imageScanDeployInfoRepository imageScanRepo.ImageScanDeployInfoRepository,
	cveStoreRepository imageScanRepo.CveStoreRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
) *SecurityOverviewServiceImpl {
	return &SecurityOverviewServiceImpl{
		logger:                        logger,
		imageScanResultRepository:     imageScanResultRepository,
		imageScanDeployInfoRepository: imageScanDeployInfoRepository,
		cveStoreRepository:            cveStoreRepository,
		ciPipelineRepository:          ciPipelineRepository,
		cdWorkflowRepository:          cdWorkflowRepository,
	}
}

func (service *SecurityOverviewServiceImpl) GetSecurityOverview(ctx context.Context, request *bean.SecurityOverviewRequest) (*bean.SecurityOverviewResponse, error) {
	service.logger.Infow("GetSecurityOverview called", "request", request)

	// Fetch all vulnerabilities with fixed_version in a single query
	vulnerabilities, err := service.imageScanResultRepository.GetVulnerabilityRawData("", nil, request.EnvIds, request.ClusterIds, request.AppIds, nil)
	if err != nil {
		service.logger.Errorw("error fetching vulnerabilities", "err", err)
		return nil, fmt.Errorf("failed to fetch vulnerabilities: %w", err)
	}

	// Calculate counts in application code
	totalCount := len(vulnerabilities)
	fixableCount := 0
	zeroDayCount := 0

	uniqueCVEs := make(map[string]bool)
	uniqueFixableCVEs := make(map[string]bool)
	uniqueZeroDayCVEs := make(map[string]bool)

	for _, vuln := range vulnerabilities {
		// Track unique CVEs
		uniqueCVEs[vuln.CveStoreName] = true

		// Check if fixable (has fixed_version)
		if vuln.FixedVersion != "" {
			fixableCount++
			uniqueFixableCVEs[vuln.CveStoreName] = true
		} else {
			// Zero-day (no fixed_version)
			zeroDayCount++
			uniqueZeroDayCVEs[vuln.CveStoreName] = true
		}
	}

	response := &bean.SecurityOverviewResponse{
		TotalVulnerabilities: &bean.VulnerabilityCount{
			Count:       totalCount,
			UniqueCount: len(uniqueCVEs),
		},
		FixableVulnerabilities: &bean.VulnerabilityCount{
			Count:       fixableCount,
			UniqueCount: len(uniqueFixableCVEs),
		},
		ZeroDayVulnerabilities: &bean.VulnerabilityCount{
			Count:       zeroDayCount,
			UniqueCount: len(uniqueZeroDayCVEs),
		},
	}

	return response, nil
}

func (service *SecurityOverviewServiceImpl) GetSeverityInsights(ctx context.Context, request *bean.SeverityInsightsRequest) (*bean.SeverityInsightsResponse, error) {
	service.logger.Infow("GetSeverityInsights called", "request", request)

	// Determine environment type filter
	// nil = all environments, true = prod only, false = non-prod only
	var isProd *bool
	if request.EnvType == bean.EnvTypeProd {
		prodValue := true
		isProd = &prodValue
	} else if request.EnvType == bean.EnvTypeNonProd {
		nonProdValue := false
		isProd = &nonProdValue
	}
	// If EnvType is "all", isProd remains nil

	// Fetch all vulnerability data with severity and execution time in a single query
	vulnerabilities, err := service.imageScanResultRepository.GetSeverityInsightDataByFilters(request.EnvIds, request.ClusterIds, request.AppIds, isProd)
	if err != nil {
		service.logger.Errorw("error fetching severity insight data", "err", err)
		return nil, fmt.Errorf("failed to fetch severity insight data: %w", err)
	}

	// Initialize counters using adapter
	severityCount := adaptor.NewSeverityCount()
	ageDistribution := adaptor.NewAgeDistribution()

	// Current time for age calculation
	now := time.Now()

	// Process vulnerabilities in a single pass
	for _, vuln := range vulnerabilities {
		severity := scanBean.Severity(vuln.Severity)

		// Count by severity
		switch severity {
		case scanBean.Critical:
			severityCount.Critical++
		case scanBean.High:
			severityCount.High++
		case scanBean.Medium:
			severityCount.Medium++
		case scanBean.Low:
			severityCount.Low++
		default:
			severityCount.Unknown++
		}

		// Calculate age in days
		age := now.Sub(vuln.ExecutionTime).Hours() / 24

		// Count by age bucket AND severity
		var ageBucket *bean.AgeBucketSeverity
		if age < 30 {
			ageBucket = ageDistribution.LessThan30Days
		} else if age < 60 {
			ageBucket = ageDistribution.Between30To60Days
		} else if age < 90 {
			ageBucket = ageDistribution.Between60To90Days
		} else {
			ageBucket = ageDistribution.MoreThan90Days
		}

		// Increment severity count within the age bucket
		switch severity {
		case scanBean.Critical:
			ageBucket.Critical++
		case scanBean.High:
			ageBucket.High++
		case scanBean.Medium:
			ageBucket.Medium++
		case scanBean.Low:
			ageBucket.Low++
		default:
			ageBucket.Unknown++
		}
	}

	response := &bean.SeverityInsightsResponse{
		SeverityDistribution: severityCount,
		AgeDistribution:      ageDistribution,
	}

	return response, nil
}

func (service *SecurityOverviewServiceImpl) GetDeploymentSecurityStatus(ctx context.Context, request *bean.DeploymentSecurityStatusRequest) (*bean.DeploymentSecurityStatusResponse, error) {
	service.logger.Infow("GetDeploymentSecurityStatus called", "request", request)

	// Get total active deployments count
	totalDeployments, err := service.imageScanDeployInfoRepository.GetActiveDeploymentCountByFilters(request.EnvIds, request.ClusterIds, request.AppIds)
	if err != nil {
		service.logger.Errorw("error getting total active deployments count", "err", err)
		return nil, fmt.Errorf("failed to get total active deployments count: %w", err)
	}

	// Get deployments with vulnerabilities count
	deploymentsWithVulnerabilities, err := service.imageScanDeployInfoRepository.GetActiveDeploymentCountWithVulnerabilitiesByFilters(request.EnvIds, request.ClusterIds, request.AppIds)
	if err != nil {
		service.logger.Errorw("error getting deployments with vulnerabilities count", "err", err)
		return nil, fmt.Errorf("failed to get deployments with vulnerabilities count: %w", err)
	}

	// Get scanned and unscanned deployment counts in a single optimized query
	scannedCounts, err := service.imageScanDeployInfoRepository.GetActiveDeploymentScannedUnscannedCountByFilters(request.EnvIds, request.ClusterIds, request.AppIds)
	if err != nil {
		service.logger.Errorw("error getting scanned/unscanned deployment counts", "err", err)
		return nil, fmt.Errorf("failed to get scanned/unscanned deployment counts: %w", err)
	}

	// Get total CI pipelines count (workflows)
	totalCiPipelines, err := service.ciPipelineRepository.GetActiveCiPipelineCount()
	if err != nil {
		service.logger.Errorw("error getting total CI pipelines count", "err", err)
		return nil, fmt.Errorf("failed to get total CI pipelines count: %w", err)
	}

	// Get scan-enabled CI pipelines count (scan_enabled=true in ci_pipeline table)
	scanEnabledCiPipelines, err := service.ciPipelineRepository.GetScanEnabledCiPipelineCount()
	if err != nil {
		service.logger.Errorw("error getting scan-enabled CI pipelines count", "err", err)
		return nil, fmt.Errorf("failed to get scan-enabled CI pipelines count: %w", err)
	}

	// Get CI pipelines with IMAGE SCAN plugin configured in POST-CI or PRE-CD stages
	pluginConfiguredPipelines, err := service.ciPipelineRepository.GetCiPipelineCountWithImageScanPluginInPostCiOrPreCd()
	if err != nil {
		service.logger.Errorw("error getting CI pipelines with IMAGE SCAN plugin in POST-CI or PRE-CD count", "err", err)
		return nil, fmt.Errorf("failed to get CI pipelines with IMAGE SCAN plugin in POST-CI or PRE-CD count: %w", err)
	}

	totalScanningEnabledPipelines := scanEnabledCiPipelines + pluginConfiguredPipelines

	// Build response with calculated percentages
	// For unscanned images: percentage = unscanned / (unscanned + scanned)
	totalScannableDeployments := scannedCounts.UnscannedCount + scannedCounts.ScannedCount
	response := &bean.DeploymentSecurityStatusResponse{
		ActiveDeploymentsWithVulnerabilities: &bean.DeploymentMetric{
			Count:      deploymentsWithVulnerabilities,
			Percentage: calculatePercentage(deploymentsWithVulnerabilities, totalDeployments),
		},
		ActiveDeploymentsWithUnscannedImages: &bean.DeploymentMetric{
			Count:      scannedCounts.UnscannedCount,
			Percentage: calculatePercentage(scannedCounts.UnscannedCount, totalScannableDeployments),
		},
		WorkflowsWithScanningEnabled: &bean.WorkflowMetric{
			Count:      totalScanningEnabledPipelines,
			Percentage: calculatePercentage(totalScanningEnabledPipelines, totalCiPipelines),
		},
	}

	return response, nil
}

func (service *SecurityOverviewServiceImpl) GetVulnerabilityTrend(ctx context.Context, currentTimeRange *utils.TimeRangeRequest, envType bean.EnvType, aggregationType constants.AggregationType) (*bean.VulnerabilityTrendResponse, error) {
	service.logger.Infow("GetVulnerabilityTrend called", "from", currentTimeRange.From, "to", currentTimeRange.To, "envType", envType, "aggregationType", aggregationType)

	// Determine environment type filter
	// nil = all environments, true = prod only, false = non-prod only
	var isProd *bool
	if envType == bean.EnvTypeProd {
		prodValue := true
		isProd = &prodValue
	} else if envType == bean.EnvTypeNonProd {
		nonProdValue := false
		isProd = &nonProdValue
	}
	// If envType is "all", isProd remains nil

	// Fetch vulnerability trend data from repository
	vulnerabilities, err := service.imageScanResultRepository.GetVulnerabilityTrendDataByFilters(
		currentTimeRange.From,
		currentTimeRange.To,
		isProd,
	)
	if err != nil {
		service.logger.Errorw("error getting vulnerability trend data", "err", err)
		return nil, fmt.Errorf("failed to get vulnerability trend data: %w", err)
	}

	// Aggregate vulnerabilities by time bucket and severity
	trendData := service.aggregateVulnerabilitiesByTime(vulnerabilities, currentTimeRange.From, currentTimeRange.To, aggregationType)

	response := &bean.VulnerabilityTrendResponse{
		Trend: trendData,
	}

	return response, nil
}

// aggregateVulnerabilitiesByTime aggregates vulnerabilities by time buckets and severity
func (service *SecurityOverviewServiceImpl) aggregateVulnerabilitiesByTime(
	vulnerabilities []*imageScanRepo.VulnerabilityTrendData,
	from, to *time.Time,
	aggregationType constants.AggregationType,
) []*bean.VulnerabilityTrendDataPoint {
	// Map to track unique CVEs per time bucket and severity: timeKey -> severity -> set of CVE names
	severityMap := make(map[string]map[int]map[string]bool)

	targetLocation := from.Location()

	// Process each vulnerability and bucket by time
	for _, vuln := range vulnerabilities {
		// Convert UTC execution time to target timezone for proper time bucketing
		localExecutionTime := vuln.ExecutionTime.In(targetLocation)

		var timeKey string
		if aggregationType == constants.AggregateByHour {
			timeKey = localExecutionTime.Truncate(time.Hour).Format("2006-01-02T15:04:05Z")
		} else if aggregationType == constants.AggregateByMonth {
			timeKey = time.Date(localExecutionTime.Year(), localExecutionTime.Month(), 1, 0, 0, 0, 0, targetLocation).Format("2006-01-02T15:04:05Z")
		} else {
			timeKey = localExecutionTime.Truncate(24 * time.Hour).Format("2006-01-02T15:04:05Z")
		}

		// Initialize maps if needed
		if severityMap[timeKey] == nil {
			severityMap[timeKey] = make(map[int]map[string]bool)
		}
		if severityMap[timeKey][vuln.Severity] == nil {
			severityMap[timeKey][vuln.Severity] = make(map[string]bool)
		}

		// Track unique CVE names per time bucket and severity
		severityMap[timeKey][vuln.Severity][vuln.CveStoreName] = true
	}

	// Generate time-series data with zero values for missing time buckets
	var trendData []*bean.VulnerabilityTrendDataPoint

	if aggregationType == constants.AggregateByHour {
		current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), to.Day(), to.Hour(), 0, 0, 0, from.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Format("2006-01-02T15:04:05Z")
			dataPoint := service.createVulnerabilityDataPoint(current, severityMap[timeKey])
			trendData = append(trendData, dataPoint)
			current = current.Add(time.Hour)
		}
	} else if aggregationType == constants.AggregateByMonth {
		current := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, to.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Format("2006-01-02T15:04:05Z")
			dataPoint := service.createVulnerabilityDataPoint(current, severityMap[timeKey])
			trendData = append(trendData, dataPoint)
			current = current.AddDate(0, 1, 0) // Add one month
		}
	} else {
		// Daily aggregation
		current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Format("2006-01-02T15:04:05Z")
			dataPoint := service.createVulnerabilityDataPoint(current, severityMap[timeKey])
			trendData = append(trendData, dataPoint)
			current = current.AddDate(0, 0, 1) // Add one day
		}
	}

	return trendData
}

// createVulnerabilityDataPoint creates a data point with counts for each severity level
func (service *SecurityOverviewServiceImpl) createVulnerabilityDataPoint(
	timestamp time.Time,
	severityCounts map[int]map[string]bool,
) *bean.VulnerabilityTrendDataPoint {
	dataPoint := &bean.VulnerabilityTrendDataPoint{
		Timestamp: timestamp,
		Critical:  0,
		High:      0,
		Medium:    0,
		Low:       0,
		Unknown:   0,
		Total:     0,
	}

	if severityCounts == nil {
		return dataPoint
	}

	// Count unique CVEs for each severity level
	for severity, cveSet := range severityCounts {
		count := len(cveSet)

		switch scanBean.Severity(severity) {
		case scanBean.Critical:
			dataPoint.Critical = count
		case scanBean.High:
			dataPoint.High = count
		case scanBean.Medium:
			dataPoint.Medium = count
		case scanBean.Low:
			dataPoint.Low = count
		default:
			dataPoint.Unknown = count
		}

		dataPoint.Total += count
	}

	return dataPoint
}

func (service *SecurityOverviewServiceImpl) GetBlockedDeploymentsTrend(ctx context.Context, currentTimeRange *utils.TimeRangeRequest, aggregationType constants.AggregationType) (*bean.BlockedDeploymentsTrendResponse, error) {
	service.logger.Infow("GetBlockedDeploymentsTrend called", "from", currentTimeRange.From, "to", currentTimeRange.To, "aggregationType", aggregationType)

	// Fetch blocked deployment data from repository
	blockedDeployments, err := service.cdWorkflowRepository.GetBlockedDeploymentsForTrend(currentTimeRange.From, currentTimeRange.To)
	if err != nil {
		service.logger.Errorw("error getting blocked deployments for trend", "err", err)
		return nil, fmt.Errorf("failed to get blocked deployments: %w", err)
	}

	// Aggregate blocked deployments by time bucket
	trendData := service.aggregateBlockedDeploymentsByTime(blockedDeployments, currentTimeRange.From, currentTimeRange.To, aggregationType)

	response := &bean.BlockedDeploymentsTrendResponse{
		Trend: trendData,
	}

	return response, nil
}

// aggregateBlockedDeploymentsByTime aggregates blocked deployments by time buckets
func (service *SecurityOverviewServiceImpl) aggregateBlockedDeploymentsByTime(
	blockedDeployments []pipelineConfig.BlockedDeploymentData,
	from, to *time.Time,
	aggregationType constants.AggregationType,
) []*bean.BlockedDeploymentDataPoint {
	// Map to track counts per time bucket: Unix timestamp -> count
	countMap := make(map[int64]int)

	targetLocation := from.Location()

	// Process each blocked deployment and bucket by time
	for _, deployment := range blockedDeployments {
		// Convert UTC started_on time to target timezone for proper time bucketing
		localStartedOn := deployment.StartedOn.In(targetLocation)

		var bucketTime time.Time
		if aggregationType == constants.AggregateByHour {
			// Truncate to hour boundary in local timezone
			bucketTime = time.Date(localStartedOn.Year(), localStartedOn.Month(), localStartedOn.Day(),
				localStartedOn.Hour(), 0, 0, 0, targetLocation)
		} else if aggregationType == constants.AggregateByMonth {
			// Truncate to month boundary (1st day of month at midnight)
			bucketTime = time.Date(localStartedOn.Year(), localStartedOn.Month(), 1, 0, 0, 0, 0, targetLocation)
		} else {
			// Daily aggregation - truncate to day boundary (midnight) in local timezone
			bucketTime = time.Date(localStartedOn.Year(), localStartedOn.Month(), localStartedOn.Day(),
				0, 0, 0, 0, targetLocation)
		}

		// Use Unix timestamp as key to avoid timezone formatting issues
		timeKey := bucketTime.Unix()
		countMap[timeKey]++
	}

	// Generate time-series data with zero values for missing time buckets
	var trendData []*bean.BlockedDeploymentDataPoint

	if aggregationType == constants.AggregateByHour {
		current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), to.Day(), to.Hour(), 0, 0, 0, from.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Unix()
			count := countMap[timeKey]

			trendData = append(trendData, &bean.BlockedDeploymentDataPoint{
				Timestamp: current,
				Count:     count,
			})

			current = current.Add(time.Hour)
		}
	} else if aggregationType == constants.AggregateByMonth {
		current := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, to.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Unix()
			count := countMap[timeKey]

			trendData = append(trendData, &bean.BlockedDeploymentDataPoint{
				Timestamp: current,
				Count:     count,
			})

			current = current.AddDate(0, 1, 0) // Add one month
		}
	} else {
		// Daily aggregation
		current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Unix()
			count := countMap[timeKey]

			trendData = append(trendData, &bean.BlockedDeploymentDataPoint{
				Timestamp: current,
				Count:     count,
			})

			current = current.AddDate(0, 0, 1) // Add one day
		}
	}

	return trendData
}

func calculatePercentage(count, total int) float64 {
	if total == 0 {
		return 0.0
	}
	return util.RoundToTwoDecimals(float64(count) / float64(total) * 100.0)
}
