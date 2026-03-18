package bean

import (
	"time"

	"github.com/devtron-labs/common-lib/utils"
)

type AllDoraMetrics struct {
	DeploymentFrequency *DoraMetric
	MeanLeadTime        *DoraMetric
	ChangeFailureRate   *DoraMetric
	MeanTimeToRecovery  *DoraMetric
}

func NewAllDoraMetrics() *AllDoraMetrics {
	return &AllDoraMetrics{
		DeploymentFrequency: &DoraMetric{},
		MeanLeadTime:        &DoraMetric{},
		ChangeFailureRate:   &DoraMetric{},
		MeanTimeToRecovery:  &DoraMetric{},
	}
}

func (r *AllDoraMetrics) WithDeploymentFrequency(deploymentFrequency *DoraMetric) *AllDoraMetrics {
	r.DeploymentFrequency = deploymentFrequency
	return r
}

func (r *AllDoraMetrics) WithMeanLeadTime(meanLeadTime *DoraMetric) *AllDoraMetrics {
	r.MeanLeadTime = meanLeadTime
	return r
}

func (r *AllDoraMetrics) WithChangeFailureRate(changeFailureRate *DoraMetric) *AllDoraMetrics {
	r.ChangeFailureRate = changeFailureRate
	return r
}

func (r *AllDoraMetrics) WithMeanTimeToRecovery(meanTimeToRecovery *DoraMetric) *AllDoraMetrics {
	r.MeanTimeToRecovery = meanTimeToRecovery
	return r
}

// LensMetrics represents the response structure from Lens API
type LensMetrics struct {
	AverageCycleTime       float64 `json:"average_cycle_time"`
	AverageLeadTime        float64 `json:"average_lead_time"`
	ChangeFailureRate      float64 `json:"change_failure_rate"`
	AverageRecoveryTime    float64 `json:"average_recovery_time"`
	AverageDeploymentSize  float32 `json:"average_deployment_size"`
	AverageLineAdded       float32 `json:"average_line_added"`
	AverageLineDeleted     float32 `json:"average_line_deleted"`
	LastFailedTime         string  `json:"last_failed_time"`
	RecoveryTimeLastFailed float64 `json:"recovery_time_last_failed"`
}

type DoraMetric struct {
	OverallAverage        *MetricValue           `json:"overallAverage"`
	ComparisonValue       int                    `json:"comparisonValue"`       // Percentage or minutes change
	ComparisonUnit        ComparisonUnit         `json:"comparisonUnit"`        // PERCENTAGE or MINUTES
	PerformanceLevelCount *PerformanceLevelCount `json:"performanceLevelCount"` // Count of pipelines in each performance category
}

func NewDoraMetric() *DoraMetric {
	return &DoraMetric{}
}
func (r *DoraMetric) WithOverallAverage(overallAverage *MetricValue) *DoraMetric {
	r.OverallAverage = overallAverage
	return r
}

func (r *DoraMetric) WithComparisonValue(comparisonValue int) *DoraMetric {
	r.ComparisonValue = comparisonValue
	return r
}
func (r *DoraMetric) WithComparisonUnit(comparisonUnit ComparisonUnit) *DoraMetric {
	r.ComparisonUnit = comparisonUnit
	return r
}
func (r *DoraMetric) WithPerformanceLevelCount(performanceLevelCount *PerformanceLevelCount) *DoraMetric {
	r.PerformanceLevelCount = performanceLevelCount
	return r
}

type MetricValue struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"` // NUMBER, PERCENTAGE, MINUTES
}

type PerformanceLevelCount struct {
	Elite  int `json:"elite"`
	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
}

// DORA Metrics Beans
type DoraMetricsRequest struct {
	TimeRangeRequest *utils.TimeRangeRequest `json:"timeRangeRequest"`
	PrevFrom         *time.Time              `json:"prevFrom,omitempty"` // Previous period start time
	PrevTo           *time.Time              `json:"prevTo,omitempty"`   // Previous period end time
}

type DoraMetricsResponse struct {
	ProdDeploymentPipelineCount int         `json:"prodDeploymentPipelineCount"`
	DeploymentFrequency         *DoraMetric `json:"deploymentFrequency"`
	MeanLeadTime                *DoraMetric `json:"meanLeadTime"`
	ChangeFailureRate           *DoraMetric `json:"changeFailureRate"`
	MeanTimeToRecovery          *DoraMetric `json:"meanTimeToRecovery"`
}

func NewDoraMetricsResponse() *DoraMetricsResponse {
	return &DoraMetricsResponse{}
}

type ComparisonUnit string

const (
	ComparisonUnitMinutes    ComparisonUnit = "MINUTES"
	ComparisonUnitPercentage ComparisonUnit = "PERCENTAGE"
)

type MetricValueUnit string

const (
	MetricValueUnitNumber     MetricValueUnit = "NUMBER"
	MetricValueUnitPercentage MetricValueUnit = "PERCENTAGE"
	MetricValueUnitMinutes    MetricValueUnit = "MINUTES"
)

func (r MetricValueUnit) ToString() string {
	return string(r)
}

type PerformanceCategory string

const (
	PerformanceElite  PerformanceCategory = "Elite"
	PerformanceHigh   PerformanceCategory = "High"
	PerformanceMedium PerformanceCategory = "Medium"
	PerformanceLow    PerformanceCategory = "Low"
)

type MetricCategory string

const (
	MetricCategoryMeanTimeToRecovery  MetricCategory = "meanTimeToRecovery"
	MetricCategoryChangeFailureRate   MetricCategory = "changeFailureRate"
	MetricCategoryMeanLeadTime        MetricCategory = "meanLeadTime"
	MetricCategoryDeploymentFrequency MetricCategory = "deploymentFrequency"
)

// IsValidMetricCategory checks if the given string is a valid metric category
func IsValidMetricCategory(category string) bool {
	switch MetricCategory(category) {
	case MetricCategoryDeploymentFrequency, MetricCategoryMeanLeadTime, MetricCategoryChangeFailureRate, MetricCategoryMeanTimeToRecovery:
		return true
	default:
		return false
	}
}

// IsValidPerformanceCategory checks if the given string is a valid performance category
func IsValidPerformanceCategory(category string) bool {
	switch PerformanceCategory(category) {
	case PerformanceElite, PerformanceHigh, PerformanceMedium, PerformanceLow:
		return true
	default:
		return false
	}
}
