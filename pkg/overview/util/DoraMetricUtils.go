package util

import (
	"time"

	"github.com/devtron-labs/devtron/pkg/overview/bean"
)

func CreateDoraMetricObject(overallAverageValue float64, overallAverageUnit bean.MetricValueUnit, comparisonValue int, comparisonUnit bean.ComparisonUnit, performanceLevelCount *bean.PerformanceLevelCount) *bean.DoraMetric {
	return &bean.DoraMetric{
		OverallAverage: &bean.MetricValue{
			Value: overallAverageValue,
			Unit:  overallAverageUnit.ToString(),
		},
		ComparisonValue:       comparisonValue,
		ComparisonUnit:        comparisonUnit,
		PerformanceLevelCount: performanceLevelCount,
	}
}

// CalculateComparison calculates the comparison value between current and previous periods
// Returns percentage for DeploymentFrequency and ChangeFailureRate, minutes for MeanLeadTime and MeanTimeToRecovery
func CalculateComparison(current, previous float64, metricCategory bean.MetricCategory) int {
	switch metricCategory {
	case bean.MetricCategoryDeploymentFrequency, bean.MetricCategoryChangeFailureRate:
		if previous == 0 {
			if current > 0 {
				return 100 // Return 100% increase when previous was 0
			}
			return 0
		}
		// Calculate percentage change for frequency and failure rate metrics
		percentageChange := ((current - previous) / previous) * 100
		return int(percentageChange)
	case bean.MetricCategoryMeanLeadTime, bean.MetricCategoryMeanTimeToRecovery:
		if previous == 0 {
			if current > 0 {
				return int(current)
			}
			return 0
		}
		// Calculate minutes difference for time-based metrics
		return int(current - previous)
	default:
		return 0
	}
}

// CalculatePerformanceLevelsForMetric calculates the count of pipelines in each performance category for a specific metric
func CalculatePerformanceLevelsForMetric(metricsData map[string]*bean.LensMetrics, metricCategory bean.MetricCategory) *bean.PerformanceLevelCount {
	performanceLevels := &bean.PerformanceLevelCount{
		Elite:  0,
		High:   0,
		Medium: 0,
		Low:    0,
	}

	if len(metricsData) == 0 {
		return performanceLevels
	}

	// Categorize each app-env pair based on the specific metric
	for _, lensMetrics := range metricsData {
		var metricValue float64

		// Get the appropriate metric value based on category
		switch metricCategory {
		case bean.MetricCategoryDeploymentFrequency:
			metricValue = lensMetrics.AverageCycleTime
		case bean.MetricCategoryMeanLeadTime:
			metricValue = lensMetrics.AverageLeadTime
		case bean.MetricCategoryChangeFailureRate:
			metricValue = lensMetrics.ChangeFailureRate
		case bean.MetricCategoryMeanTimeToRecovery:
			metricValue = lensMetrics.AverageRecoveryTime
		default:
			// Default to low performance for unknown metric categories
			performanceLevels.Low++
			continue
		}

		// Categorize based on the specific metric thresholds
		if IsInMetricCategory(metricValue, metricCategory, bean.PerformanceElite) {
			performanceLevels.Elite++
		} else if IsInMetricCategory(metricValue, metricCategory, bean.PerformanceHigh) {
			performanceLevels.High++
		} else if IsInMetricCategory(metricValue, metricCategory, bean.PerformanceMedium) {
			performanceLevels.Medium++
		} else {
			performanceLevels.Low++
		}
	}

	return performanceLevels
}

// IsInMetricCategory routes to the appropriate category checking function based on metric type
func IsInMetricCategory(value float64, metricCategory bean.MetricCategory, performanceCategory bean.PerformanceCategory) bool {
	switch metricCategory {
	case bean.MetricCategoryDeploymentFrequency:
		return IsInDeploymentFrequencyCategory(value, performanceCategory)
	case bean.MetricCategoryMeanLeadTime:
		return IsInLeadTimeCategory(value, performanceCategory)
	case bean.MetricCategoryChangeFailureRate:
		return IsInChangeFailureRateCategory(value, performanceCategory)
	case bean.MetricCategoryMeanTimeToRecovery:
		return IsInRecoveryTimeCategory(value, performanceCategory)
	default:
		return false
	}
}

// IsInDeploymentFrequencyCategory checks deployment frequency thresholds
func IsInDeploymentFrequencyCategory(value float64, category bean.PerformanceCategory) bool {
	switch category {
	case bean.PerformanceElite:
		return value >= 1.0 // On demand (multiple deploys per day)
	case bean.PerformanceHigh:
		return value >= 0.14 && value < 1.0 // Between once per day and once per week (1/7 ≈ 0.14)
	case bean.PerformanceMedium:
		return value >= 0.033 && value < 0.14 // Between once per week and once per month (1/30 ≈ 0.033)
	case bean.PerformanceLow:
		return value < 0.033 // Between once per month and once every six months
	}
	return false
}

// IsInLeadTimeCategory checks change lead time thresholds (lower is better)
func IsInLeadTimeCategory(value float64, category bean.PerformanceCategory) bool {
	switch category {
	case bean.PerformanceElite:
		return value < 24 // Less than one day
	case bean.PerformanceHigh:
		return value >= 24 && value <= 168 // Between one day and one week
	case bean.PerformanceMedium:
		return value > 168 && value <= 720 // Between one week and one month
	case bean.PerformanceLow:
		return value > 720 && value <= 4320 // Between one month and six months
	}
	return false
}

// IsInRecoveryTimeCategory checks failed deployment recovery time thresholds (lower is better)
func IsInRecoveryTimeCategory(value float64, category bean.PerformanceCategory) bool {
	switch category {
	case bean.PerformanceElite:
		return value < 1 // Less than one hour
	case bean.PerformanceHigh:
		return value >= 1 && value < 24 // Less than one day (1-24 hours)
	case bean.PerformanceMedium:
		return value >= 24 && value < 168 // Less than one day to one week (assuming this is the intended range)
	case bean.PerformanceLow:
		return value >= 168 && value <= 720 // Between one week and one month
	}
	return false
}

// IsInChangeFailureRateCategory checks change failure rate thresholds (lower is better)
func IsInChangeFailureRateCategory(value float64, category bean.PerformanceCategory) bool {
	switch category {
	case bean.PerformanceElite:
		return value <= 5 // 5% or less
	case bean.PerformanceHigh:
		return value > 5 && value <= 10 // 6-10% (interpreting the table logically)
	case bean.PerformanceMedium:
		return value > 10 && value <= 20 // 11-20%
	case bean.PerformanceLow:
		return value > 20 && value <= 40 // 21-40%
	}
	return false
}

// CalculateAverageFromValues calculates average from a slice of float64 values
func CalculateAverageFromValues(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	var total float64
	for _, value := range values {
		total += value
	}

	return total / float64(len(values))
}

// CalculatePreviousPeriod calculates the previous period dates for comparison
func CalculatePreviousPeriod(from, to *time.Time) (*time.Time, *time.Time) {
	if from == nil || to == nil {
		return nil, nil
	}

	// Calculate the duration of the current period
	duration := to.Sub(*from)

	// Previous period ends where current period starts, and starts duration before that
	previousTo := *from
	previousFrom := from.Add(-duration)

	return &previousFrom, &previousTo
}
